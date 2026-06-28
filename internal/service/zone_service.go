package service

import (
	"context"

	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
)

type ZoneStore interface {
	Create(ctx context.Context, zone *models.ParkingZone) error
	ListWithAvailability(ctx context.Context) ([]repository.ZoneAvailabilityRow, error)
	GetByIDWithAvailability(ctx context.Context, id uint) (*repository.ZoneAvailabilityRow, error)
}

type ZoneService struct {
	zones ZoneStore
}

func NewZoneService(zones ZoneStore) *ZoneService {
	return &ZoneService{zones: zones}
}

func (s *ZoneService) Create(ctx context.Context, req dto.CreateZoneRequest) (dto.ZoneResponse, error) {
	zone := &models.ParkingZone{
		Name:          req.Name,
		Type:          req.Type,
		TotalCapacity: req.TotalCapacity,
		PricePerHour:  req.PricePerHour,
	}
	if err := s.zones.Create(ctx, zone); err != nil {
		return dto.ZoneResponse{}, err
	}
	return dto.ZoneFromModel(*zone, zone.TotalCapacity), nil
}

func (s *ZoneService) List(ctx context.Context) ([]dto.ZoneResponse, error) {
	rows, err := s.zones.ListWithAvailability(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ZoneResponse, len(rows))
	for i, row := range rows {
		out[i] = dto.ZoneFromModel(row.ParkingZone, row.AvailableSpots)
	}
	return out, nil
}

func (s *ZoneService) GetByID(ctx context.Context, id uint) (dto.ZoneResponse, error) {
	row, err := s.zones.GetByIDWithAvailability(ctx, id)
	if err != nil {
		return dto.ZoneResponse{}, err
	}
	return dto.ZoneFromModel(row.ParkingZone, row.AvailableSpots), nil
}
