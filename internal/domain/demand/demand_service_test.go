package demand

import (
	"errors"
	"testing"
)

type testRepository struct {
	id  uint
	err error
}

func (r *testRepository) FindByID(int) (Demanda, error) { return Demanda{}, nil }
func (r *testRepository) FindAll() ([]Demanda, error)   { return nil, nil }
func (r *testRepository) Create(Demanda) (uint, error)  { return r.id, r.err }
func (r *testRepository) Update(int, Demanda) error     { return r.err }
func (r *testRepository) Delete(int) error              { return nil }

type failingEmbedder struct{ called uint }

func (e *failingEmbedder) ProcessDemand(id uint) error {
	e.called = id
	return errors.New("ollama unavailable")
}

func TestCreateKeepsDemandWhenEmbeddingFails(t *testing.T) {
	embedder := &failingEmbedder{}
	service := NewService(&testRepository{id: 42}, embedder)
	id, err := service.Create(Demanda{})
	if err != nil || id != 42 || embedder.called != 42 {
		t.Fatalf("id=%d error=%v embedder called for %d", id, err, embedder.called)
	}
}

func TestCreateDoesNotGenerateEmbeddingWhenSaveFails(t *testing.T) {
	embedder := &failingEmbedder{}
	service := NewService(&testRepository{err: errors.New("database unavailable")}, embedder)
	_, err := service.Create(Demanda{})
	if err == nil || embedder.called != 0 {
		t.Fatalf("error=%v embedder called for %d", err, embedder.called)
	}
}
