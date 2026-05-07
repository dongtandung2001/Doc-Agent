<!-- markdownlint-disable MD033 -->
<div align="center">

# 🧠 Doc-Agent

**AI-Powered Automated Documentation Generator for Code Repositories**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Python](https://img.shields.io/badge/Python-3.11+-3776AB?logo=python)](https://python.org/)
[![gRPC](https://img.shields.io/badge/gRPC-Enabled-9cf)](https://grpc.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?logo=postgresql)](https://postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis)](https://redis.io/)
[![ChromaDB](https://img.shields.io/badge/Vector_DB-ChromaDB-FC6D26)](https://www.trychroma.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

---

## 📋 Overview

**Doc-Agent** is a fully distributed, microservices-based system that **automatically generates high-quality documentation** for any code repository. By combining **LLM-powered code analysis**, **Retrieval-Augmented Generation (RAG)**, and **parallel document generation**, Doc-Agent transforms raw source code into structured, navigable documentation.

The system analyzes your repository structure, classifies the project type (application, library, framework, CLI tool, etc.), generates a documentation blueprint, and produces comprehensive documentation sections — all automatically and in parallel.

---

## ✨ Key Features

- **🤖 AI-Driven Code Analysis** — Uses LLMs (OpenAI GPT-4o-mini / DeepSeek) to understand project structure, classify project type, and generate accurate documentation tailored to the codebase
- **📚 RAG-Powered Responses** — Retrieval-Augmented Generation with ChromaDB vector store enables contextual Q&A over your documentation. Ask questions and get answers grounded in your project's docs
- **🏗️ True Microservices Architecture** — Independently deployable services written in **Go** and **Python**, communicating via **gRPC** and **Redis-backed message queues**
- **📄 Smart Documentation Generation** — Automatically classifies repositories into categories (Applications, Libraries, Frameworks, CLI Tools, DevOps Configurations, etc.) and generates appropriate documentation structures for each
- **🗂️ Hierarchical Document Organization** — Documentation is organized into nested, navigable section trees with a left-panel browsing experience
- **⚡ Parallel Document Generation** — Asynq task queue (backed by Redis) enables concurrent processing of multiple documentation sections
- **🔍 Vector-Searchable Knowledge Base** — All generated documentation is automatically embedded and indexed in ChromaDB for semantic search
- **🌐 HTTP/gRPC API Gateway** — Unified entry point using Connect-Go (with Fiber middleware) routing to all backend microservices
- **🔄 Idempotent Embedding Pipeline** — Crash-safe RAG pipeline with optimistic locking ensures documents are never embedded twice
- **📊 File Cache with Singleflight** — In-memory LRU file cache with request coalescing prevents redundant network fetches during analysis

---

## 🏛️ Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                        Frontend / CLI                            │
└──────────────────────────┬───────────────────────────────────────┘
                           │ HTTP/Connect-Go
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                    Gateway Service  (Go)                         │
│                    Port: 8080                                    │
│                    Fiber HTTP + gRPC Connect                     │
└──┬────────────┬──────────────┬──────────────┬───────────────────┘
   │ gRPC       │ gRPC         │ gRPC         │ gRPC
   ▼            ▼              ▼              ▼
┌────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐
│Codebase│ │ Database │ │   AI     │ │ Local Agent  │
│Service │ │ Service  │ │ Service  │ │ (Rust/CLI)   │
│(Go)    │ │ (Go)     │ │(Python)  │ │              │
│:9001   │ │ :9002     │ │ :50051   │ │ :50051       │
└────┬───┘ └──────────┘ └──────────┘ └──────────────┘
     │                              ▲
     │ Enqueues tasks               │ gRPC (file reads)
     ▼                              │
┌──────────┐               ┌────────┴────────┐
│  Redis   │               │  DocGen Worker  │
│ Task Q   │◄──────────────│  (Go / Asynq)   │
└──────────┘               │  (no port)      │
                           └────────┬────────┘
                                    │
                           ┌────────▼────────┐
                           │   AI Service    │
                           │ (OpenAI/DeepSeek)│
                           └─────────────────┘
```

### Service Communication Flow

1. **Frontend** sends analysis request to **Gateway** via HTTP/Connect-Go
2. **Gateway** proxies to **Codebase Service** (gRPC)
3. **Codebase Service** orchestrates the pipeline:
   - **Classification** — Calls AI Service to classify the project type
   - **Instruction Generation** — Generates a documentation blueprint (multi-stage AI reasoning)
   - **Enqueue** — Stores documentation sections in **Database Service** and enqueues generation tasks in **Redis**
4. **DocGen Worker** consumes tasks from Redis and for each section:
   - Reads relevant files via **Gateway → Local Agent**
   - Calls **AI Service** to generate the documentation content
   - Stores the result in **Database Service**
5. **AI Service** also handles **RAG** queries — embedding docs into ChromaDB and answering questions with context

---

## 🛠️ Technology Stack

| Component         | Language   | Framework / Libraries                                      |
|-------------------|------------|------------------------------------------------------------|
| **Gateway**       | Go         | Fiber, Connect-Go, gRPC                                    |
| **Codebase**      | Go         | gRPC, Asynq (Redis MQ client), FileCache (singleflight)    |
| **DocGen**        | Go         | Asynq worker, gRPC clients                                 |
| **Database**      | Go         | gRPC, pgx (PostgreSQL), golang-migrate                     |
| **AI Service**    | Python     | gRPC, LangChain, ChromaDB, OpenAI/DeepSeek API             |
| **Local Agent**   | Rust (CLI) | gRPC server for filesystem I/O                             |
| **Message Queue** | —          | Redis + Asynq                                              |
| **Vector DB**     | —          | ChromaDB (embedded)                                        |
| **Database**      | —          | PostgreSQL 15                                              |

---

## 📦 Services

### 1. Gateway Service (Go — Port `8080`)
The API Gateway — single entry point for all frontend communication.

- **Protocol**: HTTP/Connect-Go (supports gRPC and JSON interchangeably)
- **Middleware**: CORS, request logging, panic recovery
- **Routes**: Proxies to Codebase, Database, AI, and Local Agent services
- **Health Check**: Aggregated health of all backend services

### 2. Codebase Analysis Service (Go — Port `9001`)
The orchestrator that drives the documentation generation pipeline.

- **Pipeline**:
  1. **Repo Classification** — AI determines project type (App, Library, Framework, CLI, DevOps, etc.)
  2. **Instruction Generation** — Two-stage AI reasoning generates a JSON documentation blueprint
  3. **Task Enqueue** — Stores sections in PostgreSQL and enqueues generation tasks to Redis
- **Clients**: AI Service (gRPC/HTTP), Gateway (for Local Agent file reads), Database Service, Redis

### 3. Document Generation Worker (Go — No exposed port)
Background worker consuming tasks from Redis message queue.

- **Concurrency**: Processes up to 5 tasks in parallel (configurable)
- **For each task**: Reads source files via Local Agent → Generates content via AI Service → Stores in Database Service
- **File Cache**: In-memory LRU cache (up to 50MB) with singleflight request coalescing to minimize redundant file reads

### 4. Database Service (Go — Port `9002`)
Data persistence layer for documentation content.

- **PostgreSQL Tables**:
  - `document_sections` — Hierarchical section tree (parent/child structure)
  - `document_file_items` — Individual document content with embed status tracking
- **Auto-migration**: Runs migrations on startup with retry logic
- **Operations**: Store/get documents, manage section trees, track embedding status

### 5. AI Service (Python — Port `50051`)
The intelligence layer — handles chat, RAG, and embedding.

- **gRPC API**:
  - `Chat( messages, project_id )` — Conversational Q&A with RAG context
  - `CreateRAG( project_id )` — Triggers embedding pipeline for pending documents
  - `HealthCheck( )` — Service health
- **RAG Pipeline**:
  1. Fetch pending documents from PostgreSQL
  2. Split content into chunks (configurable size/overlap)
  3. Embed with OpenAI `text-embedding-3-small`
  4. Store in ChromaDB vector store
  5. On query: retrieve relevant chunks, generate answer with DeepSeek or OpenAI
- **LLM Support**: OpenAI GPT-4o-mini (default) or DeepSeek Chat (OpenAI-compatible)

---

## 🚀 Getting Started

### Prerequisites

- Go 1.22+
- Python 3.11+
- Docker & Docker Compose
- PostgreSQL 15
- Redis 7
- An **OpenAI API Key** (for embeddings and LLM)
- (Optional) A **DeepSeek API Key** for LLM generation

### Environment Setup

Create a `.env` file in the `backend/` directory:

```bash
# LLM Configuration
OPENAI_API_KEY=sk-your-key-here
LLM_MODEL=gpt-4o-mini

# Optional: DeepSeek (replaces OpenAI for generation, still needs OPENAI_API_KEY for embeddings)
DEEPSEEK_API_KEY=sk-your-deepseek-key
DEEPSEEK_MODEL=deepseek-chat

# PostgreSQL
DATABASE_POSTGRES_URL=postgres://user:password@localhost:5432/docagent?sslmode=disable
```

### Quick Start with Docker Compose

```bash
cd backend
docker-compose up --build
```

This starts all services:

| Service           | Port    |
|-------------------|---------|
| Gateway           | `8080`  |
| Codebase Service  | `9001`  |
| Database Service  | `9002`  |
| AI Service        | `50051` |
| PostgreSQL        | `5432`  |
| Redis             | `6379`  |
| pgAdmin           | `5050`  |

### Local Development (Without Docker)

Run each service in its own terminal:

```bash
# Terminal 1 - Database Service
cd backend/services/database
go run ./cmd/database

# Terminal 2 - Codebase Service
cd backend/services/codebase
go run ./cmd/codebase

# Terminal 3 - DocGen Worker
cd backend/services/docgen
go run ./cmd/docgen

# Terminal 4 - AI Service (Python)
cd backend/services/ai-service
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt
bash scripts/generate_grpc.sh
python main.py

# Terminal 5 - Gateway
cd backend/services/gateway
go run ./cmd/gateway
```

> **Note**: PostgreSQL and Redis must be running locally. Use Docker for those:
> ```bash
> docker run -d -p 5432:5432 -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=docagent postgres:15-alpine
> docker run -d -p 6379:6379 redis:7-alpine
> ```

---

## 🔧 Configuration

### AI Service (`services/ai-service/.env`)

```ini
# gRPC
GRPC_PORT=50051

# LLM
OPENAI_API_KEY=sk-...
LLM_MODEL=gpt-4o-mini
LLM_TEMPERATURE=0.7
LLM_MAX_TOKENS=2000

# DeepSeek (optional, replaces OpenAI for generation)
DEEPSEEK_API_KEY=sk-...
DEEPSEEK_MODEL=deepseek-chat

# Vector DB (ChromaDB)
VECTOR_DB_PATH=./chroma_db
EMBEDDING_MODEL=text-embedding-3-small
TOP_K_RETRIEVAL=5

# RAG
RAG_CHUNK_SIZE=1000
RAG_CHUNK_OVERLAP=200
RETRIEVAL_SCORE_THRESHOLD=1.2

# PostgreSQL (for embedding pipeline)
POSTGRES_URL=postgres://user:password@localhost:5432/docagent?sslmode=disable
```

### Service Configs (YAML)

Each Go service has a YAML config file with environment variable overrides:

```yaml
# Example: services/gateway/configs/gateway.yaml
server:
  host: "0.0.0.0"
  port: 8080
backends:
  codebase:
    host: "localhost"
    port: 9001
  database:
    host: "localhost"
    port: 9002
  ai:
    host: "localhost"
    port: 50051
  local_agent:
    host: "localhost"
    port: 50051
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
```

---

## 📖 API Reference

### Gateway Endpoints (Connect-Go)

| Method | Proto RPC | Description |
|--------|-----------|-------------|
| `POST` | `StartCodebaseAnalysis` | Trigger documentation generation for a repository |
| `POST` | `RequestFileContent` | Read files from the local codebase |
| `POST` | `StoreDocument` | Store a generated document |
| `POST` | `GetDocument` | Retrieve a document by project + document ID |
| `POST` | `GetDocumentSections` | Get the hierarchical section tree |
| `POST` | `Chat` | Ask a question (with RAG context from project docs) |
| `POST` | `CreateRAG` | Trigger embedding of pending documents |
| `GET` | `HealthCheck` | Health check (all services) |

### AI Service gRPC

```protobuf
service AIService {
  rpc Chat(ChatRequest) returns (ChatResponse);
  rpc CreateRAG(CreateRAGRequest) returns (CreateRAGResponse);
  rpc HealthCheck(Empty) returns (HealthCheckResponse);
}
```

---

## 💡 Usage Examples

### Generate Documentation for a Repository

```bash
# Using the Connect-Go gateway
curl -X POST http://localhost:8080/gateway.GatewayService/StartCodebaseAnalysis \
  -H "Content-Type: application/json" \
  -d '{
    "project_structure": "<compacted directory tree>",
    "readme_content": "# My Project\n..."
  }'
```

### Chat with RAG Over Your Docs

```bash
curl -X POST http://localhost:8080/gateway.GatewayService/Chat \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "How does the authentication flow work?"}],
    "project_id": "my-project-123"
  }'
```

### Programmatic Usage (Python — AI Service)

```python
from src.rag_pipeline import RAGPipeline

pipeline = RAGPipeline()

# Embed all pending documents
pipeline.process_pending_documents("my-project-123")

# Ask a question with RAG context
answer = pipeline.generate_with_rag(
    "What does the codebase service do?",
    "my-project-123"
)
print(answer)
```

See [`services/ai-service/examples/rag_usage.py`](services/ai-service/examples/rag_usage.py) for a complete set of examples.

---

## 🧪 Testing

### AI Service (Python)

```bash
cd backend/services/ai-service
pytest tests/ -v
```

### Database Service (Go)

```bash
cd backend/services/database
go test ./internal/service/ -v
```

### Integration & E2E Tests

```bash
cd backend/tests
go test ./integration/...
go test ./e2e/...
```

---

## 🧠 RAG Pipeline Deep Dive

The RAG pipeline is the core intelligence layer:

```
┌──────────────┐     ┌─────────────────┐     ┌──────────────┐
│  PostgreSQL  │────▶│  RAGPipeline     │────▶│   ChromaDB   │
│  (pending    │     │  process_pending │     │  (vector     │
│   docs)      │     │  _documents()    │     │   store)     │
└──────────────┘     └────────┬─────────┘     └──────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │  User Query      │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Similarity Search│
                    │ (ChromaDB)       │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  Re-rank by      │
                    │  Score Threshold │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  LLM (DeepSeek/  │
                    │  OpenAI) +       │
                    │  Context Prompt  │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  Answer          │
                    └──────────────────┘
```

Key design decisions:

- **Optimistic Locking**: Documents are marked `processing` before embedding, preventing duplicate work across workers
- **Crash Recovery**: On startup, documents stuck in `processing` are reset to `pending`
- **Score Threshold**: Irrelevant chunks (L2 distance > 1.2) are filtered out before LLM injection
- **Dual LLM Support**: OpenAI for embeddings (no alternative), DeepSeek or OpenAI for generation
- **Deterministic Chunk IDs**: Re-embedding overwrites existing chunks rather than creating duplicates

---

## 🗄️ Project Structure

```
backend/
├── services/                        # Microservices
│   ├── gateway/                     # API Gateway (Go / Fiber / Connect-Go)
│   │   ├── cmd/gateway/main.go      # Entry point
│   │   ├── internal/
│   │   │   ├── config/              # YAML config loading
│   │   │   ├── handlers/            # gRPC handler (proxies to backends)
│   │   │   ├── http/                # Fiber HTTP server setup
│   │   │   └── middleware/          # CORS, logging, recovery
│   │   └── configs/gateway.yaml
│   ├── codebase/                    # Codebase Analysis (Go / gRPC)
│   │   ├── cmd/codebase/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── grpc/               # gRPC server
│   │   │   ├── pipeline/           # Classification → Instructions → Enqueue
│   │   │   └── service/            # Analysis orchestrator
│   │   └── configs/codebase.yaml
│   ├── docgen/                      # Document Generation Worker (Go / Asynq)
│   │   ├── cmd/docgen/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── pipeline/           # Document generation logic
│   │   │   ├── service/            # Worker startup with Asynq server
│   │   │   └── tasks/              # Task handler (docgen:instruction)
│   │   └── configs/docgen.yaml
│   ├── database/                    # Database Service (Go / gRPC)
│   │   ├── cmd/database/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── db/                 # Migration runner
│   │   │   ├── grpc/               # gRPC server
│   │   │   ├── repository/         # PostgreSQL queries
│   │   │   └── service/            # Business logic + section tree builder
│   │   ├── migrations/             # SQL migration files
│   │   └── configs/database.yaml
│   └── ai-service/                  # AI Service (Python / gRPC / LangChain)
│       ├── main.py                  # Entry point
│       ├── src/
│       │   ├── ai_service_impl.py   # gRPC service implementation
│       │   ├── config.py            # Environment config
│       │   ├── conversation_orchestrator.py  # RAG-aware chat orchestrator
│       │   ├── db_client.py         # Direct PostgreSQL client (psycopg2)
│       │   ├── llm_client.py        # OpenAI API client
│       │   ├── rag_pipeline.py      # Full RAG lifecycle (embed + retrieve + generate)
│       │   ├── server.py            # gRPC server startup
│       │   ├── vector_store.py      # ChromaDB management
│       │   └── sqlite_patch.py      # SQLite fix for ChromaDB
│       ├── protos/                  # Proto definitions
│       ├── generated/               # Generated gRPC code
│       ├── tests/                   # Pytest test suite
│       └── examples/rag_usage.py    # Usage examples
├── shared/                          # Shared Go module
│   ├── api/proto/v1/               # Proto definitions (ai, codebase, database, gateway, localagent)
│   ├── pkg/
│   │   ├── clients/                # gRPC client factories + AIClient (agentic loop)
│   │   ├── context/                # Chat context (key-value context for prompts)
│   │   ├── errors/                 # Shared error types
│   │   └── utils/prompt/           # Template variable processor
│   ├── buf.gen.yaml                # Buf code generation config
│   ├── buf.yaml                    # Buf lint/breaking config
│   ├── go.mod
│   └── go.sum
├── configs/                         # Global service configs
├── deployments/                     # Docker, Docker Compose, K8s manifests
│   ├── docker/
│   ├── docker-compose/
│   └── k8s/                        # Kubernetes deployments per service
├── tests/
│   ├── e2e/
│   └── integration/
└── docker-compose.yml               # Full local development stack
```

---

## 🏗️ Key Design Principles

### True Microservices
Each service is a fully independent module:
- ✅ Own `go.mod` / `requirements.txt` — no cross-service compile-time dependencies
- ✅ Independent build and deployment
- ✅ Communication only via **gRPC** or **message queue**
- ✅ Independent scaling

### Modular Shared Code
The `shared/` Go module provides:
- Proto definitions and generated gRPC stubs
- Client factories for every backend service
- Chat context (key-value store for prompt template variables)
- File cache with singleflight request coalescing
- Prompt template processor (`{{$variable}}` substitution)

### Crash-Safe Pipeline
- Optimistic locking via `embed_status` column prevents duplicate processing
- Startup recovery resets stale `processing` entries to `pending`
- Failed embeddings are marked with `failed` status (manual retry)

---

## 🔮 Roadmap

- [ ] **Message Queue**: Replace in-process Redis/Asynq with RabbitMQ or Kafka for production workloads
- [ ] **Service Discovery**: Add Consul or etcd for dynamic service registration
- [ ] **Observability**: Distributed tracing with Jaeger/Zipkin, metric collection with Prometheus
- [ ] **API Documentation**: Auto-generate OpenAPI/Swagger docs from proto definitions
- [ ] **Multi-Project Support**: Full project isolation with authentication & authorization
- [ ] **CI/CD Pipelines**: Independent deployment pipelines per service (GitHub Actions)
- [ ] **Web UI**: React/Vue frontend for browsing generated documentation
- [ ] **Custom LLM Support**: Pluggable LLM backends (Anthropic, local models via Ollama)

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. **Fork** the repository
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes** following the existing code style
4. **Write tests** for new functionality
5. **Run the test suite** to ensure nothing breaks
6. **Commit** with clear, descriptive messages
7. **Open a Pull Request**

### Development Guidelines

- Each service has its own `go.mod` — treat it as an independent module
- Proto changes in `shared/api/proto/v1/` require regenerating Go stubs via `buf generate`
- For the AI service, gRPC Python code is regenerated via `bash scripts/generate_grpc.sh`
- Add environment variable documentation to the relevant service's `config.py` or `config.go`

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

<div align="center">
  <sub>Built with ❤️ using Go, Python, and a lot of ☕</sub>
</div>