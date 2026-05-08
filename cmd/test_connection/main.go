package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/stan/Projects/studies/rag/internal/rag"
)

func main() {
	// Carregar variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis de sistema")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatalf("Erro: DATABASE_URL não definida")
	}

	store, err := rag.NewPGVectorStore(dsn)
	if err != nil {
		log.Fatalf("Erro ao criar VectorStore: %v", err)
	}
	defer store.DB.Close()

	if err := store.TestConnection(); err != nil {
		log.Fatalf("Erro ao testar conexão: %v", err)
	}

	fmt.Println("Conexão com o banco de dados bem-sucedida!")
}
