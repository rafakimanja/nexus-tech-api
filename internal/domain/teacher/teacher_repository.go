package teacher

import "gorm.io/gorm"

type TeacherRepository interface {
	FindByID(id int) (Professor, error)
	FindAll() ([]Professor, error)
	Create(professor Professor) (uint, error)
	Update(id int, professor Professor) error
	Delete(id int) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) TeacherRepository {
	return &repository{db: db}
}

func (r *repository) FindByID(id int) (Professor, error) {
	var professor Professor
	err := r.db.Preload("User").Preload("ProfessorAreas.Area").First(&professor, id).Error
	return professor, err
}

func (r *repository) FindAll() ([]Professor, error) {
	var professors []Professor
	err := r.db.Preload("User").Preload("ProfessorAreas.Area").Find(&professors).Error
	return professors, err
}

func (r *repository) Create(professor Professor) (uint, error) {
	err := r.db.Create(&professor).Error
	return professor.ID, err
}

func (r *repository) Update(id int, professor Professor) error {
	var existing Professor
	if err := r.db.First(&existing, id).Error; err != nil {
		return err
	}

	professor.ID = existing.ID
	return r.db.Model(&existing).
		Select("Siape", "Is_active").
		Updates(&professor).Error
}

func (r *repository) Delete(id int) error {
	return r.db.Delete(&Professor{}, id).Error
}
