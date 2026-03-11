# rag_pipeline.py
#
# RAG implementation using LangChain + ChromaDB.
#
# Embedding model : OpenAI text-embedding-3-small
#   DeepSeek does not provide an embedding API, so OpenAI is used for this step.
#
# Generation model: DeepSeek chat API (OpenAI-compatible endpoint)
#   Configured via DEEPSEEK_API_KEY and DEEPSEEK_MODEL env vars.
#   Falls back to OpenAI if DEEPSEEK_API_KEY is not set.
#
# Pipeline flow:
#   1. process_pending_documents(project_id)
#        → db_client.get_pending_documents()      — fetch rows with embed_status='pending'
#        → db_client.mark_processing()            — lock rows so other workers skip them
#        → _embed_document() per doc              — split → embed → store in Chroma
#        → db_client.mark_completed() / failed()  — update final status
#
#   2. retrieve(query, project_id)
#        → similarity_search_with_score()         — top-K results above score threshold
#
#   3. generate_with_rag(query, project_id)
#        → process_pending_documents()            — embed anything new first
#        → retrieve()                             — fetch relevant context
#        → DeepSeek LLM via LCEL chain            — generate answer

from __future__ import annotations

import os
from typing import Dict, List, Optional

# Patch sqlite3 before importing ChromaDB (same as existing vector_store.py)
import src.sqlite_patch  # noqa: F401

from langchain.prompts import ChatPromptTemplate, PromptTemplate
from langchain.schema.output_parser import StrOutputParser
from langchain.schema.runnable import RunnablePassthrough
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain_community.vectorstores import Chroma
from langchain_openai import ChatOpenAI, OpenAIEmbeddings

from src.config import Config
from src.db_client import DatabaseClient, DocumentRecord
from src.logger import setup_logger

logger = setup_logger(__name__)

# ---------------------------------------------------------------------------
# Score threshold: ChromaDB returns L2 distances (lower = more similar).
# Chunks whose distance exceeds this value are dropped before being injected
# into the prompt so irrelevant content doesn't pollute the context.
# Override with RETRIEVAL_SCORE_THRESHOLD env var.
# ---------------------------------------------------------------------------
_DEFAULT_SCORE_THRESHOLD = 1.2


# ---------------------------------------------------------------------------
# Prompt template for RAG generation
# ---------------------------------------------------------------------------
_RAG_PROMPT = ChatPromptTemplate.from_messages(
    [
        (
            "system",
            """You are a helpful AI assistant with access to project documentation.

RELEVANT DOCUMENTATION:
{context}

Use the documentation above to answer the user's question accurately and concisely.
If the documentation does not contain the required information, say so and answer
from general knowledge. Cite which document you are referencing where relevant.""",
        ),
        ("human", "{question}"),
    ]
)

_NO_CONTEXT_PROMPT = ChatPromptTemplate.from_messages(
    [
        (
            "system",
            "You are a helpful AI assistant. No project documentation has been indexed yet. "
            "Answer from general knowledge and suggest indexing documentation for better results.",
        ),
        ("human", "{question}"),
    ]
)


def _format_docs(docs_with_scores: List[tuple]) -> str:
    """Format retrieved (Document, score) pairs into a single context string."""
    parts = []
    for i, (doc, score) in enumerate(docs_with_scores, start=1):
        title = doc.metadata.get("title", "Untitled")
        parts.append(f"[Document {i} — {title} (relevance score: {score:.3f})]\n{doc.page_content}")
    return "\n\n---\n\n".join(parts)


class RAGPipeline:
    """
    Manages the full RAG lifecycle for a given project:
    embedding new documents from Postgres and answering queries using
    DeepSeek (LLM) + ChromaDB (vector store) + OpenAI (embeddings).
    """

    def __init__(self) -> None:
        self.db = DatabaseClient()

        # --- Embedding model (OpenAI — DeepSeek has no embedding API) ---
        self.embeddings = OpenAIEmbeddings(
            model=Config.EMBEDDING_MODEL,
            openai_api_key=Config.OPENAI_API_KEY,
        )

        # --- Text splitter ---
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=Config.RAG_CHUNK_SIZE,
            chunk_overlap=Config.RAG_CHUNK_OVERLAP,
            length_function=len,
        )

        # --- Generation LLM (DeepSeek, OpenAI-compatible) ---
        self.llm = self._build_llm()

        # --- Per-project Chroma stores (lazy-loaded) ---
        self._stores: Dict[str, Chroma] = {}

        # --- Score threshold for retrieval filtering ---
        self.score_threshold = float(
            os.getenv("RETRIEVAL_SCORE_THRESHOLD", str(_DEFAULT_SCORE_THRESHOLD))
        )

        logger.info(
            f"RAGPipeline initialised — LLM: {self.llm.model_name}, "
            f"score_threshold={self.score_threshold}"
        )

    # ------------------------------------------------------------------
    # Public: embedding pipeline
    # ------------------------------------------------------------------

    def process_pending_documents(self, project_id: str) -> int:
        """
        Fetch all pending documents for *project_id*, embed them, and update
        their status in Postgres.

        Returns the number of documents successfully embedded.
        Uses optimistic locking via embed_status to prevent duplicate work
        across multiple worker instances.
        """
        pending = self.db.get_pending_documents(project_id)
        if not pending:
            logger.info(f"process_pending_documents: no pending docs for project={project_id}")
            return 0

        doc_ids = [d.id for d in pending]

        # Lock rows — only rows still in 'pending' will be updated (race protection)
        self.db.mark_processing(doc_ids)

        success_ids: List[str] = []
        failed_ids: List[str] = []

        for doc in pending:
            try:
                self._embed_document(doc, project_id)
                success_ids.append(doc.id)
                logger.info(f"Embedded doc id={doc.id} title='{doc.title}'")
            except Exception as exc:
                failed_ids.append(doc.id)
                logger.error(f"Failed to embed doc id={doc.id}: {exc}")

        if success_ids:
            self.db.mark_completed(success_ids)
        if failed_ids:
            self.db.mark_failed(failed_ids, reason="see logs above")

        logger.info(
            f"process_pending_documents: {len(success_ids)} completed, "
            f"{len(failed_ids)} failed for project={project_id}"
        )
        return len(success_ids)

    # ------------------------------------------------------------------
    # Public: retrieval
    # ------------------------------------------------------------------

    def retrieve(
        self,
        query: str,
        project_id: str,
        top_k: Optional[int] = None,
    ) -> List[Dict]:
        """
        Retrieve the most relevant document chunks for *query* within *project_id*.

        Returns a list of dicts with keys: content, score, metadata.
        Chunks whose distance exceeds self.score_threshold are excluded.
        """
        k = top_k or Config.TOP_K_RETRIEVAL
        store = self._get_store(project_id)

        try:
            raw = store.similarity_search_with_score(
                query,
                k=k,
                filter={"project_id": project_id},
            )
        except Exception as exc:
            logger.error(f"retrieve: similarity search failed: {exc}")
            return []

        results: List[Dict] = []
        seen: set = set()

        for doc, score in raw:
            # Filter by threshold (lower L2 distance = better match)
            if score > self.score_threshold:
                logger.debug(
                    f"retrieve: dropping chunk (score {score:.3f} > threshold {self.score_threshold})"
                )
                continue

            # Exact-content dedup
            if doc.page_content in seen:
                continue
            seen.add(doc.page_content)

            results.append(
                {
                    "content": doc.page_content,
                    "score": float(score),
                    "metadata": doc.metadata,
                }
            )

        logger.info(
            f"retrieve: {len(results)} chunks returned for project={project_id} "
            f"(query='{query[:60]}...')"
        )
        return results

    # ------------------------------------------------------------------
    # Public: RAG generation
    # ------------------------------------------------------------------

    def generate_with_rag(
        self,
        query: str,
        project_id: str,
        top_k: Optional[int] = None,
    ) -> str:
        """
        Full RAG query:
          1. Embed any newly added documents for the project.
          2. Retrieve relevant chunks.
          3. Generate an answer with DeepSeek using retrieved context.

        Returns the generated answer as a plain string.
        """
        # Embed new docs before answering so the answer is always up to date
        new_count = self.process_pending_documents(project_id)
        if new_count:
            logger.info(f"generate_with_rag: embedded {new_count} new docs before retrieval")

        retrieved = self.retrieve(query, project_id, top_k=top_k)

        if retrieved:
            # Build LCEL chain with context
            store = self._get_store(project_id)
            retriever = store.as_retriever(
                search_kwargs={
                    "k": top_k or Config.TOP_K_RETRIEVAL,
                    "filter": {"project_id": project_id},
                }
            )

            def _retrieve_and_format(q: str) -> str:
                docs_with_scores = store.similarity_search_with_score(
                    q,
                    k=top_k or Config.TOP_K_RETRIEVAL,
                    filter={"project_id": project_id},
                )
                filtered = [
                    (d, s) for d, s in docs_with_scores if s <= self.score_threshold
                ]
                return _format_docs(filtered) if filtered else "No relevant documentation found."

            chain = (
                {"context": _retrieve_and_format, "question": RunnablePassthrough()}
                | _RAG_PROMPT
                | self.llm
                | StrOutputParser()
            )
        else:
            # No relevant context — answer from general knowledge
            chain = (
                {"question": RunnablePassthrough()}
                | _NO_CONTEXT_PROMPT
                | self.llm
                | StrOutputParser()
            )

        answer = chain.invoke(query)
        return answer

    # ------------------------------------------------------------------
    # Internal: vector store management
    # ------------------------------------------------------------------

    def _get_store(self, project_id: str) -> Chroma:
        """Lazily create or return the Chroma collection for *project_id*."""
        if project_id not in self._stores:
            collection_name = f"project_{project_id}"
            self._stores[project_id] = Chroma(
                collection_name=collection_name,
                embedding_function=self.embeddings,
                persist_directory=f"{Config.VECTOR_DB_PATH}/{project_id}",
            )
            logger.info(f"_get_store: opened Chroma collection '{collection_name}'")
        return self._stores[project_id]

    # ------------------------------------------------------------------
    # Internal: embedding a single document
    # ------------------------------------------------------------------

    def _embed_document(self, doc: DocumentRecord, project_id: str) -> None:
        """
        Split *doc.content* into chunks and store them in the project's
        Chroma collection.  Uses deterministic chunk IDs so re-embedding the
        same document (after a manual reset) overwrites existing chunks rather
        than creating duplicates.
        """
        chunks = self.text_splitter.split_text(doc.content)
        if not chunks:
            logger.warning(f"_embed_document: no chunks produced for doc id={doc.id}")
            return

        base_meta = {
            "project_id": project_id,
            "document_id": doc.document_id,
            "doc_row_id": doc.id,
            "title": doc.title,
            "description": doc.description,
        }
        metadatas = [{**base_meta, "chunk_index": i} for i in range(len(chunks))]
        ids = [f"{doc.id}_chunk_{i}" for i in range(len(chunks))]

        store = self._get_store(project_id)

        # Delete stale chunks for this doc before re-adding
        # (handles the case where content changed and chunk count differs)
        try:
            existing = store.get(where={"doc_row_id": doc.id})
            if existing and existing.get("ids"):
                store.delete(ids=existing["ids"])
                logger.debug(
                    f"_embed_document: deleted {len(existing['ids'])} stale chunks for doc id={doc.id}"
                )
        except Exception as exc:
            # Non-fatal — Chroma may not have any entries yet
            logger.debug(f"_embed_document: could not check stale chunks: {exc}")

        store.add_texts(texts=chunks, metadatas=metadatas, ids=ids)
        logger.info(
            f"_embed_document: stored {len(chunks)} chunks for doc id={doc.id} "
            f"title='{doc.title}'"
        )

    # ------------------------------------------------------------------
    # Internal: LLM builder
    # ------------------------------------------------------------------

    @staticmethod
    def _build_llm() -> ChatOpenAI:
        """
        Build the ChatOpenAI-compatible LLM client.

        If DEEPSEEK_API_KEY is set, routes to DeepSeek's OpenAI-compatible
        endpoint (https://api.deepseek.com).  Otherwise falls back to standard
        OpenAI using Config.OPENAI_API_KEY and Config.LLM_MODEL.
        """
        deepseek_key = os.getenv("DEEPSEEK_API_KEY")

        if deepseek_key:
            model = os.getenv("DEEPSEEK_MODEL", "deepseek-chat")
            base_url = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com")
            logger.info(f"RAGPipeline LLM: DeepSeek — model={model}, base_url={base_url}")
            return ChatOpenAI(
                model=model,
                openai_api_key=deepseek_key,
                openai_api_base=base_url,
                temperature=Config.LLM_TEMPERATURE,
                max_tokens=Config.LLM_MAX_TOKENS,
            )

        logger.info(
            f"RAGPipeline LLM: OpenAI fallback — model={Config.LLM_MODEL} "
            "(set DEEPSEEK_API_KEY to use DeepSeek)"
        )
        return ChatOpenAI(
            model=Config.LLM_MODEL,
            openai_api_key=Config.OPENAI_API_KEY,
            temperature=Config.LLM_TEMPERATURE,
            max_tokens=Config.LLM_MAX_TOKENS,
        )
