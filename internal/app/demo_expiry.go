package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/realtime"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
)

const defaultDemoExpiryInterval = 30 * time.Second

// RunDemoExpiryWorker cancels expired demo reservations and publishes SSE release events.
func RunDemoExpiryWorker(ctx context.Context, hub *realtime.Hub, repo *repository.ReservationRepository, log *slog.Logger) {
	if hub == nil || repo == nil {
		return
	}
	ticker := time.NewTicker(defaultDemoExpiryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expired, err := repo.ExpireDemoReservations(ctx, time.Now())
			if err != nil {
				if log != nil {
					log.Error("demo expiry sweep failed", "error", err)
				}
				continue
			}
			for _, res := range expired {
				if res.SpotID == nil {
					continue
				}
				hub.Publish(realtime.ZoneEvent{
					Type:          realtime.EventSpotReleased,
					ZoneID:        res.ZoneID,
					SpotID:        *res.SpotID,
					UserID:        res.UserID,
					ReservationID: res.ID,
					LicensePlate:  res.LicensePlate,
				})
			}
		}
	}
}
