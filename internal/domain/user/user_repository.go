package user

import "gorm.io/gorm"

type UserRepository interface {
	FindAll() ([]User, error)
	Create(newUser User) (uint, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) UserRepository {
	return &repository{db: db}
}

// Create implements [UserRepository].
func (r *repository) Create(newUser User) (uint, error) {
	result := r.db.Create(&newUser)
	return newUser.ID, result.Error
}

// FindAll implements [UserRepository].
func (r *repository) FindAll() ([]User, error) {
	var users []User
	results := r.db.Find(&users)
	return users, results.Error
}
