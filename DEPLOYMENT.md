# Deployment Guide

This project is prepared for a low-cost public API deployment.

## Recommended Stack

- API: Render Free Web Service, Docker runtime.
- Database: Neon or Supabase free tier with pgvector.
- AI provider: Gemini API free tier.
- Demo data: bundled markdown documents seeded through a protected admin endpoint.

## Required Accounts

- GitHub repository connected to Render.
- Neon or Supabase project for PostgreSQL.
- Google AI Studio API key for Gemini.

## Secrets Policy

Never commit secrets to GitHub.

Do not commit:

- `.env`
- full `DATABASE_URL` values with real passwords
- `GEMINI_API_KEY`
- `ADMIN_TOKEN`
- provider tokens
- database passwords

This repository intentionally tracks only `.env.example`, with empty or placeholder values.

Set production secrets directly in the hosting provider dashboard, for example Render Environment Variables.

Use GitHub repository secrets only when GitHub Actions itself needs the value. The current CI does not need production secrets because it only runs tests and build.

## 1. Create The Database

Create a PostgreSQL database on Neon or Supabase.

Enable pgvector:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

Copy the production connection string. It should look like:

```text
postgres://USER:PASSWORD@HOST/DB?sslmode=require
```

This is the value for:

```env
DATABASE_URL=...
```

For the current Supabase project, use this safe template:

```text
postgresql://postgres:<YOUR-PASSWORD>@db.glkogcmzzyazhoefynui.supabase.co:5432/postgres?sslmode=require
```

Replace `<YOUR-PASSWORD>` manually only in the deployment provider secret field. Do not save the completed URL in Git.

The API also attempts to create the extension and tables on startup, but the database user must have permission to create the extension. If the provider requires extensions to be enabled through the dashboard, enable `vector` there first.

## 2. Create Gemini API Key

Create an API key in Google AI Studio.

Use it as:

```env
GEMINI_API_KEY=...
```

Set this value manually in Render as an environment variable. Do not add it to GitHub unless a future GitHub Actions deployment workflow needs it.

For public demo documents, avoid seeding private or sensitive content.

## 3. Deploy On Render

Use the repository's `render.yaml` blueprint.

Render will create a Docker web service with:

```env
AI_PROVIDER=gemini
GEMINI_BASE_URL=https://generativelanguage.googleapis.com/v1beta
EMBEDDING_MODEL=gemini-embedding-001
LLM_MODEL=gemini-2.5-flash-lite
EMBEDDING_DIMENSIONS=768
```

You must manually provide secret values:

```env
DATABASE_URL=postgresql://postgres:<YOUR-PASSWORD>@db.glkogcmzzyazhoefynui.supabase.co:5432/postgres?sslmode=require
GEMINI_API_KEY=<YOUR-GEMINI-API-KEY>
ADMIN_TOKEN=<LONG-RANDOM-SECRET>
```

In Render, mark these as secret environment variables. The values should not appear in committed files, logs, screenshots, issues, or pull requests.

Render Free does not support Blueprint `preDeployCommand` for free services. After the first successful deploy, seed demo documents by calling:

```bash
curl -X POST https://YOUR-RENDER-URL/admin/seed-demo \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

That endpoint inserts the bundled demo documents into the vector store. If `ADMIN_TOKEN` is not set, the endpoint is disabled.

## 4. Smoke Test

After deploy, test:

```bash
curl https://YOUR-RENDER-URL/health
```

Then ask a question:

```bash
curl -X POST https://YOUR-RENDER-URL/rag/ask \
  -H 'Content-Type: application/json' \
  -d '{"question":"Who is the audience for the content strategy?","top_k":3}'
```

Expected behavior:

- HTTP 200.
- Non-empty `answer`.
- At least one item in `sources`.

## Current Blockers

Deployment cannot be completed until these values exist:

- `DATABASE_URL`
- `GEMINI_API_KEY`
- `ADMIN_TOKEN`

Once those are available, the API can be deployed from the current `main` branch.
