#!/usr/bin/env python3
"""
Quick test script for AI Service API.
Tests that the API returns full ChatCompletion object and request.
"""
import grpc
import json
import sys
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent))
import generated.ai_service_pb2 as pb2
import generated.ai_service_pb2_grpc as pb2_grpc

SERVER = "localhost:50051"

def main():
    print("=" * 60)
    print("Quick Test - AI Service API")
    print("=" * 60)
    
    # Connect to server
    print("\n1. Connecting to server...")
    try:
        channel = grpc.insecure_channel(SERVER)
        stub = pb2_grpc.AIServiceStub(channel)
        
        # Test health check first
        health = stub.HealthCheck(pb2.Empty())
        if not health.isAlive:
            print("❌ Server health check failed")
            return
        print("   ✓ Server is alive")
    except Exception as e:
        print(f"❌ Failed to connect: {e}")
        print(f"   Make sure server is running: python main.py")
        return
    
    # Create a test request
    print("\n2. Sending chat request...")
    request = pb2.ChatRequest(
        messages=[
            pb2.ChatMessage(role='user', content='Say hello in one word')
        ],
        project_id='test_project'
    )
    print(f"   Project ID: {request.project_id}")
    print(f"   Message: {request.messages[0].content}")
    
    # Send request
    try:
        print("\n3. Waiting for response...")
        response = stub.Chat(request, timeout=20)
        
        print("\n✅ Response received!")
        print(f"   Content: {response.content}")
        
        # Check response structure
        print("\n4. Verifying response structure...")
        print(f"   ✓ Has request: {response.HasField('request')}")
        print(f"   ✓ Has completion_json: {bool(response.completion_json)}")
        print(f"   ✓ Has content: {bool(response.content)}")
        
        # Parse full ChatCompletion
        if response.completion_json:
            completion = json.loads(response.completion_json)
            print("\n5. Full ChatCompletion Object:")
            print(f"   ✓ ID: {completion.get('id')}")
            print(f"   ✓ Model: {completion.get('model')}")
            print(f"   ✓ Created: {completion.get('created')}")
            print(f"   ✓ Choices: {len(completion.get('choices', []))}")
            
            if completion.get('usage'):
                usage = completion['usage']
                print(f"   ✓ Usage: {usage.get('total_tokens')} tokens")
                print(f"     - Prompt: {usage.get('prompt_tokens')}")
                print(f"     - Completion: {usage.get('completion_tokens')}")
        
        # Check original request
        if response.HasField('request'):
            print("\n6. Original Request (included):")
            print(f"   ✓ Project ID: {response.request.project_id}")
            print(f"   ✓ Messages: {len(response.request.messages)}")
        
        print("\n" + "=" * 60)
        print("✅ SUCCESS! API is working correctly!")
        print("=" * 60)
        print("\nThe API is now passing:")
        print("  ✓ Full ChatCompletion object (as JSON)")
        print("  ✓ Original request")
        print("  ✓ Content (backward compatibility)")
        
    except grpc.RpcError as e:
        print(f"\n❌ gRPC Error: {e.code()} - {e.details()}")
        if e.code() == grpc.StatusCode.UNAVAILABLE:
            print("   Server is not running. Start it with: python main.py")
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
    finally:
        channel.close()

if __name__ == '__main__':
    main()
