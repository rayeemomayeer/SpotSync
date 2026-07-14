package repository

import (
	"context"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *models.Payment) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Currency == "" {
		p.Currency = "usd"
	}
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uint) (*models.Payment, error) {
	var p models.Payment
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) GetByPaymentIntent(ctx context.Context, piID string) (*models.Payment, error) {
	var p models.Payment
	err := r.db.WithContext(ctx).Where("stripe_payment_intent_id = ?", piID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) AttachReservation(ctx context.Context, id, reservationID uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"reservation_id": reservationID,
			"status":         status,
			"updated_at":     time.Now(),
		}).Error
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()}).Error
}

func (r *PaymentRepository) ListByUser(ctx context.Context, userID uint) ([]models.Payment, error) {
	var list []models.Payment
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&list).Error
	return list, err
}

func (r *PaymentRepository) ByReservationIDs(ctx context.Context, ids []uint) (map[uint]models.Payment, error) {
	out := make(map[uint]models.Payment)
	if len(ids) == 0 {
		return out, nil
	}
	var list []models.Payment
	err := r.db.WithContext(ctx).
		Where("reservation_id IN ?", ids).
		Order("id DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		if p.ReservationID == nil {
			continue
		}
		if _, exists := out[*p.ReservationID]; !exists {
			out[*p.ReservationID] = p
		}
	}
	return out, nil
}

func (r *PaymentRepository) CreateRefund(ctx context.Context, ref *models.Refund) error {
	ref.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(ref).Error
}
