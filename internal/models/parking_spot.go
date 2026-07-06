package models

import "time"

const (
	SpotStatusAvailable   = "available"
	SpotStatusUnavailable = "unavailable"
)

type ParkingSpot struct {
	ID        uint      `gorm:"primaryKey"`
	ZoneID    uint      `gorm:"not null;index:idx_parking_spots_zone_id"`
	Label     string    `gorm:"size:20;not null"`
	PosX      float64   `gorm:"not null"`
	PosY      float64   `gorm:"not null"`
	Status    string    `gorm:"not null;default:available"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Zone ParkingZone `gorm:"foreignKey:ZoneID;references:ID"`
}

func (ParkingSpot) TableName() string {
	return "parking_spots"
}
