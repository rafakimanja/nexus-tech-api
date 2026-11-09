package notification

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Notification is the durable inbox. Domains create it inside their transaction.
type Notification struct {
	ID uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"not null;index"`
	Type string `json:"type" gorm:"not null"`
	Title string `json:"title" gorm:"not null"`
	Body string `json:"body" gorm:"not null"`
	ResourceType string `json:"resource_type" gorm:"not null"`
	ResourceID uint `json:"resource_id" gorm:"not null"`
	EventKey *string `json:"-" gorm:"uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	ReadAt *time.Time `json:"read_at"`
}

type Device struct {
	ID uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"-" gorm:"not null;index"`
	Token string `json:"-" gorm:"not null;uniqueIndex;size:2048"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Delivery struct {
	ID uint `gorm:"primaryKey"`
	NotificationID uint `gorm:"not null;uniqueIndex:notification_device"`
	DeviceID uint `gorm:"not null;uniqueIndex:notification_device"`
	Attempts int
	NextAttemptAt time.Time `gorm:"index"`
	FinishedAt *time.Time
}

// Enqueue must receive the same transaction used by the calling domain.
// A unique event key makes repeatable producers (such as recommendations) safe.
func Enqueue(tx *gorm.DB, n *Notification) error {
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(n)
	if result.Error != nil || result.RowsAffected == 0 { return result.Error }
	return tx.Exec(`INSERT INTO deliveries (notification_id, device_id, attempts, next_attempt_at)
		SELECT ?, id, 0, ? FROM devices WHERE user_id = ?
		ON CONFLICT (notification_id, device_id) DO NOTHING`, n.ID, time.Now().UTC(), n.UserID).Error
}
