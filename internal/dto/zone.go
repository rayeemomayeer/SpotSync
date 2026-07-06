package dto

import "time"

type CreateZoneRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=255"`
	Type          string  `json:"type" validate:"required,oneof=general ev_charging covered"`
	TotalCapacity int     `json:"total_capacity" validate:"required,gt=0"`
	PricePerHour  float64 `json:"price_per_hour" validate:"required,gt=0"`
}

type UpdateZoneRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=255"`
	Type          string  `json:"type" validate:"required,oneof=general ev_charging covered"`
	TotalCapacity int     `json:"total_capacity" validate:"required,gt=0"`
	PricePerHour  float64 `json:"price_per_hour" validate:"required,gt=0"`
}

type ZoneListQuery struct {
	Type  string `query:"type" validate:"omitempty,oneof=general ev_charging covered"`
	Q     string `query:"q" validate:"omitempty,max=255"`
	Sort  string `query:"sort" validate:"omitempty,oneof=available_spots price_per_hour name"`
	Order string `query:"order" validate:"omitempty,oneof=asc desc"`
}

type ZoneResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	TotalCapacity  int       `json:"total_capacity"`
	PricePerHour   float64   `json:"price_per_hour"`
	AvailableSpots int       `json:"available_spots"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
