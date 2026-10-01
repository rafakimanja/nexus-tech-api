package user

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	Email    string  `json:"email" gorm:"uniqueIndex;not null"`
	Foto_url *string `json:"foto_url"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	partes := strings.Split(u.Email, "@")
	if len(partes) != 2 {
		return fmt.Errorf("e-mail malformado")
	}

	dominio := strings.ToLower(partes[1])
	email := strings.ToLower(u.Email)

	switch {
	case email == "rafaens.dev@gmail.com":
		u.Role = "admin"

	case dominio == "gmail.com" || dominio == "googlemail.com":
		u.Role = "demandante"

	case dominio == "osorio.ifrs.edu.br" ||
		dominio == "aluno.osorio.ifrs.edu.br":
		u.Role = "professor"
	}

	return nil
}
