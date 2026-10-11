package teacher

import (
	"api/internal/domain/areas"
	"api/internal/domain/user"
)

type Professor struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	Siape     uint `json:"siape" gorm:"not null"`
	Is_active bool `json:"is_active" gorm:"not null"`

	User  user.User    `gorm:"foreignKey:ID;references:ID"`
	Areas []areas.Area `gorm:"many2many:professor_areas;"`
}

type ProfessorDTO struct {
    ID        uint       `json:"id"`
    Siape     uint       `json:"siape"`
    Is_active bool       `json:"is_active"`
    Name      string     `json:"name"`
    Role      string     `json:"role"`
    Email     string     `json:"email"`
    Foto_url  *string    `json:"foto_url"`
    Areas     []areas.AreaDTO  `json:"areas"`
}

func (p *Professor) NewDTO() ProfessorDTO {

	areasDTO := make([]areas.AreaDTO, 0, len(p.Areas))

    for _, area := range p.Areas {
        areasDTO = append(areasDTO, area.NewDTO())
    }

	return ProfessorDTO{
		ID: p.ID,
		Siape: p.Siape,
		Is_active: p.Is_active,
		Name: p.User.Name,
		Role: p.User.Role,
		Email: p.User.Email,
		Foto_url: p.User.Foto_url,
	}
}