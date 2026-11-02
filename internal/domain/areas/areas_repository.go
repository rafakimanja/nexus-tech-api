package areas

import "gorm.io/gorm"

type AreasRepository interface {
	FindByID(id int) (Area, error)
	FindAll() ([]Area, error)
	Create(area Area) (uint, error)
	Update(id int, area Area) error
	Delete(id int) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) AreasRepository {
	return &repository{db: db}
}

func (r *repository) FindByID(id int) (Area, error) {
	var area Area
	err := r.db.First(&area, id).Error
	return area, err
}

func (r *repository) FindAll() ([]Area, error) {
	var areas []Area
	err := r.db.Find(&areas).Error
	return areas, err
}

func (r *repository) Create(area Area) (uint, error) {
	area.Embedding = nil
	err := r.db.Create(&area).Error
	return area.ID, err
}

func (r *repository) Update(id int, area Area) error {
	var existing Area
	if err := r.db.First(&existing, id).Error; err != nil {
		return err
	}

	return r.db.Model(&existing).Updates(map[string]any{
		"name":      area.Name,
		"embedding": gorm.Expr("CASE WHEN name IS DISTINCT FROM ? THEN NULL ELSE embedding END", area.Name),
	}).Error
}

func (r *repository) Delete(id int) error {
	return r.db.Delete(&Area{}, id).Error
}
