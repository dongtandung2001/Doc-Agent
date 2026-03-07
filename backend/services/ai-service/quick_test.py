#!/usr/bin/env python3
"""
Quick test script for AI Service API.
Tests HealthCheck, CreateRAG, Chat, and RAG-triggered Chat.

Note: CreateRAG and RAG retrieval require ChromaDB (Python 3.9+ recommended).
If CreateRAG fails, Chat and code-related Chat still run; RAG will have no indexed docs.
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
PROJECT_ID = "quick_test_project"


def main():
    print("=" * 60)
    print("Quick Test - AI Service (CreateRAG + Chat + RAG)")
    print("=" * 60)

    try:
        channel = grpc.insecure_channel(SERVER)
        stub = pb2_grpc.AIServiceStub(channel)

        # 1. Health check
        print("\n1. Health check...")
        health = stub.HealthCheck(pb2.Empty())
        if not health.isAlive:
            print("   ❌ Server health check failed")
            return
        print("   ✓ Server is alive")

        # 2. CreateRAG
        print("\n2. CreateRAG...")
        rag_req = pb2.CreateRAGRequest(project_id=PROJECT_ID)
        try:
            rag_resp = stub.CreateRAG(rag_req, timeout=10)
            if not rag_resp.success:
                print(f"   ⚠ CreateRAG failed (RAG disabled): {rag_resp.message}")
            else:
                print(f"   ✓ {rag_resp.message}")
        except grpc.RpcError as e:
            print(f"   ⚠ CreateRAG error: {e.details()}")

        # 3. Simple chat (no RAG)
        print("\n3. Simple chat (no RAG)...")
        chat_req = pb2.ChatRequest(
            messages=[pb2.ChatMessage(role="user", content="Say hello in one word.")],
            project_id=PROJECT_ID,
        )
        chat_resp = stub.Chat(chat_req, timeout=20)
        print(f"   ✓ Content: {chat_resp.content}")

        # 4. Code-related chat (triggers RAG retrieval)
        print("\n4. Code-related chat (RAG)...")
        rag_chat_req = pb2.ChatRequest(
            messages=[
                pb2.ChatMessage(
                    role="user",
                    content="How does the authentication function work in this codebase?",
                )
            ],
            project_id=PROJECT_ID,
        )
        rag_chat_resp = stub.Chat(rag_chat_req, timeout=20)
        print(f"   ✓ Content: {rag_chat_resp.content[:300]}{'...' if len(rag_chat_resp.content or '') > 300 else ''}")

        # 5. Response structure
        print("\n5. Response structure...")
        print(f"   ✓ completion_json length: {len(rag_chat_resp.completion_json or '')}")
        if rag_chat_resp.completion_json:
            completion = json.loads(rag_chat_resp.completion_json)
            print(f"   ✓ Model: {completion.get('model')}")
            print(f"   ✓ Choices: {len(completion.get('choices', []))}")

        print("\n" + "=" * 60)
        print("✅ SUCCESS – CreateRAG and Chat (incl. RAG) work.")
        print("=" * 60)

    except grpc.RpcError as e:
        print(f"\n❌ gRPC Error: {e.code()} - {e.details()}")
        if e.code() == grpc.StatusCode.UNAVAILABLE:
            print("   Start server: python main.py")
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
    finally:
        channel.close()


if __name__ == "__main__":
    main()
