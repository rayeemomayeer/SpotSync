package models

import "time"

type OutboxEvent struct {
	ID             uint       `gorm:"primaryKey"`
	AggregateType  string     `gorm:"size:64;not null"`
	AggregateID    uint       `gorm:"not null"`
	EventType      string     `gorm:"size:64;not null"`
	Payload        []byte     `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt      time.Time  `gorm:"not null"`
	ProcessedAt    *time.Time
	Attempts       int        `gorm:"not null;default:0"`
	LastError      *string
	DeadLetteredAt *time.Time
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
