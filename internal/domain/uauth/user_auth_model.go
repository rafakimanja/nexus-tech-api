package uauth

import (
	"api/internal/domain/user"

	"gorm.io/gorm"
)

type UserAuth struct {
	gorm.Model
	UserID         uint
	Provider       string
	ProviderUserID string
	User           user.User `gorm:"foreignKey:UserID"`
}
