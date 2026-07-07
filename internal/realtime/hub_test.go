package realtime

import (
	"testing"
	"time"
)

func TestHubPublishSubscribe(t *testing.T) {
	h := NewHub()
	ch := h.Subscribe(1)
	defer h.Unsubscribe(1, ch)

	h.Publish(ZoneEvent{
		Type:          EventSpotReserved,
		ZoneID:        1,
		SpotID:        9,
		UserID:        2,
		ReservationID: 42,
		LicensePlate:  "ABC-1234",
	})

	select {
	case ev := <-ch:
		if ev.SpotID != 9 || ev.Type != EventSpotReserved {
			t.Fatalf("unexpected event: %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestHubZoneIsolation(t *testing.T) {
	h := NewHub()
	ch1 := h.Subscribe(1)
	ch2 := h.Subscribe(2)
	defer h.Unsubscribe(1, ch1)
	defer h.Unsubscribe(2, ch2)

	h.Publish(ZoneEvent{Type: EventSpotReserved, ZoneID: 1, SpotID: 1})

	select {
	case <-ch1:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("zone 1 should receive event")
	}

	select {
	case <-ch2:
		t.Fatal("zone 2 should not receive zone 1 event")
	case <-time.After(100 * time.Millisecond):
	}
}
