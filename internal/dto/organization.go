package dto

import "time"

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
	Slug string `json:"slug" validate:"required,min=2,max=100"`
}

type OrganizationResponse struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	Status           string    `json:"status"`
	BillingPlan      *string   `json:"billing_plan,omitempty"`
	StripeCustomerID *string   `json:"stripe_customer_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AssignOrgMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type OrgMemberResponse struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type SetOrgPlanRequest struct {
	Plan             string `json:"plan" validate:"required,oneof=starter growth none"`
	StripeCustomerID string `json:"stripe_customer_id,omitempty" validate:"omitempty,max=255"`
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
