package areas

import "log"

type AreasService interface {
	GetAll() ([]Area, error)
	GetByID(id int) (Area, error)
	Create(area Area) (uint, error)
	Update(id int, area Area) error
	DeleteByID(id int) error
}

type service struct {
	repo     AreasRepository
	embedder interface{ ProcessArea(uint) error }
}

func NewService(repo AreasRepository, embedder interface{ ProcessArea(uint) error }) AreasService {
	return &service{repo: repo, embedder: embedder}
}

func (s *service) GetAll() ([]Area, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Area, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(area Area) (uint, error) {
	id, err := s.repo.Create(area)
	if err == nil {
		s.processEmbedding(id)
	}
	return id, err
}

func (s *service) Update(id int, area Area) error {
	err := s.repo.Update(id, area)
	if err == nil {
		s.processEmbedding(uint(id))
	}
	return err
}

func (s *service) processEmbedding(id uint) {
	if s.embedder != nil {
		if err := s.embedder.ProcessArea(id); err != nil {
			log.Printf("area %d embedding pending: %v", id, err)
		}
	}
}

func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}
