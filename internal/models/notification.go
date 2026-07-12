package models

import "time"

type Notification struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    uint       `gorm:"not null;index"`
	Type      string     `gorm:"size:64;not null"`
	Title     string     `gorm:"size:255;not null"`
	Body      string     `gorm:"not null;default:''"`
	ReadAt    *time.Time
	Metadata  []byte     `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time  `gorm:"not null"`
}

func (Notification) TableName() string {
	return "notifications"
}
