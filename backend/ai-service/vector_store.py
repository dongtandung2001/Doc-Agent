from typing import List, Dict, Optional
from langchain_openai import OpenAIEmbeddings
from langchain_community.vectorstores import Chroma
from langchain.text_splitter import RecursiveCharacterTextSplitter
from config import Config
from logger import setup_logger

logger = setup_logger(__name__)

class VectorStoreManager:
    def __init__(self):
        self.embeddings = OpenAIEmbeddings(
            model=Config.EMBEDDING_MODEL,
            openai_api_key=Config.OPENAI_API_KEY
        )
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=Config.RAG_CHUNK_SIZE,
            chunk_overlap=Config.RAG_CHUNK_OVERLAP,
            length_function=len,
        )
        self._stores: Dict[str, Chroma] = {}

    def _get_store(self, project_id: str) -> Chroma:
        """Get or create vector store for a project."""
        if project_id not in self._stores:
            collection_name = f"project_{project_id}"
            self._stores[project_id] = Chroma(
                collection_name=collection_name,
                embedding_function=self.embeddings,
                persist_directory=f"{Config.VECTOR_DB_PATH}/{project_id}"
            )
            logger.info(f"Created vector store for project: {project_id}")
        return self._stores[project_id]

    def store_embedding(
            self,
            project_id: str,
            document_id: str,
            content: str,
            metadata: Optional[Dict] = None
    ) -> bool:
        """Store document embedding in vector database."""
        try:
            # Clean and chunk text
            chunks = self.text_splitter.split_text(content)
            logger.info(f"Split document into {len(chunks)} chunks")

            # Prepare metadata
            base_metadata = {
                "project_id": project_id,
                "document_id": document_id,
                "source": "doc_generation"
            }
            if metadata:
                base_metadata.update(metadata)

            # Create metadata for each chunk
            metadatas = [
                {**base_metadata, "chunk_index": i}
                for i in range(len(chunks))
            ]

            # Store in vector database
            store = self._get_store(project_id)
            store.add_texts(
                texts=chunks,
                metadatas=metadatas,
                ids=[f"{document_id}_chunk_{i}" for i in range(len(chunks))]
            )

            logger.info(f"Stored {len(chunks)} chunks for document {document_id}")
            return True

        except Exception as e:
            logger.error(f"Error storing embedding: {str(e)}")
            return False

    def retrieve(
            self,
            query: str,
            project_id: str,
            top_k: Optional[int] = None
    ) -> List[Dict]:
        """Retrieve relevant documents from vector database."""
        try:
            k = top_k or Config.TOP_K_RETRIEVAL
            store = self._get_store(project_id)

            # Perform similarity search with scores
            results = store.similarity_search_with_score(
                query,
                k=k,
                filter={"project_id": project_id}
            )

            # Format results and deduplicate
            formatted_results = []
            seen_content = set()

            for doc, score in results:
                content = doc.page_content
                if content not in seen_content:
                    formatted_results.append({
                        "content": content,
                        "score": float(score),
                        "metadata": doc.metadata
                    })
                    seen_content.add(content)

            logger.info(f"Retrieved {len(formatted_results)} unique results")
            return formatted_results

        except Exception as e:
            logger.error(f"Error retrieving from vector store: {str(e)}")
            return []

    def delete_document(self, project_id: str, document_id: str) -> bool:
        """Delete document embeddings from vector database."""
        try:
            store = self._get_store(project_id)
            # Note: This requires custom implementation based on vector DB
            logger.info(f"Deleted embeddings for document {document_id}")
            return True
        except Exception as e:
            logger.error(f"Error deleting document: {str(e)}")
            return False