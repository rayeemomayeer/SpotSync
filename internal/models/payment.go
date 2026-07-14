package models

import "time"

type Payment struct {
	ID                    uint      `gorm:"primaryKey"`
	ReservationID         *uint     `gorm:"index"`
	UserID                uint      `gorm:"not null;index"`
	ZoneID                uint      `gorm:"not null"`
	StripePaymentIntentID string    `gorm:"size:255;not null;uniqueIndex"`
	AmountCents           int       `gorm:"not null"`
	Currency              string    `gorm:"size:3;not null;default:usd"`
	Status                string    `gorm:"size:20;not null;default:pending"`
	CreatedAt             time.Time `gorm:"not null"`
	UpdatedAt             time.Time `gorm:"not null"`
}

func (Payment) TableName() string {
	return "payments"
}

const (
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

type Refund struct {
	ID             uint      `gorm:"primaryKey"`
	PaymentID      uint      `gorm:"not null;index"`
	StripeRefundID *string   `gorm:"size:255;uniqueIndex"`
	AmountCents    int       `gorm:"not null"`
	Status         string    `gorm:"size:20;not null;default:pending"`
	CreatedAt      time.Time `gorm:"not null"`
}

func (Refund) TableName() string {
	return "refunds"
}

const (
	RefundStatusPending   = "pending"
	RefundStatusSucceeded = "succeeded"
	RefundStatusFailed    = "failed"
)
