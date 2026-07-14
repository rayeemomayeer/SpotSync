package models

import "time"

type Organization struct {
	ID               uint      `gorm:"primaryKey"`
	Name             string    `gorm:"not null"`
	Slug             string    `gorm:"uniqueIndex;not null"`
	Status           string    `gorm:"not null;default:active"`
	BillingPlan      *string   `gorm:"size:32"`
	StripeCustomerID *string   `gorm:"size:255"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
}

func (Organization) TableName() string {
	return "organizations"
}

const (
	OrgStatusPending   = "pending"
	OrgStatusActive    = "active"
	OrgStatusSuspended = "suspended"
	OrgStatusRejected  = "rejected"
)

type OrganizationMember struct {
	ID             uint      `gorm:"primaryKey"`
	OrganizationID uint      `gorm:"not null;uniqueIndex:idx_org_member"`
	UserID         uint      `gorm:"not null;uniqueIndex:idx_org_member"`
	Role           string    `gorm:"not null;default:org_admin"`
	CreatedAt      time.Time `gorm:"not null"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}

type AuditLog struct {
	ID             uint      `gorm:"primaryKey"`
	ActorUserID    *uint     `gorm:"index"`
	OrganizationID *uint     `gorm:"index"`
	Action         string    `gorm:"not null"`
	ResourceType   string    `gorm:"not null"`
	ResourceID     *uint
	Metadata       []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt      time.Time `gorm:"not null"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
