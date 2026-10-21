package main

import (
	"log"

	"api/internal/database"
	"api/internal/server"
)

func main() {
	dbService := database.New()
	defer func() {
		if err := dbService.Close(); err != nil {
			log.Printf("could not close database connection: %v", err)
		}
	}()

	db := dbService.DB()
	emb := server.NewEmbeddingService(db )

	log.Println("Gerando embeddings das areas...")

	if err := emb.UpdateMissingAreasEmbeddings(); err != nil {
		log.Fatalf("Erro ao gerar embeddings: %v", err)
	}

	log.Println("Embeddings gerados com sucesso!")
}