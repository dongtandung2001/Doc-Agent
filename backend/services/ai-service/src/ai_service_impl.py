
# ai_service_impl.py
import grpc
import sys
import json
from pathlib import Path

# Add parent directory to path for generated imports
sys.path.insert(0, str(Path(__file__).parent.parent))
import generated.ai_service_pb2 as ai_service_pb2
import generated.ai_service_pb2_grpc as ai_service_pb2_grpc

from src.conversation_orchestrator import ConversationOrchestrator
from src.rag_pipeline import RAGPipeline
from src.config import Config
from src.logger import setup_logger

logger = setup_logger(__name__)


class AIServiceServicer(ai_service_pb2_grpc.AIServiceServicer):
    """gRPC service implementation for AI Service."""

    def __init__(self):
        # Single shared RAGPipeline — owns the DB client and Chroma stores.
        # Injected into the orchestrator so both CreateRAG and Chat use the
        # same instance (shared in-process Chroma cache, single DB pool).
        self.rag_pipeline = RAGPipeline()

        # Reset any documents stuck in 'processing' from a previous crashed run.
        try:
            reset = self.rag_pipeline.db.reset_stale_processing()
            if reset:
                logger.warning(f"Startup: reset {reset} stale 'processing' docs to 'pending'")
        except Exception as exc:
            logger.error(f"Startup: could not reset stale docs (DB may be unreachable): {exc}")

        self.orchestrator = ConversationOrchestrator(rag_pipeline=self.rag_pipeline)
        logger.info("AI Service initialized")

    def Chat(self, request: ai_service_pb2.ChatRequest, context) -> ai_service_pb2.ChatResponse:
        """Handle chat requests with RAG support."""
        logger.info(f"=== Chat RPC called === project_id={request.project_id}, messages={len(request.messages)}")
        try:
            if not request.project_id:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("project_id is required")
                return ai_service_pb2.ChatResponse(completion_json="{}", content="")

            if not request.messages:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("messages cannot be empty")
                return ai_service_pb2.ChatResponse(completion_json="{}", content="")

            messages = [
                {"role": msg.role, "content": msg.content}
                for msg in request.messages
            ]

            request_name = None
            if context.invocation_metadata():
                for key, value in context.invocation_metadata():
                    if key == "request_name":
                        request_name = value
                        break

            # process_request now returns a plain string answer from load_qa_chain
            answer = self.orchestrator.process_request(
                messages=messages,
                project_id=request.project_id,
                request_name=request_name,
            )

            logger.info(f"Chat: generated answer ({len(answer)} chars)")

            # Build response — populate both fields for backward compatibility
            response = ai_service_pb2.ChatResponse()
            for msg in request.messages:
                new_msg = response.request.messages.add()
                new_msg.role = msg.role
                new_msg.content = msg.content
            response.request.project_id = request.project_id
            response.content = answer
            response.completion_json = json.dumps({
                "choices": [{"message": {"role": "assistant", "content": answer}}]
            })

            return response

        except Exception as e:
            logger.error(f"Error in Chat RPC: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Internal error: {str(e)}")
            return ai_service_pb2.ChatResponse(completion_json="{}", content="")

    def CreateRAG(self, request: ai_service_pb2.CreateRAGRequest, context) -> ai_service_pb2.CreateRAGResponse:
        """
        Trigger the RAG embedding pipeline for the given project.

        Fetches all document_file_items with embed_status='pending' from
        Postgres, embeds them into the project's Chroma vector store, and
        marks each row as 'completed' (or 'failed'). Safe to call repeatedly.
        """
        logger.info(f"=== CreateRAG RPC called === project_id={request.project_id}")
        try:
            if not request.project_id:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("project_id is required")
                return ai_service_pb2.CreateRAGResponse(success=False, message="project_id is required")

            embedded_count = self.rag_pipeline.process_pending_documents(request.project_id)

            logger.info(f"CreateRAG: embedded {embedded_count} documents for project={request.project_id}")
            return ai_service_pb2.CreateRAGResponse(
                success=True,
                message=f"Embedded {embedded_count} document(s) for project {request.project_id}",
            )

        except Exception as e:
            logger.error(f"Error in CreateRAG RPC: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return ai_service_pb2.CreateRAGResponse(success=False, message=str(e))

    def HealthCheck(self, request: ai_service_pb2.Empty, context) -> ai_service_pb2.HealthCheckResponse:
        """Health check endpoint."""
        try:
            logger.info("Health check passed")
            return ai_service_pb2.HealthCheckResponse(isAlive=True)
        except Exception as e:
            logger.error(f"Health check failed: {str(e)}")
            return ai_service_pb2.HealthCheckResponse(isAlive=False)
