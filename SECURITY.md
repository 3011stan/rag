# Security Notes

This project is a portfolio RAG API, but it is deployed publicly and should be treated as internet-facing software.

## Current Guard Rails

- Public upload is disabled in production by default.
- `POST /rag/ingest` requires either `ADMIN_TOKEN` or a scoped temporary token when `ENABLE_PUBLIC_UPLOAD=false`.
- Protected ingestion accepts PDF, Markdown, and plain text documents.
- Temporary tokens are stateless, signed, valid for 30 minutes, and limited to upload plus document listing.
- Document deletion and debug metadata require `ADMIN_TOKEN`.
- `POST /admin/seed-demo` requires `ADMIN_TOKEN`.
- Upload size is capped by `MAX_UPLOAD_BYTES`.
- Question request bodies are capped and must be JSON.
- HTTP responses include basic security headers.
- Internal errors are logged server-side and returned as generic 500 responses.
- Prompt templates tell the model to treat retrieved documents as untrusted content.
- `POST /rag/ask` can be protected with in-memory IP-based rate limiting.

## Secrets

Never commit:

- `.env`
- `DATABASE_URL`
- `GEMINI_API_KEY`
- `ADMIN_TOKEN`
- `TEMP_TOKEN_SECRET`
- provider tokens or database passwords

Production secrets belong in the hosting provider dashboard. GitHub Secrets should only be used when GitHub Actions needs the value.

## Public Demo Policy

The public demo should use curated documents only. Public upload should remain disabled unless malware scanning and stronger abuse controls are added.

## Rate Limiting

The public question endpoint uses a simple in-memory fixed-window limiter when enabled:

```env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=20
RATE_LIMIT_WINDOW_SECONDS=60
```

It is enabled by default in production and disabled by default in development. The limiter is per client IP and reads `X-Forwarded-For`, then `X-Real-IP`, then `RemoteAddr`.

This is enough for the portfolio demo on a single instance. For multi-instance production, move rate limiting to Redis, a gateway, or a CDN/WAF.

## Future Hardening

- Add automated deploy smoke tests.
- Add audit logs for admin endpoints.
- Add retrieval evaluation to detect low-quality or unsafe answers.
- Add a stricter public/private route policy before adding a frontend.
