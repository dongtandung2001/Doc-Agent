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
    mock_store = Mock()
    mock_store.add_texts = Mock(return_value=None)
    with patch.object(vector_store, "_get_store", return_value=mock_store):
        result = vector_store.store_embedding(
            project_id="test_project",
            document_id="test_doc",
            content="This is test content for embedding.",
            metadata={"test": "metadata"},
        )
    assert result is True
    mock_store.add_texts.assert_called_once()
    call_kw = mock_store.add_texts.call_args.kwargs
    assert "texts" in call_kw
    assert len(call_kw["texts"]) >= 1
    assert "metadatas" in call_kw
    assert "ids" in call_kw


def test_retrieve(vector_store):
    """Test retrieving documents from vector database."""
    mock_doc = Mock()
    mock_doc.page_content = "retrieved content"
    mock_doc.metadata = {"project_id": "test_project"}
    mock_store = Mock()
    mock_store.similarity_search_with_score = Mock(return_value=[(mock_doc, 0.5)])
    with patch.object(vector_store, "_get_store", return_value=mock_store):
        results = vector_store.retrieve(
            query="test query",
            project_id="test_project",
            top_k=5,
        )
    assert isinstance(results, list)
    assert len(results) == 1
    assert results[0]["content"] == "retrieved content"
    assert results[0]["score"] == 0.5
    mock_store.similarity_search_with_score.assert_called_once()
