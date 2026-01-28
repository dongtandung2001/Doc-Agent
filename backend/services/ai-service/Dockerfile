# Build stage
FROM python:3.11-slim AS builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy and install Python dependencies
COPY requirements.txt .
RUN pip install --user --no-cache-dir -r requirements.txt

# Generate proto code
COPY protos/ ./protos/
RUN mkdir -p generated && \
    python -m grpc_tools.protoc \
    -I./protos \
    --python_out=./generated \
    --grpc_python_out=./generated \
    ./protos/ai_service.proto

# Runtime stage
FROM python:3.11-slim

# Create non-root user
RUN groupadd -r appuser && useradd -r -g appuser appuser

WORKDIR /app

# Copy Python packages from builder
COPY --from=builder /root/.local /home/appuser/.local

# Make sure scripts in .local are in PATH
ENV PATH=/home/appuser/.local/bin:$PATH

# Copy application
COPY --from=builder /build/generated ./generated
COPY src/ ./src/
COPY main.py ./

# Create data directories with proper ownership
RUN mkdir -p /app/data/chroma_db /app/logs && \
    chown -R appuser:appuser /app

# Switch to non-root
USER appuser

EXPOSE 50051

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import grpc; import generated.ai_service_pb2 as pb2; import generated.ai_service_pb2_grpc as svc; ch=grpc.insecure_channel('localhost:50051'); svc.AIServiceStub(ch).HealthCheck(pb2.Empty())" || exit 1

CMD ["python", "main.py"]
