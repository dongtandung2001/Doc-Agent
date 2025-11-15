
# Dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy proto file and generate gRPC code
COPY protos/ ./protos/
RUN python -m grpc_tools.protoc -I./protos --python_out=. --grpc_python_out=. ./protos/ai_service.proto

# Copy application code
COPY config.py logger.py vector_store.py llm_client.py conversation_orchestrator.py ai_service_impl.py server.py main.py ./

# Create directories for data and logs
RUN mkdir -p /app/data/chroma_db /app/logs

# Expose gRPC port
EXPOSE 50051

# Health check - use a simple Python script that actually tests the gRPC endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD python -c "import grpc; import ai_service_pb2; import ai_service_pb2_grpc; channel = grpc.insecure_channel('localhost:50051'); stub = ai_service_pb2_grpc.AIServiceStub(channel); stub.HealthCheck(ai_service_pb2.Empty()); channel.close()" || exit 1

# Run the service
CMD ["python", "main.py"]