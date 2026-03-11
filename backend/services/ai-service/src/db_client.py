# db_client.py
#
# Direct PostgreSQL client for the AI service.
# Handles fetching documents that need to be vector-embedded and updating
# their embed_status so multiple workers don't process the same document.
#
# Connection string is read from POSTGRES_URL (docker-compose: DATABASE_POSTGRES_URL).
# psycopg2 accepts the standard postgres:// DSN format directly.

from __future__ import annotations

import os
from dataclasses import dataclass
from typing import List, Optional

import psycopg2
import psycopg2.extras
from psycopg2.pool import ThreadedConnectionPool

from src.logger import setup_logger

logger = setup_logger(__name__)

# Embed-status constants — shared with rag_pipeline.py
EMBED_STATUS_PENDING    = "pending"
EMBED_STATUS_PROCESSING = "processing"
EMBED_STATUS_COMPLETED  = "completed"
EMBED_STATUS_FAILED     = "failed"


@dataclass
class DocumentRecord:
    """Represents a row from document_file_items that is ready to be embedded."""
    id: str
    project_id: str
    document_id: str
    title: str
    content: str
    description: str


class DatabaseClient:
    """
    Thin psycopg2 wrapper for reading/updating document_file_items.

    Thread-safe via ThreadedConnectionPool (min=1, max=5).
    The DSN is read from the POSTGRES_URL env var which maps to
    DATABASE_POSTGRES_URL in docker-compose (format: postgres://user:pass@host:port/db).
    """

    def __init__(self) -> None:
        dsn = os.getenv(
            "POSTGRES_URL",
            os.getenv(
                "DATABASE_POSTGRES_URL",
                "postgres://user:password@localhost:5432/docagent?sslmode=disable",
            ),
        )
        # Convert postgres:// to postgresql:// if needed — psycopg2 prefers postgresql://
        dsn = dsn.replace("postgres://", "postgresql://", 1)
        self._pool = ThreadedConnectionPool(minconn=1, maxconn=5, dsn=dsn)
        logger.info("DatabaseClient: connection pool initialised")

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def get_pending_documents(self, project_id: str) -> List[DocumentRecord]:
        """
        Return all documents for *project_id* whose embed_status is 'pending'
        and that have non-empty content.  Skips rows with NULL/empty content
        since there is nothing to embed.
        """
        sql = """
            SELECT id, project_id, document_id, title, content, description
            FROM document_file_items
            WHERE project_id   = %s
              AND embed_status  = %s
              AND content      IS NOT NULL
              AND content      != ''
            ORDER BY created_at ASC
        """
        rows = self._query(sql, (project_id, EMBED_STATUS_PENDING))
        docs = [
            DocumentRecord(
                id=r["id"],
                project_id=r["project_id"],
                document_id=r["document_id"] or "",
                title=r["title"],
                content=r["content"],
                description=r["description"] or "",
            )
            for r in rows
        ]
        logger.info(
            f"get_pending_documents: found {len(docs)} pending docs for project={project_id}"
        )
        return docs

    def mark_processing(self, doc_ids: List[str]) -> None:
        """
        Atomically set embed_status = 'processing' for the given IDs.
        Only transitions rows that are still 'pending' — guards against a race
        where two workers call get_pending_documents at the same time.
        """
        if not doc_ids:
            return
        sql = """
            UPDATE document_file_items
               SET embed_status = %s,
                   updated_at   = NOW()
             WHERE id = ANY(%s)
               AND embed_status = %s
        """
        self._execute(sql, (EMBED_STATUS_PROCESSING, doc_ids, EMBED_STATUS_PENDING))
        logger.info(f"mark_processing: {len(doc_ids)} docs → processing")

    def mark_completed(self, doc_ids: List[str]) -> None:
        """
        Set embed_status = 'completed' for successfully embedded documents.
        Also sets is_embedded = TRUE so other services know the content is indexed.
        """
        if not doc_ids:
            return
        sql = """
            UPDATE document_file_items
               SET embed_status = %s,
                   is_embedded  = TRUE,
                   updated_at   = NOW()
             WHERE id = ANY(%s)
        """
        self._execute(sql, (EMBED_STATUS_COMPLETED, doc_ids))
        logger.info(f"mark_completed: {len(doc_ids)} docs → completed")

    def mark_failed(self, doc_ids: List[str], reason: str = "") -> None:
        """
        Set embed_status = 'failed'.  The optional *reason* is logged but not
        persisted (add an embed_error column if you need it stored).
        Failed documents are NOT retried automatically; a manual reset to
        'pending' is required.
        """
        if not doc_ids:
            return
        sql = """
            UPDATE document_file_items
               SET embed_status = %s,
                   updated_at   = NOW()
             WHERE id = ANY(%s)
        """
        self._execute(sql, (EMBED_STATUS_FAILED, doc_ids))
        if reason:
            logger.error(
                f"mark_failed: {len(doc_ids)} docs → failed. Reason: {reason}"
            )
        else:
            logger.warning(f"mark_failed: {len(doc_ids)} docs → failed")

    def reset_stale_processing(self, project_id: Optional[str] = None) -> int:
        """
        Reset documents stuck in 'processing' back to 'pending'.
        Useful on service startup to recover from a previous crash.

        If *project_id* is None, resets across all projects.
        Returns the number of rows reset.
        """
        if project_id:
            sql = """
                UPDATE document_file_items
                   SET embed_status = %s,
                       updated_at   = NOW()
                 WHERE embed_status = %s
                   AND project_id  = %s
            """
            count = self._execute(sql, (EMBED_STATUS_PENDING, EMBED_STATUS_PROCESSING, project_id))
        else:
            sql = """
                UPDATE document_file_items
                   SET embed_status = %s,
                       updated_at   = NOW()
                 WHERE embed_status = %s
            """
            count = self._execute(sql, (EMBED_STATUS_PENDING, EMBED_STATUS_PROCESSING))

        logger.info(f"reset_stale_processing: reset {count} stale docs to pending")
        return count

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _query(self, sql: str, params: tuple) -> List[dict]:
        """Execute a SELECT and return rows as dicts."""
        conn = self._pool.getconn()
        try:
            with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
                cur.execute(sql, params)
                return [dict(r) for r in cur.fetchall()]
        except Exception:
            conn.rollback()
            raise
        finally:
            self._pool.putconn(conn)

    def _execute(self, sql: str, params: tuple) -> int:
        """Execute an INSERT/UPDATE/DELETE and return rowcount."""
        conn = self._pool.getconn()
        try:
            with conn.cursor() as cur:
                cur.execute(sql, params)
                count = cur.rowcount
            conn.commit()
            return count
        except Exception:
            conn.rollback()
            raise
        finally:
            self._pool.putconn(conn)

    def close(self) -> None:
        """Close all connections in the pool."""
        self._pool.closeall()
        logger.info("DatabaseClient: connection pool closed")
