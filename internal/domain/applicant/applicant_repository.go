package applicant

import "gorm.io/gorm"

type ApplicantRepository interface {
	FindByID(id int) (Demandante, error)
	FindAll() ([]Demandante, error)
	Create(applicant Demandante) (uint, error)
	Update(id int, applicant Demandante) error
	Delete(id int) error
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

func (r *repository) Create(applicant Demandante) (uint, error) {
	err := r.db.Create(&applicant).Error
	return applicant.ID, err
}

func (r *repository) Update(id int, applicant Demandante) error {
	var existing Demandante
	if err := r.db.First(&existing, id).Error; err != nil {
		return err
	}

	applicant.ID = existing.ID
	return r.db.Model(&existing).
		Select("Phone", "Type", "CpfCnpj", "Adress", "Strikes").
		Updates(&applicant).Error
}

func (r *repository) Delete(id int) error {
	return r.db.Delete(&Demandante{}, id).Error
}
