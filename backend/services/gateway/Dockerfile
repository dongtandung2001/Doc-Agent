# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go.mod files first for better caching
COPY shared/go.mod shared/go.sum ./shared/
COPY services/gateway/go.mod services/gateway/go.sum ./services/gateway/

# Download dependencies (cached layer if go.mod unchanged)
WORKDIR /build/services/gateway
RUN go mod download

# Copy shared module source
WORKDIR /build
COPY shared/ ./shared/

# Copy gateway source code
COPY services/gateway/ ./services/gateway/

# Build with optimization flags and cache mount
WORKDIR /build/services/gateway
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -trimpath -o /bin/gateway ./cmd/gateway

# Runtime stage
FROM alpine:3.21

# Install CA certs and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary with proper ownership
COPY --from=builder --chown=appuser:appuser /bin/gateway /app/gateway

# Copy config if needed
COPY --chown=appuser:appuser services/gateway/configs ./configs

# Create logs directory with proper ownership
RUN mkdir -p /app/logs && chown -R appuser:appuser /app/logs

# Switch to non-root user
USER appuser

EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /app/gateway health || exit 1

CMD ["/app/gateway"]
