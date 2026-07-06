package service

import (
	"context"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
)

type ReservationStore interface {
	CreateActive(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error)
	CreateActiveWithOptions(ctx context.Context, p repository.CreateReservationParams) (*models.Reservation, error)
	Cancel(ctx context.Context, reservationID, userID uint) error
	ListByUser(ctx context.Context, userID uint) ([]models.Reservation, error)
	ListAll(ctx context.Context, page, limit int) ([]models.Reservation, error)
	CountAll(ctx context.Context) (int64, error)
	HasActiveOnSpot(ctx context.Context, spotID uint) (bool, error)
}

type ListAllResult struct {
	Items     []dto.ReservationResponse
	Total     int64
	Page      int
	Limit     int
	Paginated bool
}

type CreateReservationOptions struct {
	DemoReservation bool
}

type ReservationService struct {
	reservations ReservationStore
	zones        ZoneStore
	demoTTL      time.Duration
}

func NewReservationService(reservations ReservationStore, zones ZoneStore, demoTTL time.Duration) *ReservationService {
	if demoTTL < 1 {
		demoTTL = defaultDemoReservationTTL
	}
	return &ReservationService{reservations: reservations, zones: zones, demoTTL: demoTTL}
}

func (s *ReservationService) Create(ctx context.Context, userID uint, req dto.CreateReservationRequest, opts CreateReservationOptions) (dto.ReservationResponse, error) {
	params := repository.CreateReservationParams{
		UserID:       userID,
		ZoneID:       req.ZoneID,
		LicensePlate: req.LicensePlate,
		SpotID:       req.SpotID,
	}
	if opts.DemoReservation {
		expires := time.Now().Add(s.demoTTL)
		params.DemoExpiresAt = &expires
	}

	res, err := s.reservations.CreateActiveWithOptions(ctx, params)
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

func (s *ReservationService) ListAll(ctx context.Context, q dto.PaginationQuery) (ListAllResult, error) {
	page, limit := 0, 0
	paginated := q.Page > 0 || q.Limit > 0
	if paginated {
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
		return ListAllResult{}, err
	}
	items, err := s.mapReservationsWithZones(ctx, list)
	if err != nil {
		return ListAllResult{}, err
	}

	var total int64
	if paginated {
		total, err = s.reservations.CountAll(ctx)
		if err != nil {
			return ListAllResult{}, err
		}
	}

	return ListAllResult{
		Items:     items,
		Total:     total,
		Page:      page,
		Limit:     limit,
		Paginated: paginated,
	}, nil
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
