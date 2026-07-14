package service

import (
	"context"

	"github.com/rayeemomayeer/SpotSync/internal/cache"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/domain/spots"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
)

type ZoneStore interface {
	Create(ctx context.Context, zone *models.ParkingZone) error
	ListWithAvailability(ctx context.Context) ([]repository.ZoneAvailabilityRow, error)
	ListWithAvailabilityFiltered(ctx context.Context, f repository.ZoneListFilter) ([]repository.ZoneAvailabilityRow, error)
	GetByIDWithAvailability(ctx context.Context, id uint) (*repository.ZoneAvailabilityRow, error)
	FindByID(ctx context.Context, id uint) (*models.ParkingZone, error)
	Update(ctx context.Context, zone *models.ParkingZone) error
	Delete(ctx context.Context, id uint) error
}

type SpotZoneManager interface {
	CreateBatch(ctx context.Context, spots []models.ParkingSpot) error
	CountByZone(ctx context.Context, zoneID uint) (int64, error)
	ListByZone(ctx context.Context, zoneID uint) ([]models.ParkingSpot, error)
	DeleteByIDs(ctx context.Context, ids []uint) error
	ActiveReservationSpotIDs(ctx context.Context, zoneID uint) (map[uint]struct{}, error)
}

type ZoneActiveCounter interface {
	CountActiveByZone(ctx context.Context, zoneID uint) (int64, error)
}

type ZoneService struct {
	zones        ZoneStore
	spots        SpotZoneManager
	reservations ZoneActiveCounter
	availCache   *cache.AvailabilityCache
}

func NewZoneService(
	zones ZoneStore,
	spots SpotZoneManager,
	reservations ZoneActiveCounter,
	availCache *cache.AvailabilityCache,
) *ZoneService {
	return &ZoneService{zones: zones, spots: spots, reservations: reservations, availCache: availCache}
}

func (s *ZoneService) InvalidateAvailability(ctx context.Context, zoneID uint) {
	if s.availCache == nil {
		return
	}
	_ = s.availCache.Invalidate(ctx, zoneID)
}

func (s *ZoneService) Create(ctx context.Context, req dto.CreateZoneRequest, orgID *uint) (dto.ZoneResponse, error) {
	zone := &models.ParkingZone{
		Name:           req.Name,
		Type:           req.Type,
		TotalCapacity:  req.TotalCapacity,
		PricePerHour:   req.PricePerHour,
		OrganizationID: orgID,
		DemoSessionID:  req.DemoSessionID,
	}
	if err := s.zones.Create(ctx, zone); err != nil {
		return dto.ZoneResponse{}, err
	}

	if s.spots != nil {
		layout := spots.GridLayout(zone.ID, zone.TotalCapacity)
		if err := s.spots.CreateBatch(ctx, layout); err != nil {
			return dto.ZoneResponse{}, err
		}
		if err := s.syncZoneTotalCapacity(ctx, zone.ID); err != nil {
			return dto.ZoneResponse{}, err
		}
	}

	return s.GetByID(ctx, zone.ID)
}

func (s *ZoneService) List(ctx context.Context, q dto.ZoneListQuery) ([]dto.ZoneResponse, error) {
	filter := repository.ZoneListFilter{
		Type:          q.Type,
		Query:         q.Q,
		Sort:          q.Sort,
		Order:         q.Order,
		DemoMode:      q.DemoMode,
		DemoSessionID: q.DemoSessionID,
	}
	rows, err := s.zones.ListWithAvailabilityFiltered(ctx, filter)
	if err != nil {
		return nil, err
	}
	s.warmAvailabilityCache(ctx, rows)
	return mapZoneRows(rows), nil
}

func (s *ZoneService) GetByID(ctx context.Context, id uint) (dto.ZoneResponse, error) {
	row, err := s.zones.GetByIDWithAvailability(ctx, id)
	if err != nil {
		return dto.ZoneResponse{}, err
	}

	available := row.AvailableSpots
	if s.availCache != nil {
		if cached, ok := s.availCache.Get(ctx, id); ok {
			platform.RecordZoneAvailCacheHit()
			available = cached
		} else {
			platform.RecordZoneAvailCacheMiss()
			_ = s.availCache.Set(ctx, id, available)
		}
	}

	return dto.ZoneFromModel(row.ParkingZone, available), nil
}

func (s *ZoneService) Update(ctx context.Context, id uint, req dto.UpdateZoneRequest) (dto.ZoneResponse, error) {
	row, err := s.zones.GetByIDWithAvailability(ctx, id)
	if err != nil {
		return dto.ZoneResponse{}, err
	}

	if s.reservations != nil {
		active, err := s.reservations.CountActiveByZone(ctx, id)
		if err != nil {
			return dto.ZoneResponse{}, err
		}
		if int64(req.TotalCapacity) < active {
			return dto.ZoneResponse{}, domain.ErrCapacityBelowActive
		}
	}

	zone := row.ParkingZone
	oldCapacity := zone.TotalCapacity
	zone.Name = req.Name
	zone.Type = req.Type
	zone.TotalCapacity = req.TotalCapacity
	zone.PricePerHour = req.PricePerHour

	if err := s.zones.Update(ctx, &zone); err != nil {
		return dto.ZoneResponse{}, err
	}

	if s.spots != nil && req.TotalCapacity != oldCapacity {
		if err := s.syncSpotCapacity(ctx, id, oldCapacity, req.TotalCapacity); err != nil {
			return dto.ZoneResponse{}, err
		}
		if err := s.syncZoneTotalCapacity(ctx, id); err != nil {
			return dto.ZoneResponse{}, err
		}
	}

	return s.GetByID(ctx, id)
}

func (s *ZoneService) Delete(ctx context.Context, id uint) error {
	return s.zones.Delete(ctx, id)
}

func (s *ZoneService) syncSpotCapacity(ctx context.Context, zoneID uint, oldCapacity, newCapacity int) error {
	spotCount, err := s.spots.CountByZone(ctx, zoneID)
	if err != nil {
		return err
	}
	if spotCount == 0 {
		if newCapacity > 0 {
			return s.spots.CreateBatch(ctx, spots.GridLayout(zoneID, newCapacity))
		}
		return nil
	}

	current := int(spotCount)
	if newCapacity > current {
		additional := spots.AppendGridLayout(zoneID, current, newCapacity-current)
		return s.spots.CreateBatch(ctx, additional)
	}
	if newCapacity < current {
		return s.trimSpots(ctx, zoneID, current-newCapacity)
	}
	return nil
}

func (s *ZoneService) trimSpots(ctx context.Context, zoneID uint, removeCount int) error {
	list, err := s.spots.ListByZone(ctx, zoneID)
	if err != nil {
		return err
	}
	if removeCount <= 0 || len(list) == 0 {
		return nil
	}

	occupied, err := s.spots.ActiveReservationSpotIDs(ctx, zoneID)
	if err != nil {
		return err
	}

	var toDelete []uint
	for i := len(list) - 1; i >= 0 && len(toDelete) < removeCount; i-- {
		if _, taken := occupied[list[i].ID]; taken {
			continue
		}
		toDelete = append(toDelete, list[i].ID)
	}
	if len(toDelete) < removeCount {
		return domain.ErrCapacityBelowActive
	}
	return s.spots.DeleteByIDs(ctx, toDelete)
}

func (s *ZoneService) syncZoneTotalCapacity(ctx context.Context, zoneID uint) error {
	if s.spots == nil {
		return nil
	}
	count, err := s.spots.CountByZone(ctx, zoneID)
	if err != nil {
		return err
	}
	if count < 1 {
		return nil
	}
	zone, err := s.zones.FindByID(ctx, zoneID)
	if err != nil {
		return err
	}
	capacity := int(count)
	if zone.TotalCapacity == capacity {
		return nil
	}
	zone.TotalCapacity = capacity
	return s.zones.Update(ctx, zone)
}

func (s *ZoneService) warmAvailabilityCache(ctx context.Context, rows []repository.ZoneAvailabilityRow) {
	if s.availCache == nil {
		return
	}
	for _, row := range rows {
		_ = s.availCache.Set(ctx, row.ID, row.AvailableSpots)
	}
}

func mapZoneRows(rows []repository.ZoneAvailabilityRow) []dto.ZoneResponse {
	out := make([]dto.ZoneResponse, len(rows))
	for i, row := range rows {
		out[i] = dto.ZoneFromModel(row.ParkingZone, row.AvailableSpots)
	}
	return out
}
