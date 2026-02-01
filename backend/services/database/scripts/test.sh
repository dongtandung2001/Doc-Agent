#!/bin/bash
# Run tests for the database service
# Prerequisites: go, buf (for proto generation)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$(dirname "$SCRIPT_DIR")"
SHARED_DIR="$(dirname "$(dirname "$SERVICE_DIR")")/shared"

echo "=== Database Service Test Script ==="

# Check for go
if ! command -v go &> /dev/null; then
    echo "Error: go is not installed or not in PATH"
    exit 1
fi

# Generate proto if gen folder doesn't exist
if [ ! -d "$SHARED_DIR/gen" ]; then
    echo "Generating proto code..."
    if command -v buf &> /dev/null; then
        (cd "$SHARED_DIR" && buf generate)
    else
        echo "Warning: buf not found. Proto code may need to be generated manually."
        echo "Run: cd $SHARED_DIR && buf generate"
    fi
fi

# Run from service directory
cd "$SERVICE_DIR"

echo ""
echo "=== Running go mod tidy ==="
go mod tidy

echo ""
echo "=== Building ==="
go build ./cmd/database

echo ""
echo "=== Running tests ==="
go test -v ./...

echo ""
echo "=== All checks passed ==="
