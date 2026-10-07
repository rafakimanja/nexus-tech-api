package demand

import (
	"api/internal/domain/applicant"
	"api/internal/domain/areas"
	"api/internal/domain/teacher"

	"gorm.io/gorm"
)

type Demanda struct {
	gorm.Model
	DemandanteID    uint   `gorm:"not null"`
	ProfessorID     uint   `gorm:"not null"`
	AreaID          uint   `gorm:"not null"`
	Title           string `gorm:"not null"`
	Description     string `gorm:"not null"`
	Status          string `gorm:"not null"`
	IsCancelled     bool   `gorm:"not null"`
	ReasonCancelled string

	Professor  teacher.Professor    `gorm:"foreignKey:ProfessorID;references:ID"`
	Demandante applicant.Demandante `gorm:"foreignKey:DemandanteID;references:ID"`
	Areas      areas.Area           `gorm:"foreignKey:AreaID;references:ID"`
}
