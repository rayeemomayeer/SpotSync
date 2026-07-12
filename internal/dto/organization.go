package dto

import "time"

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
	Slug string `json:"slug" validate:"required,min=2,max=100"`
}

type OrganizationResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SuspendOrganizationRequest struct {
	Status string `json:"status" validate:"required,oneof=active suspended"`
}

type AuditLogResponse struct {
	ID             uint      `json:"id"`
	ActorUserID    *uint     `json:"actor_user_id"`
	OrganizationID *uint     `json:"organization_id"`
	Action         string    `json:"action"`
	ResourceType   string    `json:"resource_type"`
	ResourceID     *uint     `json:"resource_id"`
	CreatedAt      time.Time `json:"created_at"`
}
