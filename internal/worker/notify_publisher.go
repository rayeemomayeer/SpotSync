package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
)

const NotifyChannel = "spotsync:notify"

// RedisPublisher publishes availability (SSE) and notify-shaped email events.
type RedisPublisher struct {
	client *platform.RedisClient
}

func NewRedisPublisher(client *platform.RedisClient) *RedisPublisher {
	return &RedisPublisher{client: client}
}

func (p *RedisPublisher) PublishAvailability(ctx context.Context, zoneID uint, eventType string, payload []byte) error {
	if p.client == nil {
		return nil
	}
	if err := p.client.Publish(ctx, platform.AvailabilityChannel(zoneID), eventType, payload); err != nil {
		return err
	}
	return p.publishNotify(ctx, eventType, payload)
}

func (p *RedisPublisher) publishNotify(ctx context.Context, eventType string, payload []byte) error {
	var res domain.ReservationEventPayload
	if err := json.Unmarshal(payload, &res); err != nil {
		return err
	}
	notifyType, ok := mapNotifyType(eventType)
	if !ok {
		return nil
	}
	msg := map[string]any{
		"type":            notifyType,
		"zone_id":         strconv.FormatUint(uint64(res.ZoneID), 10),
		"user_id":         strconv.FormatUint(uint64(res.UserID), 10),
		"reservation_id":  strconv.FormatUint(uint64(res.ReservationID), 10),
		"license_plate":   res.LicensePlate,
	}
	if res.SpotID != nil {
		msg["spot_id"] = strconv.FormatUint(uint64(*res.SpotID), 10)
	}
	if res.Email != "" {
		msg["email"] = res.Email
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal notify: %w", err)
	}
	// Publish as plain JSON (eventType empty prefix unused — use dedicated PublishJSON).
	return p.client.PublishRaw(ctx, NotifyChannel, raw)
}

func mapNotifyType(eventType string) (string, bool) {
	switch eventType {
	case domain.EventReservationCreated:
		return "reservation_confirmed", true
	case domain.EventReservationCancelled, domain.EventReservationExpired:
		return "reservation_cancelled", true
	default:
		return "", false
	}
}
