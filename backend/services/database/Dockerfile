# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go.mod files first for better caching
COPY shared/go.mod shared/go.sum ./shared/
COPY services/database/go.mod services/database/go.sum ./services/database/

# Download dependencies (cached layer if go.mod unchanged)
WORKDIR /build/services/database
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

# Copy database source code
COPY services/database/ ./services/database/

# Build with optimization flags and cache mount
WORKDIR /build/services/database
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -trimpath -o /bin/database ./cmd/database

# Runtime stage
FROM alpine:3.21

# Install CA certs and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary with proper ownership
COPY --from=builder --chown=appuser:appuser /bin/database /app/database

# Copy config if needed
COPY --chown=appuser:appuser services/database/configs ./configs

# Copy migrations
COPY --chown=appuser:appuser services/database/migrations ./migrations

# Create logs directory with proper ownership
RUN mkdir -p /app/logs && chown -R appuser:appuser /app/logs

# Switch to non-root user
USER appuser

EXPOSE 9002

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /app/database health || exit 1

CMD ["/app/database"]
