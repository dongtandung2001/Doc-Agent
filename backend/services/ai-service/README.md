# AI Service Microservice

[![gRPC](https://img.shields.io/badge/gRPC-1.60-blue)](https://grpc.io/)
[![Python](https://img.shields.io/badge/Python-3.11+-brightgreen)](https://www.python.org/)
[![LangChain](https://img.shields.io/badge/LangChain-0.1-orange)](https://www.langchain.com/)
[![ChromaDB](https://img.shields.io/badge/Vector%20DB-ChromaDB-purple)](https://www.trychroma.com/)
[![License](https://img.shields.io/badge/License-Proprietary-lightgrey)]()

A production-grade **gRPC-based AI service** that powers **Retrieval-Augmented Generation (RAG)** for smart codebase documentation. This microservice is part of the **Doc-Agent** ecosystem — it ingests project documentation, indexes it into a vector database, and answers natural-language questions by retrieving and synthesizing relevant context.

Built with **LangChain**, **OpenAI embeddings**, **ChromaDB**, and **DeepSeek/OpenAI LLMs**, it provides a robust, scalable, and containerized foundation for documentation Q&A.

---

## Features

- **⚡ gRPC API** — High-performance, strongly-typed API with three RPC endpoints (`Chat`, `CreateRAG`, `HealthCheck`) defined via Protocol Buffers
- **🔍 Full RAG Pipeline** — End-to-end retrieval-augmented generation: document ingestion → chunking → embedding → similarity search → context-aware LLM answer generation
- **🧠 Dual LLM Support** — Uses **DeepSeek Chat** (OpenAI-compatible) for cost-effective generation; falls back to **OpenAI GPT-4o-mini** if DeepSeek key is not configured. Embeddings always use **OpenAI `text-embedding-3-small`**
- **🗄️ Vector Database (ChromaDB)** — Persistent, per-project vector collections for fast similarity search with configurable score thresholds and deduplication
- **📦 Automatic Document Embedding** — `CreateRAG` RPC fetches documents with `pending` status from PostgreSQL, embeds them, and updates status atomically with race-condition protection
- **🛡️ Production-Grade** — Docker multi-stage builds, Kubernetes deployment manifests, health checks, structured logging, thread-safe PostgreSQL connection pooling
- **🔄 Stale Document Recovery** — On startup, automatically resets documents stuck in `processing` state back to `pending` (crash recovery)
- **🌐 HTTP Proxy Bridge** — Included HTTP server (`scripts/serve_web_test.py`) translates REST calls to gRPC for web frontends
- **🐳 Full Containerization** — Docker Compose for local development and production deployment, with hot-reload support in development mode

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    gRPC Client                           │
│        (Gateway / Web App / Other Services)              │
└────────────────────────┬────────────────────────────────┘
                         │
                    gRPC Port 50051
                         │
┌────────────────────────▼────────────────────────────────┐
│                    AI Service                            │
│  ┌──────────────────────────────────────────────────┐   │
│  │           AIServiceServicer (gRPC impl)           │   │
│  │   Chat()  │  CreateRAG()  │  HealthCheck()       │   │
│  └─────┬──────┴──────┬───────┴──────────────────────┘   │
│        │              │                                  │
│  ┌─────▼──────┐ ┌────▼────────┐                         │
│  │Conversation│ │RAGPipeline  │                         │
│  │Orchestrator│ │             │                         │
│  │(QA Chain)  │ │▪ Embed docs│                         │
│  └─────┬──────┘ │▪ Retrieve  │                         │
│        │        │▪ Generate  │                         │
│        │        └──┬────┬────┘                         │
│  ┌─────▼──────┐    │    │                              │
│  │ LLM Client │    │    │                              │
│  │(OpenAI)    │    │    │                              │
│  └────────────┘    │    │                              │
│                    │    │                              │
│  ┌─────────────────▼┐ ┌▼────────────────┐             │
│  │  VectorStore     │ │  DatabaseClient │             │
│  │  (ChromaDB)      │ │  (PostgreSQL)   │             │
│  └──────────────────┘ └─────────────────┘             │
└────────────────────────────────────────────────────────┘
```

### Core Components

| Component | Description |
|---|---|
| **AIServiceServicer** | gRPC service implementation; validates requests, orchestrates Chat and CreateRAG flows |
| **ConversationOrchestrator** | LangChain `load_qa_chain` (stuff type) — retrieves relevant docs, injects into prompt, calls LLM |
| **RAGPipeline** | Full pipeline: embedding new docs from PostgreSQL, similarity retrieval, and DeepSeek/OpenAI generation |
| **LLMClient** | OpenAI chat completions client (used for direct LLM calls without RAG context) |
| **VectorStoreManager** | ChromaDB abstraction — lazy per-project collections, text splitting, embedding storage/retrieval/deletion |
| **DatabaseClient** | PostgreSQL (psycopg2) with thread-safe connection pool; manages document embed status lifecycle |
| **Config** | Centralized environment-based configuration (`.env` or environment variables) |

---

## Getting Started

### Prerequisites

- **Python 3.11+** (required for bundled SQLite ≥ 3.35.0 for ChromaDB compatibility)
- An **OpenAI API key** (for embeddings; also used as fallback LLM)
- (Optional) A **DeepSeek API key** for cost-effective LLM generation
- A **PostgreSQL** database with a `document_file_items` table (or deploy alongside the Doc-Agent backend stack)

### 1. Clone and Setup

```bash
git clone <repo-url>
cd backend/services/ai-service
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

### 2. Install Dependencies

```bash
pip install -r requirements.txt
```

> **macOS SQLite Note:** If you encounter SQLite version errors with ChromaDB, run `bash FIX_SQLITE.sh` or ensure Python 3.11+ is used. The service includes a `sqlite_patch.py` that attempts to use `pysqlite3` as a fallback.

### 3. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

**Required environment variables:**

| Variable | Description |
|---|---|
| `OPENAI_API_KEY` | OpenAI API key (for embeddings and optional LLM fallback) |
| `POSTGRES_URL` | PostgreSQL DSN (e.g., `postgres://user:pass@localhost:5432/docagent?sslmode=disable`) |

**Optional environment variables:**

| Variable | Default | Description |
|---|---|---|
| `DEEPSEEK_API_KEY` | — | Use DeepSeek for LLM generation (cheaper than GPT-4) |
| `DEEPSEEK_MODEL` | `deepseek-chat` | DeepSeek model name |
| `DEEPSEEK_BASE_URL` | `https://api.deepseek.com` | DeepSeek API endpoint |
| `GRPC_PORT` | `50051` | gRPC server port |
| `LLM_MODEL` | `gpt-4o-mini` | OpenAI model for generation (used if no DeepSeek key) |
| `LLM_TEMPERATURE` | `0.7` | LLM temperature |
| `LLM_MAX_TOKENS` | `2000` | Maximum tokens per response |
| `VECTOR_DB_PATH` | `./chroma_db` | ChromaDB persistence directory |
| `EMBEDDING_MODEL` | `text-embedding-3-small` | OpenAI embedding model |
| `TOP_K_RETRIEVAL` | `5` | Number of chunks to retrieve |
| `RAG_CHUNK_SIZE` | `1000` | Character chunk size for document splitting |
| `RAG_CHUNK_OVERLAP` | `200` | Overlap between chunks |
| `RETRIEVAL_SCORE_THRESHOLD` | `1.2` | Max L2 distance for retrieved chunks (lower = stricter) |
| `LOG_LEVEL` | `INFO` | Logging level |

### 4. Generate gRPC Stubs

```bash
# Using the script (recommended):
bash scripts/generate_grpc.sh

# Or manually:
mkdir -p generated
python -m grpc_tools.protoc \
    -I./protos \
    --python_out=./generated \
    --grpc_python_out=./generated \
    ./protos/ai_service.proto
```

### 5. Run the Service

```bash
# Direct Python:
python main.py

# Or using Make:
make run
```

The server starts on port **50051** (configurable via `GRPC_PORT`).

---

## Docker Deployment

### Using Docker Compose

```bash
# Production:
docker-compose up -d

# Development (with code mount for hot reload):
docker-compose -f docker-compose.dev.yml up -d
```

### Build and Run Manually

```bash
# Build image:
docker build -t ai-service:latest .

# Run container:
docker run -p 50051:50051 \
    -e OPENAI_API_KEY=sk-... \
    -e POSTGRES_URL=postgres://... \
    ai-service:latest
```

### Kubernetes

A production-ready Kubernetes deployment manifest is provided at `ai-service-deployment.yaml`, including ConfigMap, Secrets, Deployment (2 replicas), and ClusterIP Service with liveness/readiness probes.

---

## Usage

### Quick Test

Run the built-in smoke test to verify the service is running:

```bash
python quick_test.py
```

### gRPC API Endpoints

#### `HealthCheck`

Verify the service is alive:

```python
import grpc
import generated.ai_service_pb2 as pb2
import generated.ai_service_pb2_grpc as svc

channel = grpc.insecure_channel("localhost:50051")
stub = svc.AIServiceStub(channel)
response = stub.HealthCheck(pb2.Empty())
print(f"Alive: {response.isAlive}")  # True
```

#### `CreateRAG`

Trigger embedding of pending documents for a project:

```python
request = pb2.CreateRAGRequest(project_id="my-project")
response = stub.CreateRAG(request)
print(response.message)  # "Embedded 5 document(s) for project my-project"
```

This fetches all documents with `embed_status='pending'` from PostgreSQL, splits them into chunks, generates embeddings via OpenAI, and stores them in ChromaDB. Safe to call repeatedly — already-embedded documents are skipped.

#### `Chat`

Send a conversational message (with automatic RAG context retrieval):

```python
request = pb2.ChatRequest(
    messages=[pb2.ChatMessage(role="user", content="How does authentication work?")],
    project_id="my-project"
)
response = stub.Chat(request)
print(response.content)  # LLM-generated answer with context from project docs
```

The `Chat` endpoint:
1. Automatically retrieves the **top 5 relevant document chunks** from ChromaDB
2. Filters out chunks above the `RETRIEVAL_SCORE_THRESHOLD`
3. Injects them into a context-aware prompt
4. Generates an answer via the LLM (DeepSeek or OpenAI)
5. Returns both `content` (plain text answer) and `completion_json` (full OpenAI-compatible response)

### Python Example (Direct RAG Pipeline Usage)

```python
from src.rag_pipeline import RAGPipeline

pipeline = RAGPipeline()

# Embed all pending documents
count = pipeline.process_pending_documents("my-project")
print(f"Embedded {count} documents")

# Retrieve relevant chunks
results = pipeline.retrieve("How does the gateway route requests?", "my-project")
for r in results:
    print(f"Score: {r['score']:.3f} | {r['content'][:80]}...")

# Full RAG Q&A
answer = pipeline.generate_with_rag(
    "What does the codebase service do?",
    "my-project"
)
print(answer)
```

See [`examples/rag_usage.py`](./examples/rag_usage.py) for complete standalone examples.

### HTTP Bridge (for Web Frontends)

An HTTP proxy server translates REST calls to gRPC for web frontends:

```bash
# Start the AI gRPC server first (python main.py),
# then start the HTTP bridge:
python scripts/serve_web_test.py

# The bridge serves on http://localhost:8888
```

Available HTTP endpoints:
- `GET /api/health` — Health check
- `POST /api/chat` — Chat (body: `{"messages": [...], "project_id": "..."}`)
- `POST /api/create_rag` — Trigger indexing (body: `{"project_id": "..."}`)

---

## Testing

```bash
# Run all tests with coverage:
pytest tests/ -v --cov=. --cov-report=html

# Or using Make:
make test
```

### Test Suite

| Test File | Description |
|---|---|
| `tests/test_ai_service.py` | gRPC service validation — health check, chat validation, CreateRAG flow |
| `tests/test_vector_store.py` | Vector store embedding storage and retrieval with mocked ChromaDB |
| `tests/test_client.py` | End-to-end client tests |
| `tests/test_conversation_orchestrator.py` | Orchestrator logic tests |

---

## Project Structure

```
ai-service/
├── main.py                      # Entry point
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Production Docker Compose
├── docker-compose.dev.yml       # Dev Docker Compose (hot reload)
├── Makefile                     # Common commands
├── requirements.txt             # Python dependencies
├── constraints.txt              # Pinned dependency versions
├── FIX_SQLITE.sh                # SQLite fix for macOS
├── .dockerignore
├── ai-service-deployment.yaml   # Kubernetes deployment manifests
│
├── protos/
│   └── ai_service.proto         # gRPC service definition
│
├── generated/                   # Auto-generated gRPC stubs
│   ├── ai_service_pb2.py
│   ├── ai_service_pb2_grpc.py
│   └── __init__.py
│
├── src/                         # Core application code
│   ├── __init__.py
│   ├── server.py                # gRPC server setup
│   ├── ai_service_impl.py       # gRPC service implementation
│   ├── config.py                # Environment configuration
│   ├── conversation_orchestrator.py  # RAG decision logic & QA chain
│   ├── rag_pipeline.py          # Full RAG pipeline
│   ├── llm_client.py            # OpenAI LLM client
│   ├── vector_store.py          # ChromaDB abstraction
│   ├── db_client.py             # PostgreSQL client (psycopg2)
│   ├── logger.py                # Structured logging
│   └── sqlite_patch.py          # SQLite compatibility patch
│
├── tests/                       # Test suite
│   ├── test_ai_service.py
│   ├── test_client.py
│   ├── test_conversation_orchestrator.py
│   └── test_vector_store.py
│
├── examples/
│   └── rag_usage.py             # Standalone RAG pipeline examples
│
├── scripts/
│   ├── generate_grpc.sh         # gRPC code generation script
│   ├── serve_web_test.py        # HTTP→gRPC bridge proxy
│   ├── test_api_key.py          # API key validation
│   ├── test_full_response.py    # Full response integration test
│   └── verify_response_changes.py  # Response diff verification
│
├── chroma_db/                   # Local ChromaDB persistence
└── data/
    └── chroma_db/               # Docker-mounted ChromaDB data
```

---

## Database Schema

The service expects a `document_file_items` table with at least these columns:

| Column | Type | Description |
|---|---|---|
| `id` | UUID | Primary key |
| `project_id` | VARCHAR | Project identifier |
| `document_id` | VARCHAR | Unique document identifier |
| `title` | TEXT | Document title |
| `content` | TEXT | Document content to embed |
| `description` | TEXT | Document description |
| `embed_status` | VARCHAR | `pending`, `processing`, `completed`, or `failed` |
| `is_embedded` | BOOLEAN | Whether content has been indexed |
| `created_at` | TIMESTAMP | Row creation time |
| `updated_at` | TIMESTAMP | Last update time |

---

## Data Flow: From Document to Answer

```
1. Document Stored
   └─> PostgreSQL (document_file_items, embed_status='pending')

2. CreateRAG called
   └─> RAGPipeline.process_pending_documents()
       ├─> db_client.get_pending_documents()  → fetch 'pending' rows
       ├─> db_client.mark_processing()         → atomic lock
       ├─> For each document:
       │   ├─> RecursiveCharacterTextSplitter → chunks
       │   ├─> OpenAI text-embedding-3-small  → vectors
       │   └─> ChromaDB.add_texts()           → persist
       └─> db_client.mark_completed() / failed()

3. User sends Chat message
   └─> ConversationOrchestrator.process_request()
       ├─> vector_store.similarity_search()  → top-5 chunks
       ├─> PromptTemplate({context, question})
       ├─> LangChain load_qa_chain
       └─> LLM generates answer ← DeepSeek or OpenAI
```

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `pytest tests/ -v`
5. Commit with clear messages
6. Push and open a Pull Request

**Development tips:**
- Use the dev Docker Compose (`docker-compose.dev.yml`) for hot-reload during development
- Run `make proto` after modifying the `.proto` file
- Keep SQLite compatibility in mind when adding ChromaDB-dependent code
- Ensure no API-breaking changes to the gRPC proto unless absolutely necessary

---

## Available Make Commands

```bash
make install      # Install Python dependencies
make proto        # Generate gRPC code from proto file
make run          # Run the AI service
make test         # Run tests with coverage
make clean        # Clean generated files and caches
make docker-build # Build Docker image
make docker-run   # Run service in Docker
make docker-logs  # Follow service logs
make docker-stop  # Stop Docker containers
```

---

## Notes

- **Embeddings vs. Generation:** Since DeepSeek does not provide an embedding API, OpenAI's `text-embedding-3-small` is always used for vector embeddings regardless of the LLM choice.
- **Idempotent CreateRAG:** Calling `CreateRAG` multiple times is safe. Only documents with `embed_status='pending'` are processed. Already-completed or failed documents are skipped.
- **Crash Recovery:** On startup, any documents stuck in `processing` status (from a previous crash) are automatically reset to `pending` so they will be picked up on the next `CreateRAG` call.