package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/outbox"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreateReservationParams struct {
	UserID         uint
	ZoneID         uint
	LicensePlate   string
	SpotID         *uint
	DemoExpiresAt  *time.Time
	StartTime      *time.Time
	EndTime        *time.Time
	IdempotencyKey *string
}

type ReservationRepository struct {
	db     *gorm.DB
	outbox *outbox.Repository
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db, outbox: outbox.NewRepository(db)}
}

func (r *ReservationRepository) CreateActive(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error) {
	return r.CreateActiveWithOptions(ctx, CreateReservationParams{
		UserID:       userID,
		ZoneID:       zoneID,
		LicensePlate: licensePlate,
	})
}

func (r *ReservationRepository) CreateActiveWithOptions(ctx context.Context, p CreateReservationParams) (*models.Reservation, error) {
	if p.IdempotencyKey != nil && *p.IdempotencyKey != "" {
		existing, err := r.findByIdempotencyKey(ctx, *p.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

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
			UserID:         p.UserID,
			ZoneID:         p.ZoneID,
			SpotID:         spotID,
			LicensePlate:   p.LicensePlate,
			Status:         models.ReservationStatusActive,
			DemoExpiresAt:  p.DemoExpiresAt,
			StartTime:      p.StartTime,
			EndTime:        p.EndTime,
			IdempotencyKey: p.IdempotencyKey,
		}
		if err := tx.Create(&created).Error; err != nil {
			return mapReservationWriteErr(err)
		}
		payload := domain.ReservationEventPayload{
			ReservationID: created.ID,
			ZoneID:        created.ZoneID,
			SpotID:        created.SpotID,
			UserID:        created.UserID,
			LicensePlate:  created.LicensePlate,
		}
		return outbox.InsertReservationEvent(tx, r.outbox, domain.EventReservationCreated, payload)
	})
	if err != nil {
		return nil, err
	}

	var loaded models.Reservation
	if err := r.db.WithContext(ctx).Preload("Spot").First(&loaded, created.ID).Error; err != nil {
		return &created, nil
	}
	return &loaded, nil
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

func (r *ReservationRepository) Cancel(ctx context.Context, reservationID, userID uint) (*models.Reservation, error) {
	var res models.Reservation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Spot").First(&res, reservationID).Error; err != nil {
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
		if err := tx.Model(&res).Update("status", models.ReservationStatusCancelled).Error; err != nil {
			return err
		}
		res.Status = models.ReservationStatusCancelled
		payload := domain.ReservationEventPayload{
			ReservationID: res.ID,
			ZoneID:        res.ZoneID,
			SpotID:        res.SpotID,
			UserID:        res.UserID,
			LicensePlate:  res.LicensePlate,
		}
		return outbox.InsertReservationEvent(tx, r.outbox, domain.EventReservationCancelled, payload)
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepository) findByIdempotencyKey(ctx context.Context, key string) (*models.Reservation, error) {
	var res models.Reservation
	err := r.db.WithContext(ctx).Preload("Spot").Where("idempotency_key = ?", key).First(&res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepository) ExpireDemoReservations(ctx context.Context, now time.Time) ([]models.Reservation, error) {
	var list []models.Reservation
	if err := r.db.WithContext(ctx).
		Preload("Spot").
		Where("status = ? AND demo_expires_at IS NOT NULL AND demo_expires_at <= ?", models.ReservationStatusActive, now).
		Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	ids := make([]uint, len(list))
	for i, res := range list {
		ids[i] = res.ID
	}
	if err := r.db.WithContext(ctx).
		Model(&models.Reservation{}).
		Where("id IN ?", ids).
		Update("status", models.ReservationStatusCancelled).Error; err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Status = models.ReservationStatusCancelled
	}
	return list, nil
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

func (r *ReservationRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Reservation{}).Count(&count).Error
	return count, err
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
