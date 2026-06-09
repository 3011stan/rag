# RAG Project Documentation

This folder keeps operational documentation for rules and decisions that should not be hidden in code.

## Map

- [Business Rules](business-rules.md): API access rules, protected operations, and retrieval behavior.
- [Metadata](metadata.md): accepted metadata schema, reserved fields, precedence rules, and compatibility.
- [ADR-0001 Metadata Precedence](decisions/ADR-0001-metadata-precedence.md): formal decision about loader metadata vs curation metadata.

## Documentation Principle

Document what someone needs to know to avoid breaking the system or to make a consistent decision.

Prefer documentation that answers:

- What is the rule?
- Why does it exist?
- Where does it affect the code?
- What cannot change without an explicit decision?
