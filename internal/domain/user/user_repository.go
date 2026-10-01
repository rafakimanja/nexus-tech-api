package user

import "gorm.io/gorm"

type UserRepository interface {
	FindByID(id int) (User, error)
	FindByEmail(email string) (User, error)
	FindAll() ([]User, error)
	Create(newUser User) (uint, error)
	Delete(id int) error
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

// Delete implements [UserRepository].
func (r *repository) Delete(id int) error {
	result := r.db.Delete(&User{}, id)
	return result.Error
}

// FindAll implements [UserRepository].
func (r *repository) FindAll() ([]User, error) {
	var users []User
	results := r.db.Find(&users)
	return users, results.Error
}

// FindByID implements [UserRepository].
func (r *repository) FindByID(id int) (User, error) {
	var user User
	result := r.db.First(&user, id)
	return user, result.Error
}

func (r *repository) FindByEmail(email string) (User, error) {
	var user User
	result := r.db.Where("email = ?", email).First(&user)
	return user, result.Error
}
