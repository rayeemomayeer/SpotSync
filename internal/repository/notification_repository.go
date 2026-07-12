package repository

import (
	"context"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	n.CreatedAt = time.Now()
	if n.Metadata == nil {
		n.Metadata = []byte("{}")
	}
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID uint, limit int) ([]models.Notification, error) {
	if limit < 1 {
		limit = 30
	}
	var list []models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read_at", now).Error
}
