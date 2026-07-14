package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"gorm.io/gorm"
)

type OrganizationService struct {
	orgs  *repository.OrganizationRepository
	audit *repository.AuditRepository
	users *repository.UserRepository
}

func NewOrganizationService(orgs *repository.OrganizationRepository, audit *repository.AuditRepository, users *repository.UserRepository) *OrganizationService {
	return &OrganizationService{orgs: orgs, audit: audit, users: users}
}

func (s *OrganizationService) Create(ctx context.Context, actorID uint, name, slug string) (*models.Organization, error) {
	org := &models.Organization{
		Name:   strings.TrimSpace(name),
		Slug:   strings.ToLower(strings.TrimSpace(slug)),
		Status: models.OrgStatusPending,
	}
	if err := s.orgs.Create(ctx, org); err != nil {
		return nil, err
	}
	_ = s.audit.Insert(ctx, &models.AuditLog{
		ActorUserID:    &actorID,
		OrganizationID: &org.ID,
		Action:         "organization.create",
		ResourceType:   "organization",
		ResourceID:     &org.ID,
		Metadata:       mustJSON(map[string]string{"slug": org.Slug}),
	})
	return org, nil
}

func (s *OrganizationService) List(ctx context.Context) ([]models.Organization, error) {
	return s.orgs.List(ctx)
}

func (s *OrganizationService) Search(ctx context.Context, q string) ([]models.Organization, error) {
	return s.orgs.Search(ctx, q, 50)
}

func (s *OrganizationService) Get(ctx context.Context, id uint) (*models.Organization, error) {
	org, err := s.orgs.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return org, nil
}

func (s *OrganizationService) Approve(ctx context.Context, actorID, orgID uint) (*models.Organization, error) {
	org, err := s.orgs.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if org.Status != models.OrgStatusPending {
		return nil, domain.NewValidationError("Validation failed", map[string]string{
			"status": "Only pending organizations can be approved",
		})
	}
	if err := s.orgs.UpdateStatus(ctx, orgID, models.OrgStatusActive); err != nil {
		return nil, err
	}
	org.Status = models.OrgStatusActive
	_ = s.audit.Insert(ctx, &models.AuditLog{
		ActorUserID:    &actorID,
		OrganizationID: &orgID,
		Action:         "organization.approve",
		ResourceType:   "organization",
		ResourceID:     &orgID,
	})
	return org, nil
}

func (s *OrganizationService) Reject(ctx context.Context, actorID, orgID uint) (*models.Organization, error) {
	org, err := s.orgs.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if org.Status != models.OrgStatusPending {
		return nil, domain.NewValidationError("Validation failed", map[string]string{
			"status": "Only pending organizations can be rejected",
		})
	}
	if err := s.orgs.UpdateStatus(ctx, orgID, models.OrgStatusRejected); err != nil {
		return nil, err
	}
	org.Status = models.OrgStatusRejected
	_ = s.audit.Insert(ctx, &models.AuditLog{
		ActorUserID:    &actorID,
		OrganizationID: &orgID,
		Action:         "organization.reject",
		ResourceType:   "organization",
		ResourceID:     &orgID,
	})
	return org, nil
}

func (s *OrganizationService) EnsureOrgEntitled(ctx context.Context, orgID uint) error {
	org, err := s.orgs.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	if org.Status != models.OrgStatusActive {
		return domain.ErrOrgNotEntitled
	}
	if org.BillingPlan == nil {
		return domain.ErrOrgNotEntitled
	}
	plan := strings.TrimSpace(*org.BillingPlan)
	if plan != "starter" && plan != "growth" {
		return domain.ErrOrgNotEntitled
	}
	return nil
}

func (s *OrganizationService) SetStatus(ctx context.Context, actorID, orgID uint, status string) (*models.Organization, error) {
	if status != models.OrgStatusActive && status != models.OrgStatusSuspended {
		return nil, domain.NewValidationError("Validation failed", map[string]string{"status": "Must be active or suspended"})
	}
	if err := s.orgs.UpdateStatus(ctx, orgID, status); err != nil {
		return nil, err
	}
	org, err := s.orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Insert(ctx, &models.AuditLog{
		ActorUserID:    &actorID,
		OrganizationID: &orgID,
		Action:         "organization.status",
		ResourceType:   "organization",
		ResourceID:     &orgID,
		Metadata:       mustJSON(map[string]string{"status": status}),
	})
	return org, nil
}

func (s *OrganizationService) AssignOrgAdmin(ctx context.Context, actorID, orgID, userID uint) error {
	if _, err := s.orgs.GetByID(ctx, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	member := &models.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           models.RoleOrgAdmin,
	}
	if err := s.orgs.AddMember(ctx, member); err != nil {
		return err
	}
	_ = s.audit.Insert(ctx, &models.AuditLog{
		ActorUserID:    &actorID,
		OrganizationID: &orgID,
		Action:         "organization.assign_admin",
		ResourceType:   "user",
		ResourceID:     &userID,
	})
	return nil
}

func (s *OrganizationService) AssignOrgAdminByEmail(ctx context.Context, actorID, orgID uint, email string) error {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrNotFound
	}
	return s.AssignOrgAdmin(ctx, actorID, orgID, user.ID)
}

func (s *OrganizationService) ListMembers(ctx context.Context, orgID uint) ([]repository.OrgMemberView, error) {
	if _, err := s.orgs.GetByID(ctx, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return s.orgs.ListMembers(ctx, orgID)
}

func (s *OrganizationService) RemoveMember(ctx context.Context, actorID, orgID, userID uint) error {
	if err := s.orgs.RemoveMember(ctx, orgID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	_ = s.audit.Insert(ctx, &models.AuditLog{
		ActorUserID:    &actorID,
		OrganizationID: &orgID,
		Action:         "organization.remove_member",
		ResourceType:   "user",
		ResourceID:     &userID,
	})
	return nil
}

func (s *OrganizationService) SetBillingPlan(ctx context.Context, actorID, orgID uint, plan, stripeCustomerID string) (*models.Organization, error) {
	if plan != "starter" && plan != "growth" && plan != "none" {
		return nil, domain.NewValidationError("Validation failed", map[string]string{"plan": "Must be starter, growth, or none"})
	}
	clear := plan == "none"
	if err := s.orgs.UpdateBillingPlan(ctx, orgID, plan, optionalString(stripeCustomerID), clear); err != nil {
		return nil, err
	}
	org, err := s.orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Insert(ctx, &models.AuditLog{
		ActorUserID:    &actorID,
		OrganizationID: &orgID,
		Action:         "organization.billing_plan",
		ResourceType:   "organization",
		ResourceID:     &orgID,
		Metadata:       mustJSON(map[string]string{"plan": plan}),
	})
	return org, nil
}

func optionalString(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func (s *OrganizationService) ListAudit(ctx context.Context, orgID *uint, limit int) ([]models.AuditLog, error) {
	return s.audit.List(ctx, orgID, limit)
}

func (s *OrganizationService) EnsureOrgAccess(ctx context.Context, role string, userID, orgID uint) error {
	if role == models.RoleAdmin || role == models.RoleSaaSAdmin {
		return nil
	}
	if role != models.RoleOrgAdmin {
		return domain.ErrForbidden
	}
	_, err := s.orgs.Membership(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}

func (s *OrganizationService) PrimaryOrgID(ctx context.Context, userID uint) (*uint, error) {
	org, err := s.orgs.PrimaryOrgForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org.ID, nil
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
