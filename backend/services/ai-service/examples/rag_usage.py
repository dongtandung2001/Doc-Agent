"""
rag_usage.py — Standalone examples for RAGPipeline and DatabaseClient.

Run from the ai-service root:
    python -m examples.rag_usage

Requires .env with at minimum:
    OPENAI_API_KEY=sk-...          # for embeddings (text-embedding-3-small)
    DEEPSEEK_API_KEY=sk-...        # for LLM generation (deepseek-chat)
    POSTGRES_URL=postgres://user:password@localhost:5432/docagent?sslmode=disable
"""

import os
import sys
from pathlib import Path

# Allow running from the ai-service root without installing as a package
sys.path.insert(0, str(Path(__file__).parent.parent))

from src.db_client import DatabaseClient
from src.rag_pipeline import RAGPipeline


# ---------------------------------------------------------------------------
# Example 1: DatabaseClient — inspect pending documents
# ---------------------------------------------------------------------------

def example_db_client(project_id: str) -> None:
    print("\n=== Example 1: DatabaseClient ===")

    db = DatabaseClient()

    # Fetch all docs waiting to be embedded
    pending = db.get_pending_documents(project_id)
    print(f"Pending documents: {len(pending)}")
    for doc in pending:
        print(f"  id={doc.id}  title='{doc.title}'  len(content)={len(doc.content)}")

    # Recover any rows stuck in 'processing' from a previous crashed run
    reset = db.reset_stale_processing(project_id)
    print(f"Stale 'processing' rows reset to 'pending': {reset}")

    db.close()


# ---------------------------------------------------------------------------
# Example 2: RAGPipeline — embed pending documents
# ---------------------------------------------------------------------------

def example_embed(project_id: str) -> None:
    print("\n=== Example 2: Embedding pending documents ===")

    pipeline = RAGPipeline()

    # process_pending_documents:
    #   1. SELECT from document_file_items WHERE embed_status='pending'
    #   2. UPDATE embed_status='processing'   (locks rows for this worker)
    #   3. Split content → embed via OpenAI → store in Chroma
    #   4. UPDATE embed_status='completed'  (or 'failed' on error)
    embedded = pipeline.process_pending_documents(project_id)
    print(f"Documents embedded this run: {embedded}")


# ---------------------------------------------------------------------------
# Example 3: RAGPipeline — retrieve relevant chunks
# ---------------------------------------------------------------------------

def example_retrieve(project_id: str, query: str) -> None:
    print(f"\n=== Example 3: Retrieve (query='{query}') ===")

    pipeline = RAGPipeline()

    # Returns list of dicts: {content, score, metadata}
    # Chunks with L2 distance > score_threshold are filtered out automatically.
    results = pipeline.retrieve(query, project_id, top_k=3)

    if not results:
        print("No relevant chunks found (collection may be empty or threshold too strict).")
        return

    for i, r in enumerate(results, start=1):
        title = r["metadata"].get("title", "Untitled")
        score = r["score"]
        snippet = r["content"][:120].replace("\n", " ")
        print(f"  [{i}] title='{title}'  score={score:.3f}  snippet='{snippet}...'")


# ---------------------------------------------------------------------------
# Example 4: RAGPipeline — full RAG generation with DeepSeek
# ---------------------------------------------------------------------------

def example_generate(project_id: str, question: str) -> None:
    print(f"\n=== Example 4: Generate (question='{question}') ===")

    pipeline = RAGPipeline()

    # generate_with_rag:
    #   1. Calls process_pending_documents first (embeds any new content)
    #   2. Retrieves relevant chunks from Chroma
    #   3. Builds a context-aware prompt
    #   4. Calls DeepSeek chat API via LangChain LCEL chain
    #   5. Returns the answer as a plain string
    answer = pipeline.generate_with_rag(question, project_id)
    print(f"Answer:\n{answer}")


# ---------------------------------------------------------------------------
# Example 5: How CreateRAG RPC maps to the pipeline
# ---------------------------------------------------------------------------

def example_create_rag_equivalent(project_id: str) -> None:
    """
    This is exactly what the CreateRAG gRPC handler now does internally.
    Call this whenever new documents have been stored via StoreDocument so
    they get picked up and embedded.
    """
    print("\n=== Example 5: CreateRAG equivalent ===")

    pipeline = RAGPipeline()
    count = pipeline.process_pending_documents(project_id)
    print(f"CreateRAG result: embedded {count} document(s) for project={project_id}")


# ---------------------------------------------------------------------------
# Example 6: How Chat RPC retrieval maps to the pipeline
# ---------------------------------------------------------------------------

def example_chat_retrieval_equivalent(project_id: str, user_message: str) -> None:
    """
    This is what ConversationOrchestrator.process_request() does when
    should_use_rag() returns True — it calls rag_pipeline.retrieve() and
    injects the results into the system prompt before calling the LLM.
    """
    print(f"\n=== Example 6: Chat retrieval equivalent (message='{user_message}') ===")

    pipeline = RAGPipeline()
    docs = pipeline.retrieve(user_message, project_id)

    if docs:
        context = "\n\n".join(
            f"[Document {i+1} — {d['metadata'].get('title','?')}]\n{d['content']}"
            for i, d in enumerate(docs)
        )
        system_prompt = (
            "You are a helpful AI assistant with access to documentation context.\n\n"
            "RELEVANT DOCUMENTATION:\n"
            f"{context}\n\n"
            "Answer the user's question using the documentation above."
        )
    else:
        system_prompt = (
            "You are a helpful AI assistant. No relevant documentation was found. "
            "Answer from general knowledge."
        )

    print("System prompt that would be sent to LLM:")
    print("-" * 60)
    print(system_prompt[:500] + ("..." if len(system_prompt) > 500 else ""))
    print("-" * 60)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    # Use a test project ID — must match a project_id in document_file_items
    PROJECT_ID = os.getenv("EXAMPLE_PROJECT_ID", "test-project-123")

    example_db_client(PROJECT_ID)
    example_embed(PROJECT_ID)
    example_retrieve(PROJECT_ID, query="How does the gateway route requests?")
    example_generate(PROJECT_ID, question="What does the codebase service do?")
    example_create_rag_equivalent(PROJECT_ID)
    example_chat_retrieval_equivalent(PROJECT_ID, user_message="How does the API gateway work?")
