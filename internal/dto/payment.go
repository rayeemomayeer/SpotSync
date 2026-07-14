package dto

import "time"

type CreatePaymentRequest struct {
	ReservationID         uint   `json:"reservation_id" validate:"required,gt=0"`
	StripePaymentIntentID string `json:"stripe_payment_intent_id" validate:"required,min=3,max=255"`
	AmountCents           int    `json:"amount_cents" validate:"required,gt=0"`
	Currency              string `json:"currency" validate:"omitempty,len=3"`
}

type PaymentResponse struct {
	ID                    uint      `json:"id"`
	ReservationID         *uint     `json:"reservation_id,omitempty"`
	UserID                uint      `json:"user_id"`
	ZoneID                uint      `json:"zone_id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id"`
	AmountCents           int       `json:"amount_cents"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CreateRefundRequest struct {
	StripeRefundID string `json:"stripe_refund_id" validate:"required,min=3,max=255"`
	AmountCents    int    `json:"amount_cents" validate:"required,gt=0"`
}

type RefundResponse struct {
	ID             uint      `json:"id"`
	PaymentID      uint      `json:"payment_id"`
	StripeRefundID *string   `json:"stripe_refund_id,omitempty"`
	AmountCents    int       `json:"amount_cents"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
