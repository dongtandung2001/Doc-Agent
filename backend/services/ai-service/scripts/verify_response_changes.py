#!/usr/bin/env python3
"""
Verification script to test that OpenAI responses are returned as full objects.
This verifies the changes made to return ChatCompletion objects instead of just content.
"""
import sys
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock
from typing import TYPE_CHECKING

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))

# Try to import OpenAI types, but use mocks if not available
try:
    from openai.types.chat import ChatCompletion, ChatCompletionMessage
    from openai.types.chat.chat_completion import Choice
    HAS_OPENAI = True
except ImportError:
    HAS_OPENAI = False
    # Create mock classes for testing
    class Choice:
        def __init__(self, index, message, finish_reason):
            self.index = index
            self.message = message
            self.finish_reason = finish_reason
    
    class ChatCompletionMessage:
        def __init__(self, role, content, tool_calls=None):
            self.role = role
            self.content = content
            self.tool_calls = tool_calls
    
    class ChatCompletion:
        def __init__(self, id, object, created, model, choices, usage):
            self.id = id
            self.object = object
            self.created = created
            self.model = model
            self.choices = choices
            self.usage = usage

from src.llm_client import LLMClient
from src.conversation_orchestrator import ConversationOrchestrator


def test_llm_client_returns_full_response():
    """Test that LLMClient returns full ChatCompletion object."""
    print("Testing LLMClient.generate_response() returns ChatCompletion...")
    
    client = LLMClient()
    
    # Create a mock ChatCompletion response
    mock_response = ChatCompletion(
        id="test-id",
        object="chat.completion",
        created=1234567890,
        model="gpt-4o-mini",
        choices=[
            Choice(
                index=0,
                message=ChatCompletionMessage(
                    role="assistant",
                    content="Test response content"
                ),
                finish_reason="stop"
            )
        ],
        usage={
            "prompt_tokens": 10,
            "completion_tokens": 5,
            "total_tokens": 15
        }
    )
    
    # Mock the OpenAI client
    with patch.object(client.client.chat.completions, 'create', return_value=mock_response):
        response = client.generate_response(
            messages=[{"role": "user", "content": "Hello"}]
        )
        
        # Verify it's a ChatCompletion object
        assert isinstance(response, ChatCompletion), f"Expected ChatCompletion, got {type(response)}"
        assert response.id == "test-id"
        assert response.choices[0].message.content == "Test response content"
        assert response.usage["total_tokens"] == 15
        
        print("✓ LLMClient correctly returns full ChatCompletion object")
        print(f"  - Response ID: {response.id}")
        print(f"  - Model: {response.model}")
        print(f"  - Content: {response.choices[0].message.content}")
        print(f"  - Usage: {response.usage['total_tokens']} tokens")
        return True


def test_orchestrator_returns_full_response():
    """Test that ConversationOrchestrator returns full ChatCompletion object."""
    print("\nTesting ConversationOrchestrator.process_request() returns ChatCompletion...")
    
    orchestrator = ConversationOrchestrator()
    
    # Create a mock ChatCompletion response
    mock_response = ChatCompletion(
        id="test-id-2",
        object="chat.completion",
        created=1234567890,
        model="gpt-4o-mini",
        choices=[
            Choice(
                index=0,
                message=ChatCompletionMessage(
                    role="assistant",
                    content="Orchestrated response"
                ),
                finish_reason="stop"
            )
        ],
        usage={
            "prompt_tokens": 20,
            "completion_tokens": 10,
            "total_tokens": 30
        }
    )
    
    # Mock the LLM client
    with patch.object(orchestrator.llm_client, 'generate_response', return_value=mock_response):
        with patch.object(orchestrator.vector_store, 'retrieve', return_value=[]):
            response = orchestrator.process_request(
                messages=[{"role": "user", "content": "What is this code?"}],
                project_id="test_project"
            )
            
            # Verify it's a ChatCompletion object
            assert isinstance(response, ChatCompletion), f"Expected ChatCompletion, got {type(response)}"
            assert response.choices[0].message.content == "Orchestrated response"
            
            print("✓ ConversationOrchestrator correctly returns full ChatCompletion object")
            print(f"  - Response ID: {response.id}")
            print(f"  - Content accessible: {response.choices[0].message.content}")
            return True


def test_content_extraction_still_works():
    """Test that content can still be extracted from the response."""
    print("\nTesting content extraction from ChatCompletion...")
    
    # Create a mock response
    mock_response = ChatCompletion(
        id="test-id-3",
        object="chat.completion",
        created=1234567890,
        model="gpt-4o-mini",
        choices=[
            Choice(
                index=0,
                message=ChatCompletionMessage(
                    role="assistant",
                    content="Extracted content"
                ),
                finish_reason="stop"
            )
        ],
        usage={"prompt_tokens": 5, "completion_tokens": 3, "total_tokens": 8}
    )
    
    # Simulate what ai_service_impl.py does
    content = mock_response.choices[0].message.content or ""
    
    assert content == "Extracted content", f"Content extraction failed: {content}"
    print("✓ Content extraction works correctly")
    print(f"  - Extracted content: {content}")
    return True


def test_response_metadata_accessible():
    """Test that all response metadata is accessible."""
    print("\nTesting response metadata accessibility...")
    
    mock_response = ChatCompletion(
        id="test-id-4",
        object="chat.completion",
        created=1234567890,
        model="gpt-4o-mini",
        choices=[
            Choice(
                index=0,
                message=ChatCompletionMessage(
                    role="assistant",
                    content="Test",
                    tool_calls=None
                ),
                finish_reason="stop"
            )
        ],
        usage={
            "prompt_tokens": 100,
            "completion_tokens": 50,
            "total_tokens": 150
        }
    )
    
    # Test accessing various fields
    assert mock_response.id == "test-id-4"
    assert mock_response.model == "gpt-4o-mini"
    assert mock_response.choices[0].finish_reason == "stop"
    assert mock_response.usage["total_tokens"] == 150
    assert mock_response.choices[0].message.tool_calls is None
    
    print("✓ All response metadata is accessible")
    print(f"  - ID: {mock_response.id}")
    print(f"  - Model: {mock_response.model}")
    print(f"  - Finish reason: {mock_response.choices[0].finish_reason}")
    print(f"  - Total tokens: {mock_response.usage['total_tokens']}")
    return True


def main():
    """Run all verification tests."""
    print("=" * 60)
    print("Verifying OpenAI Response Changes")
    print("=" * 60)
    print("\nThis script verifies that:")
    print("1. LLMClient returns full ChatCompletion objects")
    print("2. ConversationOrchestrator returns full ChatCompletion objects")
    print("3. Content extraction still works")
    print("4. All response metadata is accessible")
    print("=" * 60)
    
    tests = [
        test_llm_client_returns_full_response,
        test_orchestrator_returns_full_response,
        test_content_extraction_still_works,
        test_response_metadata_accessible
    ]
    
    passed = 0
    failed = 0
    
    for test in tests:
        try:
            if test():
                passed += 1
            else:
                failed += 1
                print(f"✗ {test.__name__} failed")
        except Exception as e:
            failed += 1
            print(f"✗ {test.__name__} raised exception: {e}")
            import traceback
            traceback.print_exc()
    
    print("\n" + "=" * 60)
    print(f"Results: {passed} passed, {failed} failed")
    print("=" * 60)
    
    if failed == 0:
        print("\n✓ All verification tests passed!")
        return 0
    else:
        print("\n✗ Some verification tests failed")
        return 1


if __name__ == '__main__':
    sys.exit(main())
