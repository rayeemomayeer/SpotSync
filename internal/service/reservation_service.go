package service

import (
	"context"

	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
)

type ReservationStore interface {
	CreateActive(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error)
	Cancel(ctx context.Context, reservationID, userID uint) error
	ListByUser(ctx context.Context, userID uint) ([]models.Reservation, error)
	ListAll(ctx context.Context, page, limit int) ([]models.Reservation, error)
}

type ReservationService struct {
	reservations ReservationStore
	zones        ZoneStore
}

func NewReservationService(reservations ReservationStore, zones ZoneStore) *ReservationService {
	return &ReservationService{reservations: reservations, zones: zones}
}

func (s *ReservationService) Create(ctx context.Context, userID uint, req dto.CreateReservationRequest) (dto.ReservationResponse, error) {
	res, err := s.reservations.CreateActive(ctx, userID, req.ZoneID, req.LicensePlate)
	if err != nil {
		return dto.ReservationResponse{}, err
	}
	return dto.ReservationFromModel(*res), nil
}

func (s *ReservationService) Cancel(ctx context.Context, userID, reservationID uint) error {
	return s.reservations.Cancel(ctx, reservationID, userID)
}

func (s *ReservationService) ListMine(ctx context.Context, userID uint) ([]dto.ReservationResponse, error) {
	list, err := s.reservations.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.mapReservationsWithZones(ctx, list)
}

func (s *ReservationService) ListAll(ctx context.Context, q dto.PaginationQuery) ([]dto.ReservationResponse, error) {
	page, limit := 0, 0
	if q.Page > 0 || q.Limit > 0 {
		page = q.Page
		limit = q.Limit
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 20
		}
	}

	list, err := s.reservations.ListAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	return s.mapReservationsWithZones(ctx, list)
}

func (s *ReservationService) mapReservationsWithZones(ctx context.Context, list []models.Reservation) ([]dto.ReservationResponse, error) {
	out := make([]dto.ReservationResponse, len(list))
	zoneCache := make(map[uint]dto.ZoneResponse)

	for i, res := range list {
		out[i] = dto.ReservationFromModel(res)
		if res.Zone.ID == 0 {
			continue
		}
		if cached, ok := zoneCache[res.ZoneID]; ok {
			out[i].Zone = &cached
			continue
		}
		zone, err := s.zones.GetByIDWithAvailability(ctx, res.ZoneID)
		if err != nil {
			return nil, err
		}
		zr := dto.ZoneFromModel(zone.ParkingZone, zone.AvailableSpots)
		zoneCache[res.ZoneID] = zr
		out[i].Zone = &zr
	}
	return out, nil
}
