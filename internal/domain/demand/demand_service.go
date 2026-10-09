package demand

type DemandaService interface {
	GetAll() ([]Demanda, error)
	GetByID(id int) (Demanda, error)
	Create(demanda Demanda) (uint, error)
	Update(id int, demanda Demanda) error
	DeleteByID(id int) error
}

type service struct {
	repo DemandaRepository
}

func NewService(repo DemandaRepository) DemandaService {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]Demanda, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Demanda, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(demanda Demanda) (uint, error) {
	return s.repo.Create(demanda)
}

func (s *service) Update(id int, demanda Demanda) error {
	return s.repo.Update(id, demanda)
}

func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}
