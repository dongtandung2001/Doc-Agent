
"""
# AI Service Microservice

Complete implementation of the AI Service for Smart Codebase Summarization.

## Features
- gRPC API for chat interactions
- RAG (Retrieval-Augmented Generation) support
- Automatic documentation embedding
- Vector database integration (ChromaDB)
- LLM integration (OpenAI GPT-4)

## Setup

1. Create and activate a virtual environment (if not already done):
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

3. Generate gRPC code from proto file:
```bash
bash scripts/generate_grpc.sh
# Or manually:
python -m grpc_tools.protoc -I./protos --python_out=./generated --grpc_python_out=./generated ./protos/ai_service.proto
```

4. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration, especially OPENAI_API_KEY
```

5. Run the service:
```bash
python main.py
```

The server will start on port 50051 (or the port specified in GRPC_PORT environment variable).

## Testing

Run tests:
```bash
pytest tests/ -v
```

## Architecture

- **API Layer**: gRPC interface for chat requests
- **Conversation Orchestrator**: RAG decision logic and request routing
- **LLM Client**: OpenAI integration
- **Vector Store Manager**: ChromaDB integration for embeddings
- **Auto-embedding**: Automatic storage of generated documentation

## API Endpoints

### Chat
- Request: `ChatRequest{messages, project_id}`
- Response: `ChatResponse{content}`

### HealthCheck
- Request: `Empty`
- Response: `HealthCheckResponse{isAlive}`
"""