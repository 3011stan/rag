CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE rag_documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source TEXT,
  title TEXT,
  checksum TEXT UNIQUE,
  metadata JSONB,
  created_at timestamptz DEFAULT now()
);

CREATE TABLE rag_chunks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  document_id UUID REFERENCES rag_documents(id) ON DELETE CASCADE,
  chunk_index INT NOT NULL,
  content TEXT NOT NULL,
  token_count INT,
  metadata JSONB,
  embedding vector(768),
  created_at timestamptz DEFAULT now()
);

CREATE INDEX ON rag_chunks USING ivfflat (embedding vector_cosine_ops);
