# tests/test_vector_store.py
import sys
from pathlib import Path
import pytest
from unittest.mock import Mock, patch

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))
from src.vector_store import VectorStoreManager

@pytest.fixture
def vector_store():
    return VectorStoreManager()

def test_store_embedding(vector_store):
    """Test storing embeddings in vector database."""
    result = vector_store.store_embedding(
        project_id="test_project",
        document_id="test_doc",
        content="This is test content for embedding.",
        metadata={"test": "metadata"}
    )
    assert isinstance(result, bool)

def test_retrieve(vector_store):
    """Test retrieving documents from vector database."""
    results = vector_store.retrieve(
        query="test query",
        project_id="test_project",
        top_k=5
    )
    assert isinstance(results, list)
