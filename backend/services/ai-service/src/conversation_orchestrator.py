
# conversation_orchestrator.py
from typing import List, Dict, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    from openai.types.chat import ChatCompletion

from src.llm_client import LLMClient
from src.vector_store import VectorStoreManager
from src.logger import setup_logger

logger = setup_logger(__name__)

class ConversationOrchestrator:
    def __init__(self):
        self.llm_client = LLMClient()
        self.vector_store = VectorStoreManager()

    def is_code_related(self, messages: List[Dict[str, str]]) -> bool:
        """Determine if the request is code-related."""
        # Get the last user message
        user_messages = [m for m in messages if m.get("role") == "user"]
        if not user_messages:
            return False

        last_message = user_messages[-1].get("content", "").lower()

        # Keywords that indicate code-related queries
        code_keywords = [
            "code", "function", "class", "method", "api",
            "implementation", "bug", "error", "debug",
            "codebase", "repository", "file", "module"
        ]

        return any(keyword in last_message for keyword in code_keywords)

    def should_use_rag(
            self,
            messages: List[Dict[str, str]],
            request_name: Optional[str] = None
    ) -> bool:
        """Determine if RAG should be used."""
        # Use RAG if code-related or if not doc generation
        is_doc_gen = request_name and "Doc Generating" in request_name
        return self.is_code_related(messages) and not is_doc_gen

    def build_rag_prompt(
            self,
            original_messages: List[Dict[str, str]],
            retrieved_docs: List[Dict]
    ) -> str:
        """Build RAG-aware system prompt."""
        if not retrieved_docs:
            return """You are a helpful AI assistant for this project's documentation.
No project documentation has been indexed yet, so answer from general knowledge.
You can still help with coding concepts, APIs, and general questions. If the user asks about project-specific docs, suggest they add or index documentation for this project."""

        # Format retrieved documentation
        context = "\n\n".join([
            f"[Document { i +1}]\n{doc['content']}"
            for i, doc in enumerate(retrieved_docs)
        ])

        system_prompt = f"""You are a helpful AI assistant with access to documentation context.

RELEVANT DOCUMENTATION:
{context}

Use the above documentation to answer the user's question accurately. If the documentation doesn't contain relevant information, say so and provide your best answer based on general knowledge."""

        return system_prompt

    def process_request(
            self,
            messages: List[Dict[str, str]],
            project_id: str,
            request_name: Optional[str] = None
    ) -> "ChatCompletion":
        """
        Process chat request with RAG orchestration.
        Returns: Full OpenAI ChatCompletion response object
        """
        logger.info(f"Processing request for project: {project_id}")

        # Determine if RAG should be used
        use_rag = self.should_use_rag(messages, request_name)

        retrieved_docs = []
        if use_rag:
            # Get last user message for retrieval
            user_messages = [m for m in messages if m.get("role") == "user"]
            if user_messages:
                query = user_messages[-1].get("content", "")
                retrieved_docs = self.vector_store.retrieve(query, project_id)
                logger.info(f"Retrieved {len(retrieved_docs)} documents for RAG")

        # Build system prompt
        system_prompt = self.build_rag_prompt(messages, retrieved_docs)

        # Generate response - returns full ChatCompletion object
        response = self.llm_client.generate_response(
            messages,
            system_prompt
        )

        # Auto-embed if this is doc generation and no tools were executed
        if request_name and "Doc Generating" in request_name:
            # Check if tools were executed
            tools_executed = False
            if hasattr(response.choices[0].message, 'tool_calls'):
                tools_executed = response.choices[0].message.tool_calls is not None
            
            if not tools_executed:
                logger.info("Auto-embedding triggered for doc generation")
                # Extract content for embedding
                content = response.choices[0].message.content or ""
                self.auto_embed_response(
                    content,
                    project_id,
                    request_name
                )

        return response

    def auto_embed_response(
            self,
            content: str,
            project_id: str,
            request_name: str
    ) -> bool:
        """Automatically embed generated documentation."""
        try:
            # Generate document ID from request name
            document_id = request_name.replace(" ", "_").replace("Doc_Generating_", "")

            # Store embedding
            success = self.vector_store.store_embedding(
                project_id=project_id,
                document_id=document_id,
                content=content,
                metadata={"request_name": request_name}
            )

            if success:
                logger.info(f"Auto-embedded document: {document_id}")

            return success

        except Exception as e:
            logger.error(f"Error in auto-embedding: {str(e)}")
            return False