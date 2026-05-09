# Security Notes

This project is a portfolio RAG API, but it is deployed publicly and should be treated as internet-facing software.

## Current Guard Rails

- Public upload is disabled in production by default.
- `POST /rag/ingest` requires `ADMIN_TOKEN` when `ENABLE_PUBLIC_UPLOAD=false`.
- `POST /admin/seed-demo` requires `ADMIN_TOKEN`.
- Upload size is capped by `MAX_UPLOAD_BYTES`.
- Question request bodies are capped and must be JSON.
- HTTP responses include basic security headers.
- Internal errors are logged server-side and returned as generic 500 responses.
- Prompt templates tell the model to treat retrieved documents as untrusted content.

## Secrets

Never commit:

- `.env`
- `DATABASE_URL`
- `GEMINI_API_KEY`
- `ADMIN_TOKEN`
- provider tokens or database passwords

Production secrets belong in the hosting provider dashboard. GitHub Secrets should only be used when GitHub Actions needs the value.

## Public Demo Policy

The public demo should use curated documents only. Public upload should remain disabled unless rate limiting, malware scanning, and abuse controls are added.

## Future Hardening

- Add automated deploy smoke tests.
- Add API rate limiting.
- Add audit logs for admin endpoints.
- Add retrieval evaluation to detect low-quality or unsafe answers.
- Add a stricter public/private route policy before adding a frontend.
