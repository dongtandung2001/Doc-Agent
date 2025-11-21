
# ai_service_impl.py
import grpc
import sys
from typing import List
from pathlib import Path

# Add parent directory to path for generated imports
sys.path.insert(0, str(Path(__file__).parent.parent))
import generated.ai_service_pb2 as ai_service_pb2
import generated.ai_service_pb2_grpc as ai_service_pb2_grpc

from src.conversation_orchestrator import ConversationOrchestrator
from src.config import Config
from src.logger import setup_logger

logger = setup_logger(__name__)


class AIServiceServicer(ai_service_pb2_grpc.AIServiceServicer):
    """gRPC service implementation for AI Service."""

    def __init__(self):
        self.orchestrator = ConversationOrchestrator()
        logger.info("AI Service initialized")

    def Chat(self, request: ai_service_pb2.ChatRequest, context) -> ai_service_pb2.ChatResponse:
        """Handle chat requests with RAG support."""
        try:
            # Validate request
            if not request.project_id:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("project_id is required")
                return ai_service_pb2.ChatResponse(content="")

            if not request.messages:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("messages cannot be empty")
                return ai_service_pb2.ChatResponse(content="")

            # Convert messages to dict format
            messages = [
                {"role": msg.role, "content": msg.content}
                for msg in request.messages
            ]

            logger.info(f"Received chat request for project: {request.project_id}")

            # Get request_name from metadata if available
            request_name = None
            if context.invocation_metadata():
                for key, value in context.invocation_metadata():
                    if key == "request_name":
                        request_name = value
                        break

            # Process request through orchestrator
            response_content, _ = self.orchestrator.process_request(
                messages=messages,
                project_id=request.project_id,
                request_name=request_name
            )

            return ai_service_pb2.ChatResponse(content=response_content)

        except Exception as e:
            logger.error(f"Error in Chat RPC: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Internal error: {str(e)}")
            return ai_service_pb2.ChatResponse(content="")

    def HealthCheck(self, request: ai_service_pb2.Empty, context) -> ai_service_pb2.HealthCheckResponse:
        """Health check endpoint."""
        try:
            # Add any health checks here (DB connectivity, etc.)
            logger.info("Health check passed")
            return ai_service_pb2.HealthCheckResponse(isAlive=True)
        except Exception as e:
            logger.error(f"Health check failed: {str(e)}")
            return ai_service_pb2.HealthCheckResponse(isAlive=False)
