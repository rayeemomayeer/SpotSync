package models_test

import (
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/models"
)

func TestTableNames(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"users", models.User{}.TableName(), "users"},
		{"parking_zones", models.ParkingZone{}.TableName(), "parking_zones"},
		{"reservations", models.Reservation{}.TableName(), "reservations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("TableName() = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestGradedEnumValues(t *testing.T) {
	roles := []string{models.RoleDriver, models.RoleAdmin}
	zoneTypes := []string{models.ZoneTypeGeneral, models.ZoneTypeEVCharging, models.ZoneTypeCovered}
	statuses := []string{
		models.ReservationStatusActive,
		models.ReservationStatusCompleted,
		models.ReservationStatusCancelled,
	}

	for _, role := range roles {
		if role == "" {
			t.Fatal("role constant must not be empty")
		}
	}
	for _, zt := range zoneTypes {
		if zt == "" {
			t.Fatal("zone type constant must not be empty")
		}
	}
	for _, st := range statuses {
		if st == "" {
			t.Fatal("reservation status constant must not be empty")
		}
	}
}
