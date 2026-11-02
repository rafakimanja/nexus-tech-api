package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"api/internal/domain/areas"
	"api/internal/domain/demand"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Processor generates vectors for records whose embedding is pending (NULL).
type Processor struct {
	db     *gorm.DB
	client *http.Client
	url    string
}

func NewProcessor(db *gorm.DB) *Processor {
	return &Processor{db: db, client: &http.Client{Timeout: 5 * time.Second}, url: strings.TrimRight(os.Getenv("OLLAMA_URL"), "/")}
}

func (p *Processor) generate(ctx context.Context, text string) (*pgvector.Vector, error) {
	if p.url == "" {
		return nil, fmt.Errorf("OLLAMA_URL is not configured")
	}
	body, err := json.Marshal(map[string]string{"model": "mxbai-embed-large", "prompt": text})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Embedding) != 1024 {
		return nil, fmt.Errorf("expected 1024 embedding dimensions, got %d", len(result.Embedding))
	}
	var norm float64
	for _, v := range result.Embedding {
		norm += float64(v) * float64(v)
	}
	if norm == 0 || math.IsNaN(norm) || math.IsInf(norm, 0) {
		return nil, fmt.Errorf("invalid embedding vector")
	}
	vector := pgvector.NewVector(result.Embedding)
	return &vector, nil
}

func (p *Processor) ProcessDemand(id uint) error {
	var d demand.Demanda
	if err := p.db.First(&d, id).Error; err != nil {
		return err
	}
	if d.Embedding != nil {
		return nil
	}
	v, err := p.generate(context.Background(), d.ToEmbeddingText())
	if err != nil {
		return err
	}
	// Do not store a vector generated for text changed while Ollama was running.
	return p.db.Model(&d).Where("embedding IS NULL AND title = ? AND description = ?", d.Title, d.Description).Update("embedding", v).Error
}

func (p *Processor) ProcessArea(id uint) error {
	var a areas.Area
	if err := p.db.First(&a, id).Error; err != nil {
		return err
	}
	if a.Embedding != nil {
		return nil
	}
	v, err := p.generate(context.Background(), a.Name)
	if err != nil {
		return err
	}
	return p.db.Model(&a).Where("embedding IS NULL AND name = ?", a.Name).Update("embedding", v).Error
}

// ProcessPending attempts every pending item, even if one item fails.
func (p *Processor) ProcessPending() error {
	var areaIDs, demandIDs []uint
	if err := p.db.Model(&areas.Area{}).Where("embedding IS NULL").Pluck("id", &areaIDs).Error; err != nil {
		return err
	}
	if err := p.db.Model(&demand.Demanda{}).Where("embedding IS NULL").Pluck("id", &demandIDs).Error; err != nil {
		return err
	}
	var failures []string
	for _, id := range areaIDs {
		if err := p.ProcessArea(id); err != nil {
			failures = append(failures, fmt.Sprintf("area %d: %v", id, err))
		}
	}
	for _, id := range demandIDs {
		if err := p.ProcessDemand(id); err != nil {
			failures = append(failures, fmt.Sprintf("demand %d: %v", id, err))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("embedding processing failed: %s", strings.Join(failures, "; "))
	}
	return nil
}
