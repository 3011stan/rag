#!/bin/bash

# 🚀 GUIA DE EXECUÇÃO LOCAL - RAG Backend em Go
# =====================================================

# Este script contém os passos para executar o projeto RAG Backend localmente

set -e  # Exit on error

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "\n${BLUE}╔════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║ $1${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}\n"
}

print_step() {
    echo -e "${YELLOW}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# =====================================================
# PASSO 1: Verificar Pré-requisitos
# =====================================================

print_header "PASSO 1: Verificando Pré-requisitos"

print_step "Verificando Go..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    print_success "Go instalado: $GO_VERSION"
else
    print_error "Go não encontrado. Instale Go 1.21+ de https://golang.org"
    exit 1
fi

print_step "Verificando Docker..."
if command -v docker &> /dev/null; then
    print_success "Docker instalado: $(docker --version)"
else
    print_error "Docker não encontrado. Instale Docker de https://www.docker.com"
    exit 1
fi

print_step "Verificando Docker Compose..."
if command -v docker-compose &> /dev/null; then
    print_success "Docker Compose instalado: $(docker-compose --version)"
else
    print_error "Docker Compose não encontrado"
    exit 1
fi

print_step "Verificando provider de IA..."
if [ -z "$OPENAI_API_KEY" ]; then
    OLLAMA_URL="${OLLAMA_BASE_URL:-http://localhost:11434}"
    if curl -sS "$OLLAMA_URL/api/tags" >/dev/null 2>&1; then
        print_success "Ollama disponível em $OLLAMA_URL"
    else
        print_error "OPENAI_API_KEY não definida e Ollama não respondeu em $OLLAMA_URL"
        echo "Inicie o Ollama ou defina OPENAI_API_KEY."
        exit 1
    fi
else
    print_success "OPENAI_API_KEY está configurada"
fi

# =====================================================
# PASSO 2: Iniciar Infraestrutura
# =====================================================

print_header "PASSO 2: Iniciando Infraestrutura (Docker)"

print_step "Iniciando PostgreSQL + pgvector..."
docker-compose up -d

print_step "Aguardando PostgreSQL ficar pronto..."
sleep 5

# Verificar se o container está rodando
if docker ps | grep -q rag_postgres; then
    print_success "PostgreSQL está rodando"
else
    print_error "Falha ao iniciar PostgreSQL"
    docker-compose logs
    exit 1
fi

# =====================================================
# PASSO 3: Aplicar Migrations
# =====================================================

print_header "PASSO 3: Aplicando Migrations do Banco de Dados"

print_step "Aplicando schema SQL..."
PGPASSWORD=postgres psql -h localhost -U postgres -d rag -f sql/migrations/0001_create_tables.up.sql

if [ $? -eq 0 ]; then
    print_success "Migrations aplicadas com sucesso"
else
    print_error "Erro ao aplicar migrations"
    exit 1
fi

# =====================================================
# PASSO 4: Verificar Conectividade
# =====================================================

print_header "PASSO 4: Testando Conectividade com Banco de Dados"

print_step "Executando teste de conexão..."
go run ./cmd/test_connection/main.go

print_success "Conectividade verificada"

# =====================================================
# PASSO 5: Download de Dependências
# =====================================================

print_header "PASSO 5: Baixando Dependências Go"

print_step "Executando go mod download..."
go mod download

print_success "Dependências baixadas"

# =====================================================
# PASSO 6: Compilar Projeto
# =====================================================

print_header "PASSO 6: Compilando Projeto"

print_step "Compilando..."
go build -o rag-app .

if [ -f rag-app ]; then
    print_success "Compilação bem-sucedida"
    ls -lh rag-app
else
    print_error "Falha na compilação"
    exit 1
fi

# =====================================================
# PASSO 7: Executar Testes
# =====================================================

print_header "PASSO 7: Executando Testes (Opcional)"

read -p "Deseja executar os testes? (s/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Ss]$ ]]; then
    print_step "Executando testes..."
    go test ./... -v
    print_success "Testes completados"
else
    print_step "Pulando testes"
fi

# =====================================================
# PASSO 8: Iniciar Servidor
# =====================================================

print_header "PASSO 8: Iniciando Servidor"

print_success "✅ Tudo pronto! Iniciando servidor..."
echo ""
echo -e "${GREEN}O servidor estará disponível em:${NC}"
echo -e "${BLUE}  Health: http://localhost:8080/health${NC}"
echo -e "${BLUE}  Ingest: POST http://localhost:8080/rag/ingest${NC}"
echo -e "${BLUE}  Ask:    POST http://localhost:8080/rag/ask${NC}"
echo ""
echo -e "${YELLOW}Pressione Ctrl+C para parar o servidor${NC}\n"

./rag-app

# =====================================================
# Cleanup (se o script for interrompido)
# =====================================================

cleanup() {
    echo ""
    print_header "Finalizando..."
    print_step "Parando containers Docker..."
    docker-compose down
    print_success "Cleanup concluído"
}

trap cleanup EXIT
