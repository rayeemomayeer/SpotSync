package spots

import (
	"fmt"
	"math"

	"github.com/rayeemomayeer/SpotSync/internal/models"
)

const (
	defaultCols   = 6
	slotWidth     = 48.0
	slotHeight    = 80.0
	gridOriginX   = 120.0
	gridOriginY   = 100.0
)

// GridLayout generates spot positions for a zone capacity in a default grid.
func GridLayout(zoneID uint, capacity int) []models.ParkingSpot {
	if capacity < 1 {
		return nil
	}
	cols := defaultCols
	if capacity < cols {
		cols = capacity
	}
	spots := make([]models.ParkingSpot, 0, capacity)
	for i := 0; i < capacity; i++ {
		col := i % cols
		row := i / cols
		spots = append(spots, models.ParkingSpot{
			ZoneID: zoneID,
			Label:  fmt.Sprintf("A-%02d", i+1),
			PosX:   gridOriginX + float64(col)*slotWidth,
			PosY:   gridOriginY + float64(row)*slotHeight,
			Status: models.SpotStatusAvailable,
		})
	}
	return spots
}

// ShowcaseLayout returns curated coordinates for the demo EV lot (24 spots, 4x6).
func ShowcaseLayout(zoneID uint) []models.ParkingSpot {
	const capacity = 24
	cols := 6
	spots := make([]models.ParkingSpot, 0, capacity)
	for i := 0; i < capacity; i++ {
		col := i % cols
		row := i / cols
		spots = append(spots, models.ParkingSpot{
			ZoneID: zoneID,
			Label:  fmt.Sprintf("EV-%02d", i+1),
			PosX:   180 + float64(col)*52,
			PosY:   140 + float64(row)*72,
			Status: models.SpotStatusAvailable,
		})
	}
	return spots
}

// ColsForCapacity returns grid column count used for layout.
func ColsForCapacity(capacity int) int {
	if capacity <= defaultCols {
		return max(1, capacity)
	}
	return defaultCols
}

func max(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}
