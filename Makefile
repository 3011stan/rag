GOCACHE ?= /private/tmp/rag-go-build-cache
APP ?= rag-app
OLLAMA_BASE_URL ?= http://localhost:11434

.PHONY: test build run db-up db-down db-migrate db-status db-reset ollama-check api-health seed-demo smoke

test:
	GOCACHE=$(GOCACHE) go test ./...

build:
	GOCACHE=$(GOCACHE) go build -o $(APP) .

run: build
	./$(APP)

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-migrate:
	docker exec rag_postgres psql -U postgres -d rag -f /dev/stdin < sql/migrations/0001_create_tables.up.sql

db-status:
	docker exec rag_postgres psql -U postgres -d rag -c "SELECT (SELECT count(*) FROM rag_documents) AS documents, (SELECT count(*) FROM rag_chunks) AS chunks;"

db-reset:
	docker exec rag_postgres psql -U postgres -d rag -c "TRUNCATE rag_chunks, rag_documents RESTART IDENTITY CASCADE;"

ollama-check:
	curl -sS $(OLLAMA_BASE_URL)/api/tags
	curl -sS $(OLLAMA_BASE_URL)/api/embeddings -d '{"model":"nomic-embed-text","prompt":"teste"}'

api-health:
	curl -i http://localhost:8080/health

seed-demo:
	GOCACHE=$(GOCACHE) go run ./cmd/seed_demo

smoke:
	sh scripts/smoke.sh
