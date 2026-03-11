# conversation_orchestrator.py
from typing import List, Dict, Optional

from langchain.chains.question_answering import load_qa_chain
from langchain.prompts import PromptTemplate

from src.rag_pipeline import RAGPipeline
from src.logger import setup_logger

logger = setup_logger(__name__)

_PROMPT_TEMPLATE = """
Answer the question as detailed as possible using the provided context.
If the answer is not in the provided context, say "I don't have enough context to answer this" \
and answer from general knowledge if you can.

Context:
{context}

Question:
{question}

Answer:
"""


class ConversationOrchestrator:
    def __init__(self, rag_pipeline: RAGPipeline):
        self.rag_pipeline = rag_pipeline
        self._prompt = PromptTemplate(
            template=_PROMPT_TEMPLATE,
            input_variables=["context", "question"],
        )

    def process_request(
            self,
            messages: List[Dict[str, str]],
            project_id: str,
            request_name: Optional[str] = None,
    ) -> str:
        """
        Process a chat request with RAG on every call.

        Uses load_qa_chain (stuff type) — retrieves relevant docs from Chroma,
        stuffs them into the prompt, then generates an answer via the LLM.

        Returns the answer as a plain string.
        """
        user_messages = [m for m in messages if m.get("role") == "user"]
        if not user_messages:
            return "No question provided."

        question = user_messages[-1].get("content", "")
        logger.info(f"Processing request for project={project_id} question='{question[:80]}'")

        # Retrieve relevant chunks from the vector store.
        # If the collection is empty (CreateRAG not yet called) or the search
        # fails for any reason, fall back to empty docs so the chain still runs
        # and answers from general knowledge instead of crashing.
        docs = []
        try:
            store = self.rag_pipeline._get_store(project_id)
            docs = store.similarity_search(question, k=5, filter={"project_id": project_id})
        except Exception as e:
            logger.warning(f"similarity_search failed (collection may be empty): {e}")
        logger.info(f"Retrieved {len(docs)} docs for RAG")

        # Build and run the QA chain (same pattern as load_qa_chain "stuff")
        chain = load_qa_chain(
            self.rag_pipeline.llm,
            chain_type="stuff",
            prompt=self._prompt,
        )

        response = chain.invoke(
            {"input_documents": docs, "question": question},
            return_only_outputs=True,
        )

        answer = response.get("output_text", "")
        logger.info(f"Generated answer ({len(answer)} chars)")
        return answer
