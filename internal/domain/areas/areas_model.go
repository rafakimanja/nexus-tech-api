package areas

import "gorm.io/gorm"

type Area struct {
	gorm.Model
	Name string `gorm:"not null"`
	Key  string `gorm:"not null"`
	Type string `gorm:"not null"`
}
