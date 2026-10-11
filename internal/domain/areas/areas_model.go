package areas

import (
	"gorm.io/gorm"
)

type Area struct {
	gorm.Model
	Name string `gorm:"not null"`
	Key  string `gorm:"not null"`
	Type string `gorm:"not null"`
}

type AreaDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
	Type string `json:"type"`
}

func (a *Area) NewDTO() AreaDTO {
	return AreaDTO{
		ID:   a.ID,
		Name: a.Name,
		Key:  a.Key,
		Type: a.Type,
	}
}
