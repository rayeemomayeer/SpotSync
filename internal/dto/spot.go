package dto

import "time"

type SpotResponse struct {
	ID        uint      `json:"id"`
	ZoneID    uint      `json:"zone_id"`
	Label     string    `json:"label"`
	PosX      float64   `json:"pos_x"`
	PosY      float64   `json:"pos_y"`
	Status    string    `json:"status"`
	Occupied  bool      `json:"occupied"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateSpotRequest struct {
	Status string `json:"status" validate:"required,oneof=available unavailable"`
}
