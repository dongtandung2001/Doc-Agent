# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go.mod files first for better caching
COPY shared/go.mod shared/go.sum ./shared/
COPY services/docgen/go.mod services/docgen/go.sum ./services/docgen/

# Download dependencies (cached layer if go.mod unchanged)
WORKDIR /build/services/docgen
RUN go mod download

# Copy shared module source
WORKDIR /build
COPY shared/ ./shared/

# Copy docgen source code
COPY services/docgen/ ./services/docgen/

# Build with optimization flags and cache mount
WORKDIR /build/services/docgen
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -trimpath -o /bin/docgen ./cmd/docgen

# Runtime stage
FROM alpine:3.21

# Install CA certs and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary with proper ownership
COPY --from=builder --chown=appuser:appuser /bin/docgen /app/docgen

# Copy config if needed
COPY --chown=appuser:appuser services/docgen/configs ./configs

# Copy prompts directory
COPY --chown=appuser:appuser services/docgen/internal/prompts ./internal/prompts

# Create necessary directories with proper ownership
RUN mkdir -p /app/generated_docs /app/logs && \
    chown -R appuser:appuser /app/generated_docs /app/logs

# Switch to non-root user
USER appuser

# No EXPOSE - this is a worker, not a server

CMD ["/app/docgen"]
