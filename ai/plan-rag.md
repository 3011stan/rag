# Plano de implementação — RAG em Go

Resumo rápido

- Objetivo: implementar um Retrieval-Augmented Generation (RAG) backend escrito em Go.
- Arquitetura mínima: Ingest (ETL) → Vector DB (Postgres + pgvector) → API de consulta (embed pergunta → busca vetorial → prompt → LLM).

1) Componentes e responsabilidades

- Document Loader: importa PDFs, docs, banco, logs e extrai texto (usando bibliotecas externas para PDF/Office).
- Chunker: quebra texto em pedaços (chunkSize ~ 700–1000 tokens, overlap ~ 50–150 tokens).
- Embeddings Provider: interface que chama modelo de embeddings (ex.: OpenAI via go-openai).
- Vector Store: Postgres+pgvector com repository Go (pgx + pgvector-go) para inserir e buscar vetores.
- Retriever: busca top-K por similaridade, aplica filtros por metadata e reranking opcional.
- Q&A Service: monta prompt com contexto e chama LLM (OpenAI / outro) para gerar resposta.
- API HTTP: endpoints /rag/ingest e /rag/ask.

2) Tecnologias e libs recomendadas (Go)

- Router / API: net/http + github.com/go-chi/chi/v5
- HTTP client OpenAI: github.com/sashabaranov/go-openai (ou openai-go oficial)
- Postgres driver: github.com/jackc/pgx/v5
- pgvector for Go: github.com/pgvector/pgvector-go/pgx
- PDF/text: github.com/ledongthuc/pdf (ou Apache Tika via serviço)
- Config/env: github.com/joho/godotenv ou viper
- Logging/metrics: zerolog / prometheus client

3) Esquema de dados (Postgres + pgvector)

CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE rag_chunks (
  id UUID PRIMARY KEY,
  content TEXT NOT NULL,
  metadata JSONB,
  embedding vector(1536),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

Busca (similaridade):

SELECT id, content, metadata, embedding <-> $1 AS distance
FROM rag_chunks
ORDER BY embedding <-> $1
LIMIT $2;

4) Fluxo de ingestão (pipeline)

- loadDocuments(paths...)
- extractText(document) → rawText
- chunks := splitIntoChunks(rawText, chunkSize, overlap)
- Batch embeddings por N (ex.: 64) para reduzir latência/custos
- Insert em vector store: id, content, metadata, embedding
- Indexação e monitoramento de erros

Boas práticas: paralelizar I/O, respeitar rate limits do provedor, retries exponenciais, idempotência (usar document hash para detectar reingest).

5) Fluxo de consulta

- Receber POST /rag/ask { question }
- Gerar embedding da pergunta
- Buscar top-K (ex.: K=5) no Postgres
- Concatenar contextos (ordenar por score) cuidando do tamanho máximo do prompt
- Prompt template que força a resposta a usar somente o contexto; fallback: "Não encontrei informação suficiente nos documentos consultados."
- Chamar LLM com o prompt e retornar answer + sources (document, page, score)

6) Projetos/estruturas de arquivo sugeridas

cmd/server/main.go
internal/
  rag/
    ingest.go
    chunker.go
    embeddings.go (interface + impl OpenAI)
    vectorstore.go (interface + pgx impl)
    qa.go
pkg/api/handlers.go
configs/

Interfaces claras permitem trocar vector DB (Qdrant) ou provider de embeddings.

7) Observabilidade, testes e melhorias

- Métricas: latência de ingest, tamanho do índice, queries por segundo, taxa de erros.
- Testes: unitários (chunking, prompt builder), integração (inserir/recuperar do PG).
- Melhorias posteriores: re-ranking (cross-encoder), hybrid search (BM25 + vetores), filtros por metadata, caching de embeddings/perguntas frequentes, controle de custo (limitar tokens/contextos).

8) Docker / infra mínima

- docker-compose: postgres (com extensão pgvector inicializada), app service.
- Variáveis de ambiente: DATABASE_URL, OPENAI_API_KEY, CHUNK_SIZE, OVERLAP, TOP_K.

9) Exemplo rápido de prompt (template)

Você é um assistente técnico.
Responda usando apenas o contexto abaixo.
Se a resposta não estiver no contexto, responda: "Não encontrei informação suficiente nos documentos consultados."

Contexto:
${context}

Pergunta:
${question}

## Plano detalhado para avaliação

Abaixo está a versão ampliada do plano, com decisões técnicas sugeridas (versão do Go e por quê), lista de libs com propósito, módulos do projeto, parâmetros do pipeline e itens que precisará aprovar antes da implementação.

1) Versão do Go

- Recomenda-se: Go 1.21 (compatível com 1.20+). Motivos:
  - Suporte moderno a generics, melhorias de desempenho e correções de GC.
  - Ferramentas e imagens oficiais atualizadas (Docker/CI).
  - Futuro-proof para bibliotecas mais novas.

2) Dependências / bibliotecas essenciais (com propósito)

- github.com/go-chi/chi/v5 — router HTTP simples e leve.
- github.com/sashabaranov/go-openai — cliente OpenAI (embeddings + completions). Alternativa: openai-go oficial.
- github.com/jackc/pgx/v5 — driver Postgres performático.
- github.com/pgvector/pgvector-go/pgx — integração pgvector com pgx.
- github.com/pkoukk/tiktoken-go — tokenização compatível com tiktoken (chunking por tokens).
- github.com/ledongthuc/pdf — extração de texto de PDFs (simples, sem custos comerciais).
- github.com/rs/zerolog — logs estruturados e performáticos.
- github.com/spf13/viper — configuração via env/files.
- github.com/ory/dockertest/v3 — testes de integração com containers.
- github.com/prometheus/client_golang e go.opentelemetry.io/otel — métricas e tracing.
- golang.org/x/time/rate — rate-limiting local.

3) Módulos / pacotes do projeto (mapa e responsabilidades)

- cmd/server
  - main.go — bootstrap (config, logger, db, router)
- internal/config
  - config.go — load e validação de env
- internal/db
  - migrations.go / conn.go — conexão e migrações (golang-migrate)
- internal/rag/loader
  - pdf_loader.go, txt_loader.go, db_loader.go — extrair texto e metadata
- internal/rag/chunker
  - chunker.go — tokenizer-based chunk split, overlap, heurísticas
- internal/rag/embeddings
  - embeddings.go (interface)
  - openai_embeddings.go (impl) — batching, retries, caching hooks
- internal/rag/vectorstore
  - vectorstore.go (interface)
  - pgvector_store.go (pgx impl) — InsertBatch, Search
- internal/rag/retriever
  - retriever.go — top-K search + optional reranker
- internal/rag/qa
  - prompt_builder.go, llm_client.go — chama LLM e formata resposta
- internal/api
  - handlers.go — /rag/ingest, /rag/ask
- internal/worker
  - ingest_worker.go — worker pool para ingest
- internal/telemetry
  - metrics.go, tracing.go

4) Schema de banco (detalhado)

- Extensão: CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS rag_documents (
  id UUID PRIMARY KEY,
  source TEXT,
  title TEXT,
  checksum TEXT UNIQUE,
  metadata JSONB,
  created_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rag_chunks (
  id UUID PRIMARY KEY,
  document_id UUID REFERENCES rag_documents(id) ON DELETE CASCADE,
  chunk_index INT NOT NULL,
  content TEXT NOT NULL,
  token_count INT,
  metadata JSONB,
  embedding vector(1536),
  created_at timestamptz DEFAULT now()
);

Indexes sugeridos:
- ivfflat/GiST index em embedding (configurável)
- index em (document_id)
- index em (metadata->>'type') se necessário

5) Parâmetros do chunking (configuráveis)

- Unidade primária: tokens (usar tiktoken-go)
- Valores iniciais: chunk_tokens = 800, overlap_tokens = 100
- Fallback char-based: chunk_chars = 3000, overlap_chars = 300
- Estratégia: quebrar por sentenças/parágrafos, preservar tabelas/códigos como blocos inteiros quando possível.

6) Embeddings (pipeline e políticas)

- Modelo sugerido: text-embedding-3-small (ou equivalente) — dimensão configurável (ex.: 1536).
- Batch size: 64 (configurável), start com 32 se limits forem rígidos.
- Concurrency: worker pool com rate limiter (respeitar quotas OpenAI).
- Idempotência: calcular checksum do documento; se checksum igual, pular reingest ou atualizar incrementalmente.
- Cache opcional: Redis para embeddings reusados.
- Retries: exponential backoff com jitter, maxRetries = 3.

7) Fluxo de ingestão (detalhado tecnicamente)

- Recebe request com source (local path, s3, db query) ou job via fila.
- Extrai texto por página (PDF) e metadata (page number, author, date).
- Normaliza texto (unicode, remoção de headers/footers, trim, dedupe linhas repetidas).
- Tokeniza e gera chunks mantendo overlap.
- Para cada batch de chunks:
  - calcular embedding via embeddings provider (batch)
  - montar payload para inserção
  - inserir via InsertBatch em transação (ou upsert com chunk checksum)
- Emitir métricas: chunks_created, embed_calls, ingest_duration, failed_batches.

8) Fluxo de consulta (detalhado tecnicamente)

- Endpoint: POST /rag/ask
  - payload: { question: string, top_k?: int, filters?: object, model?: string }
- Steps:
  1) gerar embedding da pergunta
  2) executar Search no vectorstore com top_k (default 5)
  3) opcional: rerank por heurística ou cross-encoder
  4) montar contexto ordenado por score, truncando por token limit do modelo de geração
  5) construir prompt com template estrito (usar instrução para não inventar)
  6) chamar LLM para gerar resposta
  7) retornar resposta + sources (document_id, chunk_index, page se presente, score)

Exemplo de resposta JSON:

{
  "answer": "...",
  "sources": [
    {"document_id": "...","chunk_index": 2, "score": 0.12}
  ],
  "model": "gpt-4o-mini",
  "used_context_tokens": 1024
}

9) API e DTOs (exemplo rápido em Go)

- Request /rag/ingest: IngestRequest { Source string, Path string, DocumentID *string }
- Response: { job_id string }
- Request /rag/ask: AskRequest { Question string, TopK *int, Filters map[string]interface{} }
- Response: AskResponse { Answer string, Sources []Source, Model string }

(Se desejar eu gero os structs Go e OpenAPI spec.)

10) Testes e CI

- Unit tests: chunker logic, tokenizer, prompt builder, DTOs.
- Integration tests: com dockertest subir Postgres (+pgvector), rodar ingest/search.
- Mocks: provider OpenAI mock server para E2E em CI.
- Linters: golangci-lint no CI; staticcheck.
- GitHub Actions: jobs para test, lint, build; job opcional de integration que usa services.

11) Observabilidade, segurança e operações

- Métricas (Prometheus): ingest_rate, embed_calls, search_latency, errors_total.
- Tracing (OpenTelemetry): instrumentar chamadas externas (OpenAI), DB e handlers.
- Logs em JSON com request_id, nível por ambiente.
- Secrets: passar OPENAI_API_KEY e DATABASE_URL via secrets do ambiente/CI.
- Rate limiting por API key / IP no /rag/ask.
- Política de retenção: TTL para chunks antigos (configurável).

12) Infra mínima (docker-compose)

- postgres: imagem com pgvector (ou init script para instalar extension)
- app: imagem Go construída via Dockerfile multistage
- opcional: redis, qdrant

13) Migração inicial SQL (arquivo sql/migrations/0001_create_tables.sql)

-- migration: create documents and chunks

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE rag_documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source TEXT,
  title TEXT,
  checksum TEXT UNIQUE,
  metadata JSONB,
  created_at timestamptz DEFAULT now()
);

CREATE TABLE rag_chunks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  document_id UUID REFERENCES rag_documents(id) ON DELETE CASCADE,
  chunk_index INT NOT NULL,
  content TEXT NOT NULL,
  token_count INT,
  metadata JSONB,
  embedding vector(1536),
  created_at timestamptz DEFAULT now()
);

14) Cronograma detalhado (2 semanas, MVP)

- Semana 1:
  - Dia 1: scaffold, go.mod (Go 1.21), Docker Compose, PG+pgvector
  - Dia 2: DB layer + migrations + models
  - Dia 3: chunker + tokenizer + unit tests
  - Dia 4: loaders (PDF/txt) + normalização
  - Dia 5: embeddings adapter (mock + OpenAI)
- Semana 2:
  - Dia 6: vectorstore impl (insert/search) + tests
  - Dia 7: API endpoints /rag/ingest + /rag/ask
  - Dia 8: prompt builder + LLM client
  - Dia 9: integration tests, metrics, CI
  - Dia 10: docs, ajustes, buffer de contingência

15) Decisões pendentes (o que você deve aprovar)

- Versão exata do Go (1.21 recomendado)
- Provider de embeddings/LLM (OpenAI por padrão) — quer considerar local models? 
- Vector DB inicial: Postgres+pgvector (recomendado) vs Qdrant desde o início
- Política de retenção/privacidade (tempo de armazenamento)
- Nível de testes de integração obrigatórios no CI

### Decisão sobre índices no banco

Para o MVP, será utilizado o índice **GiST (Generalized Search Tree)** para busca vetorial. Este índice é mais simples de configurar e suficiente para o volume inicial de dados. Caso o volume cresça significativamente, podemos migrar para IVFFlat no futuro.

Exemplo de criação do índice GiST:

```sql
CREATE INDEX ON rag_chunks USING gist (embedding);
```

### Atualizações finais

- **Versão do Go**: 1.21.
- **Provider de embeddings/LLM**: OpenAI.
- **Vector DB**: Postgres + pgvector.
- **Ingestão**: Via API, upload de PDFs.
- **Formatos suportados**: Apenas PDFs.
- **Política de reingestão**: Sobrescrever documentos duplicados.
- **Chunking/embeddings**:
  - Parâmetros: `chunk_tokens = 800`, `overlap_tokens = 100`.
  - Dimensão do embedding: 1536.
- **Banco de dados**:
  - Campos obrigatórios: `document_id`, `source`, `page`, `author`, `checksum`, `created_at`.
  - Índice: GiST para busca vetorial.
- **Consulta**:
  - `top_k = 5`.
  - Sem reranker (similarity search simples).
  - Resposta inclui `sources` (document_id, page, chunk_index, score).
- **API**:
  - Sem autenticação.
  - Sem rate limit.
  - Expor `prompt_used` na resposta.
- **Deploy**: Docker Compose.
- **Configuração**: Variáveis de ambiente (env vars).
- **Logs**: Básicos.
- **Testes/CI**: Não necessários.
- **Otimização de chamadas**: Básica.
- **Ingestão**: Assíncrona (sem bloqueio).
- **Endpoint para deletar**: Não necessário.
- **Critérios de aceitação**:
  - Ingestão de PDFs funcionando.
  - Busca top-k retornando resultados relevantes.
  - API `/rag/ingest` e `/rag/ask` funcionando.
  - Logs básicos operacionais.

Próximos passos: Gerar scaffold do projeto, SQL de migração e handlers iniciais.

### Teste de Conexão ao Banco de Dados

- O arquivo `internal/rag/vectorstore.go` implementa a interface `VectorStore`.
- A função `TestConnection` foi adicionada para verificar a conectividade com o banco de dados.
- Padrão:
  - Usar `pgx` como driver PostgreSQL.
  - Implementar testes de conectividade no `vectorstore` para garantir que o banco está acessível antes de operações críticas.