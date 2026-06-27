package models

import "time"

// Reservation maps to the graded reservations table.
type Reservation struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null;index:idx_reservations_user_id"`
	ZoneID       uint      `gorm:"not null;index:idx_reservations_zone_id_status,priority:1"`
	LicensePlate string    `gorm:"size:15;not null"`
	Status       string    `gorm:"not null;default:active;index:idx_reservations_zone_id_status,priority:2"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`

	User User        `gorm:"foreignKey:UserID;references:ID"`
	Zone ParkingZone `gorm:"foreignKey:ZoneID;references:ID"`
}

// TableName returns the graded table name.
func (Reservation) TableName() string {
	return "reservations"
}
