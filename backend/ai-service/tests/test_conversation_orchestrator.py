import pytest
from conversation_orchestrator import ConversationOrchestrator


@pytest.fixture
def orchestrator():
    return ConversationOrchestrator()


def test_is_code_related(orchestrator):
    """Test code-related detection."""
    messages = [
        {"role": "user", "content": "How does the authentication function work?"}
    ]
    assert orchestrator.is_code_related(messages) is True

    messages = [
        {"role": "user", "content": "What's the weather today?"}
    ]
    assert orchestrator.is_code_related(messages) is False


def test_should_use_rag(orchestrator):
    """Test RAG usage decision."""
    messages = [
        {"role": "user", "content": "Explain the API endpoint"}
    ]
    assert orchestrator.should_use_rag(messages, None) is True
    assert orchestrator.should_use_rag(messages, "Doc Generating API") is False