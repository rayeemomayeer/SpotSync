package service

import (
	"context"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
)

const defaultDemoReservationTTL = 10 * time.Minute

type SpotStore interface {
	ListByZone(ctx context.Context, zoneID uint) ([]models.ParkingSpot, error)
	FindByIDInZone(ctx context.Context, zoneID, spotID uint) (*models.ParkingSpot, error)
	UpdateStatus(ctx context.Context, spotID uint, status string) error
	ActiveReservationSpotIDs(ctx context.Context, zoneID uint) (map[uint]struct{}, error)
	CancelExpiredDemoReservations(ctx context.Context, now time.Time) error
}

type SpotService struct {
	spots        SpotStore
	reservations ReservationStore
}

func NewSpotService(spots SpotStore, reservations ReservationStore) *SpotService {
	return &SpotService{spots: spots, reservations: reservations}
}

func (s *SpotService) ListByZone(ctx context.Context, zoneID uint) ([]dto.SpotResponse, error) {
	if err := s.spots.CancelExpiredDemoReservations(ctx, time.Now()); err != nil {
		return nil, err
	}

	spots, err := s.spots.ListByZone(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	occupied, err := s.spots.ActiveReservationSpotIDs(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	out := make([]dto.SpotResponse, len(spots))
	for i, spot := range spots {
		_, isOccupied := occupied[spot.ID]
		out[i] = dto.SpotFromModel(spot, isOccupied)
	}
	return out, nil
}

func (s *SpotService) UpdateStatus(ctx context.Context, zoneID, spotID uint, status string) (dto.SpotResponse, error) {
	spot, err := s.spots.FindByIDInZone(ctx, zoneID, spotID)
	if err != nil {
		return dto.SpotResponse{}, err
	}

	if status == models.SpotStatusUnavailable {
		hasActive, err := s.reservations.HasActiveOnSpot(ctx, spotID)
		if err != nil {
			return dto.SpotResponse{}, err
		}
		if hasActive {
			return dto.SpotResponse{}, domain.ErrConflict
		}
	}

	if err := s.spots.UpdateStatus(ctx, spot.ID, status); err != nil {
		return dto.SpotResponse{}, err
	}
	spot.Status = status

	occupied, err := s.spots.ActiveReservationSpotIDs(ctx, zoneID)
	if err != nil {
		return dto.SpotResponse{}, err
	}
	_, isOccupied := occupied[spot.ID]
	return dto.SpotFromModel(*spot, isOccupied), nil
}
