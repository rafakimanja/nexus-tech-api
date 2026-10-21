package areas

import "github.com/pgvector/pgvector-go"

type Area struct {
	ID        uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string           `gorm:"not null" json:"name"`
	Embedding *pgvector.Vector `gorm:"type:vector(1024)"`
}