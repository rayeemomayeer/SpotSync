package repository

import (
	"context"
	"errors"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
)

type ZoneRepository struct {
	db *gorm.DB
}

func NewZoneRepository(db *gorm.DB) *ZoneRepository {
	return &ZoneRepository{db: db}
}

type ZoneAvailabilityRow struct {
	models.ParkingZone
	AvailableSpots int `gorm:"column:available_spots"`
}

const availabilitySelect = `parking_zones.*, (
		total_capacity - (
			SELECT COUNT(*) FROM reservations
			WHERE zone_id = parking_zones.id AND status = 'active'
		)
	) AS available_spots`

func (r *ZoneRepository) Create(ctx context.Context, zone *models.ParkingZone) error {
	return r.db.WithContext(ctx).Create(zone).Error
}

func (r *ZoneRepository) ListWithAvailability(ctx context.Context) ([]ZoneAvailabilityRow, error) {
	var rows []ZoneAvailabilityRow
	err := r.db.WithContext(ctx).
		Model(&models.ParkingZone{}).
		Select(availabilitySelect).
		Order("id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *ZoneRepository) GetByIDWithAvailability(ctx context.Context, id uint) (*ZoneAvailabilityRow, error) {
	var row ZoneAvailabilityRow
	err := r.db.WithContext(ctx).
		Model(&models.ParkingZone{}).
		Select(availabilitySelect).
		Where("id = ?", id).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, domain.ErrNotFound
	}
	return &row, nil
}

func (r *ZoneRepository) FindByID(ctx context.Context, id uint) (*models.ParkingZone, error) {
	var zone models.ParkingZone
	err := r.db.WithContext(ctx).First(&zone, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &zone, nil
}
