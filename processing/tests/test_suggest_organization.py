"""Tests for the non-destructive folder/filename suggestion write path.

Covers PGVectorRepository.suggest_organization (the DB write) and the
classifier.suggest_folder_path helper that produces the "/"-joined nested
path stored in documents.suggested_folder_name.
"""

from types import SimpleNamespace
from unittest.mock import AsyncMock, Mock

import pytest

from processing import classifier
from processing.database import Document, PGVectorRepository


def _make_repo(session: Mock, leaf_exists: bool) -> PGVectorRepository:
    """Build a repo without touching a real engine.

    Bypasses __init__ (which would create a SQLAlchemy engine) and stubs the
    two collaborators suggest_organization relies on: get_session and the
    async leaf_folder_exists lookup.
    """
    repo = PGVectorRepository.__new__(PGVectorRepository)
    repo.get_session = Mock(return_value=session)
    repo.leaf_folder_exists = AsyncMock(return_value=leaf_exists)
    return repo


def _make_session(doc: Document | None) -> Mock:
    session = Mock()
    query = Mock()
    query.filter.return_value = query
    query.first.return_value = doc
    session.query.return_value = query
    return session


def _make_doc() -> Document:
    """A Document row in its pre-suggestion state.

    folder_id and title are pre-set so the test can assert they are not
    mutated by the suggestion write.
    """
    return Document(
        id="doc-1",
        tenant_id="tenant-1",
        org_id="org-1",
        folder_id="folder-original",
        owner_id="owner-1",
        title="scan-0001.pdf",
        status="processing",
    )


@pytest.mark.asyncio
async def test_suggest_organization_writes_suggestion_columns_and_stage() -> None:
    doc = _make_doc()
    session = _make_session(doc)
    repo = _make_repo(session, leaf_exists=False)

    await repo.suggest_organization(
        document_id="doc-1",
        tenant_id="tenant-1",
        org_id="org-1",
        suggested_folder_name="Clients/Acme/Invoices",
        suggested_filename="Invoice INV-123 Acme.pdf",
        suggestion_confidence=0.9,
    )

    # Suggestion columns are written, path stored as the "/"-joined nested path.
    assert doc.suggested_folder_name == "Clients/Acme/Invoices"
    assert doc.suggested_filename == "Invoice INV-123 Acme.pdf"
    assert doc.suggestion_confidence == 0.9
    # Leaf ("Invoices") did not exist -> accepting requires creating it.
    assert doc.suggestion_create_new is True
    # Stage advances to "suggesting".
    assert doc.processing_stage == "suggesting"
    # Non-destructive: folder_id and title are untouched.
    assert doc.folder_id == "folder-original"
    assert doc.title == "scan-0001.pdf"

    session.commit.assert_called_once()
    session.rollback.assert_not_called()
    # create_new decision derives from the leaf segment of the nested path.
    repo.leaf_folder_exists.assert_awaited_once_with("tenant-1", "org-1", "Invoices")


@pytest.mark.asyncio
async def test_suggest_organization_normalizes_nested_path_segments() -> None:
    """Per the contract: split on "/", strip each segment, drop empties."""
    doc = _make_doc()
    session = _make_session(doc)
    repo = _make_repo(session, leaf_exists=False)

    await repo.suggest_organization(
        document_id="doc-1",
        tenant_id="tenant-1",
        org_id="org-1",
        suggested_folder_name="  Clients / / Acme /  Invoices  /",
        suggested_filename="file.pdf",
        suggestion_confidence=0.7,
    )

    assert doc.suggested_folder_name == "Clients/Acme/Invoices"
    # Leaf is the last surviving (stripped, non-empty) segment.
    repo.leaf_folder_exists.assert_awaited_once_with("tenant-1", "org-1", "Invoices")


@pytest.mark.asyncio
async def test_suggest_organization_create_new_false_when_leaf_exists() -> None:
    doc = _make_doc()
    session = _make_session(doc)
    repo = _make_repo(session, leaf_exists=True)

    await repo.suggest_organization(
        document_id="doc-1",
        tenant_id="tenant-1",
        org_id="org-1",
        suggested_folder_name="Invoices/Acme",
        suggested_filename="file.pdf",
        suggestion_confidence=0.9,
    )

    assert doc.suggestion_create_new is False
    assert doc.processing_stage == "suggesting"


@pytest.mark.asyncio
async def test_suggest_organization_missing_document_is_a_noop() -> None:
    session = _make_session(doc=None)
    repo = _make_repo(session, leaf_exists=False)

    await repo.suggest_organization(
        document_id="missing",
        tenant_id="tenant-1",
        org_id="org-1",
        suggested_folder_name="Invoices",
        suggested_filename="file.pdf",
        suggestion_confidence=0.7,
    )

    session.commit.assert_not_called()


@pytest.mark.parametrize(
    ("metadata", "expected"),
    [
        # invoice nests under "Invoices/{issuer}"; path is "/"-joined.
        ({"doc_type": "invoice", "issuer": "Acme Corp"}, "Invoices/Acme Corp"),
        ({"doc_type": "invoice"}, "Invoices"),
        # contract nests under "Contracts/{issuer}".
        ({"doc_type": "contract", "issuer": "Globex"}, "Contracts/Globex"),
        ({"doc_type": "contract"}, "Contracts"),
        # identity is a fixed single-segment path (no issuer nesting).
        ({"doc_type": "identity", "issuer": "Passport Office"}, "Identity"),
        # other / unknown falls back to "Documents".
        ({"doc_type": "other", "issuer": "Acme"}, "Documents"),
        ({}, "Documents"),
    ],
)
def test_suggest_folder_path_doc_types(metadata, expected) -> None:
    path = classifier.suggest_folder_path(metadata)
    assert path == expected
    # Every produced segment is non-empty and free of stray separators.
    segments = path.split("/")
    assert all(seg.strip() == seg and seg for seg in segments)


def test_suggest_folder_path_returns_slash_joined_string() -> None:
    """suggest_folder_path output is a SimpleNamespace-free plain string."""
    path = classifier.suggest_folder_path({"doc_type": "invoice", "issuer": "Acme"})
    assert isinstance(path, str)
    assert path == "Invoices/Acme"
    assert not isinstance(path, SimpleNamespace)
