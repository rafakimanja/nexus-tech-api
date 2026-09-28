package uauth

type UAuthService interface {
	Create(newAuth UserAuth) (uint, error)
	GetByID(id int) (UserAuth, error)
}

type service struct {
	repo UAuthRepository
}

func (s *service) Create(newAuth UserAuth) (uint, error) {
	return s.repo.Create(newAuth)
}

func (s *service) GetByID(id int) (UserAuth, error){
	return s.repo.FindByID(id)
}