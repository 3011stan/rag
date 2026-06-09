# StanOS: Obsidian Knowledge System Plan

## Context

StanOS is the working name for a personal knowledge and automation system built around Obsidian, local Markdown files, RAG, and AI agents such as Codex/Claude.

The current source of truth for notes, tasks, and project management is mostly Notion. The goal is not to blindly recreate Notion inside Obsidian, but to gradually move the parts that benefit from local-first Markdown, search, backlinks, automation, and agent-friendly editing.

The user does not use iPhone or iPad. This makes Syncthing a strong synchronization option because the expected device set is closer to:

- macOS/Linux/Windows desktop or laptop
- Android phone
- possibly multiple computers

The implementation of this plan will happen in another chat. This document should provide enough context for that chat to continue without re-discovering the product direction.

## Product Direction

The system should support four long-term goals:

1. Build a second brain for technical learning, content production, personal projects, and career transition into Machine Learning Engineering.
2. Make notes usable by humans and AI agents through plain Markdown files and predictable structure.
3. Connect Obsidian knowledge with the existing RAG portfolio project.
4. Eventually support personal automations, task/project intelligence, and data collection pipelines.

## Current Recommendation

Use Obsidian as the primary knowledge base and Syncthing as the main cross-device sync layer.

Recommended baseline:

```txt
Obsidian = writing, reading, linking, knowledge management
Syncthing = file sync between computers and Android
Git = optional backup/versioning from the main computer
Codex/Claude = agents operating on the local vault files
RAG API = external knowledge retrieval/QA layer for selected documents
```

Avoid using multiple sync engines on the same vault at the same time. For example, do not mix Syncthing, Dropbox, Google Drive, and Obsidian Sync on the same folder unless there is a very deliberate reason.

## Why Obsidian

Obsidian makes sense for this project because:

- Notes are Markdown files on disk.
- The vault can be edited by Codex/Claude as a normal folder.
- It is easier to version, inspect, transform, and refactor knowledge.
- Backlinks and internal links help connect ideas across ML Engineering, software engineering, content, projects, and personal systems.
- The system can grow into a personal data/RAG pipeline without being locked into a SaaS data model.

Notion can still remain useful for operational dashboards, collaboration, or highly visual databases during the transition.

## Migration Strategy From Notion

Do not migrate everything at once.

Start with content that benefits from being local and agent-friendly:

- ML Engineering notes
- RAG notes
- project logs
- content ideas
- article/video outlines
- study notes
- portfolio project notes

Keep in Notion temporarily:

- complex databases
- dashboards
- active task/project views if they still work well
- anything that would slow down the migration too much

Later, evaluate whether task/project management should move fully into StanOS.

## Initial Vault Structure

Suggested first structure:

```txt
StanOS/
  00-inbox/
  10-projects/
  20-areas/
  30-resources/
  40-content/
  50-rag/
  90-archive/
  _templates/
  _attachments/
```

### Folder Roles

`00-inbox`

Fast capture area for ideas, rough notes, links, and unsorted thoughts.

`10-projects`

Active projects with defined outcomes. Examples:

- RAG portfolio project
- Obsidian migration
- ML Engineering transition
- frontend for RAG API
- content production system

`20-areas`

Ongoing responsibilities without a fixed finish line. Examples:

- career
- health
- finance
- studies
- content
- personal operations

`30-resources`

Reusable knowledge notes. Examples:

- RAG
- embeddings
- retrieval evaluation
- Go
- system design
- ML Engineering
- LLM applications

`40-content`

Content production workspace. Examples:

- article drafts
- short-form scripts
- video ideas
- post outlines
- published content index

`50-rag`

Material intentionally prepared for ingestion into the RAG system. This folder can later become the bridge between Obsidian and the RAG API.

`90-archive`

Inactive or completed material.

`_templates`

Reusable templates for notes, projects, content, meetings, and daily logs.

`_attachments`

Images, PDFs, exported files, and other attachments.

## Naming Conventions

Use simple, readable file names:

```txt
YYYY-MM-DD daily note.md
RAG Evaluation.md
Frontend for RAG API.md
ML Engineering Content Strategy.md
```

Prefer stable names over clever names. Links should remain easy to understand.

## Metadata Convention

Use YAML frontmatter only when it adds value.

Example:

```yaml
---
type: project
status: active
area: portfolio
tags:
  - rag
  - ml-engineering
created: 2026-05-11
---
```

Avoid over-modeling the system too early. Metadata should support retrieval, filtering, automation, and content production.

## Initial Templates

Create templates later for:

- project
- resource note
- content idea
- article draft
- daily note
- weekly review
- RAG ingestion note

Example project template:

```md
---
type: project
status: active
area:
tags: []
created:
---

# Project Name

## Outcome

## Context

## Next Actions

## Decisions

## References
```

## Syncthing Setup Plan

The recommended sync plan is:

1. Create the Obsidian vault on the main computer.
2. Add the vault folder to Syncthing.
3. Share it with the Android device and any other computers.
4. On Android, configure Syncthing to sync to a local folder accessible by Obsidian mobile.
5. Open that local folder as the Obsidian vault on Android.
6. Test creating and editing small notes from both sides.
7. Only then migrate larger Notion exports.

Important operating rules:

- Do not edit the same note on two devices at the same time.
- Let Syncthing finish syncing before switching devices.
- Treat Syncthing as sync, not backup.
- Keep a separate backup/versioning strategy.

## Git Strategy

Git can be used as backup and history, especially from the main computer.

Recommended:

- private GitHub repository for the vault or selected parts of it
- regular commits for meaningful changes
- ignore volatile Obsidian workspace files

Potential `.gitignore`:

```gitignore
.obsidian/workspace.json
.obsidian/workspace-mobile.json
.obsidian/cache/
.trash/
```

Do not commit secrets, private credentials, API keys, database URLs, or sensitive personal data.

## Relationship With The RAG Portfolio Project

The current RAG API already supports:

- PDF ingestion
- Markdown ingestion
- plain text ingestion
- protected upload
- demo document seeding
- Gemini/Ollama provider configuration
- deployed API on Render

StanOS can become a real data source for that RAG project.

Near-term idea:

```txt
Obsidian note -> selected folder/file -> RAG ingestion -> ask questions over curated personal knowledge
```

Recommended initial bridge:

- Only ingest notes from `50-rag/`.
- Keep private/sensitive notes out of the RAG ingestion path.
- Add explicit metadata to notes intended for ingestion.
- Prefer intentional ingestion over automatic whole-vault ingestion.

Retrieval metadata should be treated as preference, not exclusion, by default.

The RAG should avoid hard metadata filters for normal questions because useful answers may need cooperation between layers. For example, a question about Reels may still benefit from foundational storytelling notes, self-knowledge, and editorial decisions.

Initial retrieval direction:

- Use metadata preferences for soft boosting.
- Prefer matching layers, categories, or platforms without excluding other relevant chunks.
- Keep layered retrieval as a future evolution if stronger context composition control becomes necessary.

## Future Features

These are product ideas for later. They should be recorded now but not implemented in the first setup.

### 1. Task, Agenda, And Project Tool

Build a custom tool to manage tasks, agenda, projects, and personal planning.

Current behavior:

- The user uses a bullet journal manually and by hand.
- The goal is not to replace the handwritten bullet journal.
- The digital system should complement it by producing metrics, insights, reminders, and automation.

Future ideal:

The user wakes up and asks:

```txt
What do we have today?
```

The automation returns:

- tasks scheduled for the day
- agenda items
- project next actions
- reminders
- context-aware suggestions

The system should eventually be context-aware. Example:

- If it is cloudy or raining, some tasks may be replaced.
- If an outdoor task becomes impractical, the system may suggest an indoor alternative.
- If an appointment becomes impossible, the system may recommend rescheduling.

Possible future inputs:

- Obsidian tasks
- calendar events
- weather
- habit logs
- project metadata
- bullet journal manual review
- repository task files such as `ai/tasks.md`
- Git branches and commits linked to task IDs

Future developer workflow integration:

- Project tasks should have stable IDs.
- Git branches should include the task ID that originated the work.
- Remote Git branches should be accompanied by a pull request with a clear description of the change, validation performed, and review context.
- The task manager should eventually connect daily planning with software delivery work.
- Example: a day plan may include `T099 - Document RAG content feeding strategy`, and the related Git branch may be `docs/T099-rag-content-guidelines`.
- This connects personal planning, project management, Git history, and portfolio narrative.

### 2. Personal Data Generation Tool

Create a process/tool that generates owned data for analysis and future RAG usage.

Possible sources:

- liked tweets/posts
- comments written on social platforms
- liked YouTube videos
- Discord community interactions
- saved links
- reading history
- content ideas
- notes from conversations

Goal:

- produce structured personal data
- analyze behavior, interests, and content patterns
- generate input for RAG
- support content production and learning loops

This should be designed carefully to avoid collecting noisy data without purpose.

### 3. Frontend For The RAG API

Build a frontend to interact with the existing RAG API.

This is likely a near-term portfolio task.

Expected features:

- ask questions
- display answer as Markdown
- show retrieved sources/chunks
- show latency/debug metadata
- list demo documents
- possibly upload protected documents in local/admin mode

This frontend should be portfolio-ready and explain the project clearly.

### 4. Intentional Send-To-RAG Tool

Create a tool to intentionally send selected information to the RAG system.

Possible implementation:

- Chrome extension
- browser bookmarklet
- local script
- Obsidian command/plugin later

Expected workflow:

```txt
select text on any web page
send selection to RAG/StanOS
store source URL, title, timestamp, and selected text
optionally create an Obsidian note
optionally ingest into RAG
```

This may overlap with the personal data generation tool. The distinction should be:

- Personal data generation tool = passive or semi-automatic collection of owned interaction data.
- Send-to-RAG tool = intentional capture of specific content selected by the user.

## First Implementation Milestones

### Milestone 1: Local Vault Setup

- Create the Obsidian vault.
- Add the initial folder structure.
- Add basic templates.
- Configure Syncthing between computer and Android.
- Validate editing from both devices.

### Milestone 2: Notion Migration Trial

- Export a small subset from Notion.
- Import only study/content/project notes.
- Clean links and file names.
- Decide what remains in Notion for now.

### Milestone 3: Agent-Friendly Structure

- Ensure notes are readable by Codex/Claude.
- Add conventions for metadata, links, and folders.
- Create an index/MOC for ML Engineering.
- Create an index/MOC for content production.

### Milestone 4: RAG Integration

- Use `50-rag/` as the curated ingestion folder.
- Create a manual ingestion script or workflow.
- Keep sensitive/private notes excluded.
- Test ask/answer over selected Obsidian notes.

### Milestone 5: Portfolio Frontend

- Build a simple UI for the deployed RAG API.
- Render answers as Markdown.
- Display sources/chunks.
- Use the existing demo dataset and selected Obsidian-derived notes where appropriate.

## Open Questions For The Implementation Chat

- What operating systems and Android device/folder paths will be used?
- Should `.obsidian/` config sync through Syncthing initially?
- Should the vault be backed by a private GitHub repo from day one?
- Which Notion pages should be migrated first?
- Which notes are allowed to be ingested into the RAG API?
- What privacy boundary should exist between personal notes and portfolio/demo data?

## Guiding Principles

- Start simple.
- Prefer plain Markdown over complex plugin-dependent workflows.
- Do not migrate everything at once.
- Keep private data out of public demos.
- Build workflows that help produce content and engineering artifacts.
- Let the system evolve from real use, not from an overly detailed upfront taxonomy.
