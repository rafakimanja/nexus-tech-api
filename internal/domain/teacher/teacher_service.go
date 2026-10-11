package teacher

import (
	"api/internal/domain/user"
	"fmt"
)

type TeacherService interface {
	GetAll() ([]Professor, error)
	GetByID(id int) (Professor, error)
	Create(professor Professor) (uint, error)
	Update(id int, professor Professor) error
	UpdateAreas(id int, areasIDs []uint) error
	DeleteByID(id int) error
}

type service struct {
	repo TeacherRepository
	userRepo user.UserRepository
}

func NewService(repo TeacherRepository, userRepo user.UserRepository) TeacherService {
	return &service{
		repo: repo,
		userRepo: userRepo,
	}
}

func (s *service) GetAll() ([]Professor, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Professor, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(professor Professor) (uint, error) {
	user, erro := s.userRepo.FindByID(int(professor.ID))

	if erro != nil {
		return 0, erro
	}

	if user.Role != "professor" {
		return 0, fmt.Errorf("User dosen't a teacher")
	}

	return s.repo.Create(professor)
}

func (s *service) Update(id int, professor Professor) error {
	return s.repo.Update(id, professor)
}

func (s *service) UpdateAreas(id int, areasIDs []uint) error {
	return nil
}

func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}
