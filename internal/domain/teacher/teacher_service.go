package teacher

import (
	"api/internal/domain/areas_teacher"
	"api/internal/domain/user"
	"fmt"
)

type TeacherService interface {
	GetAll() ([]Professor, error)
	GetByID(id int) (Professor, error)
	Create(professor Professor) (uint, error)
	Update(id int, professor Professor) error
	UpdateAreas(id int, areas_id uint, tipe string) error
	DeleteByID(id int) error
}

type service struct {
	repo                TeacherRepository
	userRepo            user.UserRepository
	areasTeacherService areas_teacher.AreasTeacherService
}

func NewService(repo TeacherRepository, userRepo user.UserRepository, areasTeacherService areas_teacher.AreasTeacherService) TeacherService {
	return &service{
		repo:                repo,
		userRepo:            userRepo,
		areasTeacherService: areasTeacherService,
	}
}

func (s *service) GetAll() ([]Professor, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Professor, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(professor Professor) (uint, error) {
	user, err := s.userRepo.FindByID(int(professor.ID))
	if err != nil {
		return 0, err
	}
	if user.Role != "professor" {
		return 0, fmt.Errorf("user is not a professor")
	}
	return s.repo.Create(professor)
}

func (s *service) Update(id int, professor Professor) error {
	return s.repo.Update(id, professor)
}

func (s *service) UpdateAreas(id int, areas_id uint, tipe string) error {
	association := areas_teacher.AssociationInput{
		AreasID: areas_id,
		Type:    tipe,
	}

	return s.areasTeacherService.ReplaceByProfessorID(id, []areas_teacher.AssociationInput{association})
}

func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}
