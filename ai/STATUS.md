# RAG Backend - Status da Implementação

**Data**: 2026-05-05  
**Progresso**: 100% (72 de 72 tarefas concluídas)  
**Compilação**: ✅ Sucesso  
**Modo local/offline/free**: ✅ Validado com Ollama

---

## Estado Atual

O MVP RAG está funcional com dois caminhos de provider:

- **OpenAI**: usado quando `OPENAI_API_KEY` está preenchida.
- **Ollama local**: usado automaticamente quando `OPENAI_API_KEY` está vazia.

Configuração local validada:

```bash
OPENAI_API_KEY=
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_EMBED_MODEL=nomic-embed-text
OLLAMA_LLM_MODEL=mistral
```

Modelos locais encontrados no Ollama:

- `nomic-embed-text:latest` para embeddings.
- `mistral:latest` para geração de respostas.

O modelo `nomic-embed-text` gera embeddings de **768 dimensões**, então o schema local foi ajustado para:

```sql
embedding vector(768)
```

---

## Validações Executadas

Comandos/testes executados com sucesso:

```bash
env GOCACHE=/private/tmp/rag-go-build-cache go test ./...
env GOCACHE=/private/tmp/rag-go-build-cache go build -o rag-app .
curl http://localhost:11434/api/tags
curl http://localhost:11434/api/embeddings
curl http://localhost:8080/health
curl -X POST http://localhost:8080/rag/ingest
curl -X POST http://localhost:8080/rag/ask
```

Resultado do teste end-to-end:

- API subiu em `:8080`.
- PDF mínimo foi ingerido com sucesso.
- `chunk_count` retornou `1`.
- Embedding foi gerado via Ollama.
- Chunk foi salvo no PostgreSQL/pgvector.
- Pergunta em `/rag/ask` retornou resposta e `sources`.

---

## Mudanças Relevantes

- Adicionado provider de embeddings Ollama.
- Adicionado QA Service Ollama.
- Adicionada seleção automática OpenAI/Ollama no startup.
- Adicionadas variáveis Ollama em `internal/config`.
- Ajustadas migrations para `vector(768)`.
- Ajustado banco local para `vector(768)`.
- Corrigido ID de documento para UUID compatível com PostgreSQL.
- Corrigida resposta de ingestão para retornar `chunk_count` real.
- Corrigidos testes quebrados de Chunker e PDF Loader.
- Adicionados testes unitários dos providers Ollama.
- Atualizado `run-local.sh` para aceitar Ollama sem OpenAI API Key.

---

## Limitações Conhecidas do MVP

- O schema atual está otimizado para Ollama/local com embeddings de 768 dimensões.
- Para voltar a OpenAI `text-embedding-3-small`, será necessário usar `vector(1536)` ou tornar a dimensão configurável por migration.
- O índice `ivfflat` em tabela pequena emite aviso de baixa relevância; isso é esperado em ambiente vazio/de estudo.
- O PDF Loader cobre PDFs simples, mas PDFs escaneados ou com layout complexo podem precisar de OCR ou parser mais robusto.
