package dto

import "time"

type OutboxEventResponse struct {
	ID             uint       `json:"id"`
	AggregateType  string     `json:"aggregate_type"`
	AggregateID    uint       `json:"aggregate_id"`
	EventType      string     `json:"event_type"`
	Attempts       int        `json:"attempts"`
	LastError      *string    `json:"last_error"`
	DeadLetteredAt *time.Time `json:"dead_lettered_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
