package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertTx(tx *gorm.DB, aggregateType string, aggregateID uint, eventType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ev := models.OutboxEvent{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       raw,
		CreatedAt:     time.Now(),
	}
	return tx.Create(&ev).Error
}

func (r *Repository) FetchUnprocessed(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	if limit < 1 {
		limit = 50
	}
	var list []models.OutboxEvent
	err := r.db.WithContext(ctx).
		Where("processed_at IS NULL AND dead_lettered_at IS NULL").
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *Repository) MarkProcessed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ? AND processed_at IS NULL", id).
		Update("processed_at", now).Error
}

const maxOutboxAttempts = 5

func (r *Repository) RecordFailure(ctx context.Context, id uint, errMsg string) error {
	return r.db.WithContext(ctx).Exec(`
UPDATE outbox_events
SET attempts = attempts + 1,
    last_error = ?,
    dead_lettered_at = CASE WHEN attempts + 1 >= ? THEN NOW() ELSE dead_lettered_at END
WHERE id = ? AND processed_at IS NULL
`, errMsg, maxOutboxAttempts, id).Error
}

func InsertReservationEvent(tx *gorm.DB, repo *Repository, eventType string, payload domain.ReservationEventPayload) error {
	return repo.InsertTx(tx, domain.AggregateReservation, payload.ReservationID, eventType, payload)
}
