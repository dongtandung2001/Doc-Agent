# DocGen — AI-Powered Documentation Generation Service

[![Go Version](https://img.shields.io/badge/Go-1.25-blue?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://docker.com)

**DocGen** is a high-performance, event-driven microservice within the [Doc-Agent](https://github.com/dongtandung2001/Doc-Agent) ecosystem that automates the generation of comprehensive technical documentation for software projects. By combining LLM-powered AI analysis with a distributed task queue architecture, DocGen transforms source code into structured, developer-ready documentation — including architecture diagrams, component deep-dives, and system design explanations.

---

## Features

- **🤖 AI-Driven Documentation Generation** — Leverages a forward-thinking LLM (via `AIClient`) to produce rich, context-aware technical documentation from source code files.
- **⚡ Asynchronous Task Queue** — Built on [Asynq](https://github.com/hibiken/asynq) and Redis for reliable, distributed task processing with configurable concurrency.
- **🧩 Hierarchical Document Modeling** — Supports parent-child item structures for generating documentation across nested modules and sub-systems.
- **🔗 Multi-Service Orchestration** — Coordinates with AI, Gateway, Database, and Redis services through a clean client abstraction layer in the shared module.
- **📦 Optimized Docker Deployment** — Multi-stage Docker build with protobuf code generation, caching, and a minimal runtime image running as a non-root user.
- **💾 Persistent Document Storage** — Automatically stores generated documentation with structured metadata (ID, project ID, title, description, content) in the database service.
- **🔧 Configurable & Environment-Aware** — YAML-based configuration with automatic environment variable override support (e.g., `REDIS_HOST` → `redis.host`).
- **🗂️ Prompt-Driven Generation** — Uses a templated markdown prompt system (`generate_doc.md`) to guide the AI through a rigorous, multi-phase documentation process.
- **🛡️ Graceful Shutdown** — Handles OS signals (SIGTERM/SIGINT) for clean resource teardown and task finalization.

---

## Architecture Overview

```
┌────────────┐     ┌──────────────────────┐     ┌─────────────┐
│  Gateway   │◄────│     DocGen Worker     │────►│  AI Service │
│  (REST)    │     │  (Asynq Server)       │     │  (gRPC)     │
└────────────┘     └──────────┬────────────┘     └─────────────┘
                              │
                     ┌────────▼────────┐
                     │     Redis       │
                     │  (Task Queue)   │
                     └────────┬────────┘
                              │
                     ┌────────▼────────┐
                     │   Database Svc  │
                     │  (gRPC Storage) │
                     └─────────────────┘
```

The service operates as a **background worker** — it does not expose an HTTP server. Tasks are enqueued by other services via Redis, and DocGen consumes them, orchestrates the AI pipeline, and persists results.

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.25 or later
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/) (for containerized deployment)
- Access to the shared module (`backend/shared`) — see [repository structure](#repository-structure)
- Running instances of:
  - **Redis** (task queue backend)
  - **AI Service** (gRPC-based LLM inference)
  - **Gateway Service** (REST/gRPC proxy)
  - **Database Service** (document storage)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/dongtandung2001/Doc-Agent.git
   cd Doc-Agent/backend/services/docgen
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Configure the service**

   Edit `configs/docgen.yaml` or set environment variables:

   ```yaml
   ai:
     host: "localhost"
     port: 50051
   gateway:
     host: "localhost"
     port: 8080
   redis:
     host: "localhost"
     port: 6379
     db: 0
   database:
     host: "localhost"
     port: 9002
     user: "user"
     password: "password"
   ```

   Environment variables override YAML values automatically (e.g., `AI_HOST=ai.example.com`).

4. **Build and run**

   ```bash
   go run ./cmd/docgen
   ```

### Docker Deployment

Build and run using the provided Dockerfile:

```bash
docker build -t docgen:latest -f Dockerfile ../../
docker run --rm \
  -e REDIS_HOST=redis \
  -e AI_HOST=ai-service \
  -e DATABASE_HOST=db-service \
  docgen:latest
```

The multi-stage Docker build also generates protobuf Go code automatically, ensuring the shared module is always in sync.

---

## Usage

### Enqueuing a Documentation Task

Tasks are enqueued via Redis using the Asynq client with the task type `docgen:instruction`. The payload is a JSON object representing the documentation item:

```json
{
  "title": "API Reference",
  "name": "api-reference",
  "prompt": "Generate comprehensive API documentation...",
  "children": [
    {
      "title": "Authentication Module",
      "name": "auth-module",
      "prompt": "Document all auth endpoints...",
      "code_files": "path/to/auth/files"
    }
  ],
  "projectType": "backend",
  "code_files": "path/to/main/files"
}
```

### Task Processing Flow

1. **Task Receipt** — The Asynq server picks up a `docgen:instruction` task from the Redis queue.
2. **Context Initialization** — A `ChatContext` is populated with title, name, prompt, project type, and code file references.
3. **Prompt Injection** — The AI prompt template (`internal/prompts/generate_doc.md`) is loaded and populated with context variables (`{{$prompt}}`, `{{$title}}`, `{{$projectType}}`, `{{$code_files}}`).
4. **AI Orchestration** — The AI client is called with agentic mode and tool usage enabled, allowing the LLM to read files and produce structured documentation.
5. **Result Persistence** — The generated documentation is stored in the database with metadata (ID, project ID, title, description, content).
6. **Completion** — A success message with the document ID is returned, and the task is marked as processed.

### Worker Concurrency

By default, the worker processes up to **5 tasks concurrently**. This can be tuned in the service initialization:

```go
asynq.Config{
    Concurrency: 5,
    Queues: map[string]int{
        "default": 1,
    },
}
```

---

## Repository Structure

```
backend/services/docgen/
├── cmd/docgen/            # Application entrypoint
│   └── main.go            # Service bootstrap, client connections, worker startup
├── configs/
│   └── docgen.yaml        # Service configuration (YAML)
├── internal/
│   ├── config/
│   │   └── config.go      # Configuration loading with Viper & env overrides
│   ├── pipeline/
│   │   └── generate_doc.go # Core documentation generation pipeline
│   ├── prompts/
│   │   └── generate_doc.md # AI prompt template (7-phase documentation process)
│   ├── service/
│   │   └── docgen.go      # Service layer: Asynq server setup & lifecycle
│   └── tasks/
│       └── handler.go     # Task types, payload structures, and handlers
├── Dockerfile             # Multi-stage production build
├── .dockerignore          # Build context exclusions
├── go.mod                 # Module definition & dependencies
└── go.sum                 # Dependency checksums
```

---

## Configuration Reference

| Key              | Environment Variable | Default       | Description                  |
|------------------|----------------------|---------------|------------------------------|
| `ai.host`        | `AI_HOST`            | `localhost`   | AI gRPC service host         |
| `ai.port`        | `AI_PORT`            | `50051`       | AI gRPC service port         |
| `gateway.host`   | `GATEWAY_HOST`       | `localhost`   | Gateway service host         |
| `gateway.port`   | `GATEWAY_PORT`       | `8080`        | Gateway service port         |
| `redis.host`     | `REDIS_HOST`         | `localhost`   | Redis host                   |
| `redis.port`     | `REDIS_PORT`         | `6379`        | Redis port                   |
| `redis.password` | `REDIS_PASSWORD`     | `""`          | Redis password               |
| `redis.db`       | `REDIS_DB`           | `0`           | Redis database number        |
| `database.host`  | `DATABASE_HOST`      | `localhost`   | Database service host        |
| `database.port`  | `DATABASE_PORT`      | `9002`        | Database service port        |
| `database.user`  | `DATABASE_USER`      | `user`        | Database authentication user |
| `database.password` | `DATABASE_PASSWORD` | `""`        | Database authentication password |

---

## Development

### Adding a New Task Type

1. Define the task type constant in `internal/tasks/handler.go`:
   ```go
   const TaskTypeMyNewTask = "docgen:my_task"
   ```

2. Create a handler method on `TaskHandler`:
   ```go
   func (h *TaskHandler) HandleMyTask(ctx context.Context, task *asynq.Task) error {
       // Process the task
   }
   ```

3. Register the handler in `RegisterHandlers()`:
   ```go
   mux.HandleFunc(TaskTypeMyNewTask, h.HandleMyTask)
   ```

### Modifying the AI Prompt

The prompt template is located at `internal/prompts/generate_doc.md`. It uses Go template syntax for variable injection. The prompt orchestrates a 7-phase documentation process:

1. **Strategic Planning** — Task analysis, code assessment, documentation budgeting
2. **Deep Code Analysis** — Systematic file review, pattern recognition, dependency mapping
3. **Comprehensive Documentation** — Full document generation with architecture, components, and technical deep-dives

### Testing

Test files (`*_test.go`) are excluded from the Docker build context via `.dockerignore` to keep images lean. Run tests locally:

```bash
go test ./...
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork** the repository and create a feature branch from `main`.
2. **Code style** — Follow standard Go formatting (`gofmt`) and linting practices.
3. **Commit messages** — Use clear, descriptive commit messages referencing issue numbers where applicable.
4. **Pull requests** — Open a PR against `main` with a description of changes, motivation, and any breaking changes noted.
5. **Review** — All PRs require review before merging. Be responsive to feedback.

For major changes, please open an issue first to discuss what you'd like to change.

---

## License

This project is licensed under the [MIT License](LICENSE). See the `LICENSE` file for details.

---

## Acknowledgments

- [Asynq](https://github.com/hibiken/asynq) — Distributed task queue for Go
- [Viper](https://github.com/spf13/viper) — Go configuration management
- [Connect-Go](https://connectrpc.com/) — gRPC-compatible HTTP APIs
- [Buf](https://buf.build/) — Protobuf code generation