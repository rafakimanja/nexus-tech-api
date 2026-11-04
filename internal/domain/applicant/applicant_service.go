package applicant

type ApplicantService interface {
	GetAll() ([]Demandante, error)
	GetByID(id int) (Demandante, error)
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
