package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/outbox"
	"gorm.io/gorm"
)

type ExpiryEngine struct {
	db     *gorm.DB
	outbox *outbox.Repository
	log    *slog.Logger
}

func NewExpiryEngine(db *gorm.DB, outboxRepo *outbox.Repository, log *slog.Logger) *ExpiryEngine {
	return &ExpiryEngine{db: db, outbox: outboxRepo, log: log}
}

func (e *ExpiryEngine) RunOnce(ctx context.Context, now time.Time) (int, error) {
	var expired []models.Reservation
	err := e.db.WithContext(ctx).
		Where("status = ? AND end_time IS NOT NULL AND end_time <= ?", models.ReservationStatusActive, now).
		Find(&expired).Error
	if err != nil {
		return 0, err
	}
	if len(expired) == 0 {
		return 0, nil
	}

	count := 0
	for _, res := range expired {
		if err := e.completeOne(ctx, res); err != nil {
			if e.log != nil {
				e.log.Error("expiry failed", "reservation_id", res.ID, "error", err)
			}
			continue
		}
		count++
	}
	return count, nil
}

func (e *ExpiryEngine) completeOne(ctx context.Context, res models.Reservation) error {
	return e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked models.Reservation
		if err := tx.First(&locked, res.ID).Error; err != nil {
			return err
		}
		if locked.Status != models.ReservationStatusActive {
			return nil
		}
		if locked.EndTime == nil || locked.EndTime.After(time.Now()) {
			return nil
		}
		if err := tx.Model(&locked).Update("status", models.ReservationStatusCompleted).Error; err != nil {
			return err
		}
		payload := domain.ReservationEventPayload{
			ReservationID: locked.ID,
			ZoneID:        locked.ZoneID,
			SpotID:        locked.SpotID,
			UserID:        locked.UserID,
			LicensePlate:  locked.LicensePlate,
		}
		return outbox.InsertReservationEvent(tx, e.outbox, domain.EventReservationExpired, payload)
	})
}

func RunExpiryLoop(ctx context.Context, engine *ExpiryEngine, interval time.Duration, log *slog.Logger) {
	if interval < time.Second {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := engine.RunOnce(ctx, time.Now())
			if err != nil && log != nil {
				log.Error("expiry tick failed", "error", err)
			} else if n > 0 && log != nil {
				log.Info("reservations expired", "count", n)
			}
		}
	}
}
