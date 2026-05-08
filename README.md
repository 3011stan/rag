# RAG Lab

RAG Lab is a local-first Retrieval-Augmented Generation API built in Go.

The project is being shaped as a Machine Learning Engineering portfolio project. It demonstrates document ingestion, chunking, embeddings, semantic retrieval with PostgreSQL/pgvector, answer generation with sources, local model support, and deploy-ready provider configuration.

## Current Capabilities

- PDF ingestion through `POST /rag/ingest`.
- Question answering through `POST /rag/ask`.
- PostgreSQL + pgvector vector store.
- OpenAI, Ollama, and Gemini provider paths.
- Local/offline development with Ollama.
- Free-tier friendly deploy path with Gemini.
- Automatic schema initialization on startup.
- Demo content seeding command.
- GitHub Actions CI for tests and build.
- Docker runtime for API deployment.

## Provider Strategy

Local development:

```env
AI_PROVIDER=ollama
OLLAMA_BASE_URL=http://localhost:11434
EMBEDDING_MODEL=nomic-embed-text
LLM_MODEL=mistral
EMBEDDING_DIMENSIONS=768
```

Public demo/deploy:

```env
AI_PROVIDER=gemini
GEMINI_API_KEY=your-key
GEMINI_BASE_URL=https://generativelanguage.googleapis.com/v1beta
EMBEDDING_MODEL=gemini-embedding-001
LLM_MODEL=gemini-2.5-flash-lite
EMBEDDING_DIMENSIONS=768
```

Optional OpenAI path:

```env
AI_PROVIDER=openai
OPENAI_API_KEY=your-key
```

## Local Setup

Start PostgreSQL with pgvector:

```bash
docker compose up -d postgres
```

Start Ollama and pull the local models:

```bash
ollama serve
ollama pull nomic-embed-text
ollama pull mistral
```

Create `.env` from the example:

```bash
cp .env.example .env
```

Run tests and build:

```bash
make test
make build
```

Run the API:

```bash
make run
```

## API Examples

Health check:

```bash
curl http://localhost:8080/health
```

Ingest a PDF:

```bash
curl -X POST http://localhost:8080/rag/ingest \
  -F 'file=@/path/to/document.pdf'
```

Ask a question:

```bash
curl -X POST http://localhost:8080/rag/ask \
  -H 'Content-Type: application/json' \
  -d '{"question":"What is the document about?","top_k":5}'
```

Seed demo content:

```bash
make seed-demo
```

## Deployment Plan

The first public deployment is planned around:

- API: Render Free Web Service using Docker.
- Database: Neon or Supabase free tier with pgvector.
- AI provider: Gemini API free tier.
- Demo data: preloaded documents about content production for ML Engineering topics.

Required production environment variables:

```env
DATABASE_URL=postgres://...
AI_PROVIDER=gemini
GEMINI_API_KEY=...
GEMINI_BASE_URL=https://generativelanguage.googleapis.com/v1beta
EMBEDDING_MODEL=gemini-embedding-001
LLM_MODEL=gemini-2.5-flash-lite
EMBEDDING_DIMENSIONS=768
CHUNK_TOKENS=800
OVERLAP_TOKENS=100
TOP_K=5
ENVIRONMENT=production
LOG_LEVEL=info
```

The app initializes the RAG schema on startup. The database user must be allowed to create the `vector` extension or the extension must already be enabled.

## Portfolio Direction

This project is not intended to stop at "chat with PDF".

The roadmap is to evolve it into a RAG engineering lab with:

- retrieval evaluation;
- pipeline observability;
- chunking experiments;
- provider comparisons;
- seeded demo documents;
- deployable public API;
- later, a small UI for the portfolio.

See [ai/PORTFOLIO_ROADMAP.md](ai/PORTFOLIO_ROADMAP.md) for the full plan.
