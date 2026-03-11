import os
from pathlib import Path
from dotenv import load_dotenv

# Load .env from project root
env_path = Path(__file__).parent.parent / '.env'
load_dotenv(dotenv_path=env_path)


class Config:
    # gRPC Server
    GRPC_PORT = os.getenv("GRPC_PORT", "50051")

    # LLM Configuration
    OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
    LLM_MODEL = os.getenv("LLM_MODEL", "gpt-4o-mini")  # Cheaper: $0.075/$0.30 per 1M tokens vs $1.25/$5.00 for gpt-4o
    LLM_TEMPERATURE = float(os.getenv("LLM_TEMPERATURE", "0.7"))
    LLM_MAX_TOKENS = int(os.getenv("LLM_MAX_TOKENS", "2000"))

    # Vector Database
    VECTOR_DB_PATH = os.getenv("VECTOR_DB_PATH", "./chroma_db")
    EMBEDDING_MODEL = os.getenv("EMBEDDING_MODEL", "text-embedding-3-small")
    TOP_K_RETRIEVAL = int(os.getenv("TOP_K_RETRIEVAL", "5"))

    # RAG Configuration
    RAG_CHUNK_SIZE = int(os.getenv("RAG_CHUNK_SIZE", "1000"))
    RAG_CHUNK_OVERLAP = int(os.getenv("RAG_CHUNK_OVERLAP", "200"))
    RETRIEVAL_SCORE_THRESHOLD = float(os.getenv("RETRIEVAL_SCORE_THRESHOLD", "1.2"))

    # DeepSeek LLM (OpenAI-compatible; used by RAGPipeline for generation)
    # NOTE: DeepSeek has no embedding API — OPENAI_API_KEY is still required for embeddings.
    DEEPSEEK_API_KEY = os.getenv("DEEPSEEK_API_KEY")
    DEEPSEEK_MODEL = os.getenv("DEEPSEEK_MODEL", "deepseek-chat")
    DEEPSEEK_BASE_URL = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com")

    # PostgreSQL (direct connection for the embedding pipeline)
    # Accepts the same DSN as DATABASE_POSTGRES_URL in docker-compose.
    POSTGRES_URL = os.getenv(
        "POSTGRES_URL",
        os.getenv(
            "DATABASE_POSTGRES_URL",
            "postgres://user:password@localhost:5432/docagent?sslmode=disable",
        ),
    )

    # Logging
    LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")