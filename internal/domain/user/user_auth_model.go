package user

import "gorm.io/gorm"

type UserAuth struct {
	gorm.Model
	UserID         uint
	Provider       string
	ProviderUserID string
	User           User `gorm:"foreignKey:UserID"`
}
