I now have a thorough understanding of all the files. Let me compile the improved README.

<readme>

# Database Service

> A gRPC microservice for structured document storage, retrieval, and organization — the persistence backbone of the Doc-Agent ecosystem.

[![Go Version](https://img.shields.io/badge/Go-1.25-blue)](https://go.dev/)
[![gRPC](https://img.shields.io/badge/gRPC-enabled-brightgreen)](https://grpc.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)

---

## Overview

The **Database Service** is a gRPC-based microservice that manages all persistent storage within the Doc-Agent platform. Built with Go and PostgreSQL, it provides a robust API for storing, retrieving, and organizing documentation content. The service supports hierarchical document section trees, automatic database migrations, and an embedding status tracking system for RAG (Retrieval-Augmented Generation) vector pipelines.

## Features

- **📄 Document Storage & Retrieval** — Store and fetch rich document content with metadata (title, description, extra fields) scoped by project.
- **🌳 Hierarchical Section Tree** — Organize documents into nested folder/category structures with support for parent-child relationships. The tree structure is computed server-side from a flat data model.
- **🔁 Upsert Semantics** — All write operations use "upsert" (INSERT ... ON CONFLICT UPDATE), making the service idempotent and safe for concurrent use.
- **🧩 Embedding Pipeline Integration** — Each document tracks an `embed_status` field (`pending` → `processing` → `completed` / `failed`) to coordinate with downstream RAG vector embedding workers.
- **⚡ gRPC API** — High-performance, strongly-typed communication via Protocol Buffers and gRPC, with health check endpoint.
- **🗄️ Automatic Migrations** — Schema migrations run automatically on startup with retry logic (up to 10 attempts for database readiness).
- **🐳 Production-Ready Docker** — Multi-stage Docker build with non-root user, build cache optimization, zero-dependency runtime image, and built-in healthcheck.
- **🧪 Mock Repository for Testing** — In-memory mock implementations of all repository interfaces enable fast, deterministic unit tests without a database.

## Architecture

```
┌─────────────┐     gRPC      ┌────────────────────────────────┐
│  Other       │◄────────────►│     Database Service (Go)      │
│  Services    │              │                                │
│  (Agent,     │              │  ┌──────────┐  ┌───────────┐  │
│   Embedding) │              │  │ gRPC     │  │ Business  │  │
│              │              │  │ Server   │──►│ Logic     │──┤
└─────────────┘              │  └──────────┘  │ (Service) │  │
                              │               └─────┬─────┘  │
                              │                     │         │
                              │  ┌──────────────────▼──────┐  │
                              │  │   Repository Layer       │  │
                              │  │  (PostgreSQL / Mock)     │  │
                              │  └──────────────────┬──────┘  │
                              └─────────────────────┼─────────┘
                                                     │
                                          ┌──────────▼──────────┐
                                          │     PostgreSQL       │
                                          │   (Document Store)   │
                                          └─────────────────────┘
```

## Schema

### `document_sections`
Represents folders/categories in the documentation tree. Supports nested hierarchies via `parent_id`.

| Column | Type | Description |
|--------|------|-------------|
| `id` | `VARCHAR(255)` | Primary key |
| `project_id` | `VARCHAR(255)` | Scoping project identifier |
| `name` | `VARCHAR(500)` | Section display name |
| `description` | `TEXT` | Optional description |
| `url` | `VARCHAR(500)` | URL slug for the section |
| `order` | `INTEGER` | Sort order within siblings |
| `parent_id` | `VARCHAR(255)` | Parent section (self-referencing FK) |
| `is_completed` | `BOOLEAN` | Completion status |
| `prompt` | `TEXT` | Generation prompt for this section |
| `document_id` | `VARCHAR(255)` | Associated document reference |
| `created_at` | `TIMESTAMPTZ` | Row creation timestamp |
| `updated_at` | `TIMESTAMPTZ` | Last update timestamp |

### `document_file_items`
Represents individual document files or entries.

| Column | Type | Description |
|--------|------|-------------|
| `id` | `VARCHAR(255)` | Primary key |
| `project_id` | `VARCHAR(255)` | Scoping project identifier |
| `title` | `VARCHAR(500)` | Document title |
| `content` | `TEXT` | Document body / content |
| `description` | `TEXT` | Optional description |
| `document_section_id` | `VARCHAR(255)` | FK → `document_sections(id)` (CASCADE) |
| `document_id` | `VARCHAR(255)` | Alternative document identifier |
| `extra` | `TEXT` | Arbitrary extra metadata |
| `is_embedded` | `BOOLEAN` | Whether stored in vector DB |
| `embed_status` | `VARCHAR(20)` | `pending` · `processing` · `completed` · `failed` |
| `created_at` | `TIMESTAMPTZ` | Row creation timestamp |
| `updated_at` | `TIMESTAMPTZ` | Last update timestamp |

## gRPC API

### Service: `DatabaseService`

| RPC | Request | Response | Description |
|-----|---------|----------|-------------|
| `StoreDocument` | `StoreDocumentRequest` | `StoreDocumentResponse` | Create or update a document with its section |
| `StoreSection` | `StoreSectionRequest` | `StoreSectionResponse` | Create or update a documentation section |
| `GetDocument` | `GetDocumentRequest` | `GetDocumentResponse` | Retrieve document content by project + document ID |
| `GetDocumentSections` | `GetDocumentSectionsRequest` | `GetDocumentSectionsResponse` | Return the nested table-of-contents tree for a project |
| `HealthCheck` | `HealthCheckRequest` | `HealthCheckResponse` | Liveness probe (returns `is_alive: true`) |

## Installation

### Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [PostgreSQL](https://www.postgresql.org/download/) 17+
- [Buf](https://buf.build/docs/installation) (for protobuf code generation)
- [Docker](https://docs.docker.com/get-docker/) (optional, for containerized deployment)

### 1. Generate Proto Code

Protobuf definitions are shared across services. Generate the Go code before building:

```bash
cd backend/shared
buf generate
```

### 2. Install Dependencies

```bash
cd backend/services/database
go mod tidy
```

### 3. Configure

Edit `configs/database.yaml` or set environment variables:

```yaml
server:
  host: "0.0.0.0"
  port: 9002

database:
  postgres_url: "postgres://user:password@localhost:5432/docagent?sslmode=disable"
  vector_db_url: "http://localhost:8000"
```

### 4. Run Migrations

Migrations run **automatically** on service startup. The service retries up to 10 times (with 3s delays) until the database is reachable.

### 5. Run the Service

```bash
# Using default config file
go run ./cmd/database

# Using environment variables
POSTGRES_URL="postgres://user:password@localhost:5432/docagent?sslmode=disable" \
  go run ./cmd/database
```

### 6. Run Tests

```bash
# Run all tests
go test -v ./...

# Or use the convenience script
./scripts/test.sh
```

### 7. Docker Deployment

```bash
# Build the Docker image
make docker-build

# Or use docker-compose from the project root
cd backend
docker-compose up database postgres
```

## Configuration

All settings can be configured via YAML file (`configs/database.yaml`) or environment variables. Environment variables take precedence.

| Environment Variable | YAML Path | Description | Default |
|---------------------|-----------|-------------|---------|
| `POSTGRES_URL` | `database.postgres_url` | PostgreSQL connection string | From config file |
| `VECTOR_DB_URL` | `database.vector_db_url` | Vector database endpoint | `http://localhost:8000` |
| `SERVER_HOST` | `server.host` | gRPC listen address | `0.0.0.0` |
| `SERVER_PORT` | `server.port` | gRPC listen port | `9002` |

> **Note:** Environment variables follow the pattern `DATABASE_POSTGRES_URL` → `database.postgres_url` (dots replaced with underscores). The config file is optional — if not found, all values must be provided via environment variables.

## Usage Examples

### Store a Document (via gRPC)

```go
client := apiv1.NewDatabaseServiceClient(grpcConn)

resp, err := client.StoreDocument(ctx, &apiv1.StoreDocumentRequest{
    Id:          "getting-started",
    Title:       "Getting Started",
    Content:     "# Getting Started\n\nWelcome to the documentation.",
    Description: "A beginner's guide",
    ProjectId:   "my-project",
})
// resp.Success == true
```

The service automatically creates a corresponding section (derived from the title) if one doesn't already exist.

### Retrieve a Document

```go
resp, err := client.GetDocument(ctx, &apiv1.GetDocumentRequest{
    ProjectId:  "my-project",
    DocumentId: "getting-started",
})
// resp.Doc == "# Getting Started\n\nWelcome to the documentation."
```

### Get the Section Tree

```go
resp, err := client.GetDocumentSections(ctx, &apiv1.GetDocumentSectionsRequest{
    ProjectId: "my-project",
})
// resp.Sections contains the nested tree:
// Introduction
// ├── Getting Started
// API Reference
// ├── Authentication
// ├── Endpoints
```

### Store a Section (with hierarchy)

```go
client.StoreSection(ctx, &apiv1.StoreSectionRequest{
    Id:        "endpoints",
    Title:     "API Endpoints",
    ProjectId: "my-project",
    ParentId:  "api-reference",
    Order:     1,
})
```

### Health Check

```go
resp, err := client.HealthCheck(ctx, &apiv1.HealthCheckRequest{})
// resp.IsAlive == true
```

## Project Structure

```
services/database/
├── cmd/database/
│   └── main.go              # Entrypoint: config, migrations, gRPC server startup
├── configs/
│   └── database.yaml         # Default configuration
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration loading (Viper + env vars)
│   ├── db/
│   │   └── migrate.go        # Automatic migration runner with retry logic
│   ├── grpc/
│   │   └── server.go         # gRPC handler registration and RPC implementations
│   ├── repository/
│   │   ├── repository.go     # Domain models & repository interfaces
│   │   ├── postgres.go       # PostgreSQL repository implementations
│   │   └── mock.go           # In-memory mock repositories for testing
│   └── service/
│       ├── database.go       # Business logic: Store, Get, Tree-building
│       └── database_test.go  # Unit tests
├── migrations/
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_add_embed_status.up.sql
│   └── 000002_add_embed_status.down.sql
├── scripts/
│   └── test.sh               # Convenience test runner
├── Dockerfile                # Multi-stage production Docker build
├── Makefile                  # Build, test, run, docker-build targets
├── go.mod / go.sum           # Go module dependencies
└── README.md
```

## Makefile Commands

| Target | Description |
|--------|-------------|
| `make proto` | Generate protobuf code via `buf generate` |
| `make build` | Compile the service binary to `bin/database` |
| `make test` | Run all tests with verbose output |
| `make run` | Build and run the service |
| `make docker-build` | Build a production Docker image |

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork** the repository and create a feature branch from `main`.
2. **Ensure proto code is generated** before building: `make proto`.
3. **Write tests** for any new functionality — the project uses mock repositories for fast, isolated unit tests.
4. **Run the full test suite**: `make test`.
5. **Submit a pull request** with a clear description of changes.

### Development Setup

```bash
# Clone the monorepo and navigate to the service
cd backend/services/database

# Generate proto (one-time)
make proto

# Install dependencies
go mod tidy

# Run tests
go test -v ./...
```

## License

This project is part of the Doc-Agent ecosystem. See the [LICENSE](../../../LICENSE) file for details.