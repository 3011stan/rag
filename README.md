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
- Demo content seeding command and protected admin seed endpoint.
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

Ingest a PDF, Markdown, or text file:

```bash
curl -X POST http://localhost:8080/rag/ingest \
  -F 'file=@/path/to/document.pdf'
```

```bash
curl -X POST http://localhost:8080/rag/ingest \
  -F 'file=@/path/to/notes.md;type=text/markdown'
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

On a deployed demo, seed the bundled documents through the protected admin endpoint:

```bash
curl -X POST https://YOUR-RENDER-URL/admin/seed-demo \
  -H "Authorization: Bearer $ADMIN_TOKEN"
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
ADMIN_TOKEN=...
ENABLE_PUBLIC_UPLOAD=false
MAX_UPLOAD_BYTES=10485760
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

Never commit real secrets to GitHub. Set `DATABASE_URL`, `GEMINI_API_KEY`, and `ADMIN_TOKEN` in the hosting provider dashboard.

For Supabase on Render, use the Supabase Session pooler connection string instead of the direct database URL. The direct Supabase hostname can resolve to IPv6, which can fail on Render with `network is unreachable`.

The value must include the `postgresql://` prefix:

```env
DATABASE_URL=postgresql://postgres.<PROJECT-REF>:<PASSWORD>@aws-0-<REGION>.pooler.supabase.com:5432/postgres?sslmode=require
```

The app initializes the RAG schema on startup. The database user must be allowed to create the `vector` extension or the extension must already be enabled.

Render Free does not support pre-deploy commands in Blueprint services. After the service is live, call `POST /admin/seed-demo` with `ADMIN_TOKEN` to seed the bundled demo documents.

## Security Defaults

- Public PDF upload is disabled in production with `ENABLE_PUBLIC_UPLOAD=false`.
- Supported protected upload formats are PDF, Markdown, and plain text.
- Protected endpoints accept `Authorization: Bearer <ADMIN_TOKEN>`.
- Uploads are limited by `MAX_UPLOAD_BYTES`, defaulting to 10 MB.
- JSON question payloads are limited and must use `Content-Type: application/json`.
- Basic security headers are set on every response.
- Prompt templates explicitly treat retrieved documents as untrusted content.

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

See [ARCHITECTURE.md](ARCHITECTURE.md) for the current backend package layout.
