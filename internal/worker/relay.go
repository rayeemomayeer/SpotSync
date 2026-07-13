package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/outbox"
)

const relayBatchSize = 50

type EventPublisher interface {
	PublishAvailability(ctx context.Context, zoneID uint, eventType string, payload []byte) error
}

type Relay struct {
	outbox    *outbox.Repository
	publisher EventPublisher
	log       *slog.Logger
}

func NewRelay(outboxRepo *outbox.Repository, publisher EventPublisher, log *slog.Logger) *Relay {
	return &Relay{outbox: outboxRepo, publisher: publisher, log: log}
}

func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	events, err := r.outbox.FetchUnprocessed(ctx, relayBatchSize)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, ev := range events {
		if err := r.processOne(ctx, ev); err != nil {
			if r.log != nil {
				r.log.Error("outbox relay failed", "event_id", ev.ID, "error", err)
			}
			if markErr := r.outbox.RecordFailure(ctx, ev.ID, err.Error()); markErr != nil && r.log != nil {
				r.log.Error("outbox failure record failed", "event_id", ev.ID, "error", markErr)
			}
			continue
		}
		if err := r.outbox.MarkProcessed(ctx, ev.ID); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}

func (r *Relay) processOne(ctx context.Context, ev models.OutboxEvent) error {
	if r.publisher == nil {
		return nil
	}
	var payload domain.ReservationEventPayload
	if err := json.Unmarshal(ev.Payload, &payload); err != nil {
		return err
	}
	return r.publisher.PublishAvailability(ctx, payload.ZoneID, ev.EventType, ev.Payload)
}

func RunRelayLoop(ctx context.Context, relay *Relay, interval time.Duration, log *slog.Logger) {
	if interval < time.Second {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := relay.RunOnce(ctx)
			if err != nil && log != nil {
				log.Error("relay tick failed", "error", err)
			} else if n > 0 && log != nil {
				log.Debug("outbox relay processed", "count", n)
			}
		}
	}
}

type NoopPublisher struct{}

func (NoopPublisher) PublishAvailability(context.Context, uint, string, []byte) error {
	return nil
}
