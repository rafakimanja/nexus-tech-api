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

	log.Println("Gerando embeddings das areas...")

	if err := server.UpdateMissingAreasEmbeddings(db); err != nil {
		log.Fatalf("Erro ao gerar embeddings: %v", err)
	}

	log.Println("Embeddings gerados com sucesso!")
}