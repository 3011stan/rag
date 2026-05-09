# Architecture

The backend is organized around small packages with explicit responsibilities.

## Request Flow

```text
HTTP handler -> RAG pipeline -> retriever/vector store/AI providers -> HTTP response
```

Handlers are responsible for HTTP concerns only: authentication, request limits, validation, logging, and response mapping.

The RAG pipeline coordinates application behavior:

- PDF ingestion
- chunking
- embedding generation
- vector storage
- retrieval
- answer generation

## Package Layout

- `internal/api`: HTTP contracts, handlers, auth, validation, errors, and middleware.
- `internal/ai`: provider factory for embedding and answering services.
- `internal/rag`: domain models and PostgreSQL/pgvector storage.
- `internal/rag/answering`: prompt construction and LLM-backed answer generation.
- `internal/rag/embeddings`: embedding provider implementations.
- `internal/rag/loader`: document loader strategies for PDF, Markdown, and plain text.
- `internal/rag/pipeline`: application service that coordinates RAG workflows.
- `internal/rag/retriever`: semantic retrieval over the vector store.
- `internal/demo/seed`: curated demo document seeding.

## Code Style

Comments are reserved for exported API documentation and non-obvious decisions. Implementation code should be readable through names, package boundaries, and small functions.
