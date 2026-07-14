package domain

import (
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
