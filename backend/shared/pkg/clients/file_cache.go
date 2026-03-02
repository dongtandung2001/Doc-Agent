package clients

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

// Default cache limits
const (
	defaultMaxTotalSize int64 = 50 * 1024 * 1024 // 50MB total cache
	defaultMaxFileSize  int64 = 5 * 1024 * 1024  // 5MB per file max
	defaultStatsFile          = "/var/log/docgen/cache_stats.json"
	defaultPersistSec         = 30
)

// Global singleton cache
var (
	globalFileCache *FileCache
	cacheOnce       sync.Once
)

// Monotonic counter for LRU ordering. Cheaper than time.Now() (no syscall)
// and immune to clock drift. Higher value = more recently accessed.
var accessCounter atomic.Int64

// GetGlobalFileCache returns the singleton FileCache instance.
// Thread-safe via sync.Once.
func GetGlobalFileCache() *FileCache {
	cacheOnce.Do(func() {
		maxTotal := envInt64("FILE_CACHE_MAX_TOTAL_SIZE", defaultMaxTotalSize)
		maxFile := envInt64("FILE_CACHE_MAX_FILE_SIZE", defaultMaxFileSize)
		globalFileCache = NewFileCache(maxTotal, maxFile)
	})
	return globalFileCache
}

// FileCache is a memory-bounded cache with singleflight request coalescing.
//
// Concurrency design (RWMutex + atomic counters):
//
//	The read path (cache hit) uses RLock so all workers read concurrently.
//	LRU ordering uses a lock-free atomic counter instead of a linked list,
//	avoiding the need for a write lock on every cache hit. Eviction scans
//	entries for the smallest counter value — O(n) per eviction, but eviction
//	is rare (only when cache exceeds maxTotalSize) and n is small (hundreds
//	of entries at most for a 50MB cache of typical source files).
//
//	Timeline for 50 concurrent workers requesting cached files:
//	  Mutex approach:  workers serialize, #50 waits ~49 lock/unlock cycles
//	  RWMutex approach: all 50 read concurrently, ~0.5μs each
//
//	Network I/O (50-500ms) always happens outside any lock via singleflight.
type FileCache struct {
	entries      map[string]*CacheEntry
	totalSize    int64
	maxTotalSize int64
	maxFileSize  int64
	mu           sync.RWMutex
	group        singleflight.Group
	stats        CacheStats

	// Persistence
	statsFile     string
	persistTicker *time.Ticker
	stopPersist   chan struct{}
}

// CacheEntry holds a cached file's content and its LRU access order.
type CacheEntry struct {
	path       string
	content    string
	size       int64
	lastAccess atomic.Int64 // monotonic counter, updated atomically on read (no lock needed)
}

// CacheStats holds atomic counters for lock-free telemetry.
type CacheStats struct {
	Hits            atomic.Int64
	Misses          atomic.Int64
	Evictions       atomic.Int64
	BytesCached     atomic.Int64
	BytesSaved      atomic.Int64
	CoalescedReqs   atomic.Int64
	TotalRequests   atomic.Int64
	SkippedTooLarge atomic.Int64
	FetchErrors     atomic.Int64
	FetchLatencyMs  atomic.Int64
	FetchCount      atomic.Int64
}

// CacheStatsSnapshot is a point-in-time snapshot for reporting and persistence.
type CacheStatsSnapshot struct {
	Timestamp       string  `json:"timestamp"`
	Hits            int64   `json:"hits"`
	Misses          int64   `json:"misses"`
	HitRate         float64 `json:"hit_rate_percent"`
	Evictions       int64   `json:"evictions"`
	BytesCached     int64   `json:"bytes_cached"`
	BytesSaved      int64   `json:"bytes_saved"`
	CoalescedReqs   int64   `json:"coalesced_requests"`
	TotalRequests   int64   `json:"total_requests"`
	SkippedTooLarge int64   `json:"skipped_too_large"`
	FetchErrors     int64   `json:"fetch_errors"`
	AvgFetchLatency float64 `json:"avg_fetch_latency_ms"`
}

// NewFileCache creates a FileCache with the given size limits.
// It loads any persisted stats from disk and starts a background goroutine
// for periodic stats persistence.
func NewFileCache(maxTotalSize, maxFileSize int64) *FileCache {
	statsFile := os.Getenv("CACHE_STATS_FILE")
	if statsFile == "" {
		statsFile = defaultStatsFile
	}

	fc := &FileCache{
		entries:      make(map[string]*CacheEntry),
		maxTotalSize: maxTotalSize,
		maxFileSize:  maxFileSize,
		statsFile:    statsFile,
		stopPersist:  make(chan struct{}),
	}

	// Wipe stale stats from previous run so each startup begins with a clean slate.
	_ = os.Remove(statsFile)

	persistSec := envInt64("CACHE_PERSIST_INTERVAL_SEC", defaultPersistSec)
	fc.startPeriodicPersist(time.Duration(persistSec) * time.Second)

	log.Printf("[FileCache] Initialized: maxTotal=%dMB maxFile=%dMB statsFile=%s",
		maxTotalSize/(1024*1024), maxFileSize/(1024*1024), statsFile)

	return fc
}

// GetOrFetch returns cached content for the given path, or calls fetchFn exactly
// once (even under concurrent access) via singleflight. The result is cached if
// it fits within the per-file size limit.
func (fc *FileCache) GetOrFetch(path string, fetchFn func() (string, error)) (string, error) {
	fc.stats.TotalRequests.Add(1)

	// Fast path: RLock allows all workers to check cache concurrently.
	// The atomic lastAccess update requires no write lock.
	fc.mu.RLock()
	if entry, ok := fc.entries[path]; ok {
		content := entry.content
		size := entry.size
		entry.lastAccess.Store(accessCounter.Add(1))
		fc.mu.RUnlock()

		fc.stats.Hits.Add(1)
		fc.stats.BytesSaved.Add(size)
		return content, nil
	}
	fc.mu.RUnlock()

	fc.stats.Misses.Add(1)

	// Slow path: fetch via singleflight. Network I/O happens outside all locks.
	// If multiple goroutines request the same path concurrently, only one fetches.
	result, err, shared := fc.group.Do(path, func() (interface{}, error) {
		// Double-check cache after acquiring singleflight slot.
		fc.mu.RLock()
		if entry, ok := fc.entries[path]; ok {
			content := entry.content
			entry.lastAccess.Store(accessCounter.Add(1))
			fc.mu.RUnlock()
			return content, nil
		}
		fc.mu.RUnlock()

		// Network fetch — no locks held.
		start := time.Now()
		content, err := fetchFn()
		latencyMs := time.Since(start).Milliseconds()

		fc.stats.FetchLatencyMs.Add(latencyMs)
		fc.stats.FetchCount.Add(1)

		if err != nil {
			fc.stats.FetchErrors.Add(1)
			return "", err
		}

		contentSize := int64(len(content))

		// Skip caching for oversized files but still return the content.
		if contentSize > fc.maxFileSize {
			fc.stats.SkippedTooLarge.Add(1)
			return content, nil
		}

		// Cache the result. Write lock held for ~1-5μs (map insert + possible evictions).
		fc.mu.Lock()
		evicted := fc.set(path, content, contentSize)
		fc.stats.Evictions.Add(int64(evicted))
		fc.stats.BytesCached.Store(fc.totalSize)
		fc.mu.Unlock()

		return content, nil
	})

	if shared {
		fc.stats.CoalescedReqs.Add(1)
	}

	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// set adds or updates an entry in the cache. Caller must hold fc.mu write lock.
// Returns the number of entries evicted to make room.
func (fc *FileCache) set(path, content string, size int64) int {
	// If already cached, remove the old entry first.
	if existing, ok := fc.entries[path]; ok {
		fc.totalSize -= existing.size
		delete(fc.entries, path)
	}

	// Evict least-recently-used entries until there's room.
	evicted := 0
	for fc.totalSize+size > fc.maxTotalSize && len(fc.entries) > 0 {
		fc.evictOldest()
		evicted++
	}

	// If the file still doesn't fit after full eviction, don't cache it.
	if fc.totalSize+size > fc.maxTotalSize {
		return evicted
	}

	entry := &CacheEntry{
		path:    path,
		content: content,
		size:    size,
	}
	entry.lastAccess.Store(accessCounter.Add(1))
	fc.entries[path] = entry
	fc.totalSize += size

	return evicted
}

// evictOldest removes the entry with the smallest lastAccess counter.
// O(n) scan, but n is small (bounded by maxTotalSize / avg file size) and
// eviction is infrequent. Caller must hold fc.mu write lock.
func (fc *FileCache) evictOldest() {
	var oldestPath string
	var oldestAccess int64 = math.MaxInt64

	for path, entry := range fc.entries {
		if t := entry.lastAccess.Load(); t < oldestAccess {
			oldestAccess = t
			oldestPath = path
		}
	}

	if oldestPath == "" {
		return
	}

	entry := fc.entries[oldestPath]
	fc.totalSize -= entry.size
	delete(fc.entries, oldestPath)
}

// --- Telemetry ---

// GetStats returns a point-in-time snapshot of cache statistics.
func (fc *FileCache) GetStats() CacheStatsSnapshot {
	return CacheStatsSnapshot{
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		Hits:            fc.stats.Hits.Load(),
		Misses:          fc.stats.Misses.Load(),
		HitRate:         fc.hitRate(),
		Evictions:       fc.stats.Evictions.Load(),
		BytesCached:     fc.stats.BytesCached.Load(),
		BytesSaved:      fc.stats.BytesSaved.Load(),
		CoalescedReqs:   fc.stats.CoalescedReqs.Load(),
		TotalRequests:   fc.stats.TotalRequests.Load(),
		SkippedTooLarge: fc.stats.SkippedTooLarge.Load(),
		FetchErrors:     fc.stats.FetchErrors.Load(),
		AvgFetchLatency: fc.avgFetchLatency(),
	}
}

func (fc *FileCache) hitRate() float64 {
	total := fc.stats.TotalRequests.Load()
	if total == 0 {
		return 0
	}
	return float64(fc.stats.Hits.Load()) / float64(total) * 100
}

func (fc *FileCache) avgFetchLatency() float64 {
	count := fc.stats.FetchCount.Load()
	if count == 0 {
		return 0
	}
	return float64(fc.stats.FetchLatencyMs.Load()) / float64(count)
}

// LogStats writes a summary line to the log.
func (fc *FileCache) LogStats(prefix string) {
	s := fc.GetStats()
	log.Printf("[%s] Cache Stats: hits=%d misses=%d hit_rate=%.1f%% "+
		"coalesced=%d evictions=%d cached=%dKB saved=%dKB avg_latency=%.1fms",
		prefix,
		s.Hits, s.Misses, s.HitRate,
		s.CoalescedReqs, s.Evictions,
		s.BytesCached/1024, s.BytesSaved/1024,
		s.AvgFetchLatency)
}

// ResetStats clears all counters. BytesCached reflects current state and is not reset.
func (fc *FileCache) ResetStats() {
	fc.stats.Hits.Store(0)
	fc.stats.Misses.Store(0)
	fc.stats.Evictions.Store(0)
	fc.stats.BytesSaved.Store(0)
	fc.stats.CoalescedReqs.Store(0)
	fc.stats.TotalRequests.Store(0)
	fc.stats.SkippedTooLarge.Store(0)
	fc.stats.FetchErrors.Store(0)
	fc.stats.FetchLatencyMs.Store(0)
	fc.stats.FetchCount.Store(0)
}

// --- Persistence ---


func (fc *FileCache) persistStats() error {
	snap := fc.GetStats()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}

	dir := filepath.Dir(fc.statsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create stats dir: %w", err)
	}

	// Atomic write: temp file → rename.
	tmp := fc.statsFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmp, fc.statsFile); err != nil {
		return fmt.Errorf("rename stats file: %w", err)
	}
	return nil
}

func (fc *FileCache) startPeriodicPersist(interval time.Duration) {
	fc.persistTicker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-fc.persistTicker.C:
				if err := fc.persistStats(); err != nil {
					log.Printf("[FileCache] Periodic persist failed: %v", err)
				}
			case <-fc.stopPersist:
				fc.persistTicker.Stop()
				return
			}
		}
	}()
}

// Shutdown stops the periodic persistence goroutine and writes final stats.
func (fc *FileCache) Shutdown() {
	close(fc.stopPersist)
	if err := fc.persistStats(); err != nil {
		log.Printf("[FileCache] Final persist failed: %v", err)
	} else {
		log.Printf("[FileCache] Stats persisted to %s", fc.statsFile)
	}
	fc.LogStats("Shutdown")
}

// --- Helpers ---

func envInt64(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int64
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return fallback
	}
	return n
}
