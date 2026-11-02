package applicant

import (
	"api/internal/domain/user"
	"time"
)

type Demandante struct {
	ID      uint `gorm:"primaryKey"`
	Phone   string
	Type    string    `gorm:"not null"`
	CpfCnpj string    `json:"cpf_cnoj" gorm:"not null"`
	Adress  string    `gorm:"not null"`
	Strikes uint      `gorm:"not null"`
	User    user.User `gorm:"foreignKey:ID;references:ID"`
}

type DemandanteDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	PhotoURL  *string   `json:"photo_url"`
	CreatedAt time.Time `json:"created_at"`
	Phone     string    `json:"phone"`
	Type      string    `json:"type"`
	CpfCnpj   string    `json:"cpf_cnpj"`
	Address   string    `json:"address"`
	Strikes   uint      `json:"strikes"`
	Blocked   bool      `json:"blocked"`
}

func (d Demandante) NewDTO() DemandanteDTO {
	return DemandanteDTO{ID: d.ID, Name: d.User.Name, Email: d.User.Email, Role: d.User.Role, PhotoURL: d.User.Foto_url, CreatedAt: d.User.CreatedAt, Phone: d.Phone, Type: d.Type, CpfCnpj: d.CpfCnpj, Address: d.Adress, Strikes: d.Strikes, Blocked: d.Strikes >= 3}
}
