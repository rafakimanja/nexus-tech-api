package applicant

type ApplicantService interface {
	GetAll() ([]Demandante, error)
	GetByID(id int) (Demandante, error)
	Create(applicant Demandante) (uint, error)
	Update(id int, applicant Demandante) error
	DeleteByID(id int) error
}

type service struct {
	repo ApplicantRepository
}

func NewService(repo ApplicantRepository) ApplicantService {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]Demandante, error) {
	return s.repo.FindAll()
}

func (s *service) GetByID(id int) (Demandante, error) {
	return s.repo.FindByID(id)
}

func (s *service) Create(applicant Demandante) (uint, error) {
	return s.repo.Create(applicant)
}

func (s *service) Update(id int, applicant Demandante) error {
	return s.repo.Update(id, applicant)
}

func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}
