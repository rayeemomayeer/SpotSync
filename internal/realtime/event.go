package realtime

// EventType is a zone availability stream event name (SSE).
type EventType string

const (
	EventSpotReserved EventType = "spot_reserved"
	EventSpotReleased EventType = "spot_released"
)

// ZoneEvent is pushed to SSE subscribers for a parking zone.
type ZoneEvent struct {
	Type          EventType `json:"type"`
	ZoneID        uint      `json:"zone_id"`
	SpotID        uint      `json:"spot_id"`
	UserID        uint      `json:"user_id"`
	ReservationID uint      `json:"reservation_id"`
	LicensePlate  string    `json:"license_plate,omitempty"`
}
