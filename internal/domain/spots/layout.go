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

// ShowcaseLayout returns coordinates aligned with spotsync-web reference-layout.ts (1264×843 map).
func ShowcaseLayout(zoneID uint) []models.ParkingSpot {
	const capacity = 24
	spots := make([]models.ParkingSpot, 0, capacity)
	anchors := [][2]float64{
		{290, 350}, {320, 350}, {350, 350}, {275, 425}, {305, 425}, {335, 425},
		{920, 350}, {965, 350}, {995, 350}, {860, 380}, {890, 380}, {905, 405},
		{304, 552}, {328, 552}, {364, 552}, {280, 600}, {268, 636}, {340, 636},
		{844, 552}, {916, 552}, {940, 552}, {868, 576}, {976, 576}, {1000, 564},
	}
	for i := 0; i < capacity; i++ {
		spots = append(spots, models.ParkingSpot{
			ZoneID: zoneID,
			Label:  fmt.Sprintf("A-%02d", i+1),
			PosX:   anchors[i][0],
			PosY:   anchors[i][1],
			Status: models.SpotStatusAvailable,
		})
	}
	return spots
}

// AppendGridLayout generates additional spots starting after existingCount.
func AppendGridLayout(zoneID uint, existingCount, additional int) []models.ParkingSpot {
	if additional < 1 {
		return nil
	}
	all := GridLayout(zoneID, existingCount+additional)
	if existingCount >= len(all) {
		return nil
	}
	return all[existingCount:]
}

func ColsForCapacity(capacity int) int {
	if capacity <= defaultCols {
		return max(1, capacity)
	}
	return defaultCols
}

func max(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}
