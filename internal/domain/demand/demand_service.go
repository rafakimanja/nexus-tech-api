package demand

import "log"

type DemandaService interface {
	GetAll() ([]Demanda, error)
	GetByID(id int) (Demanda, error)
	Create(demanda Demanda) (uint, error)
	Update(id int, demanda Demanda) error
	DeleteByID(id int) error
}

type service struct {
	repo     DemandaRepository
	embedder interface{ ProcessDemand(uint) error }
}

func NewService(repo DemandaRepository, embedder interface{ ProcessDemand(uint) error }) DemandaService {
	return &service{repo: repo, embedder: embedder}
}

func (s *service) GetAll() ([]Demanda, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Demanda, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(demanda Demanda) (uint, error) {
	id, err := s.repo.Create(demanda)
	if err == nil {
		s.processEmbedding(id)
	}
	return id, err
}

func (s *service) Update(id int, demanda Demanda) error {
	err := s.repo.Update(id, demanda)
	if err == nil {
		s.processEmbedding(uint(id))
	}
	return err
}

func (s *service) processEmbedding(id uint) {
	if s.embedder != nil {
		if err := s.embedder.ProcessDemand(id); err != nil {
			log.Printf("demand %d embedding pending: %v", id, err)
		}
	}
}

func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}
