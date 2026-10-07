package applicant

import "api/internal/domain/user"

type Demandante struct {
	ID      uint `gorm:"primaryKey"`
	Phone   string
	Type    string    `gorm:"not null"`
	CpfCnpj string    `json:"cpf_cnoj" gorm:"not null"`
	Adress  string    `gorm:"not null"`
	Strikes uint      `gorm:"not null"`
	User    user.User `gorm:"foreignKey:ID;references:ID"`
}
