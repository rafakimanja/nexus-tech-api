package teacher

import (
	"api/internal/domain/areas_teacher"
	"api/internal/domain/user"
)

type Professor struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	Siape     uint `json:"siape" gorm:"not null"`
	Is_active bool `json:"is_active" gorm:"not null"`

	User           user.User                     `gorm:"foreignKey:ID;references:ID"`
	ProfessorAreas []areas_teacher.ProfessorArea `gorm:"foreignKey:ProfessorID;references:ID" json:"areas"`
}

type ProfessorAreaDTO struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ProfessorDTO struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	Role     string             `json:"role"`
	Email    string             `json:"email"`
	FotoURL  *string            `json:"foto_url"`
	Siape    uint               `json:"siape"`
	IsActive bool               `json:"is_active"`
	Areas    []ProfessorAreaDTO `json:"areas"`
}

func (p Professor) NewDTO() ProfessorDTO {
	areas := make([]ProfessorAreaDTO, 0, len(p.ProfessorAreas))
	for _, association := range p.ProfessorAreas {
		areas = append(areas, ProfessorAreaDTO{
			Type: association.Type,
			Name: association.Area.Name,
		})
	}

	return ProfessorDTO{
		ID:       p.ID,
		Name:     p.User.Name,
		Role:     p.User.Role,
		Email:    p.User.Email,
		FotoURL:  p.User.Foto_url,
		Siape:    p.Siape,
		IsActive: p.Is_active,
		Areas:    areas,
	}
}
