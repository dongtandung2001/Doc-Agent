# Doc-Agent API Gateway

> A high-performance, microservice API gateway built with Go and Fiber — serving as the central entry point for the Doc-Agent documentation automation platform.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ACC1?style=for-the-badge&logo=go)](https://gofiber.io/)
[![Connect-Go](https://img.shields.io/badge/Connect-Go-7C3AED?style=for-the-badge)](https://connectrpc.com/)
[![gRPC](https://img.shields.io/badge/gRPC-proto-FF6C37?style=for-the-badge)](https://grpc.io/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)
![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)

---

## 📋 Overview

The **Doc-Agent API Gateway** is the unified entry point for the Doc-Agent ecosystem — a modern documentation automation platform that leverages AI-powered code analysis, RAG (Retrieval-Augmented Generation) chat, and intelligent document management. Built on **Fiber** (a Go web framework inspired by Express.js) and **Connect-Go** (a gRPC-compatible framework), the gateway seamlessly proxies requests to multiple backend microservices, providing a single, cohesive API surface.

It supports dual communication protocols — **gRPC** for high-performance internal service calls and **HTTP** for flexible AI service integration — all wrapped in a clean, idiomatic Go codebase with proper middleware, graceful shutdown, and containerized deployment.

---

## ✨ Features

- **🔄 Unified API Gateway** — Single entry point routing to multiple backend microservices (Codebase Analysis, Database, Local Agent, AI Service)
- **🔌 Dual Protocol Support** — Native gRPC via Connect-Go for low-latency internal calls; HTTP fallback for flexible AI service integration
- **🤖 AI-Powered Chat & RAG** — Proxies conversational AI requests and RAG embedding pipeline triggers to the AI service with optional HTTP wrapper
- **📄 Intelligent Document Management** — Store, retrieve, and query document sections through a dedicated database service
- **🔍 Codebase Analysis Orchestration** — Triggers and manages automated codebase analysis pipelines
- **🩺 Unified Health Checking** — Aggregate health endpoint that monitors all backend services in a single call
- **🛡️ Enterprise-Grade Middleware** — CORS, structured request logging, and panic recovery with stack traces
- **🌐 Local Agent Integration** — Bridges to local agent services for file content requests and on-premise operations
- **🐳 Containerized Deployment** — Multi-stage Docker build with non-root user, health checks, and build cache optimization
- **⚡ High Performance** — Built on Fiber (fasthttp-based) with zero-allocation, low-latency request handling
- **🔧 Environment-Aware Configuration** — YAML-based configuration with automatic environment variable overrides
- **🪶 Graceful Shutdown** — Proper signal handling for clean shutdown on SIGTERM/SIGINT

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Client / Frontend                  │
└──────────────────┬──────────────────────────────────┘
                   │ HTTP / Connect-Go
                   ▼
┌─────────────────────────────────────────────────────┐
│              DOC-AGENT API GATEWAY                     │
│              (Port 8080 · Fiber + Connect-Go)          │
├─────────────────────────────────────────────────────┤
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐  │
│  │  CORS   │ │  Logger  │ │ Recovery │ │ Health  │  │
│  │  Middle │ │  Middle  │ │  Middle  │ │ Check   │  │
│  └────┬────┘ └────┬─────┘ └────┬─────┘ └────┬────┘  │
│       └───────────┴────────────┴────────────┘        │
└──────────────────────────────────────────────────────┘
         │          │           │           │
        gRPC       gRPC       gRPC     gRPC/HTTP
         ▼          ▼           ▼           ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│ Codebase │ │ Database │ │  Local   │ │    AI    │
│ Analysis │ │  Service │ │  Agent   │ │ Service  │
│ Service  │ │          │ │  Service │ │ (RAG/Chat)│
│ :9001    │ │ :9002    │ │ :50051   │ │ :50051   │
└──────────┘ └──────────┘ └──────────┘ └──────────┘
```

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.25+ (for local development)
- **Docker** 24+ (for containerized deployment)
- **Protobuf toolchain** (for generating proto code)
- **Access to backend services** (Codebase, Database, Local Agent, AI)

### Installation

#### 1. Clone the Repository

```bash
git clone https://github.com/dongtandung2001/Doc-Agent.git
cd Doc-Agent/backend/services/gateway
```

#### 2. Install Dependencies

```bash
go mod download
```

#### 3. Configure the Gateway

Edit `configs/gateway.yaml`:

```yaml
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
    base_url: "http://localhost:8888"   # Optional HTTP fallback
  local_agent:
    host: "localhost"
    port: 50051
```

> **Tip:** All configuration values can be overridden via environment variables using underscore notation (e.g., `BACKENDS_CODEBASE_HOST`).

#### 4. Run Locally

```bash
go run ./cmd/gateway
```

#### 5. Run with Docker

```bash
# Build the image
docker build -t doc-agent-gateway -f Dockerfile ../../..

# Run the container
docker run -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  doc-agent-gateway
```

---

## 📖 Usage

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "service": "gateway"
}
```

### Aggregate Backend Health

```bash
# Uses Connect-Go protocol
grpcurl -plaintext -d '{}' \
  localhost:8080 apiv1.GatewayService/HealthCheck
```

### Start Codebase Analysis

```bash
grpcurl -plaintext -d '{"project_id": "my-project"}' \
  localhost:8080 apiv1.GatewayService/StartCodebaseAnalysis
```

### Document Management

**Store a document:**
```bash
grpcurl -plaintext -d '{
  "project_id": "my-project",
  "file_path": "/src/main.go",
  "content": "package main\nfunc main() {}"
}' localhost:8080 apiv1.GatewayService/StoreDocument
```

**Retrieve a document:**
```bash
grpcurl -plaintext -d '{
  "project_id": "my-project",
  "file_path": "/src/main.go"
}' localhost:8080 apiv1.GatewayService/GetDocument
```

### AI Chat & RAG

**Chat with AI:**
```bash
grpcurl -plaintext -d '{
  "project_id": "my-project",
  "messages": [{"role": "user", "content": "Explain the main entry point"}]
}' localhost:8080 apiv1.GatewayService/Chat
```

**Trigger RAG pipeline:**
```bash
grpcurl -plaintext -d '{
  "project_id": "my-project"
}' localhost:8080 apiv1.GatewayService/CreateRAG
```

### Request Local File Content

```bash
grpcurl -plaintext -d '{
  "project_id": "my-project",
  "file_path": "/etc/config.yaml"
}' localhost:8080 apiv1.GatewayService/RequestFileContent
```

---

## 🧩 API Reference

The gateway exposes a single **Connect-Go** RPC service — `GatewayService` — mounted at the default Connect path. All service definitions are in `apiv1` protobuf schema.

| RPC Method | Service Target | Description |
|---|---|---|
| `HealthCheck` | All backends | Aggregate health check across all configured services |
| `StartCodebaseAnalysis` | Codebase Analysis Service | Initiate code analysis pipeline |
| `RequestFileContent` | Local Agent Service | Request file content from local agent |
| `GetDocument` | Database Service | Retrieve stored document by project + path |
| `StoreDocument` | Database Service | Store a document for a project |
| `GetDocumentSections` | Database Service | Retrieve sections of a stored document |
| `Chat` | AI Service | Send chat messages to the AI assistant |
| `CreateRAG` | AI Service | Trigger the RAG embedding pipeline |

---

## 🛠️ Project Structure

```
services/gateway/
├── cmd/gateway/
│   └── main.go                  # Application entry point
├── configs/
│   └── gateway.yaml             # YAML configuration file
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration loader (Viper)
│   ├── handlers/
│   │   ├── gateway.go           # Gateway service handler (gRPC proxy)
│   │   └── http_ai_client.go    # HTTP AI client wrapper
│   ├── http/
│   │   └── server.go            # Fiber HTTP server setup
│   └── middleware/
│       ├── cors.go              # CORS middleware
│       ├── logger.go            # Request logging middleware
│       └── recovery.go          # Panic recovery middleware
├── Dockerfile                   # Multi-stage Docker build
├── .dockerignore                # Docker build context exclusions
├── go.mod
└── go.sum
```

---

## 🤝 Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit your changes** (`git commit -m 'Add amazing feature'`)
4. **Push to the branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

### Development Setup

```bash
# Install protoc plugins for code generation
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
go install github.com/bufbuild/buf/cmd/buf@latest

# Generate proto code from shared module
(cd ../../shared && buf generate)
```

### Code Quality

- Follow idiomatic Go conventions (`gofmt`, `go vet`)
- Ensure all tests pass (`go test ./...`)
- Use meaningful commit messages
- Update documentation for any new features

---

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgements

- [Fiber](https://gofiber.io/) — Express-inspired Go web framework
- [Connect-Go](https://connectrpc.com/) — gRPC-compatible HTTP framework
- [Viper](https://github.com/spf13/viper) — Configuration management
- [Buf](https://buf.build/) — Protobuf code generation
- [Asynq](https://github.com/hibiken/asynq) — Distributed task queue

---

<div align="center">
  <sub>Built with ❤️ by the Doc-Agent Team</sub>
</div>