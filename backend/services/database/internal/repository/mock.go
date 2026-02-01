package repository

import (
	"context"
	"sync"
)

// MockSectionRepository is an in-memory implementation for testing
type MockSectionRepository struct {
	mu       sync.RWMutex
	sections map[string]*DocumentSection
}

// MockFileItemRepository is an in-memory implementation for testing
type MockFileItemRepository struct {
	mu    sync.RWMutex
	items map[string]*DocumentFileItem // key: projectID+":"+id
}

// NewMockSectionRepository creates a new mock section repository
func NewMockSectionRepository() *MockSectionRepository {
	return &MockSectionRepository{
		sections: make(map[string]*DocumentSection),
	}
}

// NewMockFileItemRepository creates a new mock file item repository
func NewMockFileItemRepository() *MockFileItemRepository {
	return &MockFileItemRepository{
		items: make(map[string]*DocumentFileItem),
	}
}

func (m *MockSectionRepository) UpsertSection(ctx context.Context, section *DocumentSection) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := *section
	m.sections[section.ID] = &s
	return nil
}

func (m *MockSectionRepository) GetSectionsByProjectID(ctx context.Context, projectID string) ([]*DocumentSection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*DocumentSection
	for _, s := range m.sections {
		if s.ProjectID == projectID {
			sc := *s
			result = append(result, &sc)
		}
	}
	return result, nil
}

func (m *MockFileItemRepository) UpsertFileItem(ctx context.Context, item *DocumentFileItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	i := *item
	m.items[item.ProjectID+":"+item.ID] = &i
	return nil
}

func (m *MockFileItemRepository) GetFileItemByID(ctx context.Context, projectID, documentID string) (*DocumentFileItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := projectID + ":" + documentID
	if item, ok := m.items[key]; ok {
		i := *item
		return &i, nil
	}
	return nil, ErrNotFound
}
