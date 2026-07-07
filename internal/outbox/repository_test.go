package outbox_test

import (
	"encoding/json"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
)

func TestReservationEventPayloadJSON(t *testing.T) {
	payload := domain.ReservationEventPayload{
		ReservationID: 1,
		ZoneID:        2,
		UserID:        3,
		LicensePlate:  "ABC",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded domain.ReservationEventPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ReservationID != 1 || decoded.ZoneID != 2 {
		t.Fatalf("unexpected payload: %+v", decoded)
	}
}
