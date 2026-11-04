package teacher

import (
	"api/internal/domain/areas_teacher"
	"api/internal/domain/user"
	"time"
)

type Professor struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Siape     string `json:"siape" gorm:"type:text;not null"`
	Is_active bool   `json:"is_active" gorm:"not null"`

	User           user.User                     `gorm:"foreignKey:ID;references:ID"`
	ProfessorAreas []areas_teacher.ProfessorArea `gorm:"foreignKey:ProfessorID;references:ID" json:"areas"`
}

type ProfessorAreaDTO struct {
	AreasID uint   `json:"areas_id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

type ProfessorDTO struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Role      string             `json:"role"`
	Email     string             `json:"email"`
	FotoURL   *string            `json:"photo_url"`
	CreatedAt time.Time          `json:"created_at"`
	Siape     string             `json:"siape"`
	IsActive  bool               `json:"is_active"`
	Areas     []ProfessorAreaDTO `json:"areas"`
}

func (p Professor) NewDTO() ProfessorDTO {
	areas := make([]ProfessorAreaDTO, 0, len(p.ProfessorAreas))
	for _, association := range p.ProfessorAreas {
		areas = append(areas, ProfessorAreaDTO{
			AreasID: association.AreasID,
			Type:    association.Type,
			Name:    association.Area.Name,
		})
	}

	return ProfessorDTO{
		ID:        p.ID,
		Name:      p.User.Name,
		Role:      p.User.Role,
		Email:     p.User.Email,
		FotoURL:   p.User.Foto_url,
		CreatedAt: p.User.CreatedAt,
		Siape:     p.Siape,
		IsActive:  p.Is_active,
		Areas:     areas,
	}
}
