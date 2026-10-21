package demand

import (
	"api/internal/domain/applicant"
	"api/internal/domain/areas"
	"api/internal/domain/teacher"
	"fmt"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type Demanda struct {
	gorm.Model
	DemandanteID    uint             `gorm:"not null"`
	ProfessorID     uint             `gorm:"not null"`
	AreaID          uint             `gorm:"not null"`
	Title           string           `gorm:"not null"`
	Description     string           `gorm:"not null"`
	Status          string           `gorm:"not null"`
	IsCancelled     bool             `gorm:"not null"`
	Embedding       *pgvector.Vector `gorm:"type:vector(1024)"`
	ReasonCancelled string

	Professor  teacher.Professor    `gorm:"foreignKey:ProfessorID;references:ID"`
	Demandante applicant.Demandante `gorm:"foreignKey:DemandanteID;references:ID"`
	Areas      areas.Area           `gorm:"foreignKey:AreaID;references:ID"`
}

func(d Demanda) ToEmbeddingText() string {
	return fmt.Sprintf(
		"Titulo: %s\nDescrição: %s",
		d.Title,
		d.Description,
	)
}