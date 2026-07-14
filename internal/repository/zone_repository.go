package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"gorm.io/gorm"
)

type ZoneListFilter struct {
	Type  string
	Query string
	Sort  string
	Order string
}

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

const zoneColumns = `parking_zones.id, parking_zones.name, parking_zones.type, parking_zones.total_capacity, parking_zones.price_per_hour, parking_zones.organization_id, parking_zones.created_at, parking_zones.updated_at`

const availabilitySelect = zoneColumns + `, (
		total_capacity - (
			SELECT COUNT(*) FROM reservations
			WHERE zone_id = parking_zones.id AND status = 'active'
		)
	) AS available_spots`

func (r *ZoneRepository) Create(ctx context.Context, zone *models.ParkingZone) error {
	return r.db.WithContext(ctx).Create(zone).Error
}

func (r *ZoneRepository) ListWithAvailability(ctx context.Context) ([]ZoneAvailabilityRow, error) {
	return r.ListWithAvailabilityFiltered(ctx, ZoneListFilter{})
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

func (r *ZoneRepository) Update(ctx context.Context, zone *models.ParkingZone) error {
	res := r.db.WithContext(ctx).Model(zone).Updates(map[string]any{
		"name":           zone.Name,
		"type":           zone.Type,
		"total_capacity": zone.TotalCapacity,
		"price_per_hour": zone.PricePerHour,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ZoneRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var active int64
		if err := tx.Model(&models.Reservation{}).
			Where("zone_id = ? AND status = ?", id, models.ReservationStatusActive).
			Count(&active).Error; err != nil {
			return err
		}
		if active > 0 {
			return domain.ErrZoneHasActiveReservations
		}

		if err := tx.Where("zone_id = ?", id).Delete(&models.Reservation{}).Error; err != nil {
			return err
		}

		res := tx.Delete(&models.ParkingZone{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
}

func (r *ZoneRepository) ListWithAvailabilityFiltered(ctx context.Context, f ZoneListFilter) ([]ZoneAvailabilityRow, error) {
	q := r.db.WithContext(ctx).Model(&models.ParkingZone{}).Select(availabilitySelect)

	if f.Type != "" {
		q = q.Where("parking_zones.type = ?", f.Type)
	}
	if strings.TrimSpace(f.Query) != "" {
		q = q.Where(
			"parking_zones.search_vector @@ plainto_tsquery('english', ?) OR parking_zones.name ILIKE ?",
			strings.TrimSpace(f.Query),
			"%"+strings.TrimSpace(f.Query)+"%",
		)
	}

	order := buildZoneOrderClause(f.Sort, f.Order)
	q = q.Order(order)

	var rows []ZoneAvailabilityRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func buildZoneOrderClause(sort, order string) string {
	dir := "ASC"
	if strings.EqualFold(order, "desc") {
		dir = "DESC"
	}

	switch sort {
	case "name":
		return fmt.Sprintf("parking_zones.name %s", dir)
	case "price_per_hour":
		return fmt.Sprintf("parking_zones.price_per_hour %s", dir)
	case "available_spots":
		return fmt.Sprintf("available_spots %s", dir)
	default:
		return "parking_zones.id ASC"
	}
}

// FindByID is kept for callers that need the raw zone row without availability.
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
