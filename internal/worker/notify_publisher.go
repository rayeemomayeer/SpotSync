package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
)

const NotifyChannel = "spotsync:notify"

// RedisPublisher publishes availability (SSE) and notify-shaped email events.
type RedisPublisher struct {
	client *platform.RedisClient
	notifs *repository.NotificationRepository
}

func NewRedisPublisher(client *platform.RedisClient, notifs *repository.NotificationRepository) *RedisPublisher {
	return &RedisPublisher{client: client, notifs: notifs}
}

func (p *RedisPublisher) PublishAvailability(ctx context.Context, zoneID uint, eventType string, payload []byte) error {
	if p.client == nil {
		return p.writeInApp(ctx, eventType, payload)
	}
	if err := p.client.Publish(ctx, platform.AvailabilityChannel(zoneID), eventType, payload); err != nil {
		return err
	}
	if err := p.publishNotify(ctx, eventType, payload); err != nil {
		return err
	}
	return p.writeInApp(ctx, eventType, payload)
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
		"type":           notifyType,
		"zone_id":        strconv.FormatUint(uint64(res.ZoneID), 10),
		"user_id":        strconv.FormatUint(uint64(res.UserID), 10),
		"reservation_id": strconv.FormatUint(uint64(res.ReservationID), 10),
		"license_plate":  res.LicensePlate,
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
	return p.client.PublishRaw(ctx, NotifyChannel, raw)
}

func (p *RedisPublisher) writeInApp(ctx context.Context, eventType string, payload []byte) error {
	if p.notifs == nil {
		return nil
	}
	var res domain.ReservationEventPayload
	if err := json.Unmarshal(payload, &res); err != nil {
		return err
	}
	if res.UserID == 0 {
		return nil
	}
	var typ, title, body string
	switch eventType {
	case domain.EventReservationCreated:
		typ = "reservation_confirmed"
		title = "Reservation confirmed"
		body = fmt.Sprintf("Zone #%d · reservation #%d", res.ZoneID, res.ReservationID)
	case domain.EventReservationCancelled, domain.EventReservationExpired:
		typ = "reservation_cancelled"
		title = "Reservation cancelled"
		body = fmt.Sprintf("Zone #%d · reservation #%d", res.ZoneID, res.ReservationID)
	default:
		return nil
	}
	return p.notifs.Create(ctx, &models.Notification{
		UserID: res.UserID,
		Type:   typ,
		Title:  title,
		Body:   body,
	})
}

func (p *RedisPublisher) PublishOrgApproved(ctx context.Context, org *models.Organization, email string) error {
	if p.client == nil {
		return nil
	}
	msg := map[string]any{
		"type":     "org_approved",
		"zone_id":  "0",
		"org_name": org.Name,
	}
	if email != "" {
		msg["email"] = email
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
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
