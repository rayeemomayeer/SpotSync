package dto

import (
	"github.com/rayeemomayeer/SpotSync/internal/models"
)

// UserFromModel maps a GORM user to the wire DTO (password excluded).
func UserFromModel(u models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ZoneFromModel maps a zone model to the wire DTO with computed availability.
func ZoneFromModel(z models.ParkingZone, availableSpots int) ZoneResponse {
	return ZoneResponse{
		ID:             z.ID,
		Name:           z.Name,
		Type:           z.Type,
		TotalCapacity:  z.TotalCapacity,
		PricePerHour:   z.PricePerHour,
		AvailableSpots: availableSpots,
		CreatedAt:      z.CreatedAt,
		UpdatedAt:      z.UpdatedAt,
	}
}

// ReservationFromModel maps a reservation model to the wire DTO.
func ReservationFromModel(r models.Reservation) ReservationResponse {
	resp := ReservationResponse{
		ID:           r.ID,
		UserID:       r.UserID,
		ZoneID:       r.ZoneID,
		LicensePlate: r.LicensePlate,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}

	if r.Zone.ID != 0 {
		zone := ZoneFromModel(r.Zone, 0)
		resp.Zone = &zone
	}
	if r.User.ID != 0 {
		user := UserFromModel(r.User)
		resp.User = &user
	}

	return resp
}
