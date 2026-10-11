package areas

type AreasService interface {
	GetAll() ([]Area, error)
	GetByID(id int) (Area, error)
	Create(area Area) (uint, error)
	Update(id int, area Area) error
	DeleteByID(id int) error
}

type service struct {
	repo AreasRepository
}

func NewService(repo AreasRepository) AreasService {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]Area, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Area, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(area Area) (uint, error) {
	return s.repo.Create(area)
}

func (s *service) Update(id int, area Area) error {
	return s.repo.Update(id, area)
}

func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}
