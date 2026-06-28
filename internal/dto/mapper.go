package dto

import (
	"github.com/rayeemomayeer/SpotSync/internal/models"
)

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
