package areas_teacher

import (
	"api/internal/domain/areas"

	"gorm.io/gorm"
)

const (
	AreaTypeTechnical = "técnica"
	AreaTypeInterest  = "interesse"
)

// ProfessorArea is the explicit link between a professor and an area.
// Its default GORM table name is professor_areas.
type ProfessorArea struct {
	gorm.Model
	ProfessorID uint       `gorm:"not null;index" json:"professor_id"`
	AreasID     uint       `gorm:"not null;index" json:"areas_id"`
	Type        string     `gorm:"not null" json:"type"`
	Area        areas.Area `gorm:"foreignKey:AreasID;references:ID" json:"area,omitempty"`
}

type AssociationInput struct {
	AreasID uint   `json:"areas_id"`
	Type    string `json:"type"`
}

type ProfessorAreaDTO struct {
	AreasID uint   `json:"areas_id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}

func (pa ProfessorArea) NewDTO() ProfessorAreaDTO {
	return ProfessorAreaDTO{AreasID: pa.AreasID, Name: pa.Area.Name, Type: pa.Type}
}
