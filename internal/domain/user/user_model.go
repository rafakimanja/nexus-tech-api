package user

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Role     string
	Email    string
	Foto_url *string
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	fmt.Println("EMAIL:", u.Email)
	partes := strings.Split(u.Email, "@")
	if len(partes) != 2 {
		return fmt.Errorf("e-mail malformado")
	}
	dominio := strings.ToLower(partes[1])

	if u.Email == "rafaens.dev@gmail.com" {
		u.Role = "admin"
	}

	switch {
	case dominio == "gmail.com" || dominio == "googlemail.com":
		u.Role = "demandante"
	case dominio == "osorio.ifrs.edu.br" || dominio == "aluno.osorio.ifrs.edu.br":
		u.Role = "professor"
	}

	return nil
}
