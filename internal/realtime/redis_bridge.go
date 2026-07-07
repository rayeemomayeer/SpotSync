package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/redis/go-redis/v9"
)

// BridgeRedis forwards Redis pub/sub messages into the in-process SSE hub.
func BridgeRedis(ctx context.Context, client *platform.RedisClient, hub *Hub, log *slog.Logger) {
	if client == nil || hub == nil {
		return
	}
	pubsub := client.PSubscribe(ctx, "spotsync:zone:*")
	if pubsub == nil {
		return
	}
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			ev := mapRedisMessage(msg)
			if ev != nil {
				hub.Publish(*ev)
			}
		}
	}
}

func mapRedisMessage(msg *redis.Message) *ZoneEvent {
	if msg == nil {
		return nil
	}
	parts := strings.SplitN(msg.Payload, "|", 2)
	if len(parts) != 2 {
		return nil
	}
	eventType := parts[0]
	var payload domain.ReservationEventPayload
	if err := json.Unmarshal([]byte(parts[1]), &payload); err != nil {
		return nil
	}
	var t EventType
	switch eventType {
	case domain.EventReservationCreated:
		t = EventSpotReserved
	case domain.EventReservationCancelled, domain.EventReservationExpired:
		t = EventSpotReleased
	default:
		return nil
	}
	if payload.SpotID == nil {
		return nil
	}
	return &ZoneEvent{
		Type:          t,
		ZoneID:        payload.ZoneID,
		SpotID:        *payload.SpotID,
		UserID:        payload.UserID,
		ReservationID: payload.ReservationID,
		LicensePlate:  payload.LicensePlate,
	}
}
