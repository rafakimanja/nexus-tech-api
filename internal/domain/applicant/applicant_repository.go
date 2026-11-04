package applicant

import "gorm.io/gorm"

type ApplicantRepository interface {
	FindByID(id int) (Demandante, error)
	FindAll() ([]Demandante, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) ApplicantRepository {
	return &repository{db: db}
}

func (r *repository) FindByID(id int) (Demandante, error) {
	var applicant Demandante
	err := r.db.Preload("User").First(&applicant, id).Error
	return applicant, err
}

func (r *repository) FindAll() ([]Demandante, error) {
	var applicants []Demandante
	err := r.db.Preload("User").Find(&applicants).Error
	return applicants, err
}
