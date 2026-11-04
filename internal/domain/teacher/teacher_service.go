package teacher

type TeacherService interface {
	GetAll() ([]Professor, error)
	GetByID(id int) (Professor, error)
}

type service struct{ repo TeacherRepository }

func NewService(repo TeacherRepository) TeacherService {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]Professor, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Professor, error) {
	return s.repo.FindByID(id)
}
