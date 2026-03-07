# Test client for AI Service

import grpc
import sys
import os
import pytest
from dotenv import load_dotenv

# Import generated protobuf files
from pathlib import Path

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))
import generated.ai_service_pb2 as ai_service_pb2
import generated.ai_service_pb2_grpc as ai_service_pb2_grpc

# Load environment variables
load_dotenv()

# Default server address
DEFAULT_SERVER = os.getenv("GRPC_SERVER", "localhost:50051")


@pytest.fixture
def project_id():
    """Default project ID for integration tests."""
    return "test_project_123"


@pytest.fixture
def server_address():
    """Default gRPC server address for integration tests."""
    return DEFAULT_SERVER


def test_health_check(server_address: str = DEFAULT_SERVER):
    """Test the health check endpoint."""
    print("Testing Health Check...")
    
    try:
        with grpc.insecure_channel(server_address) as channel:
            stub = ai_service_pb2_grpc.AIServiceStub(channel)
            response = stub.HealthCheck(ai_service_pb2.Empty())
            print(f"✓ Health Check Response: isAlive={response.isAlive}")
            return response.isAlive
    except grpc.RpcError as e:
        print(f"✗ Health Check failed: {e.code()} - {e.details()}")
        return False
    except Exception as e:
        print(f"✗ Connection error: {str(e)}")
        print(f"  Make sure the server is running on {server_address}")
        return False


def test_chat(project_id: str, server_address: str, user_message: str = "Hello! Can you help me understand this project?"):
    """Test the chat endpoint."""
    print(f"\nTesting Chat Endpoint...")
    print(f"Project ID: {project_id}")
    print(f"User Message: {user_message}")

    try:
        with grpc.insecure_channel(server_address) as channel:
            stub = ai_service_pb2_grpc.AIServiceStub(channel)
            
            request = ai_service_pb2.ChatRequest(
                messages=[
                    ai_service_pb2.ChatMessage(role="user", content=user_message)
                ],
                project_id=project_id
            )
            
            print("Waiting for response...")
            response = stub.Chat(request)
            print(f"\n✓ AI Response:\n{response.content}")
            return response.content
    except grpc.RpcError as e:
        print(f"✗ Chat request failed: {e.code()} - {e.details()}")
        return None
    except Exception as e:
        print(f"✗ Error: {str(e)}")
        return None


def test_rag_query(project_id: str, server_address: str):
    """Test a code-related query that should trigger RAG."""
    print("\nTesting RAG (Code-related query)...")

    user_message = "How does the authentication function work in this codebase?"
    print(f"Query: {user_message}")

    try:
        with grpc.insecure_channel(server_address) as channel:
            stub = ai_service_pb2_grpc.AIServiceStub(channel)
            
            request = ai_service_pb2.ChatRequest(
                messages=[
                    ai_service_pb2.ChatMessage(role="user", content=user_message)
                ],
                project_id=project_id
            )
            
            # Add metadata for request_name if needed
            metadata = [('request_name', 'Code Query')]
            print("Waiting for response (with RAG)...")
            response = stub.Chat(request, metadata=metadata)
            print(f"\n✓ AI Response (with RAG):\n{response.content}")
            return response.content
    except grpc.RpcError as e:
        print(f"✗ RAG query failed: {e.code()} - {e.details()}")
        return None
    except Exception as e:
        print(f"✗ Error: {str(e)}")
        return None


def test_doc_generation(project_id: str, server_address: str):
    """Test document generation that should trigger auto-embedding."""
    print("\nTesting Doc Generation (Auto-embedding)...")

    user_message = "Generate documentation for the user authentication module."
    print(f"Query: {user_message}")

    try:
        with grpc.insecure_channel(server_address) as channel:
            stub = ai_service_pb2_grpc.AIServiceStub(channel)
            
            request = ai_service_pb2.ChatRequest(
                messages=[
                    ai_service_pb2.ChatMessage(role="user", content=user_message)
                ],
                project_id=project_id
            )
            
            # Add metadata to indicate doc generation
            metadata = [('request_name', 'Doc Generating Authentication')]
            print("Waiting for response (will auto-embed)...")
            response = stub.Chat(request, metadata=metadata)
            print(f"\n✓ Generated Documentation:\n{response.content}")
            print("\n✓ Documentation should be auto-embedded in vector DB")
            return response.content
    except grpc.RpcError as e:
        print(f"✗ Doc generation failed: {e.code()} - {e.details()}")
        return None
    except Exception as e:
        print(f"✗ Error: {str(e)}")
        return None


def interactive_mode(server_address: str = DEFAULT_SERVER):
    """Interactive chat mode."""
    print("\n" + "=" * 60)
    print("AI Service Interactive Client")
    print("=" * 60)
    print(f"Server: {server_address}")

    project_id = input("\nEnter project ID: ").strip()
    if not project_id:
        project_id = "test_project"
        print(f"Using default project ID: {project_id}")

    print("\nType your messages (or 'quit' to exit)")
    print("-" * 60)

    conversation_history = []

    try:
        channel = grpc.insecure_channel(server_address)
        stub = ai_service_pb2_grpc.AIServiceStub(channel)
        
        while True:
            user_input = input("\nYou: ").strip()

            if user_input.lower() in ['quit', 'exit', 'q']:
                print("Goodbye!")
                break

            if not user_input:
                continue

            # Add to conversation history
            conversation_history.append({
                "role": "user",
                "content": user_input
            })

            try:
                # Build request with conversation history
                messages = [
                    ai_service_pb2.ChatMessage(role=msg["role"], content=msg["content"])
                    for msg in conversation_history
                ]
                
                request = ai_service_pb2.ChatRequest(
                    messages=messages,
                    project_id=project_id
                )
                
                print("AI: ", end="", flush=True)
                response = stub.Chat(request)
                print(response.content)
                
                # Add AI response to history
                conversation_history.append({
                    "role": "assistant",
                    "content": response.content
                })
                
            except grpc.RpcError as e:
                print(f"\n✗ Error: {e.code()} - {e.details()}")
            except Exception as e:
                print(f"\n✗ Error: {str(e)}")
        
        channel.close()
    except Exception as e:
        print(f"\n✗ Failed to connect to server: {str(e)}")
        print(f"  Make sure the server is running on {server_address}")


def main():
    """Main test runner."""
    print("\n" + "=" * 60)
    print("AI Service gRPC Client - Test Suite")
    print("=" * 60)
    
    # Parse command line arguments
    server_address = DEFAULT_SERVER
    mode = 'test'  # 'test' or 'interactive'
    
    if len(sys.argv) > 1:
        if sys.argv[1] in ['--help', '-h']:
            print("\nUsage:")
            print("  python client.py                    # Run test suite")
            print("  python client.py interactive        # Interactive mode")
            print("  python client.py <server_address>   # Run tests against specific server")
            print("  python client.py <server> interactive  # Interactive mode with custom server")
            print("\nExamples:")
            print("  python client.py")
            print("  python client.py localhost:50051")
            print("  python client.py interactive")
            print("  python client.py localhost:50051 interactive")
            return
        elif sys.argv[1] == 'interactive':
            mode = 'interactive'
        elif len(sys.argv) > 2 and sys.argv[2] == 'interactive':
            server_address = sys.argv[1]
            mode = 'interactive'
        else:
            server_address = sys.argv[1]
    
    print(f"Server: {server_address}")
    print("=" * 60)

    if mode == 'interactive':
        interactive_mode(server_address)
    else:
        # Run test suite
        project_id = "test_project_123"

        print("\n1. Health Check")
        if not test_health_check(server_address):
            print("\n⚠ Server health check failed. Make sure the server is running.")
            print(f"  Start the server with: python main.py")
            return

        print("\n2. Simple Chat")
        test_chat(project_id, "Hello! Can you help me understand this project?", server_address)

        print("\n3. Code-related Query (RAG)")
        test_rag_query(project_id, server_address)

        print("\n4. Documentation Generation (Auto-embedding)")
        test_doc_generation(project_id, server_address)

        print("\n" + "=" * 60)
        print("Test suite complete!")
        print("\nTo use interactive mode, run:")
        print("python client.py interactive")
        print("=" * 60)


if __name__ == '__main__':
    main()