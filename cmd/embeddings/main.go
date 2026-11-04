package main

import (
	"log"

	"api/internal/database"
	"api/internal/embedding"
)

func main() {
	dbService := database.New()
	defer func() {
		if err := dbService.Close(); err != nil {
			log.Printf("could not close database connection: %v", err)
		}
	}()

	db := dbService.DB()

	log.Println("Gerando embeddings pendentes de áreas e demandas...")

	if err := embedding.NewProcessor(db).ProcessPending(); err != nil {
		log.Fatalf("Erro ao gerar embeddings: %v", err)
	}

	log.Println("Embeddings gerados com sucesso!")
}
