#!/usr/bin/env python3
"""
HTTP server that proxies to the AI service gRPC (Chat, CreateRAG, HealthCheck).
Used by the gateway so the web chat can call the Python RAG/Chat service.

Run the gRPC server first: python main.py
Then run this script and point the gateway config ai.base_url to http://localhost:8888
"""
import json
import os
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from urllib.parse import urlparse

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT))

import grpc
import generated.ai_service_pb2 as pb2
import generated.ai_service_pb2_grpc as pb2_grpc

GRPC_SERVER = "localhost:50051"
HTTP_PORT = int(os.environ.get("AI_HTTP_PORT", "8888"))


def call_chat(messages, project_id):
    channel = grpc.insecure_channel(GRPC_SERVER)
    stub = pb2_grpc.AIServiceStub(channel)
    request = pb2.ChatRequest(
        messages=[
            pb2.ChatMessage(role=m.get("role", "user"), content=m.get("content", ""))
            for m in messages
        ],
        project_id=project_id or "web_test_project",
    )
    response = stub.Chat(request, timeout=30)
    channel.close()
    return {"content": response.content or ""}


def call_create_rag(project_id):
    channel = grpc.insecure_channel(GRPC_SERVER)
    stub = pb2_grpc.AIServiceStub(channel)
    request = pb2.CreateRAGRequest(project_id=project_id or "web_test_project")
    response = stub.CreateRAG(request, timeout=10)
    channel.close()
    return {"success": response.success, "message": response.message}


def call_health():
    channel = grpc.insecure_channel(GRPC_SERVER)
    stub = pb2_grpc.AIServiceStub(channel)
    response = stub.HealthCheck(pb2.Empty())
    channel.close()
    return {"isAlive": response.isAlive}


class Handler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        print("[%s] %s" % (self.log_date_time_string(), format % args))

    def send_cors(self):
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")

    def do_OPTIONS(self):
        self.send_response(204)
        self.send_cors()
        self.end_headers()

    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path == "/api/health":
            try:
                out = call_health()
            except Exception as e:
                out = {"isAlive": False}
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_cors()
            self.end_headers()
            self.wfile.write(json.dumps(out).encode("utf-8"))
        elif self.path in ("/", "/index.html"):
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_cors()
            self.end_headers()
            self.wfile.write(b"<html><body><p>AI service HTTP proxy. Use gateway or POST /api/chat.</p></body></html>")
        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        parsed = urlparse(self.path)
        if parsed.path == "/api/chat":
            try:
                length = int(self.headers.get("Content-Length", 0))
                body = self.rfile.read(length).decode("utf-8")
                data = json.loads(body) if body else {}
                messages = data.get("messages", [])
                project_id = data.get("project_id", "web_test_project")
                out = call_chat(messages, project_id)
            except Exception as e:
                print("ERROR /api/chat:", e)
                self.send_response(500)
                self.send_header("Content-Type", "application/json")
                self.send_cors()
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode("utf-8"))
                return
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_cors()
            self.end_headers()
            self.wfile.write(json.dumps(out).encode("utf-8"))
        elif parsed.path == "/api/create_rag":
            try:
                length = int(self.headers.get("Content-Length", 0))
                body = self.rfile.read(length).decode("utf-8")
                data = json.loads(body) if body else {}
                project_id = data.get("project_id", "web_test_project")
                out = call_create_rag(project_id)
            except Exception as e:
                self.send_response(500)
                self.send_header("Content-Type", "application/json")
                self.send_cors()
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode("utf-8"))
                return
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_cors()
            self.end_headers()
            self.wfile.write(json.dumps(out).encode("utf-8"))
        else:
            self.send_response(404)
            self.end_headers()


def main():
    server = HTTPServer(("", HTTP_PORT), Handler)
    print("AI service HTTP proxy: http://localhost:%s" % HTTP_PORT)
    print("Ensure gRPC AI service is running: python main.py")
    server.serve_forever()


if __name__ == "__main__":
    main()
