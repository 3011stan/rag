# Evaluation

This project uses deterministic retrieval evaluation in CI before adding heavier LLM evaluation tools.

The goal is to catch regressions in:

- ingestion;
- curated metadata;
- deterministic embeddings;
- pgvector retrieval;
- soft boost ranking;
- source metadata returned by `/rag/ask`.

It does not evaluate final LLM answer quality yet.

## CI Workflow

The workflow lives in:

```txt
.github/workflows/rag-eval.yml
```

It runs on every pull request and on manual dispatch.

The job runs fully inside the GitHub Actions runner:

1. Start a Postgres database with pgvector.
2. Build the API from the pull request branch.
3. Start the API locally with `AI_PROVIDER=test`.
4. Ingest stable eval fixtures from `eval/fixtures`.
5. Run `go run ./cmd/eval`.
6. Write a GitHub Step Summary.
7. Fail the job if deterministic retrieval thresholds are not met.

No Render, Supabase, Gemini, OpenAI, Ollama, or LangSmith dependency is required for this CI eval.

## Deterministic Provider

The eval workflow uses:

```txt
AI_PROVIDER=test
EMBEDDING_DIMENSIONS=256
```

This provider uses a deterministic hashing vectorizer for embeddings and a deterministic test answerer.

This keeps CI:

- offline;
- reproducible;
- fast;
- free of external model/API secrets.

The provider is not intended to measure real semantic quality. It is intended to validate retrieval mechanics and metadata-driven ranking behavior.

## Dataset

The eval dataset lives in:

```txt
eval/questions.yaml
```

Each question defines:

- `id`;
- `question`;
- `top_k`;
- `preferences`;
- `expected_metadata`;
- optional `expected_preview_contains`.

## Metrics

The runner reports:

- `metadata_hit_rate`: at least one returned source matches expected metadata.
- `top_source_metadata_hit_rate`: the first returned source matches expected metadata.
- `empty_source_count`: number of questions with fewer than the minimum expected sources.

Initial thresholds are intentionally modest:

```yaml
metadata_hit_rate: 0.80
top_source_metadata_hit_rate: 0.60
min_sources_per_question: 1
```

## Adding Evaluation Coverage

Add a new fixture only when the CI needs to cover a stable retrieval behavior.

Do not mirror every real corpus document in `eval/fixtures`.

Use fixtures for stable regression tests. Use the real corpus for product quality evaluation and human review.

Add a new question when:

- a new metadata value becomes important;
- a retrieval regression is found;
- a new retrieval strategy is introduced;
- a portfolio-critical user flow needs a guardrail.

## Future Work

Future evaluation tasks may add:

- automated comparison against previous runs;
- post-deploy eval against Render;
- human-in-the-loop scoring;
- LLM-as-judge;
- LangSmith or Langfuse integration.
