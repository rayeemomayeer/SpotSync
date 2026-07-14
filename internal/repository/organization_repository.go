package repository

import (
	"context"
	"strings"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now
	org.Slug = strings.ToLower(strings.TrimSpace(org.Slug))
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id uint) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).Where("slug = ?", strings.ToLower(slug)).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) List(ctx context.Context) ([]models.Organization, error) {
	var list []models.Organization
	err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *OrganizationRepository) Search(ctx context.Context, q string, limit int) ([]models.Organization, error) {
	if limit < 1 {
		limit = 20
	}
	q = strings.TrimSpace(q)
	var list []models.Organization
	db := r.db.WithContext(ctx).Order("id ASC").Limit(limit)
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(slug) LIKE ?", like, like)
	}
	err := db.Find(&list).Error
	return list, err
}

func (r *OrganizationRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Organization{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()}).Error
}

func (r *OrganizationRepository) UpdateBillingPlan(ctx context.Context, id uint, plan string, stripeCustomerID *string) error {
	updates := map[string]any{
		"billing_plan": plan,
		"updated_at":   time.Now(),
	}
	if stripeCustomerID != nil {
		updates["stripe_customer_id"] = *stripeCustomerID
	}
	return r.db.WithContext(ctx).Model(&models.Organization{}).
		Where("id = ?", id).
		Updates(updates).Error
}

type OrgMemberView struct {
	UserID    uint
	Email     string
	Name      string
	Role      string
	CreatedAt time.Time
}

func (r *OrganizationRepository) ListMembers(ctx context.Context, orgID uint) ([]OrgMemberView, error) {
	var rows []OrgMemberView
	err := r.db.WithContext(ctx).Table("organization_members m").
		Select("m.user_id, u.email, u.name, m.role, m.created_at").
		Joins("JOIN users u ON u.id = m.user_id").
		Where("m.organization_id = ?", orgID).
		Order("m.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID uint) error {
	res := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&models.OrganizationMember{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *OrganizationRepository) AddMember(ctx context.Context, member *models.OrganizationMember) error {
	member.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *OrganizationRepository) Membership(ctx context.Context, orgID, userID uint) (*models.OrganizationMember, error) {
	var m models.OrganizationMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *OrganizationRepository) PrimaryOrgForUser(ctx context.Context, userID uint) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).
		Joins("JOIN organization_members m ON m.organization_id = organizations.id").
		Where("m.user_id = ?", userID).
		Order("organizations.id ASC").
		First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Insert(ctx context.Context, entry *models.AuditLog) error {
	entry.CreatedAt = time.Now()
	if entry.Metadata == nil {
		entry.Metadata = []byte("{}")
	}
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *AuditRepository) List(ctx context.Context, orgID *uint, limit int) ([]models.AuditLog, error) {
	if limit < 1 {
		limit = 50
	}
	q := r.db.WithContext(ctx).Order("id DESC").Limit(limit)
	if orgID != nil {
		q = q.Where("organization_id = ?", *orgID)
	}
	var list []models.AuditLog
	err := q.Find(&list).Error
	return list, err
}
