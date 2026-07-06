package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreateReservationParams struct {
	UserID        uint
	ZoneID        uint
	LicensePlate  string
	SpotID        *uint
	DemoExpiresAt *time.Time
}

type ReservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) CreateActive(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error) {
	return r.CreateActiveWithOptions(ctx, CreateReservationParams{
		UserID:       userID,
		ZoneID:       zoneID,
		LicensePlate: licensePlate,
	})
}

func (r *ReservationRepository) CreateActiveWithOptions(ctx context.Context, p CreateReservationParams) (*models.Reservation, error) {
	var created models.Reservation

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var zone models.ParkingZone
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&zone, p.ZoneID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		var active int64
		if err := tx.Model(&models.Reservation{}).
			Where("zone_id = ? AND status = ?", p.ZoneID, models.ReservationStatusActive).
			Count(&active).Error; err != nil {
			return err
		}
		if active >= int64(zone.TotalCapacity) {
			return domain.ErrZoneFull
		}

		spotID, err := resolveSpotForReservation(tx, p.ZoneID, p.SpotID)
		if err != nil {
			return err
		}

		created = models.Reservation{
			UserID:        p.UserID,
			ZoneID:        p.ZoneID,
			SpotID:        spotID,
			LicensePlate:  p.LicensePlate,
			Status:        models.ReservationStatusActive,
			DemoExpiresAt: p.DemoExpiresAt,
		}
		if err := tx.Create(&created).Error; err != nil {
			return mapReservationWriteErr(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func resolveSpotForReservation(tx *gorm.DB, zoneID uint, requested *uint) (*uint, error) {
	var spots []models.ParkingSpot
	if err := tx.Where("zone_id = ?", zoneID).Order("id ASC").Find(&spots).Error; err != nil {
		return nil, err
	}
	if len(spots) == 0 {
		return requested, nil
	}

	occupied := make(map[uint]struct{})
	var occupiedIDs []uint
	if err := tx.Model(&models.Reservation{}).
		Where("zone_id = ? AND status = ? AND spot_id IS NOT NULL", zoneID, models.ReservationStatusActive).
		Pluck("spot_id", &occupiedIDs).Error; err != nil {
		return nil, err
	}
	for _, id := range occupiedIDs {
		occupied[id] = struct{}{}
	}

	if requested != nil {
		for _, spot := range spots {
			if spot.ID != *requested {
				continue
			}
			if spot.Status == models.SpotStatusUnavailable {
				return nil, domain.ErrSpotUnavailable
			}
			if _, taken := occupied[spot.ID]; taken {
				return nil, domain.ErrSpotTaken
			}
			id := spot.ID
			return &id, nil
		}
		return nil, domain.ErrNotFound
	}

	for _, spot := range spots {
		if spot.Status != models.SpotStatusAvailable {
			continue
		}
		if _, taken := occupied[spot.ID]; taken {
			continue
		}
		id := spot.ID
		return &id, nil
	}

	return nil, domain.ErrZoneFull
}

func mapReservationWriteErr(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrSpotTaken
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrSpotTaken
	}
	return err
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
	if err := r.CancelExpiredDemoReservations(ctx, time.Now()); err != nil {
		return nil, err
	}
	var list []models.Reservation
	err := r.db.WithContext(ctx).
		Preload("Zone").
		Preload("Spot").
		Where("user_id = ?", userID).
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *ReservationRepository) ListAll(ctx context.Context, page, limit int) ([]models.Reservation, error) {
	q := r.db.WithContext(ctx).
		Preload("User").
		Preload("Zone").
		Preload("Spot").
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

func (r *ReservationRepository) CancelExpiredDemoReservations(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Reservation{}).
		Where("status = ? AND demo_expires_at IS NOT NULL AND demo_expires_at <= ?", models.ReservationStatusActive, now).
		Update("status", models.ReservationStatusCancelled).Error
}

func (r *ReservationRepository) HasActiveOnSpot(ctx context.Context, spotID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Reservation{}).
		Where("spot_id = ? AND status = ?", spotID, models.ReservationStatusActive).
		Count(&count).Error
	return count > 0, err
}
