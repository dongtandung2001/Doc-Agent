#!/usr/bin/env python3
"""
Test script to verify that the API returns full ChatCompletion object and request.
"""
import grpc
import sys
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))
import generated.ai_service_pb2 as ai_service_pb2
import generated.ai_service_pb2_grpc as ai_service_pb2_grpc

# Default server address
DEFAULT_SERVER = "localhost:50051"


def test_full_response():
    """Test that the API returns full ChatCompletion and request."""
    print("=" * 60)
    print("Testing Full ChatCompletion Response")
    print("=" * 60)
    
    try:
        with grpc.insecure_channel(DEFAULT_SERVER) as channel:
            stub = ai_service_pb2_grpc.AIServiceStub(channel)
            
            # Create a test request
            request = ai_service_pb2.ChatRequest(
                messages=[
                    ai_service_pb2.ChatMessage(
                        role="user",
                        content="Hello! Say 'test' in one word."
                    )
                ],
                project_id="test_project_123"
            )
            
            print("\n1. Sending request...")
            print(f"   Project ID: {request.project_id}")
            print(f"   Message: {request.messages[0].content}")
            
            print("\n2. Waiting for response...")
            response = stub.Chat(request)
            
            print("\n3. Verifying response structure...")
            
            # Check that request is included
            assert response.HasField('request'), "❌ Response should include 'request' field"
            assert response.request.project_id == request.project_id, "❌ Request project_id mismatch"
            print("   ✓ Request included in response")
            
            # Check that completion_json is included
            assert response.completion_json, "❌ Response should include 'completion_json' field"
            print("   ✓ Completion JSON included in response")
            
            # Parse and check completion JSON
            import json
            completion = json.loads(response.completion_json)
            assert completion.get('id'), f"❌ Completion should have 'id', got: {completion.get('id')}"
            assert completion.get('model'), f"❌ Completion should have 'model', got: {completion.get('model')}"
            assert completion.get('created', 0) > 0, f"❌ Completion should have 'created' timestamp"
            print(f"   ✓ Completion ID: {completion.get('id')}")
            print(f"   ✓ Model: {completion.get('model')}")
            print(f"   ✓ Created: {completion.get('created')}")
            
            # Check choices
            assert len(completion.get('choices', [])) > 0, "❌ Completion should have at least one choice"
            choice = completion['choices'][0]
            assert choice.get('message', {}).get('role'), "❌ Choice should have message role"
            print(f"   ✓ Choices: {len(completion.get('choices', []))}")
            print(f"   ✓ Message role: {choice['message']['role']}")
            
            # Check usage
            usage = completion.get('usage', {})
            assert usage.get('total_tokens', 0) > 0, f"❌ Usage should have total_tokens > 0, got: {usage.get('total_tokens')}"
            print(f"   ✓ Usage - Total tokens: {usage.get('total_tokens')}")
            print(f"   ✓ Usage - Prompt tokens: {usage.get('prompt_tokens')}")
            print(f"   ✓ Usage - Completion tokens: {usage.get('completion_tokens')}")
            
            # Check content (backward compatibility)
            assert response.content, "❌ Response should have content field"
            print(f"   ✓ Content (backward compat): {response.content[:50]}...")
            
            # Check that content matches completion content
            completion_content = choice['message'].get('content', '')
            assert response.content == completion_content, "❌ Content field should match completion message content"
            print("   ✓ Content matches completion message content")
            
            print("\n" + "=" * 60)
            print("✅ All tests passed! Full ChatCompletion object is working.")
            print("=" * 60)
            
            # Print full response structure
            print("\n📋 Response Structure:")
            print(f"   - Request included: ✓")
            print(f"   - Completion ID: {completion.get('id')}")
            print(f"   - Completion Model: {completion.get('model')}")
            print(f"   - Completion Choices: {len(completion.get('choices', []))}")
            print(f"   - Completion Usage: {usage.get('total_tokens')} tokens")
            print(f"   - Content: {len(response.content)} chars")
            print(f"   - Completion JSON: {len(response.completion_json)} chars")
            
            return True
            
    except grpc.RpcError as e:
        print(f"\n❌ gRPC Error: {e.code()} - {e.details()}")
        return False
    except AssertionError as e:
        print(f"\n❌ Assertion failed: {e}")
        return False
    except Exception as e:
        print(f"\n❌ Error: {str(e)}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == '__main__':
    success = test_full_response()
    sys.exit(0 if success else 1)
