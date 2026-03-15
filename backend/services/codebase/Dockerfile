# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go.mod files first for better caching
COPY shared/go.mod shared/go.sum ./shared/
COPY services/codebase/go.mod services/codebase/go.sum ./services/codebase/

# Download dependencies (cached layer if go.mod unchanged)
WORKDIR /build/services/codebase
RUN go mod download

# Copy shared module source (no gen/ - it is gitignored)
WORKDIR /build
COPY shared/ ./shared/

# Generate shared proto Go code (gen/) so the shared module can be built
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest && \
    go install github.com/bufbuild/buf/cmd/buf@latest && \
    (cd /build/shared && buf generate)

# Copy codebase source code
COPY services/codebase/ ./services/codebase/

# Build with optimization flags and cache mount
WORKDIR /build/services/codebase
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -trimpath -o /bin/codebase ./cmd/codebase

# Runtime stage
FROM alpine:3.21

# Install CA certs, wget for healthcheck, and create non-root user
RUN apk --no-cache add ca-certificates wget && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary with proper ownership
COPY --from=builder --chown=appuser:appuser /bin/codebase /app/codebase

# Copy config if needed
COPY --chown=appuser:appuser services/codebase/configs ./configs

# Copy prompts directory
COPY --chown=appuser:appuser services/codebase/internal/prompts ./internal/prompts

# Create logs directory with proper ownership
RUN mkdir -p /app/logs && chown -R appuser:appuser /app/logs

# Switch to non-root user
USER appuser

EXPOSE 9001

# Healthcheck - verify the command exists first
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9001/health || exit 1

CMD ["/app/codebase"]
