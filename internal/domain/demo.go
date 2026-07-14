package domain

import (
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/models"
)

// ValidateDemoZoneWrite ensures demo-mode mutations target demo or session-owned zones.
func ValidateDemoZoneWrite(zone *models.ParkingZone, sessionID string) error {
	sid := strings.TrimSpace(sessionID)
	if sid == "" {
		return ErrForbidden
	}
	if zone.IsDemo {
		return nil
	}
	if zone.DemoSessionID != nil && strings.TrimSpace(*zone.DemoSessionID) == sid {
		return nil
	}
	return ErrForbidden
}

// ValidateDemoZoneVisible ensures a zone is visible in demo mode.
func ValidateDemoZoneVisible(zone *models.ParkingZone, sessionID string) error {
	sid := strings.TrimSpace(sessionID)
	if sid == "" {
		return ErrNotFound
	}
	if zone.IsDemo {
		return nil
	}
	if zone.DemoSessionID != nil && strings.TrimSpace(*zone.DemoSessionID) == sid {
		return nil
	}
	return ErrNotFound
}

// ValidateLiveZoneVisible hides sandbox-only zones outside demo mode.
func ValidateLiveZoneVisible(zone *models.ParkingZone) error {
	if zone.DemoSessionID != nil && strings.TrimSpace(*zone.DemoSessionID) != "" && !zone.IsDemo {
		return ErrNotFound
	}
	return nil
}
