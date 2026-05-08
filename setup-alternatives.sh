#!/bin/bash

# 🚀 SETUP DO RAG BACKEND SEM OPENAI API KEY
# ============================================
# Este script ajuda você a configurar o projeto para rodar
# com modelos locais (Ollama) ou gratuitos (Groq, Hugging Face)

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║                                                            ║"
echo "║  🚀 RAG Backend - Setup SEM OpenAI API Key               ║"
echo "║                                                            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

echo "Qual alternativa você quer usar?"
echo ""
echo "1️⃣  Ollama (Local, Offline, Recomendado ⭐⭐⭐)"
echo "2️⃣  Groq (Gratuito, Online, Muito Rápido ⭐⭐⭐)"
echo "3️⃣  Hugging Face (Gratuito, Online ⭐⭐)"
echo ""
read -p "Escolha (1/2/3): " choice

case $choice in
    1)
        echo ""
        echo "📦 OLLAMA - Configuração"
        echo "========================"
        echo ""
        echo "1. Instalar Ollama (se não tiver):"
        echo "   macOS: brew install ollama"
        echo "   Ou baixe de: https://ollama.ai"
        echo ""
        echo "2. Em um terminal, iniciar o servidor:"
        echo "   ollama serve"
        echo ""
        echo "3. Em outro terminal, baixar os modelos:"
        echo "   ollama pull nomic-embed-text"
        echo "   ollama pull mistral"
        echo ""
        echo "4. Verificar que está funcionando:"
        echo "   curl http://localhost:11434/api/tags"
        echo ""
        echo "Preparar .env? (s/n)"
        read -p "Resposta: " setup_env
        
        if [[ "$setup_env" == "s" || "$setup_env" == "S" ]]; then
            cat > .env << 'ENVFILE'
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/rag?sslmode=disable

# OpenAI (deixar vazio - não será usado)
OPENAI_API_KEY=

# Ollama
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_EMBED_MODEL=nomic-embed-text
OLLAMA_LLM_MODEL=mistral

# Application
PORT=:8080
ENVIRONMENT=development
LOG_LEVEL=info

# RAG Configuration
CHUNK_TOKENS=800
OVERLAP_TOKENS=100
TOP_K=5
ENVFILE
            echo "✅ .env criado com config do Ollama!"
            echo ""
            echo "📝 Próximas ações:"
            echo "   1. Certifique-se de que 'ollama serve' está rodando"
            echo "   2. Execute: docker-compose up -d"
            echo "   3. Execute: go run main.go"
        fi
        ;;
        
    2)
        echo ""
        echo "⚡ GROQ - Configuração"
        echo "====================="
        echo ""
        echo "1. Criar conta em: https://console.groq.com"
        echo "2. Gerar API key em: https://console.groq.com/keys"
        echo "3. Copiar a chave"
        echo ""
        read -p "Cole sua Groq API Key: " groq_key
        
        if [ -z "$groq_key" ]; then
            echo "❌ API Key vazia. Cancelando..."
            exit 1
        fi
        
        cat > .env << ENVFILE
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/rag?sslmode=disable

# OpenAI (deixar vazio - não será usado)
OPENAI_API_KEY=

# Groq (para LLM/Respostas)
GROQ_API_KEY=$groq_key
GROQ_MODEL=mixtral-8x7b-32768

# Para Embeddings, usar Hugging Face ou Ollama
# Se escolher Ollama, descomente abaixo:
# OLLAMA_BASE_URL=http://localhost:11434
# OLLAMA_EMBED_MODEL=nomic-embed-text

# Application
PORT=:8080
ENVIRONMENT=development
LOG_LEVEL=info

# RAG Configuration
CHUNK_TOKENS=800
OVERLAP_TOKENS=100
TOP_K=5
ENVFILE
        
        echo "✅ .env criado com config do Groq!"
        echo ""
        echo "📝 Próximas ações:"
        echo "   1. Execute: docker-compose up -d"
        echo "   2. Execute: go run main.go"
        ;;
        
    3)
        echo ""
        echo "🤗 HUGGING FACE - Configuração"
        echo "=============================="
        echo ""
        echo "1. Criar conta em: https://huggingface.co"
        echo "2. Gerar token em: https://huggingface.co/settings/tokens"
        echo "3. Copiar o token"
        echo ""
        read -p "Cole seu Hugging Face Token: " hf_token
        
        if [ -z "$hf_token" ]; then
            echo "❌ Token vazio. Cancelando..."
            exit 1
        fi
        
        cat > .env << ENVFILE
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/rag?sslmode=disable

# OpenAI (deixar vazio - não será usado)
OPENAI_API_KEY=

# Hugging Face (Embeddings + LLM)
HUGGINGFACE_API_KEY=$hf_token
HUGGINGFACE_EMBED_MODEL=sentence-transformers/all-MiniLM-L6-v2
HUGGINGFACE_LLM_MODEL=mistralai/Mistral-7B-Instruct-v0.1

# Application
PORT=:8080
ENVIRONMENT=development
LOG_LEVEL=info

# RAG Configuration
CHUNK_TOKENS=800
OVERLAP_TOKENS=100
TOP_K=5
ENVFILE
        
        echo "✅ .env criado com config do Hugging Face!"
        echo ""
        echo "📝 Próximas ações:"
        echo "   1. Execute: docker-compose up -d"
        echo "   2. Execute: go run main.go"
        ;;
        
    *)
        echo "❌ Opção inválida"
        exit 1
        ;;
esac

echo ""
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "📚 Documentação completa: ALTERNATIVES_WITHOUT_OPENAI.md"
echo ""
echo "❓ Dúvidas?"
echo "   - Ollama: https://ollama.ai"
echo "   - Groq: https://console.groq.com"
echo "   - Hugging Face: https://huggingface.co"
echo ""
