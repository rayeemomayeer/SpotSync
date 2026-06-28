package repository

import (
	"context"
	"errors"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) CreateActive(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error) {
	var created models.Reservation

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var zone models.ParkingZone
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&zone, zoneID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		var active int64
		if err := tx.Model(&models.Reservation{}).
			Where("zone_id = ? AND status = ?", zoneID, models.ReservationStatusActive).
			Count(&active).Error; err != nil {
			return err
		}
		if active >= int64(zone.TotalCapacity) {
			return domain.ErrZoneFull
		}

		created = models.Reservation{
			UserID:       userID,
			ZoneID:       zoneID,
			LicensePlate: licensePlate,
			Status:       models.ReservationStatusActive,
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *ReservationRepository) FindByID(ctx context.Context, id uint) (*models.Reservation, error) {
	var res models.Reservation
	err := r.db.WithContext(ctx).First(&res, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepository) Cancel(ctx context.Context, reservationID, userID uint) error {
	var res models.Reservation
	if err := r.db.WithContext(ctx).First(&res, reservationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	if res.UserID != userID {
		return domain.ErrNotOwner
	}
	if res.Status == models.ReservationStatusCancelled {
		return nil
	}
	return r.db.WithContext(ctx).Model(&res).Update("status", models.ReservationStatusCancelled).Error
}

func (r *ReservationRepository) ListByUser(ctx context.Context, userID uint) ([]models.Reservation, error) {
	var list []models.Reservation
	err := r.db.WithContext(ctx).
		Preload("Zone").
		Where("user_id = ?", userID).
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *ReservationRepository) ListAll(ctx context.Context, page, limit int) ([]models.Reservation, error) {
	q := r.db.WithContext(ctx).
		Preload("User").
		Preload("Zone").
		Order("id DESC")

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		q = q.Offset(offset).Limit(limit)
	}

	var list []models.Reservation
	err := q.Find(&list).Error
	return list, err
}

func (r *ReservationRepository) CountActiveByZone(ctx context.Context, zoneID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Reservation{}).
		Where("zone_id = ? AND status = ?", zoneID, models.ReservationStatusActive).
		Count(&count).Error
	return count, err
}
