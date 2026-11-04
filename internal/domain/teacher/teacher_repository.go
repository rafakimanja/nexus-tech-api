package teacher

import "gorm.io/gorm"

type TeacherRepository interface {
	FindByID(id int) (Professor, error)
	FindAll() ([]Professor, error)
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
