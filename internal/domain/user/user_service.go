package user

type UserService interface {
	GetAll() ([]User, error)
	Create(newUser User) (uint, error)
}

type service struct {
	repo UserRepository
}

// Create implements [UserService].
func (s *service) Create(newUser User) (uint, error) {
	return s.repo.Create(newUser)
}

// GetAll implements [UserService].
func (s *service) GetAll() ([]User, error) {
	return s.repo.FindAll()
}

func NewService(repo UserRepository) UserService {
	return &service{repo: repo}
}
