
# ai_service_impl.py
import grpc
import sys
import json
from typing import List, TYPE_CHECKING
from pathlib import Path

# Add parent directory to path for generated imports
sys.path.insert(0, str(Path(__file__).parent.parent))
import generated.ai_service_pb2 as ai_service_pb2
import generated.ai_service_pb2_grpc as ai_service_pb2_grpc

if TYPE_CHECKING:
    from openai.types.chat import ChatCompletion

from src.conversation_orchestrator import ConversationOrchestrator
from src.config import Config
from src.logger import setup_logger

logger = setup_logger(__name__)


def serialize_chat_completion_to_json(completion: "ChatCompletion") -> str:
    """Convert OpenAI ChatCompletion object to JSON string.
    
    OpenAI SDK v1.x ChatCompletion objects are Pydantic models that can be serialized.
    """
    try:
        # OpenAI SDK v1.x uses Pydantic - try model_dump_json() first (most efficient)
        if hasattr(completion, 'model_dump_json') and callable(getattr(completion, 'model_dump_json', None)):
            result = completion.model_dump_json()
            # Ensure it's a string, not a Mock or other object
            if isinstance(result, str):
                return result
        if hasattr(completion, 'model_dump') and callable(getattr(completion, 'model_dump', None)):
            # Pydantic v2 style - convert to dict then JSON
            return json.dumps(completion.model_dump(), default=str)
        if hasattr(completion, 'dict') and callable(getattr(completion, 'dict', None)):
            # Pydantic v1 style
            return json.dumps(completion.dict(), default=str)
        else:
            # Fallback: manually convert to dict (handles all cases)
            completion_dict = {
                "id": getattr(completion, 'id', ''),
                "object": getattr(completion, 'object', 'chat.completion'),
                "created": getattr(completion, 'created', 0),
                "model": getattr(completion, 'model', ''),
                "choices": [
                    {
                        "index": getattr(choice, 'index', 0),
                        "message": {
                            "role": getattr(choice.message, 'role', ''),
                            "content": getattr(choice.message, 'content', None),
                            "tool_calls": [
                                {
                                    "id": getattr(tc, 'id', ''),
                                    "type": getattr(tc, 'type', ''),
                                    "function": {
                                        "name": getattr(tc.function, 'name', ''),
                                        "arguments": getattr(tc.function, 'arguments', '')
                                    } if hasattr(tc, 'function') and getattr(tc, 'function', None) else None
                                }
                                for tc in (getattr(choice.message, 'tool_calls', None) or [])
                            ] if hasattr(choice.message, 'tool_calls') and getattr(choice.message, 'tool_calls', None) else None,
                            "refusal": getattr(choice.message, 'refusal', None)
                        },
                        "finish_reason": getattr(choice, 'finish_reason', None)
                    }
                    for choice in getattr(completion, 'choices', [])
                ],
                "usage": {
                    "prompt_tokens": getattr(completion.usage, 'prompt_tokens', 0),
                    "completion_tokens": getattr(completion.usage, 'completion_tokens', 0),
                    "total_tokens": getattr(completion.usage, 'total_tokens', 0)
                } if hasattr(completion, 'usage') and getattr(completion, 'usage', None) else None,
                "system_fingerprint": getattr(completion, 'system_fingerprint', None)
            }
            return json.dumps(completion_dict, default=str)
    except Exception as e:
        logger.error(f"Error serializing ChatCompletion to JSON: {str(e)}")
        import traceback
        logger.error(traceback.format_exc())
        # Fallback: return minimal JSON
        return json.dumps({
            "id": getattr(completion, 'id', ''),
            "model": getattr(completion, 'model', ''),
            "error": f"Serialization error: {str(e)}"
        }, default=str)


class AIServiceServicer(ai_service_pb2_grpc.AIServiceServicer):
    """gRPC service implementation for AI Service."""

    def __init__(self):
        self.orchestrator = ConversationOrchestrator()
        logger.info("AI Service initialized")

    def Chat(self, request: ai_service_pb2.ChatRequest, context) -> ai_service_pb2.ChatResponse:
        """Handle chat requests with RAG support."""
        logger.info(f"=== Chat RPC called ===")
        logger.info(f"Project ID: {request.project_id}, Messages: {len(request.messages)}")
        try:
            # Validate request
            if not request.project_id:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("project_id is required")
                return ai_service_pb2.ChatResponse(
                    request=request,
                    completion_json="{}",
                    content=""
                )

            if not request.messages:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("messages cannot be empty")
                return ai_service_pb2.ChatResponse(
                    request=request,
                    completion_json="{}",
                    content=""
                )

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

            # Process request through orchestrator - returns full OpenAI response
            completion = self.orchestrator.process_request(
                messages=messages,
                project_id=request.project_id,
                request_name=request_name
            )

            # Serialize ChatCompletion to JSON (much simpler!)
            try:
                logger.info(f"Serializing ChatCompletion: id={completion.id}, model={completion.model}")
                completion_json = serialize_chat_completion_to_json(completion)
                logger.info(f"Serialization successful: {len(completion_json)} chars")
            except Exception as serialize_error:
                logger.error(f"Error serializing ChatCompletion to JSON: {str(serialize_error)}")
                import traceback
                logger.error(traceback.format_exc())
                raise
            
            # Extract content for backward compatibility
            content = completion.choices[0].message.content or ""
            
            # Build response with full ChatCompletion (as JSON) and original request
            try:
                response = ai_service_pb2.ChatResponse()
                # Include original request
                for msg in request.messages:
                    new_msg = response.request.messages.add()
                    new_msg.role = msg.role
                    new_msg.content = msg.content
                response.request.project_id = request.project_id
                # Full ChatCompletion as JSON string
                response.completion_json = completion_json
                # Convenience field for backward compatibility
                response.content = content
                logger.info(f"Successfully created response with completion ID: {completion.id}")
                logger.debug(f"Response fields - request: {response.HasField('request')}, completion_json length: {len(completion_json)}, content length: {len(content)}")
            except Exception as response_error:
                logger.error(f"Error creating ChatResponse: {str(response_error)}")
                import traceback
                logger.error(traceback.format_exc())
                raise
            
            return response

        except Exception as e:
            logger.error(f"Error in Chat RPC: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Internal error: {str(e)}")
            return ai_service_pb2.ChatResponse(
                request=request if 'request' in locals() else ai_service_pb2.ChatRequest(),
                completion_json="{}",
                content=""
            )

    def CreateRAG(self, request: ai_service_pb2.CreateRAGRequest, context) -> ai_service_pb2.CreateRAGResponse:
        """Create or ensure RAG index exists for the given project. Idempotent."""
        logger.info(f"=== CreateRAG RPC called === project_id={request.project_id}")
        try:
            if not request.project_id:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("project_id is required")
                return ai_service_pb2.CreateRAGResponse(success=False, message="project_id is required")

            success = self.orchestrator.vector_store.ensure_rag_index(request.project_id)
            if success:
                logger.info(f"RAG index ready for project: {request.project_id}")
                return ai_service_pb2.CreateRAGResponse(
                    success=True,
                    message=f"RAG index created or already exists for project {request.project_id}"
                )
            return ai_service_pb2.CreateRAGResponse(
                success=False,
                message=f"Failed to create RAG index for project {request.project_id}"
            )
        except Exception as e:
            logger.error(f"Error in CreateRAG RPC: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return ai_service_pb2.CreateRAGResponse(success=False, message=str(e))

    def HealthCheck(self, request: ai_service_pb2.Empty, context) -> ai_service_pb2.HealthCheckResponse:
        """Health check endpoint."""
        try:
            # Add any health checks here (DB connectivity, etc.)
            logger.info("Health check passed")
            return ai_service_pb2.HealthCheckResponse(isAlive=True)
        except Exception as e:
            logger.error(f"Health check failed: {str(e)}")
            return ai_service_pb2.HealthCheckResponse(isAlive=False)
