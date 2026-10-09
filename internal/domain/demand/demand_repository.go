package demand

import "gorm.io/gorm"

type DemandaRepository interface {
	FindByID(id int) (Demanda, error)
	FindAll() ([]Demanda, error)
	Create(demanda Demanda) (uint, error)
	Update(id int, demanda Demanda) error
	Delete(id int) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) DemandaRepository {
	return &repository{db: db}
}

func (r *repository) FindByID(id int) (Demanda, error) {
	var demanda Demanda
	err := r.db.Preload("Professor").Preload("Demandante").Preload("Areas").First(&demanda, id).Error
	return demanda, err
}

func (r *repository) FindAll() ([]Demanda, error) {
	var demandas []Demanda
	err := r.db.Preload("Professor").Preload("Demandante").Preload("Areas").Find(&demandas).Error
	return demandas, err
}

func (r *repository) Create(demanda Demanda) (uint, error) {
	err := r.db.Create(&demanda).Error
	return demanda.ID, err
}

func (r *repository) Update(id int, demanda Demanda) error {
	var existing Demanda
	if err := r.db.First(&existing, id).Error; err != nil {
		return err
	}

	demanda.ID = existing.ID
	return r.db.Model(&existing).
		Select("DemandanteID", "ProfessorID", "AreaID", "Title", "Description", "Status", "IsCancelled", "ReasonCancelled").
		Updates(&demanda).Error
}

func (r *repository) Delete(id int) error {
	return r.db.Delete(&Demanda{}, id).Error
}
