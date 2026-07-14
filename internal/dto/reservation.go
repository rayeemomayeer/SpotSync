package dto

import "time"

type CreateReservationRequest struct {
	ZoneID       uint   `json:"zone_id" validate:"required,gt=0"`
	LicensePlate string `json:"license_plate" validate:"required,min=1,max=15"`
	SpotID       *uint  `json:"spot_id" validate:"omitempty,gt=0"`
	StartTime    *time.Time `json:"start_time" validate:"omitempty"`
	EndTime      *time.Time `json:"end_time" validate:"omitempty"`
}

type ReservationResponse struct {
	ID            uint          `json:"id"`
	UserID        uint          `json:"user_id"`
	ZoneID        uint          `json:"zone_id"`
	SpotID        *uint         `json:"spot_id,omitempty"`
	LicensePlate  string        `json:"license_plate"`
	Status        string        `json:"status"`
	PaymentStatus *string       `json:"payment_status,omitempty"`
	PaymentID     *uint         `json:"payment_id,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Zone          *ZoneResponse `json:"zone,omitempty"`
	User          *UserResponse `json:"user,omitempty"`
	Spot          *SpotResponse `json:"spot,omitempty"`
}

type PaginationQuery struct {
	Page  int `query:"page" validate:"omitempty,min=1"`
	Limit int `query:"limit" validate:"omitempty,min=1,max=100"`
}
