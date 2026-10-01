package uauth

import (
	"api/internal/domain/user"

	"gorm.io/gorm"
)

type UserAuth struct {
	gorm.Model
	UserID         uint      `json:"user_id" gorm:"not null"`
	Provider       string    `json:"provider" gorm:"not null"`
	ProviderUserID string    `json:"provider_user_id" gorm:"not null"`
	User           user.User `gorm:"foreignKey:UserID;references:ID"`
}
