package models

// User roles (graded contract).
const (
	RoleDriver = "driver"
	RoleAdmin  = "admin"
)

// Parking zone types (graded contract).
const (
	ZoneTypeGeneral    = "general"
	ZoneTypeEVCharging = "ev_charging"
	ZoneTypeCovered    = "covered"
)

// Reservation lifecycle states (graded contract).
// Phase 0 uses active/cancelled; completed is used from Phase 1 expiry.
const (
	ReservationStatusActive    = "active"
	ReservationStatusCompleted = "completed"
	ReservationStatusCancelled = "cancelled"
)
