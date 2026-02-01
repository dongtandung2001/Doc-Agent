FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY ../../shared /shared

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /database ./cmd/database

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /database .
COPY configs ./configs
COPY migrations ./migrations

EXPOSE 9002

CMD ["./database"]
