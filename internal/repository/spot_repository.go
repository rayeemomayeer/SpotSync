package repository

import (
	"context"
	"errors"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
)

type SpotRepository struct {
	db *gorm.DB
}

func NewSpotRepository(db *gorm.DB) *SpotRepository {
	return &SpotRepository{db: db}
}

func (r *SpotRepository) CreateBatch(ctx context.Context, spots []models.ParkingSpot) error {
	if len(spots) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&spots).Error
}

func (r *SpotRepository) CountByZone(ctx context.Context, zoneID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ParkingSpot{}).Where("zone_id = ?", zoneID).Count(&count).Error
	return count, err
}

func (r *SpotRepository) ListByZone(ctx context.Context, zoneID uint) ([]models.ParkingSpot, error) {
	var spots []models.ParkingSpot
	err := r.db.WithContext(ctx).
		Where("zone_id = ?", zoneID).
		Order("id ASC").
		Find(&spots).Error
	return spots, err
}

func (r *SpotRepository) FindByIDInZone(ctx context.Context, zoneID, spotID uint) (*models.ParkingSpot, error) {
	var spot models.ParkingSpot
	err := r.db.WithContext(ctx).Where("zone_id = ? AND id = ?", zoneID, spotID).First(&spot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &spot, nil
}

func (r *SpotRepository) UpdateStatus(ctx context.Context, spotID uint, status string) error {
	res := r.db.WithContext(ctx).Model(&models.ParkingSpot{}).Where("id = ?", spotID).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SpotRepository) ActiveReservationSpotIDs(ctx context.Context, zoneID uint) (map[uint]struct{}, error) {
	var ids []uint
	err := r.db.WithContext(ctx).
		Model(&models.Reservation{}).
		Where("zone_id = ? AND status = ? AND spot_id IS NOT NULL", zoneID, models.ReservationStatusActive).
		Pluck("spot_id", &ids).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		out[id] = struct{}{}
	}
	return out, nil
}

func (r *SpotRepository) CancelExpiredDemoReservations(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Reservation{}).
		Where("status = ? AND demo_expires_at IS NOT NULL AND demo_expires_at <= ?", models.ReservationStatusActive, now).
		Update("status", models.ReservationStatusCancelled).Error
}
