package domain

const (
	EventReservationCreated   = "ReservationCreated"
	EventReservationCancelled = "ReservationCancelled"
	EventReservationExpired   = "ReservationExpired"

	AggregateReservation = "reservation"
)

type ReservationEventPayload struct {
	ReservationID uint   `json:"reservation_id"`
	ZoneID        uint   `json:"zone_id"`
	SpotID        *uint  `json:"spot_id,omitempty"`
	UserID        uint   `json:"user_id"`
	LicensePlate  string `json:"license_plate,omitempty"`
	Email         string `json:"email,omitempty"`
}
