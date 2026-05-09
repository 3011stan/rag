# RAG Lab - Roadmap de Portfolio

## Visao

Transformar o MVP atual em um projeto de portfolio de Machine Learning Engineering:

**RAG Lab**: uma plataforma local-first para ingestao, busca semantica, resposta com fontes, avaliacao e comparacao de pipelines RAG usando modelos locais e cloud.

O objetivo nao e apenas demonstrar "chat com PDF", mas mostrar capacidade de construir, medir, operar e melhorar um sistema RAG.

---

## Posicionamento Para Portfolio

Descricao curta:

> Plataforma RAG em Go com PostgreSQL/pgvector e suporte a modelos locais via Ollama. O projeto implementa ingestao de PDFs, chunking, embeddings, retrieval semantico, respostas com fontes, observabilidade e um harness de avaliacao para comparar configuracoes de RAG.

O que este projeto deve comunicar:

- Engenharia de software aplicada a ML.
- Arquitetura extensivel para providers de embeddings e LLM.
- Capacidade de rodar local/offline/free.
- Entendimento de retrieval, chunking, avaliacao e trade-offs.
- Preocupacao com qualidade, observabilidade e reproducibilidade.

---

## Estado Atual

Ja existe:

- API Go com endpoints `GET /health`, `POST /rag/ingest`, `POST /rag/ask`.
- PostgreSQL com pgvector.
- Ingestao de PDF.
- Chunking por tokens.
- Embeddings com OpenAI e Ollama.
- QA com OpenAI e Ollama.
- Selecao automatica: OpenAI quando `OPENAI_API_KEY` existe, Ollama quando vazia.
- Teste end-to-end local validado com:
  - `nomic-embed-text` para embeddings.
  - `mistral` para resposta.
  - `vector(768)` no pgvector.

Limitacoes atuais:

- Estrutura ainda parece MVP/prototipo.
- Pouca separacao entre configuracao, providers e pipeline.
- Sem UI.
- Sem avaliacao formal de qualidade.
- Sem metricas por etapa do pipeline.
- Sem historico de experimentos.
- Documentacao ainda precisa ser simplificada.

---

## Decisoes De Produto E Portfolio

Estas decisoes guiam as proximas refatoracoes e evitam otimizar o projeto para um caminho errado.

### Demo Publica

- A demo publica deve usar documentos pre-carregados.
- Upload publico livre deve ficar desabilitado por padrao.
- Upload pode existir no modo local/dev ou em modo protegido.
- Motivo: reduzir risco de abuso, custo, arquivos inadequados e complexidade operacional.

### Providers

- Local/dev: Ollama.
- Deploy/demo: `AI_PROVIDER` configuravel.
- Provider cloud pode ser usado na demo se Ollama for pesado demais para hospedar.
- O backend deve continuar Docker-first e provider-agnostic.

### Frontend

- Frontend fica para depois que o backend estiver pronto e estavel.
- Quando chegar a hora, a UI deve ser simples e demonstravel:
  - lista de documentos;
  - pergunta/resposta;
  - fontes/chunks recuperados;
  - painel de debug/latencia.

### Deploy Target

Deploy target significa o ambiente/plataforma onde o projeto sera publicado.

Nao sao bibliotecas Go. Exemplos:

- Render, Fly.io e Railway: plataformas para hospedar a API/backend.
- Vercel e Netlify: plataformas para hospedar frontend.
- Postgres gerenciado: banco hospedado por um provedor cloud.
- Docker Compose: ambiente local/reproduzivel para desenvolvimento.

Decisao atual:

- Preparar o backend de forma agnostica com Docker.
- Escolher a plataforma final depois que backend e demo estiverem prontos.
- Usar GitHub Actions para uma pipeline simples de CI/CD.

### Secrets

- Secrets nunca devem ser commitados no GitHub.
- `.env` e `.env.*` devem permanecer ignorados pelo Git.
- `.env.example` pode ser versionado apenas com placeholders.
- `DATABASE_URL`, `GEMINI_API_KEY`, tokens de deploy e senhas devem ser configurados no provedor de hospedagem.
- GitHub Secrets devem ser usados apenas quando a pipeline precisar acessar o valor.
- A CI atual nao precisa de secrets porque roda apenas testes e build.

### Dataset De Demonstracao

O dataset da demo nao sera apenas sobre conteudo tecnico puro.

Direcao escolhida:

> RAG para apoiar producao de conteudo sobre Machine Learning Engineering, RAG, arquitetura de software e temas relacionados.

Isso conecta melhor com uma ideia futura de produto/conteudo e ainda demonstra conhecimento tecnico.

Exemplos de documentos/dados para a demo:

- notas sobre conceitos de RAG;
- resumos de papers;
- outlines de posts/artigos;
- guias internos ficticios de producao de conteudo tecnico;
- documentos sobre estrategia editorial para temas de ML Engineering.

### Idioma

- README principal, commits, nomes de features e documentacao final de portfolio devem ser em ingles.
- Podemos manter notas internas em portugues quando fizer sentido durante o desenvolvimento.

### Git

- Usar Conventional Commits.
- Commits pequenos e narrativos.
- Antes de refatorar, versionar o baseline atual em blocos para preservar a historia de evolucao.

---

## O Que Importa Em Um Projeto RAG Forte

### 1. Ingestao

- Parsing de documentos.
- Limpeza e normalizacao de texto.
- Metadata por documento e chunk.
- Deduplicacao por checksum.
- Reingestao/versionamento.
- Tratamento claro de PDFs sem texto extraivel.

### 2. Chunking

- Chunking por tokens.
- Overlap configuravel.
- Estrategias alternativas:
  - tamanho fixo;
  - por pagina;
  - por secoes/titulos;
  - chunking recursivo.
- Comparacao entre estrategias.

### 3. Embeddings

- Providers plugaveis.
- Dimensao vetorial documentada por modelo.
- Comparacao local vs cloud.
- Medicao de latencia.
- Cache opcional de embeddings.

### 4. Retrieval

- Top-k configuravel.
- Filtros por metadata/documento.
- Score threshold.
- Hybrid search: keyword + vector.
- Reranking opcional.
- Retorno transparente dos chunks usados.

### 5. Geracao

- Prompt templates versionados.
- Resposta com fontes.
- Recusa quando contexto for insuficiente.
- Controle de temperatura/modelo.
- Protecao contra alucinacao via instrucoes e avaliacao.

### 6. Avaliacao

- Dataset pequeno de perguntas e respostas esperadas.
- Expected sources por pergunta.
- Metricas:
  - retrieval hit rate;
  - context precision;
  - answer faithfulness;
  - latencia media;
  - tokens/custo quando aplicavel.
- Comparacao entre configuracoes:
  - chunk size;
  - overlap;
  - top-k;
  - provider;
  - reranker.

### 7. Observabilidade

- Request ID.
- Latencia por etapa:
  - parsing;
  - chunking;
  - embeddings;
  - retrieval;
  - generation.
- Modelo usado.
- Chunks recuperados.
- Scores.
- Logs estruturados.

### 8. Produto/Demo

- UI simples para upload e perguntas.
- Lista de documentos ingeridos.
- Visualizacao dos chunks recuperados.
- Historico de perguntas.
- Pagina de experimentos/avaliacao.

---

## Roadmap De Refatoracao

### Fase 1 - Consolidacao Do MVP

Objetivo: deixar o backend limpo, reproduzivel e facil de explicar.

- [x] Consolidar documentacao em poucos arquivos.
- [x] Criar `Makefile` com:
  - `make test`
  - `make build`
  - `make run`
  - `make db-up`
  - `make db-migrate`
  - `make ollama-check`
- [x] Separar configuracao de provider:
  - `AI_PROVIDER=ollama|openai|auto`
  - `EMBEDDING_MODEL`
  - `LLM_MODEL`
  - `EMBEDDING_DIMENSIONS`
- [ ] Criar interfaces mais claras:
  - `EmbeddingProvider`
  - `LLMProvider`
  - `Retriever`
  - `RAGPipeline`
- [ ] Remover codigo duplicado entre QA OpenAI e QA Ollama.
- [ ] Padronizar erros e respostas da API.
- [ ] Criar README principal enxuto.

Entregavel:

- Backend local rodando com um comando documentado.
- Testes verdes.
- Documentacao curta e confiavel.

Status em 2026-05-07:

- `make test` passou.
- `make build` passou.
- API em `:8080` continuou respondendo `GET /health` e `POST /rag/ask`.
- Tentativa de subir uma segunda instancia em `:8081` para validar o binario novo foi recusada na aprovacao do ambiente, entao nao foi repetida.

Status em 2026-05-08:

- Baseline funcional versionado em Git com commits pequenos.
- Provider Gemini implementado para demo/deploy sem OpenAI.
- CI inicial criado com GitHub Actions.
- Dockerfile de API criado para deploy.
- `render.yaml` adicionado como blueprint inicial.
- Schema RAG agora e garantido no startup via `DATABASE_URL`.

### Fase 2 - Observabilidade De Pipeline

Objetivo: mostrar maturidade de ML Engineering.

- Instrumentar tempos por etapa.
- Retornar metadata opcional de debug em `/rag/ask`.
- Persistir logs ou traces simples por request.
- Criar endpoint de estatisticas:
  - total de documentos;
  - total de chunks;
  - modelos configurados;
  - dimensao de embedding;
  - tempo medio de resposta.

Entregavel:

- Resposta ou logs mostrando exatamente onde o tempo foi gasto e quais chunks foram usados.

### Fase 3 - Evaluation Harness

Objetivo: transformar o projeto em laboratorio RAG.

- Criar pasta `eval/`.
- Definir dataset JSON/YAML:
  - pergunta;
  - resposta esperada;
  - fontes esperadas;
  - tags.
- Criar comando:
  - `make eval`
  - ou `go run ./cmd/eval`
- Medir:
  - hit rate de fontes;
  - presenca de termos esperados;
  - latencia;
  - configuracao usada.
- Gerar relatorio em Markdown/JSON.

Entregavel:

- Um relatorio comparando pelo menos duas configuracoes, por exemplo:
  - chunk 400/top_k 3;
  - chunk 800/top_k 5.

### Fase 4 - Melhorias De Retrieval

Objetivo: mostrar conhecimento alem do basico.

- Adicionar filtros por `document_id`.
- Adicionar score threshold.
- Adicionar busca lexical simples.
- Implementar hybrid retrieval.
- Opcional: adicionar reranker local ou cloud.

Entregavel:

- Experimento demonstrando melhora ou trade-off entre retrieval vetorial e hybrid.

### Fase 5 - UI E Demo

Objetivo: tornar o projeto demonstravel para recrutadores e avaliadores.

- Esta fase deve comecar apenas depois que o backend estiver pronto e estavel.
- Criar UI simples:
  - upload de PDF;
  - lista de documentos;
  - pergunta/resposta;
  - sources e previews;
  - painel de debug do retrieval.
- Proteger demo publica:
  - documentos de exemplo pre-carregados;
  - upload desabilitado ou limitado;
  - rate limit basico.

Entregavel:

- Demo local com UI.
- Opcional: deploy publico limitado.

### Fase 6 - Portfolio

Objetivo: empacotar o projeto como case tecnico.

- README com:
  - problema;
  - arquitetura;
  - como rodar local;
  - trade-offs;
  - resultados de avaliacao;
  - screenshots/GIF.
- Artigo curto:
  - "Building a local-first RAG system with Go, pgvector and Ollama".
- Video/GIF de 2 minutos.
- Link no portfolio para:
  - GitHub;
  - demo ou video;
  - artigo tecnico.

---

## Deploy: Faz Sentido?

Sim, mas a melhor estrategia e separar **demo publica** de **modo local completo**.

Recomendacao:

- Manter o modo completo local com Ollama no README.
- Para demo publica, usar documentos pre-carregados e limites fortes.
- Evitar upload publico irrestrito.
- Se Ollama ficar pesado no servidor, usar provider cloud barato apenas na demo.
- Um video/GIF pode ser suficiente se o custo de deploy for alto.
- A plataforma final de deploy sera escolhida depois; por enquanto o projeto deve ser Docker-first e cloud-agnostic.

Opcoes:

- **Portfolio minimo**: GitHub + README forte + video curto.
- **Portfolio intermediario**: UI local + screenshots + eval report.
- **Portfolio avancado**: demo hospedada + artigo + relatorio de experimentos.

### Estrategia De Deploy Recomendada

O projeto deve ser preparado para deploy sem depender de uma plataforma especifica.

Arquitetura de deploy:

- API Go empacotada em Docker.
- Banco PostgreSQL com extensao pgvector.
- Documentos de demo pre-carregados por script/job de seed.
- `AI_PROVIDER` configuravel por ambiente.
- Upload publico desabilitado por padrao em producao.

Ambientes:

- `local`: Docker Compose + Ollama + Postgres local.
- `staging` ou `demo`: API hospedada + Postgres gerenciado + provider cloud ou Ollama remoto, se viavel.
- `production/portfolio`: demo estavel com documentos pre-carregados.

Pipeline simples com GitHub Actions:

1. Rodar testes em todo push/PR:
   - checkout;
   - setup Go;
   - cache de dependencias;
   - `go test ./...`;
   - `go build`.
2. Em merge na `main`:
   - build da imagem Docker;
   - push para registry, por exemplo GitHub Container Registry;
   - disparar deploy na plataforma escolhida.
3. Depois do deploy:
   - health check em `/health`;
   - opcionalmente rodar smoke test em `/rag/ask` com dataset pre-carregado.

Primeira versao da pipeline:

- CI apenas: testes e build.
- Sem deploy automatico ate escolhermos a plataforma.

Segunda versao:

- CD: build/push Docker image e deploy automatico.

Decisao pendente:

- Escolher plataforma de hospedagem final para a API e o banco.
- Confirmar se a demo usara provider cloud ou Ollama hospedado.

Decisao operacional atual:

- API: preparar deploy inicial no Render Free Web Service.
- Banco: Supabase Postgres com pgvector.
- IA: Gemini API free tier para embeddings e geracao.
- Secrets de producao devem ser configurados manualmente no Render:
  - `DATABASE_URL`
  - `GEMINI_API_KEY`

---

## Estrutura De Documentacao Desejada

Manter poucos documentos:

- `ai/PORTFOLIO_ROADMAP.md`: este roadmap.
- `ai/tasks.md`: backlog e status das tarefas.
- `ai/STATUS.md`: estado operacional atual.
- `ai/plan-rag.md`: plano original, mantido como historico.
- `ai/README.md`: referencia atual do MVP, a ser revisada depois.

Evitar novos markdowns soltos para cada descoberta. Quando houver nova decisao, atualizar este roadmap, `tasks.md` ou `STATUS.md`.

---

## Git Flow E Convencao De Commits

Objetivo: usar o historico Git como narrativa de evolucao do projeto, com commits pequenos, revisaveis e alinhados ao portfolio.

### Branches

- `main`: estado estavel e demonstravel.
- `feature/<escopo-curto>`: novas funcionalidades ou refatoracoes.
- `fix/<escopo-curto>`: correcoes de bug.
- `docs/<escopo-curto>`: documentacao.
- `chore/<escopo-curto>`: tarefas de manutencao, tooling e organizacao.

Exemplos:

- `feature/rag-pipeline`
- `feature/eval-harness`
- `fix/pdf-document-id`
- `docs/portfolio-readme`
- `chore/makefile`

### Convencao De Commits

Usar Conventional Commits:

```text
<tipo>(<escopo>): <mensagem curta no imperativo>
```

Tipos principais:

- `feat`: nova funcionalidade.
- `fix`: correcao de bug.
- `refactor`: mudanca interna sem alterar comportamento esperado.
- `test`: testes.
- `docs`: documentacao.
- `chore`: tooling, configuracao ou manutencao.
- `perf`: melhoria de performance.

Exemplos:

```text
docs(portfolio): add roadmap and git workflow
chore(git): add ignore rules for local artifacts
feat(config): support explicit ai provider selection
refactor(api): extract rag pipeline from handlers
test(ollama): cover local embedding provider errors
fix(loader): generate uuid document ids
```

### Regra Para Commits Pequenos

Cada commit deve responder a uma pergunta simples:

- O que mudou?
- Por que mudou?
- Como validar?

Evitar commits que misturam categorias diferentes. Por exemplo, nao juntar refatoracao de pipeline, ajuste de README e mudanca de schema no mesmo commit.

### Template Mental De Commit

Antes de commitar:

```text
make test
make build
```

Quando tocar nos handlers, pipeline, providers ou banco, validar tambem:

```text
curl http://localhost:8080/health
curl -X POST http://localhost:8080/rag/ask ...
```

### Sequencia Recomendada A Partir Daqui

1. `docs(portfolio): add roadmap and git workflow`
2. `chore(git): add ignore rules for local artifacts`
3. `chore(makefile): add project automation commands`
4. `feat(config): support explicit ai provider selection`
5. `refactor(api): extract rag pipeline from handlers`
6. `refactor(qa): introduce llm provider interface`
7. `feat(observability): record pipeline stage timings`
8. `feat(eval): add rag evaluation harness`

Como o Git sera iniciado depois de parte do MVP ja existir, o historico deve contar a evolucao a partir da profissionalizacao do projeto, nao tentar recriar artificialmente tudo que aconteceu antes.

---

## Proximas Tasks Sugeridas

- Criar `RAGPipeline` para concentrar ingestao e ask fora dos handlers HTTP.
- Refatorar providers para remover duplicacao entre OpenAI e Ollama.
- Adicionar metricas por etapa.
- Criar dataset inicial de avaliacao.
- Criar comando `cmd/eval`.
- Escrever README de portfolio.
