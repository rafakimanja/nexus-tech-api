package user

type UserService interface {
	GetAll() ([]User, error)
	GetByID(id int) (User, error)
	GetByEmail(email string) (User, error)
	Create(newUser User) (uint, error)
	DeleteByID(id int) error
}

type service struct {
	repo UserRepository
}

// Create implements [UserService].
func (s *service) Create(newUser User) (uint, error) {
	return s.repo.Create(newUser)
}

// DeleteByID implements [UserService].
func (s *service) DeleteByID(id int) error {
	return s.repo.Delete(id)
}

// GetAll implements [UserService].
func (s *service) GetAll() ([]User, error) {
	return s.repo.FindAll()
}

// GetByID implements [UserService].
func (s *service) GetByID(id int) (User, error) {
	return s.repo.FindByID(id)
}

func (s *service) GetByEmail(email string) (User, error){
	return s.repo.FindByEmail(email)
}

func NewService(repo UserRepository) UserService {
	return &service{repo: repo}
}
