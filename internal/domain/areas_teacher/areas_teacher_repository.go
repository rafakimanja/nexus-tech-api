package areas_teacher

import (
	"api/internal/domain/areas"

	"gorm.io/gorm"
)

type AreasTeacherRepository interface {
	FindByProfessorID(professorID int) ([]ProfessorArea, error)
	ReplaceByProfessorID(professorID int, associations []AssociationInput) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) AreasTeacherRepository {
	return &repository{db: db}
}

func (r *repository) FindByProfessorID(professorID int) ([]ProfessorArea, error) {
	var associations []ProfessorArea
	err := r.db.Preload("Area").Where("professor_id = ?", professorID).Find(&associations).Error
	return associations, err
}

func (r *repository) ReplaceByProfessorID(professorID int, associations []AssociationInput) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("professors").Where("id = ?", professorID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		areaIDs := make([]uint, 0, len(associations))
		for _, association := range associations {
			areaIDs = append(areaIDs, association.AreasID)
		}
		if len(areaIDs) > 0 {
			if err := tx.Model(&areas.Area{}).Where("id IN ?", areaIDs).Count(&count).Error; err != nil {
				return err
			}
			if count != int64(len(areaIDs)) {
				return gorm.ErrRecordNotFound
			}
		}

		if err := tx.Unscoped().Where("professor_id = ?", professorID).Delete(&ProfessorArea{}).Error; err != nil {
			return err
		}
		links := make([]ProfessorArea, 0, len(associations))
		for _, association := range associations {
			links = append(links, ProfessorArea{
				ProfessorID: uint(professorID),
				AreasID:     association.AreasID,
				Type:        association.Type,
			})
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Create(&links).Error
	})
}
