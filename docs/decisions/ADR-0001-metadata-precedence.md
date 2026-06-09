# ADR-0001: Metadata Precedence

## Status

Accepted

## Context

The ingestion pipeline produces technical metadata through document loaders. Examples include filename, content type, source type, PDF page count, and checksum.

The user may also provide curation metadata during ingestion. Examples include layer, category, platform, source quality, author, tags, and visibility.

Some fields can exist in both places. For example, a PDF loader may expose `author`, while the user may provide a curated `author` value that is more useful for the corpus.

Without an explicit precedence rule, one metadata source can silently overwrite the other.

## Decision

Use the following precedence rule:

```txt
loader wins technical/reserved fields;
curation wins semantic/editorial fields when provided;
curation is optional.
```

Reserved technical fields:

- `filename`
- `content_type`
- `source_type`
- `pages`
- `checksum`

Curation metadata must not contain reserved technical fields. If it does, ingestion should fail with a validation error.

Semantic/editorial fields may be provided by curation metadata and may override loader-provided values when both exist.

## Consequences

- Technical document identity stays consistent.
- Curated metadata can improve weak or missing loader metadata.
- Existing ingestion without curation metadata remains compatible.
- Metadata can support future soft boost retrieval without requiring hard filters.

## Alternatives Considered

### Loader Always Wins

Rejected.

This preserves technical consistency but discards curated semantic fields when loaders provide weak or empty values.

### Curation Always Wins

Rejected.

This allows users to overwrite technical metadata such as filename, source type, page count, or checksum, which can make document records inconsistent.

### Hard Separation With No Overlap

Rejected for now.

It is cleaner but too rigid. Some fields, such as `author`, can reasonably be inferred by loaders and improved by curation.
