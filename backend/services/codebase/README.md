# Doc-Agent Codebase Analysis Service

An intelligent, AI-powered microservice that automates the analysis of software repositories and generates structured, high-quality documentation. Part of the **Doc-Agent** ecosystem, this service orchestrates a multi-stage pipeline to classify repositories, generate documentation instructions, and enqueue processing tasks for automated document generation.

## Overview

The Codebase Analysis Service is the core intelligence engine of Doc-Agent. It ingests repository structures and README content, then runs a three-phase pipeline powered by large language models (LLMs) to produce comprehensive documentation catalogs tailored to the specific project type — whether it's an application, library, framework, CLI tool, DevOps configuration, or documentation project.

Built with **Go** and **gRPC**, this service is designed for scalability, reliability, and seamless integration into the broader Doc-Agent microservices architecture.

---

## Features

- **🔍 Intelligent Repository Classification** – Automatically categorizes projects into one of seven types (Applications, Libraries, Frameworks, CLI Tools, Development Tools, DevOps Configuration, Documentation) using AI-driven analysis.
- **📝 Context-Aware Instruction Generation** – Generates structured, hierarchical documentation catalogs with "Getting Started" and "Deep Dive" modules, adapted to the project's complexity and technology stack.
- **⚡ Asynchronous Task Enqueuing** – Transforms generated documentation instructions into distributed task queues using [Asynq](https://github.com/hibiken/asynq) and Redis, enabling scalable, fault-tolerant processing.
- **🧩 Multi-Stage AI Pipeline** – Orchestrates a three-phase pipeline: classification → instruction generation → task enqueuing, with full context propagation between stages.
- **🔌 gRPC API** – Exposes a robust gRPC interface for `StartCodebaseAnalysis` and `HealthCheck` endpoints, enabling seamless inter-service communication.
- **🏗️ Microservice-Ready Architecture** – Designed with clean separation of concerns: configuration management, service layer, gRPC server, and pipeline stages.
- **🛡️ Graceful Shutdown & Containerization** – Production-ready with Docker multi-stage builds, health checks, non-root user execution, and graceful signal handling.
- **📋 Persistent State Management** – Stores generated documentation sections in a database service and uses Redis for message queuing and caching.

---

## Architecture

The service follows a modular pipeline architecture:

```
Client Request
     │
     ▼
┌─────────────────────┐
│   gRPC Server       │
│   (internal/grpc)   │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Analysis Service   │
│  (internal/service) │
└─────────┬───────────┘
          │
    ┌─────┴─────┐
    │           │
    ▼           ▼
┌─────────┐ ┌─────────┐
│ Step 1  │ │ Step 2  │
│Classify │ │Generate │
│  Repo   │ │Instruct.│
└─────────┘ └────┬────┘
                 │
                 ▼
           ┌─────────┐
           │ Step 3  │
           │Enqueue  │
           │  Tasks  │
           └─────────┘
```

### Pipeline Stages

1. **Repo Classification** – Analyzes project structure and README to classify the repository type.
2. **Instruction Generation** – Based on classification, generates a tailored, hierarchical JSON documentation catalog.
3. **Task Enqueuing** – Parses the instruction catalog, stores sections in the database, and enqueues individual documentation generation tasks to Redis/Asynq.

---

## Installation

### Prerequisites

- **Go** 1.25+ (for development/building)
- **Docker** (recommended for deployment)
- **Redis** (for task queuing)
- Access to Doc-Agent ecosystem services (AI Service, Gateway, Database)

### From Source

```bash
# Clone the repository
git clone https://github.com/dongtandung2001/Doc-Agent.git
cd Doc-Agent/backend/services/codebase

# Download dependencies
go mod download

# Build the binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -trimpath -o bin/codebase ./cmd/codebase
```

### Using Docker

```bash
# Build the Docker image
docker build -t doc-agent/codebase-service -f Dockerfile ../../..

# Run the container
docker run -p 9001:9001 \
  -e REDIS_HOST=your-redis-host \
  -e AI_HOST=your-ai-service \
  doc-agent/codebase-service
```

### Configuration

Configuration is managed via YAML (default: `configs/codebase.yaml`) with environment variable overrides:

| Config Key       | Environment Variable | Default       | Description                     |
|------------------|---------------------|---------------|----------------------------------|
| `server.host`    | `SERVER_HOST`       | `0.0.0.0`     | gRPC server bind address         |
| `server.port`    | `SERVER_PORT`       | `9001`        | gRPC server port                 |
| `ai.host`        | `AI_HOST`           | `localhost`   | AI service host                  |
| `ai.port`        | `AI_PORT`           | `50051`       | AI service port                  |
| `gateway.host`   | `GATEWAY_HOST`      | `localhost`   | Gateway service host             |
| `gateway.port`   | `GATEWAY_PORT`      | `8080`        | Gateway service port             |
| `redis.host`     | `REDIS_HOST`        | `localhost`   | Redis host                       |
| `redis.port`     | `REDIS_PORT`        | `6379`        | Redis port                       |
| `redis.password` | `REDIS_PASSWORD`    | `""`          | Redis password                   |
| `redis.db`       | `REDIS_DB`          | `0`           | Redis database index             |

---

## Usage

### Starting the Service

```bash
# Run directly
go run ./cmd/codebase

# Or using the built binary
./bin/codebase
```

### gRPC API

The service exposes the `CodebaseAnalysisService` gRPC service with two RPCs:

#### `StartCodebaseAnalysis`

Initiates the full documentation generation pipeline for a repository.

**Request:**
```protobuf
message StartCodebaseAnalysisRequest {
  string project_structure = 1;  // Directory tree or file listing
  string readme_content = 2;     // Existing README content
}
```

**Response:**
```protobuf
message StartCodebaseAnalysisResponse {
  bool success = 1;
}
```

**Example (with grpcurl):**
```bash
grpcurl -plaintext -d '{
  "project_structure": "[F] src/main.go\n[D] src/pkg\n[D] docs",
  "readme_content": "# My Project\nA sample project..."
}' localhost:9001 apiv1.CodebaseAnalysisService/StartCodebaseAnalysis
```

#### `HealthCheck`

Returns the health status of the service.

```bash
grpcurl -plaintext localhost:9001 apiv1.CodebaseAnalysisService/HealthCheck
```

### Integration Example

```go
package main

import (
    "context"
    "log"
    "time"
    
    apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    conn, err := grpc.Dial("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := apiv1.NewCodebaseAnalysisServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, err := client.StartCodebaseAnalysis(ctx, &apiv1.StartCodebaseAnalysisRequest{
        ProjectStructure: "...",  // Repository file structure
        ReadmeContent:    "...",  // Repository README
    })
    if err != nil {
        log.Fatalf("Analysis failed: %v", err)
    }
    log.Printf("Analysis started: success=%v", resp.Success)
}
```

---

## Project Structure

```
├── cmd/codebase/
│   └── main.go                 # Entry point: initializes dependencies, starts gRPC server
├── configs/
│   └── codebase.yaml           # Default service configuration
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration loading with Viper + env overrides
│   ├── grpc/
│   │   └── server.go           # gRPC server implementation & handler registration
│   ├── pipeline/
│   │   ├── repo_classification.go      # Stage 1: AI-powered repo classification
│   │   ├── generate_instructions.go    # Stage 2: Documentation structure generation
│   │   └── enqueue_instruction.go      # Stage 3: Task parsing & enqueuing
│   ├── prompts/
│   │   ├── generate_classfication.md   # Classification prompt template
│   │   ├── generate_instruction.md     # Instruction generation prompt template
│   │   └── generate_document.md        # Document generation prompt template
│   ├── service/
│   │   └── analyzer.go                # Business logic orchestrating the pipeline
│   └── utils/                          # Utility packages
├── Dockerfile                   # Multi-stage Docker build
├── go.mod / go.sum             # Go module dependencies
└── .dockerignore               # Docker build exclusion rules
```

---

## Internal Pipeline Details

### 1. Classification (Stage 1)

The `ClassifyRepo` function reads a classification prompt template and uses the AI service to categorize the repository. The classification determines which documentation generation protocol to use:

| Classification     | Protocol Focus                                   |
|--------------------|--------------------------------------------------|
| `Applications`     | User workflows, architecture, deployment         |
| `Libraries`        | API reference, integration patterns, best practices |
| `Frameworks`       | Core concepts, extension points, developer onboarding |
| `CLITools`         | Command reference, scripting, automation          |
| `DevelopmentTools` | Feature guides, workflow integration, build tools |
| `DevOpsConfiguration` | Infrastructure architecture, IaC, monitoring   |
| `Documentation`    | Content architecture, contribution, quality assurance |
| *(default)*        | General project documentation                    |

### 2. Instruction Generation (Stage 2)

Uses a multi-stage AI conversation to first analyze the repository in-depth, then outputs a structured JSON catalog of documentation sections with nested hierarchy, each containing specific generation prompts.

### 3. Task Enqueuing (Stage 3)

The generated JSON catalog is parsed, flattened, and each documentation section is:
1. Stored in the database via the Database service
2. Enqueued as an [Asynq](https://github.com/hibiken/asynq) task (`docgen:instruction`) to Redis for distributed processing

---

## Contributing

Contributions are welcome! To contribute:

1. **Fork** the repository
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit your changes**: `git commit -m 'Add amazing feature'`
4. **Push to the branch**: `git push origin feature/amazing-feature`
5. **Open a Pull Request**

### Development Setup

```bash
# Clone and navigate
git clone https://github.com/dongtandung2001/Doc-Agent.git
cd Doc-Agent/backend/services/codebase

# Install dependencies
go mod download

# Run tests (if applicable)
go test ./...

# Build
go build -o bin/codebase ./cmd/codebase
```

### Coding Guidelines

- **No unnecessary comments** – Code should be self-documenting
- **Follow Go conventions** – Use `gofmt` and standard Go project layout
- **Security first** – Never commit secrets or credentials
- **Graceful error handling** – All errors should be logged and properly propagated
- **Parallel operations** – Batch independent operations where possible

---

## License

This project is licensed under the terms specified in the repository's [LICENSE](../../../../LICENSE) file.

---

## Acknowledgments

- Built with [Go](https://go.dev/), [gRPC](https://grpc.io/), and [Asynq](https://github.com/hibiken/asynq)
- Powered by AI-driven code analysis and documentation generation
- Part of the **Doc-Agent** ecosystem for automated software documentation