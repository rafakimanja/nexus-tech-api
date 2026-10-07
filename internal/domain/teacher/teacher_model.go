package teacher

import "api/internal/domain/user"

type Professor struct {
	ID        uint `gorm:"primaryKey"`
	Siape     uint `gorm:"not null"`
	Is_active bool `gorm:"not null"`

	User user.User `gorm:"foreignKey:ID;references:ID"`
}
