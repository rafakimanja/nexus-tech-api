package teacher

type TeacherService interface {
	GetAll() ([]Professor, error)
	GetByID(id int) (Professor, error)
	Create(professor Professor) (uint, error)
	Update(id int, professor Professor) error
	DeleteByID(id int) error
}

type service struct {
	repo TeacherRepository
}

func NewService(repo TeacherRepository) TeacherService {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]Professor, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Professor, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(professor Professor) (uint, error) {
	return s.repo.Create(professor)
}

func (s *service) Update(id int, professor Professor) error {
	return s.repo.Update(id, professor)
}

func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}
