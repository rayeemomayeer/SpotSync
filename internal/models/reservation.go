package models

import "time"

// Reservation maps to the graded reservations table.
type Reservation struct {
	ID            uint       `gorm:"primaryKey"`
	UserID        uint       `gorm:"not null;index:idx_reservations_user_id"`
	ZoneID        uint       `gorm:"not null;index:idx_reservations_zone_id_status,priority:1"`
	SpotID        *uint      `gorm:"index"`
	LicensePlate  string     `gorm:"size:15;not null"`
	Status        string     `gorm:"not null;default:active;index:idx_reservations_zone_id_status,priority:2"`
	DemoExpiresAt   *time.Time `gorm:"index:idx_reservations_demo_expires_at"`
	StartTime       *time.Time
	EndTime         *time.Time `gorm:"index:idx_reservations_end_time_active"`
	IdempotencyKey  *string    `gorm:"size:128;uniqueIndex:idx_reservations_idempotency_key"`
	Version         int        `gorm:"not null;default:0"`
	OrganizationID  *uint      `gorm:"index"`
	DemoSessionID   *string    `gorm:"size:64;index"`
	CreatedAt       time.Time  `gorm:"not null"`
	UpdatedAt     time.Time  `gorm:"not null"`

	User User        `gorm:"foreignKey:UserID;references:ID"`
	Zone ParkingZone `gorm:"foreignKey:ZoneID;references:ID"`
	Spot *ParkingSpot `gorm:"foreignKey:SpotID;references:ID"`
}

// TableName returns the graded table name.
func (Reservation) TableName() string {
	return "reservations"
}
