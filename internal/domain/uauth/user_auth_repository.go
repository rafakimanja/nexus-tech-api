package uauth

import "gorm.io/gorm"

type UAuthRepository interface {
	Create (newAuth UserAuth) (uint, error)
	FindByID(id int) (UserAuth, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) UAuthRepository {
	return &repository{db: db}
}

func (r *repository) Create(newAuth UserAuth) (uint, error){
	result := r.db.Create(&newAuth)
	return newAuth.ID, result.Error
}

func (r *repository) FindByID(id int) (UserAuth, error){
	var uAuth UserAuth
	result := r.db.First(&uAuth, id)
	return uAuth, result.Error
}