# Database Service

Manages PostgreSQL for structured document storage and retrieval per the [Design Specification](../../../docs/).

## Responsibilities

- **StoreDocument**: Store generated documentation in the SQL database (DocumentFileItems + optional DocumentSections)
- **GetDocument**: Retrieve document content by project_id and document_id
- **GetDocumentSections**: Return nested table-of-contents structure for a project

## Schema

### DocumentSections
Represents folders/categories in the documentation tree.
- `id`, `project_id`, `name`, `description`, `url`, `order`, `parent_id`, `is_completed`, `prompt`, `document_id`

### DocumentFileItems
Represents individual document files.
- `id`, `project_id`, `content`, `title`, `description`, `document_section_id`, `document_id`, `extra`, `is_embedded`

## Setup

### 1. Generate Proto (required before first build)
```bash
cd backend/shared
buf generate
```

### 2. Install Dependencies
```bash
cd backend/services/database
go mod tidy
```

### 3. Run Migrations
Migrations run automatically on startup. Ensure PostgreSQL is running and `POSTGRES_URL` is set.

### 4. Local Development
```bash
# With default config (configs/database.yaml)
go run ./cmd/database

# Or with env vars
POSTGRES_URL=postgres://user:password@localhost:5432/docagent?sslmode=disable go run ./cmd/database
```

### 5. Run Tests
```bash
go test -v ./...
# Or use the test script:
./scripts/test.sh
```

### 6. Docker
```bash
cd backend
docker-compose up database postgres
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_URL` | PostgreSQL connection string | From `configs/database.yaml` |
| `SERVER_HOST` | gRPC listen host | 0.0.0.0 |
| `SERVER_PORT` | gRPC listen port | 9002 |
