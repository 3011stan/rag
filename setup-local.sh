#!/bin/bash

# 🚀 GUIA DE EXECUÇÃO LOCAL - RAG BACKEND EM GO
# ================================================

set -e  # Exit on error

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                                                                ║"
echo "║         🚀 INICIANDO RAG BACKEND EM GO - SETUP LOCAL          ║"
echo "║                                                                ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# 1. Verificar se estamos no diretório correto
echo "📍 Verificando diretório..."
if [ ! -f "go.mod" ]; then
    echo "❌ Erro: arquivo go.mod não encontrado"
    echo "   Por favor, execute este script a partir do diretório /src"
    exit 1
fi
echo "✅ Diretório correto: $(pwd)"
echo ""

# 2. Verificar se Go está instalado
echo "🔍 Verificando Go..."
if ! command -v go &> /dev/null; then
    echo "❌ Go não está instalado"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo "✅ Go instalado: $GO_VERSION"
echo ""

# 3. Carregar arquivo .env
echo "📂 Carregando variáveis de ambiente..."
if [ ! -f ".env" ]; then
    echo "⚠️  Arquivo .env não encontrado. Criando com valores padrão..."
    cat > .env << 'ENVFILE'
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/rag?sslmode=disable

# OpenAI
OPENAI_API_KEY=sk-your-api-key-here

# Application
PORT=:8080
ENVIRONMENT=development
LOG_LEVEL=info

# RAG Configuration
CHUNK_TOKENS=800
OVERLAP_TOKENS=100
TOP_K=5
ENVFILE
    echo "✅ Arquivo .env criado"
    echo "⚠️  IMPORTANTE: Edite .env e adicione sua OPENAI_API_KEY antes de continuar!"
fi
echo ""

# 4. Verificar OPENAI_API_KEY
echo "🔑 Verificando variáveis de ambiente..."
export $(cat .env | grep -v '^#' | xargs)

# Check if OPENAI_API_KEY is empty or still has the placeholder value
if [[ -z "${OPENAI_API_KEY}" ]] || [[ "${OPENAI_API_KEY}" == "sk-your-api-key-here" ]]; then
    echo "❌ OPENAI_API_KEY não está configurada corretamente"
    echo "   Por favor, edite o arquivo .env e adicione sua chave OpenAI"
    echo ""
    echo "   Como obter sua chave:"
    echo "   1. Acesse https://platform.openai.com/api-keys"
    echo "   2. Crie uma nova chave API"
    echo "   3. Copie e cole em .env (OPENAI_API_KEY=sk-...)"
    exit 1
fi
echo "✅ OPENAI_API_KEY configurada"
echo ""

# 5. Verificar se Docker está instalado
echo "🐳 Verificando Docker..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker não está instalado"
    echo "   Por favor, instale Docker Desktop em https://www.docker.com/products/docker-desktop"
    exit 1
fi
DOCKER_VERSION=$(docker --version)
echo "✅ Docker instalado: $DOCKER_VERSION"
echo ""

# 6. Iniciar containers Docker
echo "🚀 Iniciando containers Docker..."
if docker ps --filter "name=rag" &> /dev/null; then
    echo "   Parando containers existentes..."
    docker-compose down 2>/dev/null || true
fi

echo "   Iniciando PostgreSQL + pgvector..."
docker-compose up -d

# Aguardar PostgreSQL estar pronto
echo "   Aguardando PostgreSQL ficar pronto..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U postgres &> /dev/null; then
        echo "✅ PostgreSQL pronto"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ PostgreSQL não ficou pronto a tempo"
        exit 1
    fi
    echo -n "."
    sleep 1
done
echo ""

# 7. Aplicar migrations
echo "🔄 Aplicando migrations SQL..."
docker-compose exec -T postgres psql -U postgres -d rag < sql/migrations/0001_create_tables.up.sql
echo "✅ Migrations aplicadas"
echo ""

# 8. Compilar projeto
echo "🔨 Compilando projeto Go..."
go build -o rag-app .
if [ $? -ne 0 ]; then
    echo "❌ Erro ao compilar"
    exit 1
fi
echo "✅ Compilação bem-sucedida"
echo ""

# 9. Testar conexão com banco
echo "🧪 Testando conectividade com banco..."
go run ./cmd/test_connection/main.go
echo "✅ Banco de dados acessível"
echo ""

# 10. Exibir instruções de uso
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                                                                ║"
echo "║              ✅ SETUP CONCLUÍDO COM SUCESSO!                  ║"
echo "║                                                                ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

echo "🚀 PRÓXIMAS AÇÕES:"
echo ""
echo "1️⃣  Iniciar o servidor:"
echo "    $ ./rag-app"
echo "    ou"
echo "    $ go run main.go"
echo ""
echo "2️⃣  Em outro terminal, testar a API:"
echo ""
echo "   Health Check:"
echo "   $ curl http://localhost:8080/health"
echo ""
echo "   Ingerir um PDF:"
echo "   $ curl -X POST http://localhost:8080/rag/ingest \\"
echo "     -F \"file=@document.pdf\""
echo ""
echo "   Fazer uma pergunta:"
echo "   $ curl -X POST http://localhost:8080/rag/ask \\"
echo "     -H \"Content-Type: application/json\" \\"
echo "     -d '{\"question\": \"What is this about?\"}'"
echo ""
echo "3️⃣  Rodar testes:"
echo "    $ go test ./... -v"
echo ""
echo "4️⃣  Parar containers Docker:"
echo "    $ docker-compose down"
echo ""
echo "📚 Para mais informações, veja README.md ou EXAMPLES.md"
echo ""
