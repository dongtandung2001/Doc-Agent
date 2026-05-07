<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://img.shields.io/badge/Status-Active-00E676?style=for-the-badge">
    <img alt="Doc-Agent Logo" src="https://img.shields.io/badge/Doc--Agent-v1.0-135bec?style=for-the-badge">
  </picture>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Language-Rust-orange?logo=rust" alt="Rust">
  <img src="https://img.shields.io/badge/Language-Go-00ADD8?logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Language-Python-3776AB?logo=python" alt="Python">
  <img src="https://img.shields.io/badge/Architecture-Microservices-6a0dad" alt="Microservices">
  <img src="https://img.shields.io/badge/LLM-OpenAI%20%7C%20DeepSeek-brightgreen" alt="LLM">
  <img src="https://img.shields.io/badge/Vector_Store-ChromaDB-purple" alt="ChromaDB">
  <img src="https://img.shields.io/badge/Database-PostgreSQL-4169E1?logo=postgresql" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Message_Queue-Redis-DC382D?logo=redis" alt="Redis">
  <img src="https://img.shields.io/badge/License-MIT-yellow" alt="MIT License">
</p>

<h1 align="center">Doc-Agent 🤖📝</h1>

<p align="center">
  <strong>An intelligent, multi-agent documentation generation system</strong><br>
  Automatically analyzes codebases, classifies project types, and generates comprehensive documentation — powered by AI, RAG pipelines, and a true microservice architecture.
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-architecture">Architecture</a> •
  <a href="#-getting-started">Getting Started</a> •
  <a href="#-usage">Usage</a> •
  <a href="#-api-reference">API Reference</a> •
  <a href="#-contributing">Contributing</a>
</p>

---

## ✨ Features

### 🤖 AI-Powered Interactive Chat (Rust CLI)
- Full-duplex chat session with any OpenAI-compatible LLM API (OpenAI, DeepSeek, etc.)
- State-machine-driven conversation loop with robust error recovery and streaming-disabled mode
- **Three built-in file system tools** exposed to the LLM via OpenAI function calling:
  - `fs_read` — Read file contents with optional line range selection
  - `fs_scan` — List directory contents with configurable depth control
  - `ignore_scan` — Scan directories respecting nested `.gitignore` rules
- **Parallel tool execution** via `tokio::spawn` with performance instrumentation (`tracing`-based timing logs)
- **Tool approval workflow** — non-read tools require user confirmation by default (`TRUST_ALL_TOOLS` to bypass)
- **Slash command system** — `/quit`, `/init`, `/doc_generation`, `/readme`

### 📄 Automated Document Generation
- **Codebase classification** — AI analyzes project structure and classifies it (Application, Library, Framework, CLI Tool, DevOps, etc.)
- **Intelligent instruction generation** — Hierarchical documentation blueprint tailored to project type via multi-stage AI reasoning
- **Task queue architecture** — Documentation tasks enqueued to Redis (asynq) and consumed by worker services
- **Section-by-section generation** — Each documentation section generated independently using AI with access to actual source files
- **Auto-embedding** — Generated docs are automatically embedded into a vector database for RAG retrieval

### 🔍 RAG-Powered Q&A (Python AI Service)
- **LangChain + ChromaDB** vector store for document embeddings
- **OpenAI `text-embedding-3-small`** for high-quality embeddings
- **DeepSeek / OpenAI GPT** for LLM generation (dual-LLM support)
- **Automatic chunking** with configurable chunk size and overlap
- **Score-based retrieval filtering** — irrelevant chunks (L2 distance > threshold) are dropped before prompting
- **Optimistic locking** — worker-safe document embedding with status tracking (pending → processing → completed/failed)
- **Crash recovery** — documents stuck in `processing` state are automatically reset on startup

### 🧩 True Microservice Architecture (Go & Python)
- **Gateway Service** (Fiber + Connect-Go) — Single entry point, proxies to all backend services, supports HTTP/gRPC interchangeably
- **Codebase Analysis Service** — Orchestrates the document generation pipeline (classify → instruct → enqueue)
- **Database Service** — Manages PostgreSQL for structured document storage with nested section trees
- **Document Generation Worker** — Consumes tasks from Redis message queue, generates documentation sections concurrently (up to 5 parallel tasks)
- **AI Service** — Python gRPC server for LLM interactions, RAG embedding pipeline, and ChromaDB vector management
- **Local Agent** (Rust) — Secure local filesystem access via gRPC, runs on the user's machine

### 🗄️ Persistent Storage & Caching
- **PostgreSQL 15** for structured data: document sections, file items, embedding status
- **ChromaDB** for vector embeddings and semantic search
- **Redis** for task queuing (asynq), caching, and pub/sub
- **File-level LRU cache** (up to 50MB) with singleflight request coalescing to avoid redundant network fetches

### 🌐 Modern Web Frontend
- Single-page documentation viewer built with **Tailwind CSS**
- **Markdown rendering** with syntax highlighting (highlight.js)
- **Mermaid diagram** support for inline architecture diagrams
- **Nested table of contents** with search functionality
- Resizable AI assistant chat panel with drag-to-resize support

---

## 🏗️ Architecture

```
                    ┌──────────────┐
                    │  Web Frontend│
                    │  (HTML/CSS)  │
                    └──────┬───────┘
                           │ HTTP/Connect-Go
                    ┌──────▼───────┐
                    │   Gateway    │ ◄── Connect-Go (HTTP/gRPC)
                    │  (Go/Fiber)  │
                    └──┬──┬──┬──┬──┘
                       │  │  │  │
          ┌────────────┘  │  │  └──────────────┐
          ▼               ▼  ▼                  ▼
  ┌──────────────┐ ┌──────────────┐  ┌─────────────────┐
  │ Codebase     │ │ Database     │  │ AI Service      │
  │ Analysis     │ │ Service      │  │ (Python/gRPC)   │
  │ (Go)         │ │ (Go/Postgres)│  │ - Chat / RAG    │
  └──────┬───────┘ └──────────────┘  │ - ChromaDB      │
         │                           │ - Embeddings    │
         │ Redis MQ                  └─────────────────┘
         ▼
  ┌──────────────┐
  │ DocGen       │
  │ Worker (Go)  │
  └──────────────┘
         │
         ▼ Filesystem (via Gateway)
  ┌──────────────┐
  │ Local Agent  │
  │ (Rust/CLI)   │
  │ - fs_read    │
  │ - fs_scan    │
  │ - ignore_scan│
  └──────────────┘
```

### Service Communication Flow

```
User runs: /doc_generation
       │
       ▼
┌──────────────┐    1. Scan directory (ignore_scan)      ┌──────────────┐
│  Local Agent │ ─────────────────────────────────────►   │  Filesystem  │
│   (Rust)     │ ◄─────────────────────────────────────   │              │
│              │    2. Return project structure           └──────────────┘
│              │
│              │    3. POST /analyze (project_structure)
│              │ ─────────────────────────────────────►
└──────────────┘                                         ┌──────────────┐
                                                         │   Gateway    │
                                                         │    (Go)      │
                                                         └──────┬───────┘
                                                                │ 4. gRPC StartCodebaseAnalysis
                                                                ▼
                                                         ┌──────────────────┐
                                                         │  Codebase        │
                                                         │  Analysis (Go)   │
                                                         └──────┬───────────┘
                                                   5. Classify │ 6. Generate Instructions
                                                          │    │
                                                          ▼    ▼
                                                    ┌──────────────┐
                                                    │ AI Service   │
                                                    │ (Python/LLM) │
                                                    └──────────────┘
                                                           │
                                                   7. Enqueue tasks to Redis MQ
                                                           │
                                                           ▼
                                                    ┌──────────────┐
                                                    │  DocGen      │
                                                    │  Worker (Go) │
                                                    └──────┬───────┘
                                                   8. Generate │ 9. Store
                                                    per-section │ Document
                                                    documentation│
                                                          │    │
                                                          ▼    ▼
                                                    ┌──────────────┐   ┌──────────────┐
                                                    │ AI Service   │   │  Database    │
                                                    │ (with RAG)   │   │  Service     │
                                                    └──────────────┘   └──────────────┘
                                                           │
                                                   10. Auto-embed into ChromaDB
```

---

## 🛠️ Technology Stack

| Component              | Technology                                                                 |
|------------------------|----------------------------------------------------------------------------|
| **Chat CLI**           | Rust 1.89+, Tokio, Clap, rustyline, Tonic (gRPC), reqwest                  |
| **API Gateway**        | Go 1.25+, Fiber, Connect-Go                                                 |
| **Codebase Analysis**  | Go, gRPC, Redis (asynq), Viper (config)                                    |
| **Document Generation**| Go, asynq (task queue), singleflight file cache                             |
| **Database Service**   | Go, gRPC, pgx (PostgreSQL), golang-migrate                                  |
| **AI Service**         | Python 3.11+, LangChain, ChromaDB, OpenAI, DeepSeek, gRPC                  |
| **LLM Models**         | OpenAI GPT-4o-mini, text-embedding-3-small, DeepSeek Chat                  |
| **Vector Store**       | ChromaDB                                                                   |
| **Message Queue**      | Redis + hibiken/asynq                                                      |
| **SQL Database**       | PostgreSQL 15+                                                             |
| **Frontend**           | Tailwind CSS, marked.js, highlight.js, Mermaid, Font Awesome Material Icons|
| **Deployment**         | Docker, Docker Compose, Kubernetes                                         |

---

## 🚀 Getting Started

### Prerequisites

| Component    | Required                          |
|--------------|-----------------------------------|
| Rust         | 1.70+ (for Local Agent CLI)       |
| Go           | 1.21+ (for backend microservices) |
| Python       | 3.10+ (for AI Service)            |
| PostgreSQL   | 15+                               |
| Redis        | Latest                            |
| Docker       | (optional, for containerized deployment) |

### Quick Start (Docker Compose)

```bash
# Clone the repository
git clone https://github.com/your-org/doc-agent.git
cd doc-agent

# Set up environment variables
cp .env.example .env
# Edit .env with your API keys (OPENAI_API_KEY, DEEPSEEK_API_KEY, etc.)

# Start all services
cd backend
docker-compose up --build
```

This starts the full stack:

| Service           | Port    |
|-------------------|---------|
| Gateway           | `8080`  |
| Codebase Service  | `9001`  |
| Database Service  | `9002`  |
| AI Service        | `50051` |
| PostgreSQL        | `5432`  |
| Redis             | `6379`  |
| pgAdmin           | `5050`  |

### Manual Setup

#### 1. Backend Services (Go)

```bash
# Generate protobuf code
cd backend/shared
buf generate

# Build each service
cd services/gateway && go build ./cmd/gateway
cd services/codebase && go build ./cmd/codebase
cd services/database && go build ./cmd/database
cd services/docgen && go build ./cmd/docgen
```

#### 2. AI Service (Python)

```bash
cd backend/services/ai-service

# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Generate gRPC code
bash scripts/generate_grpc.sh

# Configure and run
cp .env.example .env
# Edit .env with your OPENAI_API_KEY
python main.py
```

#### 3. Local Agent CLI (Rust)

```bash
cd local-agent

# Build release binary
cargo build --release

# Configure
cp .env.example .env
# Edit .env with your API settings

# Start interactive chat
cargo run --release -- chat
```

### Environment Configuration

**Root `.env` (backend/):**
```ini
# LLM Configuration
OPENAI_API_KEY=sk-your-key-here
LLM_MODEL=gpt-4o-mini

# DeepSeek (optional, replaces OpenAI for generation)
DEEPSEEK_API_KEY=sk-your-deepseek-key
DEEPSEEK_MODEL=deepseek-chat

# API Gateway
CHAT_API_URL=https://api.deepseek.com/chat/completions
CHAT_API_KEY=sk-your-api-key
CHAT_MODEL=deepseek-chat

# PostgreSQL
DATABASE_POSTGRES_URL=postgres://user:password@localhost:5432/docagent?sslmode=disable
```

**AI Service (`backend/services/ai-service/.env`):**
```ini
# gRPC
GRPC_PORT=50051

# LLM
OPENAI_API_KEY=sk-...
LLM_MODEL=gpt-4o-mini
LLM_TEMPERATURE=0.7
LLM_MAX_TOKENS=2000

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

**Local Agent (`local-agent/.env`):**
```ini
# API Configuration
CHAT_API_URL=https://api.token-ai.cn/v1/chat/completions
CHAT_API_KEY=your-api-key-here
CHAT_MODEL=DeepSeek-V3

# Logging
RUST_LOG=info

# gRPC server
GRPC_HOST=127.0.0.1
GRPC_PORT=50051

# Tool trust (set to true to skip approval prompts)
TRUST_ALL_TOOLS=false
```

---

## 💻 Usage

### Start an Interactive Chat Session

```bash
cd local-agent
cargo run --release -- chat
```

```
Welcome to Chat CLI!
Type /quit to exit

You: Show me the structure of this project
Doc-agent: I'll scan that for you.

Tools to execute:
  - Scan . (depth: 2)

Approve? [y/N]: y

Executing tools...
[D] src
  [F] main.rs
  [F] lib.rs
  ...
```

### Generate Documentation

```bash
# Inside the chat session
You: /doc_generation
Starting document generation...
✓ gRPC server started successfully on 127.0.0.1:50051
✓ Codebase analysis started successfully.

# Or use the /readme command to generate/improve README
You: /readme
Found existing README.md — improving it...
✓ README.md written to ./README.md
```

### Using Slash Commands

| Command                        | Description                                       |
|--------------------------------|---------------------------------------------------|
| `/quit`, `/exit`, `/q`        | Exit the chat session                             |
| `/init`, `/start`             | Launch document generation pipeline               |
| `/doc_generation [path]`      | Full document generation with gRPC server spawn   |
| `/readme [path]`              | Generate or improve a README.md for a directory   |

### Programmatic Usage (Rust)

```rust
use local_agent::cli::chat::{ChatArgs, ChatSession};

let args = ChatArgs::default();
let mut session = ChatSession::new(args).await?;
let response = session.send_message("What's in this project?").await?;
println!("{}", response);
```

### gRPC Service (on-demand, after `/init`)

```bash
# Check health
grpcurl -plaintext 127.0.0.1:50051 api.proto.v1.LocalAgentService/HealthCheck

# Request file content
grpcurl -plaintext -d '{"args":[{"id":"1","path":"README.md"}]}' \
  127.0.0.1:50051 api.proto.v1.LocalAgentService/RequestFileContent
```

### RAG-Powered Q&A (Python)

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

---

## 📖 API Reference

### Gateway Endpoints (Connect-Go)

Base URL: `http://localhost:8080`

| Method | Endpoint                                             | Description                                       |
|--------|------------------------------------------------------|---------------------------------------------------|
| `POST` | `/api.proto.v1.GatewayService/Chat`                  | Chat with AI assistant with RAG context           |
| `POST` | `/api.proto.v1.GatewayService/CreateRAG`             | Trigger RAG embedding pipeline                    |
| `POST` | `/api.proto.v1.GatewayService/StartCodebaseAnalysis` | Start automatic document generation               |
| `POST` | `/api.proto.v1.GatewayService/GetDocument`           | Retrieve generated document                       |
| `POST` | `/api.proto.v1.GatewayService/GetDocumentSections`   | Get nested document tree for navigation           |
| `POST` | `/api.proto.v1.GatewayService/RequestFileContent`    | Read file content via Local Agent                 |
| `GET`  | `/health`                                            | Simple gateway health check                       |
| `POST` | `/api.proto.v1.GatewayService/HealthCheck`           | Full health check (all backend services)          |

### AI Service gRPC

```protobuf
service AIService {
  rpc Chat(ChatRequest) returns (ChatResponse);
  rpc CreateRAG(CreateRAGRequest) returns (CreateRAGResponse);
  rpc HealthCheck(Empty) returns (HealthCheckResponse);
}
```

### Chat API Example

```bash
curl -X POST http://localhost:8080/api.proto.v1.GatewayService/Chat \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "How does authentication work?"}],
    "project_id": "my-project-123"
  }'
```

### Request File Content

```bash
curl -X POST http://localhost:8080/api.proto.v1.GatewayService/RequestFileContent \
  -H "Content-Type: application/json" \
  -d '{"paths": ["src/main.rs", "src/lib.rs"]}'
```

### Start Codebase Analysis

```bash
curl -X POST http://localhost:8080/api.proto.v1.GatewayService/StartCodebaseAnalysis \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

## 🧠 RAG Pipeline Deep Dive

The RAG pipeline is the core intelligence layer, implementing a complete Retrieval-Augmented Generation lifecycle:

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

### Key Design Decisions

- **Optimistic Locking**: Documents are marked `processing` before embedding, preventing duplicate work across workers
- **Crash Recovery**: On startup, documents stuck in `processing` are reset to `pending`
- **Score Threshold**: Irrelevant chunks (L2 distance > 1.2) are filtered out before LLM injection
- **Dual LLM Support**: OpenAI for embeddings (no alternative), DeepSeek or OpenAI for generation
- **Deterministic Chunk IDs**: Re-embedding overwrites existing chunks rather than creating duplicates
- **Content Deduplication**: Exact-content deduplication prevents duplicate context in answers

---

## 🏗️ Project Structure

```
Doc-Agent/
├── backend/                          # Backend microservices (Go + Python)
│   ├── docker-compose.yml            # Local development orchestration
│   ├── configs/                      # Shared service configurations
│   ├── services/
│   │   ├── gateway/                  # API Gateway (Go / Fiber + Connect-Go)
│   │   ├── codebase/                 # Codebase Analysis (Go / gRPC)
│   │   ├── docgen/                   # Document Generation Worker (Go / asynq)
│   │   ├── database/                 # Database Service (Go / gRPC + PostgreSQL)
│   │   └── ai-service/               # AI Service (Python / gRPC + LangChain + ChromaDB)
│   │       └── src/
│   │           ├── rag_pipeline.py   # RAG lifecycle (embed → retrieve → generate)
│   │           ├── ai_service_impl.py# gRPC service implementation
│   │           ├── vector_store.py   # ChromaDB management
│   │           ├── llm_client.py     # OpenAI/DeepSeek API client
│   │           └── db_client.py      # Direct PostgreSQL client (psycopg2)
│   ├── shared/                       # Shared proto definitions, clients, utilities
│   │   ├── api/proto/v1/             # Protocol Buffer definitions
│   │   └── pkg/                      # Shared Go packages (clients, errors, utils)
│   ├── deployments/                  # Docker, Docker Compose, K8s manifests
│   └── tests/                        # Integration & E2E tests
│       ├── e2e/
│       └── integration/
├── local-agent/                      # Rust CLI chat agent
│   └── src/
│       ├── main.rs                   # Entry point
│       ├── api/client.rs             # LLM API HTTP client
│       ├── cli/
│       │   ├── chat/                 # Chat session, state machine, tools
│       │   │   ├── session.rs        # State machine engine
│       │   │   ├── tool_manager.rs   # Tool orchestration & parallel execution
│       │   │   ├── tools/            # fs_read, fs_scan, ignore_scan
│       │   │   └── cli/slash_commands/  # /quit, /doc_generation, /readme
│       │   └── root.rs              # Cli struct & subcommands
│       └── grpc/                     # On-demand gRPC server (tonic)
│           ├── server.rs
│           └── service.rs
├── web/                              # Web frontend
│   ├── web.html                      # Single-page documentation viewer
│   └── test.md                       # Test markdown file
├── docs/                             # Architecture diagrams & documentation
│   ├── architecture.mermaid          # System architecture diagram
│   ├── sequence_diagram.mermaid      # Request flow diagram
│   └── *.png                         # Rendered diagram images
├── CLAUDE.md                         # AI coding assistant guide
├── INTEGRATION.md                    # AI Service integration guide
└── endpoints.md                      # API endpoint documentation
```

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

### Rust Local Agent

```bash
cd local-agent
cargo test
```

### Integration & E2E Tests

```bash
cd backend/tests
go test ./integration/...
go test ./e2e/...
```

---

## 🤝 Contributing

Contributions are welcome! To get started:

1. **Fork** the repository
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes** — key areas for contribution:
   - Add new tools for the CLI agent in `local-agent/src/cli/chat/tools/` by implementing the `Tool` trait
   - Add new documentation prompt templates in `backend/services/codebase/internal/prompts/`
   - Improve the RAG pipeline in `backend/services/ai-service/src/rag_pipeline.py`
   - Add new slash commands in `local-agent/src/cli/chat/cli/slash_commands/`
4. **Run tests** to ensure nothing breaks
5. **Submit a pull request** with a clear description of your changes

### Development Guidelines

- Each service has its own `go.mod` — treat it as an independent module
- Proto changes in `shared/api/proto/v1/` require regenerating Go stubs via `buf generate`
- For the AI service, gRPC Python code is regenerated via `bash scripts/generate_grpc.sh`
- Communication between services must be via **gRPC** or **message queue** only — never add direct code dependencies between services

### Development Setup

```bash
# Clone all components
git clone https://github.com/your-org/doc-agent.git
cd doc-agent

# Backend services (separate terminals)
cd backend
docker-compose up postgres redis  # Start dependencies
cd services/database && go run ./cmd/database
cd services/codebase && go run ./cmd/codebase
cd services/gateway && go run ./cmd/gateway

# AI Service
cd services/ai-service && python main.py

# Local Agent CLI
cd ../../local-agent && cargo run -- chat
```

---

## 🔮 Roadmap

- [ ] **Message Queue**: Replace in-process Redis/Asynq with RabbitMQ or Kafka for production workloads
- [ ] **Service Discovery**: Add Consul or etcd for dynamic service registration
- [ ] **Observability**: Distributed tracing with Jaeger/Zipkin, metric collection with Prometheus
- [ ] **API Documentation**: Auto-generate OpenAPI/Swagger docs from proto definitions
- [ ] **Multi-Project Support**: Full project isolation with authentication & authorization
- [ ] **CI/CD Pipelines**: Independent deployment pipelines per service (GitHub Actions)
- [ ] **Web UI**: React/Vue frontend for enhanced documentation browsing
- [ ] **Custom LLM Support**: Pluggable LLM backends (Anthropic, local models via Ollama)

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

<p align="center">
  <sub>Built with ❤️ using Rust, Go, Python, and a lot of ☕</sub>
</p>