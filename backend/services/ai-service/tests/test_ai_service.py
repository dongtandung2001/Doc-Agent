# tests/test_ai_service.py
import sys
from pathlib import Path
import pytest

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))
import generated.ai_service_pb2 as ai_service_pb2
from src.ai_service_impl import AIServiceServicer

@pytest.fixture
def service():
    return AIServiceServicer()

def test_health_check(service):
    """Test health check endpoint."""
    from unittest.mock import Mock
    request = ai_service_pb2.Empty()
    context = Mock()
    response = service.HealthCheck(request, context)
    assert response.isAlive is True

def test_chat_validation(service):
    """Test chat request validation."""
    from unittest.mock import Mock
    # Test missing project_id
    request = ai_service_pb2.ChatRequest(
        messages=[ai_service_pb2.ChatMessage(role="user", content="test")],
        project_id=""
    )
    context = Mock()
    response = service.Chat(request, context)
    context.set_code.assert_called()

def test_create_rag_success(service):
    """Test CreateRAG with valid project_id."""
    from unittest.mock import Mock, patch
    with patch.object(service.orchestrator.vector_store, 'ensure_rag_index', return_value=True):
        request = ai_service_pb2.CreateRAGRequest(project_id="test-project-rag")
        context = Mock()
        response = service.CreateRAG(request, context)
    assert response.success is True
    assert "test-project-rag" in response.message

def test_create_rag_missing_project_id(service):
    """Test CreateRAG with missing project_id."""
    from unittest.mock import Mock
    request = ai_service_pb2.CreateRAGRequest(project_id="")
    context = Mock()
    response = service.CreateRAG(request, context)
    assert response.success is False
    assert "project_id" in response.message
    context.set_code.assert_called()