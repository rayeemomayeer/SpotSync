package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
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
	var ev models.OutboxEvent
	if err := r.db.WithContext(ctx).First(&ev, id).Error; err != nil {
		return err
	}
	if ev.ProcessedAt != nil {
		return nil
	}
	nextAttempts := ev.Attempts + 1
	deadLetter := nextAttempts >= maxOutboxAttempts
	err := r.db.WithContext(ctx).Exec(`
UPDATE outbox_events
SET attempts = ?,
    last_error = ?,
    dead_lettered_at = CASE WHEN ? THEN NOW() ELSE dead_lettered_at END
WHERE id = ? AND processed_at IS NULL
`, nextAttempts, errMsg, deadLetter, id).Error
	if err != nil {
		return err
	}
	if deadLetter && ev.DeadLetteredAt == nil {
		platform.RecordOutboxDeadLetter()
	}
	return nil
}

func (r *Repository) FetchDeadLettered(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	if limit < 1 {
		limit = 50
	}
	var list []models.OutboxEvent
	err := r.db.WithContext(ctx).
		Where("dead_lettered_at IS NOT NULL").
		Order("dead_lettered_at DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *Repository) ReplayDeadLetter(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Exec(`
UPDATE outbox_events
SET attempts = 0,
    last_error = NULL,
    dead_lettered_at = NULL,
    processed_at = NULL
WHERE id = ? AND dead_lettered_at IS NOT NULL
`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}


func InsertReservationEvent(tx *gorm.DB, repo *Repository, eventType string, payload domain.ReservationEventPayload) error {
	return repo.InsertTx(tx, domain.AggregateReservation, payload.ReservationID, eventType, payload)
}
