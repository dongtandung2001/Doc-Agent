# Microservices Architecture

This directory contains all the backend microservices for the Doc-Agent project, reorganized into true independent services that match the architecture diagram.

## Structure

```
backend/
├── services/              # Independent microservices
│   ├── gateway/          # API Gateway (HTTP/gRPC proxy)
│   ├── codebase/         # Codebase Analysis Service
│   ├── docgen/           # Document Generation Worker
│   └── database/         # Database Service (SQL + Vector DB)
├── shared/               # Shared code and proto definitions
│   ├── gen/             # Generated proto code
│   ├── api/proto/v1/    # Proto definitions
│   └── pkg/             # Common utilities
└── docker-compose.yml    # Local development setup
```

## Services

### 1. Gateway Service
- **Port**: 8080
- **Type**: HTTP/gRPC API Gateway
- **Responsibilities**:
  - Exposes HTTP/Connect-RPC endpoints to frontend
  - Proxies requests to backend microservices
  - Handles CORS, logging, recovery middleware

### 2. Codebase Analysis Service
- **Port**: 9001
- **Type**: gRPC service
- **Responsibilities**:
  - Analyzes project structure
  - Classifies codebase using AI
  - Generates documentation tasks
  - Enqueues tasks to message queue for DocGen

### 3. Document Generation Worker
- **Port**: None (worker service)
- **Type**: Background worker (no server interface)
- **Clients**: AI Service, Gateway (for Local Agent access)
- **Responsibilities**:
  - Consumes tasks from message queue (future: RabbitMQ/Kafka)
  - Calls AI Service to generate documentation sections
  - Calls Gateway to request file content from Local Agent
  - Stores completed sections in Database service

### 4. Database Service
- **Port**: 9002
- **Type**: gRPC service
- **Responsibilities**:
  - Manages PostgreSQL for structured data
  - Manages Vector DB for embeddings
  - Provides document CRUD operations

## Key Principles

### True Microservices
Each service:
- ✅ Has its own `go.mod` (independent module)
- ✅ Can be built independently
- ✅ Can be deployed independently
- ✅ Only communicates via gRPC/message queue
- ✅ Has no compile-time dependencies on other services

### Shared Module
The `shared/` module contains:
- Proto definitions and generated code
- Common data models
- Utilities that are truly shared

Each service imports from shared using:
```go
import apiv1 "github.com/dongtandung2001/Doc-Agent/backend/shared/gen/api/proto/v1"
```

## Development

### Build All Services
```bash
# Build gateway
cd services/gateway && go build ./cmd/gateway

# Build codebase
cd services/codebase && go build ./cmd/codebase

# Build docgen
cd services/docgen && go build ./cmd/docgen

# Build database
cd services/database && go build ./cmd/database
```

### Run with Docker Compose
```bash
cd backend
docker-compose up --build
```

This will start:
- Gateway on :8080
- Codebase service on :9001
- DocGen worker (no exposed port - background worker)
- Database service on :9002
- PostgreSQL on :5432

### Local Development
Each service can be run independently:

```bash
# Terminal 1 - Database service
cd services/database
go run ./cmd/database

# Terminal 2 - Codebase service
cd services/codebase
go run ./cmd/codebase

# Terminal 3 - DocGen worker
cd services/docgen
go run ./cmd/docgen

# Terminal 4 - Gateway
cd services/gateway
go run ./cmd/gateway
```

## Service Communication

```
Frontend/CLI
    ↓
Gateway (HTTP/Connect)
    ↓
├─→ Codebase Service (gRPC)
├─→ Database Service (gRPC)
└─→ Local Agent (gRPC)

Codebase → Message Queue → DocGen Worker
                            ↓
                     ┌──────┴──────┐
                     ↓              ↓
              AI Service      Gateway → Local Agent
                     ↓
              Database Service
```

## Next Steps

1. **Message Queue**: Add RabbitMQ/Kafka for Codebase → DocGen communication
2. **Service Discovery**: Add Consul or etcd for dynamic service discovery
3. **Observability**: Add distributed tracing (Jaeger/Zipkin)
4. **API Documentation**: Generate OpenAPI/Swagger docs from proto files
5. **Testing**: Add integration tests for each service
6. **CI/CD**: Setup independent deployment pipelines per service

