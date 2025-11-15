# tests/test_ai_service.py
import pytest
from ai_service_impl import AIServiceServicer, ChatRequest, ChatMessage

@pytest.fixture
def service():
    return AIServiceServicer()

def test_health_check(service):
    """Test health check endpoint."""
    from ai_service_impl import Empty
    response = service.HealthCheck(Empty(), None)
    assert response.isAlive is True

def test_chat_validation(service):
    """Test chat request validation."""
    # Test missing project_id
    request = ChatRequest(messages=[ChatMessage("user", "test")], project_id="")
    context = Mock()
    response = service.Chat(request, context)
    context.set_code.assert_called()