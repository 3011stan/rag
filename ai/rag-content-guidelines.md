# RAG Content Feeding Guidelines

## Purpose

This document defines how content should be selected, prepared, classified, and ingested into the RAG system.

The RAG should not become a generic "chat with documents" or a repository of shallow creator advice. It should become a curated knowledge layer for content production, technical communication, ML Engineering learning, and StanOS workflows.

The core principle is:

```txt
The RAG should learn creation systems, not generic creator content.
```

## What The RAG Should Learn

The corpus should teach durable systems and decision-making patterns:

- Narrative structure
- Retention
- Technical storytelling
- Knowledge transformation
- Creator workflows
- Attention psychology
- Distribution strategy
- Positioning
- Multi-platform adaptation
- Documentary and authentic creation
- Personal editorial decisions

The goal is not to imitate generic creators. The goal is to help the system understand how the user thinks, creates, approves, rejects, and adapts ideas.

## What Should Be Avoided

Avoid feeding the RAG with low-quality or noisy material:

- Generic viral hook lists
- Growth hacks
- "How to grow on TikTok" style posts
- Copywriting tricks without context
- Trend-dependent advice
- Large random PDF dumps
- Low-signal social media advice
- Content that would make the system sound generic, motivational, or inauthentic

The system should not become a content coach with generic advice. It should preserve the user's positioning and voice.

## Corpus Strategy

Use a small, highly curated corpus before scaling.

Initial target:

```txt
10-20 high-quality documents
```

Quality matters more than quantity. Adding hundreds of documents too early will make retrieval noisier and evaluation harder.

## Knowledge Layers

The corpus should be organized in layers.

### Layer 1: Foundations

Universal principles with long shelf life.

This layer teaches:

- Narrative
- Retention
- Storytelling
- Attention
- Communication
- Content structure

Examples:

- Storytelling frameworks
- Retention analysis
- Communication principles
- Educational writing principles

### Layer 2: Platform Specific

Platform-aware content strategy.

This layer captures how different media formats behave:

- Reels
- TikTok
- YouTube
- LinkedIn
- Medium
- Podcasts

Each platform may have different pacing, density, audience expectations, algorithmic constraints, and language.

### Layer 3: Self Knowledge

The user's own content, voice, decisions, and results.

This should become the most important layer over time.

It includes:

- Approved scripts
- Published posts
- Reels
- Content ideas
- Personal tone and voice
- Formats that worked
- Formats that were rejected
- Review notes

This layer helps the RAG learn identity, not generic advice.

### Layer 4: Editorial Decisions

Explicit decisions made during the content and product process.

Examples:

- Use documentary tone instead of hype.
- Prefer process over results.
- Use insider humor when it fits the audience.
- Keep technical explanations practical and grounded.
- Avoid toxic productivity aesthetics.
- Less artificial editing can improve authenticity.

Editorial decisions are valuable because they encode judgment. They should be ingested intentionally.

## Initial Corpus Categories

### Creator Systems

Purpose:

- Teach how creators design workflows, pipelines, and operating systems.

Good sources:

- Workflow breakdowns
- Production systems
- Creative pipelines
- Creator operating systems
- Behind-the-scenes essays
- Akita-style process articles

Avoid:

- Surface-level productivity advice
- Generic "creator tips"

### Storytelling And Retention

Purpose:

- Teach structure, pacing, tension, payoff, and audience attention.

Good sources:

- YouTube essay analysis
- Retention breakdowns
- Narrative frameworks
- Story structure notes
- Educational content structure

### Technical Communication

Purpose:

- Teach how to transform complexity into clarity.

Good sources:

- Engineering blogs
- Architecture explainers
- Strong technical documentation
- Technical creators who explain complex systems clearly

### Platform Dynamics

Purpose:

- Teach how ideas change across platforms and formats.

Good sources:

- YouTube format analysis
- Reels/TikTok pacing notes
- LinkedIn writing patterns
- Medium/article structure
- Podcast-to-article adaptation

### Self Knowledge

Purpose:

- Teach the system the user's own voice, preferences, and content strategy.

Good sources:

- Scripts
- Published posts
- Drafts approved by the user
- Content retrospectives
- Personal principles
- Voice and tone guidelines

## Metadata Schema

Every ingested document should have metadata.

The metadata can live in Markdown frontmatter in StanOS/Obsidian and later be sent to the RAG backend during ingestion.

Metadata fields must remain backward-compatible. Existing documents may have only basic loader metadata such as filename, content type, source type, page count, or checksum. The RAG must continue to retrieve those documents even when the new schema fields are missing.

Missing metadata should mean:

- No soft boost for that field.
- No ingestion failure.
- No retrieval failure.
- No exclusion from normal semantic search.

Initial schema:

```yaml
---
type: knowledge_asset
layer: foundations
category: storytelling
platform: general
source_kind: article
source_quality: high
evergreen: true
visibility: private
source_url:
author:
created:
captured_at:
tags:
  - narrative
  - retention
---
```

Recommended fields:

- `type`: document role, such as `knowledge_asset`, `editorial_decision`, `script`, `workflow`, or `reference`.
- `layer`: `foundations`, `platform_specific`, `self_knowledge`, or `editorial_decisions`.
- `category`: corpus category, such as `creator_systems`, `storytelling`, `technical_communication`, `platform_dynamics`, or `self_knowledge`.
- `platform`: `general`, `youtube`, `reels`, `tiktok`, `linkedin`, `medium`, or `podcast`.
- `source_kind`: source format, such as `article`, `transcript`, `note`, `script`, `decision`, `pdf_extract`, `workflow`, or `reference`.
- `source_quality`: `high`, `medium`, or `low`.
- `evergreen`: whether the material is expected to remain useful over time.
- `visibility`: `private`, `portfolio_demo`, or `public`.
- `source_url`: original URL when available.
- `author`: original author or creator when available.
- `created`: original creation date when known.
- `captured_at`: date when the material was captured into StanOS.
- `tags`: focused tags that help retrieval and review.

`ingestion_status` is a StanOS/Obsidian workflow field, not RAG backend metadata. It can be used before ingestion with values such as `candidate`, `approved`, or `rejected`, but it should not be persisted in the RAG database. If a document is present in the RAG database, its ingestion status is already implicit.

Avoid over-modeling the schema too early. Add fields only when they support retrieval, curation, evaluation, or automation.

## Existing Document Metadata Backfill

The current RAG database already persists document records and chunk records:

```txt
rag_documents = document identity, source, title, checksum, metadata, created_at
rag_chunks = chunk content, token count, metadata, embedding, created_at
```

The system does not persist the original uploaded file as a binary artifact, but it does persist enough document and chunk information to list, inspect, retrieve, and enrich existing records.

Existing documents should continue working even if they only have loader metadata.

Because the current corpus is small and will be replaced by more relevant material, manual backfill is not required now. Prefer ingesting new relevant documents with curated metadata from the start.

Backfill may be revisited later if the corpus grows with valuable legacy documents.

Future backfill process, if needed:

1. List existing documents.
2. Inspect each title, source, metadata, and representative chunks.
3. Assign `layer`, `category`, `platform`, `source_quality`, `visibility`, and useful tags.
4. Update `rag_documents.metadata`.
5. Propagate relevant metadata to the document chunks, or re-ingest the source with metadata when safer.
6. Ask validation questions and confirm the enriched documents receive appropriate soft boosts.

Backward compatibility rule:

```txt
Legacy documents must continue to work before and after metadata backfill.
```

Backfill should improve retrieval behavior, not become a prerequisite for the API to work.

## Corpus Evaluation Process

Before adding many new documents, maintain a small evaluation set.

Each evaluation question should record:

- Question.
- Expected source or expected source metadata.
- Preferred metadata, when relevant.
- Human usefulness score.
- Human groundedness/fidelity note.
- Observed failure or regression, if any.

Suggested human scoring:

- `0`: not useful or unsupported by retrieved context.
- `1`: partially useful but incomplete, generic, or weakly grounded.
- `2`: useful, grounded, and actionable.

Suggested lightweight metrics:

- Source hit rate: expected source appears in `sources`.
- Metadata hit rate: expected layer/category/platform appears in `sources`.
- Answer usefulness average.
- Groundedness/fidelity notes.
- Regression count after new ingestion.

The goal is to know whether new content improves retrieval and answers, not only whether ingestion succeeds.

## Document Preparation Process

When possible, prefer clean Markdown or plain text over raw PDFs.

Recommended preparation flow:

1. Select a candidate source.
2. Evaluate whether it teaches a durable system, workflow, principle, or decision.
3. Extract clean text.
4. Remove navigation, ads, comments, duplicated text, and irrelevant boilerplate.
5. Preserve title, source URL, author, and capture date.
6. Add metadata frontmatter.
7. Transform the source into a knowledge asset when useful.
8. Save it in the appropriate StanOS/Obsidian folder.
9. Mark `ingestion_status: approved` only when it should enter the RAG.
10. Ingest into the RAG.
11. Test with representative questions.
12. Record whether retrieval and answers were useful.

## Knowledge Asset Transformation

Not every source should be ingested as raw text.

For high-value sources, convert them into knowledge assets by extracting:

- Principles
- Heuristics
- Patterns
- Frameworks
- Checklists
- Workflows
- Editorial decisions
- Anti-patterns

This makes the corpus easier to retrieve and more useful for generation.

Example:

```md
# Creator Workflow: Technical Documentary Video

## Source

- URL:
- Author:

## Core Principles

- Principle 1
- Principle 2

## Workflow

1. Research
2. Outline
3. Script
4. Record
5. Edit
6. Publish
7. Review

## Useful For

- YouTube essays
- Technical storytelling
- Behind-the-scenes content
```

## Manual Feeding Session

Use this process whenever time is reserved in the agenda for RAG feeding.

Suggested session length:

```txt
60-90 minutes
```

Checklist:

1. Pick 1-3 candidate sources.
2. Decide whether each source is worth keeping.
3. Reject low-signal or generic material quickly.
4. Extract clean text.
5. Add metadata.
6. Convert important sources into knowledge assets.
7. Save approved assets in the curated StanOS/RAG area.
8. Ingest approved files into the RAG.
9. Ask 3-5 validation questions.
10. Record what worked, what failed, and what needs better metadata.

The goal of the session is not volume. The goal is corpus quality.

## Retrieval Strategy Decision

Metadata should be used as preference, not exclusion, by default.

The system should avoid hard metadata filters for normal questions. A useful answer may require cooperation between multiple layers.

Example:

```txt
A question about Reels may need:
- platform_specific/reels
- foundations/storytelling
- self_knowledge
- editorial_decisions
```

Initial implementation direction:

- Use metadata preferences for soft boosting.
- Prefer chunks that match requested layers, categories, or platforms.
- Do not exclude other semantically relevant chunks by default.
- Keep layered retrieval as a future evolution if stronger context composition control becomes necessary.

The API should eventually expose this as `preferences`, not `filters`, to avoid implying strict exclusion.

Example future request:

```json
{
  "question": "How should I structure a technical reel?",
  "top_k": 5,
  "preferences": {
    "platform": "reels",
    "category": "storytelling",
    "preferred_layers": ["platform_specific", "self_knowledge"]
  }
}
```

## Relationship With StanOS

StanOS should act as the curation and operating layer.

Recommended relationship:

```txt
StanOS/Obsidian = capture, curation, metadata, editorial decisions
RAG backend = ingestion, embeddings, retrieval, answer generation
Frontend = interaction, debugging, source inspection
```

The initial StanOS folder for approved RAG material should be:

```txt
50-rag/
```

Only approved, intentional material should be ingested. Private notes should not enter the portfolio/demo corpus by accident.

## Future Direction

Later improvements:

- Automate ingestion from approved StanOS folders.
- Build a frontend/admin panel for corpus management.
- Add metadata preferences with soft boosting in the backend.
- Show metadata and source quality in retrieved sources.
- Build a Chrome extension or browser capture tool.
- Build a content knowledge graph connecting creators, formats, techniques, hooks, styles, and outcomes.
- Connect task planning and content production through StanOS.

## Current Decision Summary

- Start with a small, curated corpus.
- Prefer systems, workflows, principles, and decisions over generic advice.
- Store metadata for every ingested document.
- Prefer Markdown/plain text when possible.
- Use StanOS as the curation layer.
- Use metadata as retrieval preferences, not hard filters.
- Treat self knowledge and editorial decisions as high-value future corpus layers.
