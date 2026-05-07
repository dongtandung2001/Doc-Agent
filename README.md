<p align="center">
  <img src="https://img.shields.io/badge/Language-Rust-orange.svg" alt="Rust">
  <img src="https://img.shields.io/badge/Language-Go-blue.svg" alt="Go">
  <img src="https://img.shields.io/badge/Language-Python-blue.svg" alt="Python">
  <img src="https://img.shields.io/badge/Architecture-Microservices-6a0dad" alt="Microservices">
  <img src="https://img.shields.io/badge/LLM-OpenAI_%7C_DeepSeek-brightgreen" alt="LLM">
  <img src="https://img.shields.io/badge/Vector_Store-ChromaDB-purple" alt="ChromaDB">
  <img src="https://img.shields.io/badge/Database-PostgreSQL-blue" alt="PostgreSQL">
</p>

# Doc-Agent 🤖📝

**An intelligent, multi-agent documentation generation system** that automatically analyzes codebases, classifies project types, and generates high-quality README files, API references, and structured documentation — powered by LLMs, RAG pipelines, and a microservice architecture.

Doc-Agent combines a **Rust-powered CLI chat agent**, a **Python AI service** with RAG capabilities, **Go-based microservices** for codebase analysis and document generation, and a modern **web frontend** for browsing generated documentation. It uses Retrieval-Augmented Generation (RAG) with ChromaDB to index documentation and provide context-aware answers.

---

## Architecture Overview

```
                    ┌──────────────┐
                    │  Web UI      │
                    │  (HTML/CSS)  │
                    └──────┬───────┘
                           │ HTTP/Connect
                    ┌──────▼───────┐
                    │   Gateway    │ ◄── gRPC
                    │  (Go/Fiber)  │
                    └──┬──┬──┬──┬──┘
                       │  │  │  │
          ┌────────────┘  │  │  └──────────────┐
          ▼               ▼  ▼                  ▼
  ┌──────────────┐ ┌──────────────┐  ┌─────────────────┐
  │ Codebase     │ │ Database     │  │ AI Service      │
  │ Analysis     │ │ Service      │  │ (Python/gRPC)   │
  │ (Go)         │ │ (Go/Postgres)│  │ - Chat          │
  └──────┬───────┘ └──────────────┘  │ - RAG Pipeline  │
         │                           │ - ChromaDB      │
         │ Redis MQ                  └─────────────────┘
         ▼
  ┌──────────────┐
  │ DocGen       │
  │ Worker (Go)  │
  └──────────────┘
         │
         ▼
  ┌──────────────┐
  │ Local Agent  │
  │ (Rust/CLI)   │
  │ - Chat       │
  │ - File Tools │
  │ - gRPC Server│
  └──────────────┘
```

---

## ✨ Features

### 🤖 AI-Powered Interactive Chat (Rust CLI)
- Full-duplex chat session with any OpenAI-compatible LLM API (OpenAI, DeepSeek, etc.)
- State-machine-driven conversation loop with robust error recovery
- **Three built-in file system tools** exposed to the LLM:
  - `fs_read` — Read file contents with optional line range
  - `fs_scan` — List directory contents with configurable depth
  - `ignore_scan` — Scan directories respecting nested `.gitignore` rules
- **Parallel tool execution** via `tokio::spawn` with performance instrumentation
- **Tool approval workflow** — non-read tools require user confirmation by default
- **Slash command system** — `/quit`, `/init`, `/doc_generation`, `/readme`

### 📄 Automated Document Generation
- **Codebase classification** — AI analyzes project structure and classifies it (Application, Library, Framework, CLI Tool, DevOps, etc.)
- **Intelligent instruction generation** — Generates a hierarchical documentation plan tailored to the project type
- **Task queue architecture** — Documentation tasks are enqueued to Redis (via asynq) and consumed by worker services
- **Section-by-section generation** — Each documentation section is generated independently using AI with access to actual source files
- **Auto-embedding** — Generated docs are automatically embedded into a vector database for RAG retrieval

### 🔍 RAG-Powered Q&A (Python AI Service)
- **LangChain + ChromaDB** vector store for document embeddings
- **OpenAI `text-embedding-3-small`** for embeddings
- **DeepSeek / OpenAI GPT** for LLM generation
- **Automatic chunking** with configurable chunk size and overlap
- **Score-based retrieval filtering** — irrelevant chunks are dropped before prompting
- **Optimistic locking** — worker-safe document embedding with status tracking (pending → processing → completed/failed)

### 🧩 Microservice Architecture (Go)
- **Gateway Service** (Fiber/Connect-RPC) — Single entry point, proxies to all backend services
- **Codebase Analysis Service** — Orchestrates the document generation pipeline (classify → instruct → enqueue)
- **Database Service** — Manages PostgreSQL for structured document storage with nested section trees
- **Document Generation Worker** — Consumes tasks from Redis message queue, generates documentation sections

### 🗄️ Persistent Storage
- **PostgreSQL** for structured data: document sections, file items, embedding status
- **ChromaDB** for vector embeddings
- **Redis** for task queuing and caching
- **File-level caching** with singleflight coalescing to avoid redundant network fetches

### 🌐 Web Frontend
- Modern, responsive documentation viewer built with Tailwind CSS
- Markdown rendering with syntax highlighting (highlight.js)
- Mermaid diagram support
- Dark/light mode
- Search functionality and nested document navigation

---

## 🚀 Getting Started

### Prerequisites

| Component     | Required                                       |
|---------------|------------------------------------------------|
| Rust          | 1.70+ (for Local Agent CLI)                    |
| Go            | 1.21+ (for backend microservices)              |
| Python        | 3.10+ (for AI Service)                         |
| PostgreSQL    | 15+                                            |
| Redis         | Latest                                         |
| Docker        | (optional, for containerized deployment)       |

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

This starts:
- **Gateway** on `:8080`
- **Codebase Service** on `:9001`
- **Database Service** on `:9002`
- **AI Service** (Python) on `:50051`
- **PostgreSQL** on `:5432`
- **Redis** on `:6379`
- **PGAdmin** on `:5050`

### Manual Setup

#### 1. Backend Services (Go)

```bash
# Generate protobuf code
cd backend/shared
buf generate

# Build each service
cd ../services/gateway && go build ./cmd/gateway
cd ../services/codebase && go build ./cmd/codebase
cd ../services/database && go build ./cmd/database
cd ../services/docgen && go build ./cmd/docgen
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

# Start chat
cargo run --release -- chat
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

### Access Generated Documentation

Open the web frontend at `http://localhost:8080` to browse generated documentation with:
- Nested table of contents
- Full-text search
- Markdown rendering with syntax highlighting
- Mermaid diagram support

### API Overview

The Gateway exposes endpoints via Connect-RPC (HTTP/gRPC):

| Endpoint                    | Description                              |
|-----------------------------|------------------------------------------|
| `POST /gateway.v1.GatewayService/Chat` | Chat with AI assistant with RAG context |
| `POST /gateway.v1.GatewayService/CreateRAG` | Trigger RAG embedding pipeline |
| `POST /gateway.v1.GatewayService/StartCodebaseAnalysis` | Start automatic document generation |
| `GET /gateway.v1.GatewayService/GetDocument` | Retrieve generated document |
| `GET /gateway.v1.GatewayService/GetDocumentSections` | Get nested document tree |
| `POST /gateway.v1.GatewayService/RequestFileContent` | Read file content via Local Agent |
| `GET /health` | Health check |

---

## 🏗️ Project Structure

```
doc-agent/
├── backend/                          # Backend microservices
│   ├── docker-compose.yml            # Local development orchestration
│   ├── configs/                      # Shared service configs
│   ├── services/
│   │   ├── gateway/                  # API Gateway (Go/Fiber + Connect-RPC)
│   │   ├── codebase/                 # Codebase Analysis (Go/gRPC)
│   │   ├── docgen/                   # Document Generation Worker (Go/asynq)
│   │   ├── database/                 # Database Service (Go/gRPC + PostgreSQL)
│   │   └── ai-service/               # AI Service (Python/gRPC + LangChain + ChromaDB)
│   ├── shared/                       # Shared proto definitions, clients, utilities
│   │   ├── api/proto/v1/             # Protocol Buffer definitions
│   │   └── pkg/                      # Shared Go packages
│   ├── deployments/                  # Docker, Docker Compose, K8s manifests
│   └── tests/                        # Integration & E2E tests
├── local-agent/                      # Rust CLI chat agent
│   └── src/
│       ├── cli/chat/                 # Chat session, state machine, tools
│       ├── grpc/                     # On-demand gRPC server
│       └── api/                      # LLM API client
├── web/                              # Web frontend
│   └── web.html                      # Single-page documentation viewer
├── docs/                             # Architecture diagrams
└── llm_wrapper.py                    # LLM wrapper utility
```

---

## 🛠️ Technology Stack

| Component              | Technology                                                    |
|------------------------|---------------------------------------------------------------|
| **Chat CLI**           | Rust, Tokio, Clap, rustyline, Tonic (gRPC), reqwest           |
| **API Gateway**        | Go, Fiber, Connect-Go                                          |
| **Codebase Analysis**  | Go, gRPC, Redis (asynq)                                       |
| **Document Generation**| Go, asynq (task queue)                                         |
| **Database Service**   | Go, gRPC, PostgreSQL, psycopg2                                 |
| **AI Service**         | Python, LangChain, ChromaDB, OpenAI, DeepSeek, gRPC           |
| **LLM Models**         | OpenAI GPT-4o-mini, text-embedding-3-small, DeepSeek Chat     |
| **Vector Store**       | ChromaDB                                                      |
| **Message Queue**      | Redis + asynq                                                 |
| **SQL Database**       | PostgreSQL 15+                                                 |
| **Frontend**           | Tailwind CSS, marked.js, highlight.js, Mermaid                |
| **Deployment**         | Docker, Docker Compose, Kubernetes                             |

---

## 🤝 Contributing

Contributions are welcome! To get started:

1. **Fork** the repository
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes** — key areas for contribution:
   - Add new tools for the CLI agent in `local-agent/src/cli/chat/tools/`
   - Add new documentation prompt templates in `backend/services/codebase/internal/prompts/`
   - Improve the RAG pipeline in `backend/services/ai-service/src/rag_pipeline.py`
   - Add new slash commands in `local-agent/src/cli/chat/cli/slash_commands/`
4. **Run tests**
   ```bash
   # Rust
   cd local-agent && cargo test
   
   # Go
   cd backend/services/database && go test ./...
   
   # Python
   cd backend/services/ai-service && pytest tests/ -v
   ```
5. **Submit a pull request** with a clear description of your changes

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

## 📊 Service Communication Flow

The diagram below illustrates how a documentation generation request flows through the system:

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

## 📝 License

This project is licensed under the terms of the MIT license. See `LICENSE` for details.