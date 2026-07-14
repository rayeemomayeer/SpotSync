package domain

import (
	"errors"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/models"
)

func TestValidateDemoZoneWrite(t *testing.T) {
	sid := "sess-1"
	other := "sess-2"
	zoneDemo := &models.ParkingZone{IsDemo: true}
	zoneOwned := &models.ParkingZone{DemoSessionID: &sid}
	zoneOther := &models.ParkingZone{DemoSessionID: &other}
	zoneLive := &models.ParkingZone{}

	if err := ValidateDemoZoneWrite(zoneDemo, sid); err != nil {
		t.Fatalf("demo zone: %v", err)
	}
	if err := ValidateDemoZoneWrite(zoneOwned, sid); err != nil {
		t.Fatalf("owned zone: %v", err)
	}
	if ValidateDemoZoneWrite(zoneOther, sid) == nil {
		t.Fatal("expected forbidden for other session")
	}
	if ValidateDemoZoneWrite(zoneLive, sid) == nil {
		t.Fatal("expected forbidden for live zone")
	}
}

func TestValidateDemoZoneVisible(t *testing.T) {
	sid := "sess-1"
	if err := ValidateDemoZoneVisible(&models.ParkingZone{IsDemo: true}, sid); err != nil {
		t.Fatalf("is_demo zone must be visible: %v", err)
	}
	// Regression: zero-value IsDemo (as when SELECT omits is_demo) hides the zone.
	if !errors.Is(ValidateDemoZoneVisible(&models.ParkingZone{ID: 4}, sid), ErrNotFound) {
		t.Fatal("expected not found when is_demo missing/false")
	}
}
