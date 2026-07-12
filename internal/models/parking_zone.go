package models

import "time"

// ParkingZone maps to the graded parking_zones table.
type ParkingZone struct {
	ID             uint      `gorm:"primaryKey"`
	Name           string    `gorm:"not null"`
	Type           string    `gorm:"column:type;not null"`
	TotalCapacity  int       `gorm:"not null"`
	PricePerHour   float64   `gorm:"type:numeric(10,2);not null"`
	OrganizationID *uint     `gorm:"index"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

// TableName returns the graded table name.
func (ParkingZone) TableName() string {
	return "parking_zones"
}
