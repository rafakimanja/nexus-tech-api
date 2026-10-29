package server

import (
	"api/internal/domain/areas"
	"api/internal/domain/demand"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)


const OLLAMA_URL = "OLLAMA_URL"

func GenerateEmbedding(text string)(*pgvector.Vector, error){
	url := os.Getenv(OLLAMA_URL)

	payload, _ := json.Marshal(map[string]string{
		"model": "mxbai-embed-large",
		"prompt": text,
	})


	resp, err := http.Post(
		url+"/api/embeddings",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama retornou status %d", resp.StatusCode)
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	vector := pgvector.NewVector(result.Embedding)
	return &vector, nil
}

func UpdateDemandEmbedding(demanda *demand.Demanda, db *gorm.DB) error {
	text := demanda.ToEmbeddingText()

	embedding, err := GenerateEmbedding(text)
	if err != nil {
		return err
	}

	return db.Model(demanda).Update("embedding", embedding).Error
}

func UpdateAreaEmbeddings(area *areas.Area, db *gorm.DB) error {
	embedding, err := GenerateEmbedding(area.Name)
	if err != nil {
		return err
	}

	return db.Model(area).Update("embedding", embedding).Error
}

func UpdateMissingDemandEmbeddings(db *gorm.DB) error {
	var demandas []demand.Demanda

	if err := db.Where("embedding IS NULL").
		Find(&demandas).Error; err != nil {
		return err
	}

	for i := range demandas {
		if err := UpdateDemandEmbedding(&demandas[i], db); err != nil {
			return fmt.Errorf(
				"erro na demanda %d: %w",
				demandas[i].ID,
				err,
			)
		}
	}

	return nil
}

func UpdateMissingAreasEmbeddings(db *gorm.DB) error {
	var areas []areas.Area

	if err := db.Where("embedding IS NULL").
		Find(&areas).Error; err != nil {
		return err
	}

	for i := range areas {
		if err := UpdateAreaEmbeddings(&areas[i], db); err != nil {
			return fmt.Errorf(
				"erro na demanda %d: %w",
				areas[i].ID,
				err,
			)
		}
	}

	return nil
}

func CompareEmbeddings(demandEmbed *pgvector.Vector){

}