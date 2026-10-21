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

type Emb struct {
	db *gorm.DB
}

func NewEmbeddingService(db *gorm.DB) Emb {
	return Emb{
		db: db,
	}
}

func(e Emb) GenerateEmbedding(text string)(*pgvector.Vector, error){
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

func(e Emb) UpdateDemandEmbedding(demanda *demand.Demanda) error {
	text := demanda.ToEmbeddingText()

	embedding, err := e.GenerateEmbedding(text)
	if err != nil {
		return err
	}

	return e.db.Model(demanda).Update("embedding", embedding).Error
}

func(e Emb) UpdateAreaEmbeddings(area *areas.Area) error {
	embedding, err := e.GenerateEmbedding(area.Name)
	if err != nil {
		return err
	}

	return e.db.Model(area).Update("embedding", embedding).Error
}

func(e Emb) UpdateMissingDemandEmbeddings() error {
	var demandas []demand.Demanda

	if err := e.db.Where("embedding IS NULL").
		Find(&demandas).Error; err != nil {
		return err
	}

	for i := range demandas {
		if err := e.UpdateDemandEmbedding(&demandas[i]); err != nil {
			return fmt.Errorf(
				"erro na demanda %d: %w",
				demandas[i].ID,
				err,
			)
		}
	}

	return nil
}

func(e Emb) UpdateMissingAreasEmbeddings() error {
	var areas []areas.Area

	if err := e.db.Where("embedding IS NULL").
		Find(&areas).Error; err != nil {
		return err
	}

	for i := range areas {
		if err := e.UpdateAreaEmbeddings(&areas[i]); err != nil {
			return fmt.Errorf(
				"erro na demanda %d: %w",
				areas[i].ID,
				err,
			)
		}
	}

	return nil
}