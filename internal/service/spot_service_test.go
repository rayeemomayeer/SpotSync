package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

type mockSpotStore struct {
	listFn     func(ctx context.Context, zoneID uint) ([]models.ParkingSpot, error)
	findFn     func(ctx context.Context, zoneID, spotID uint) (*models.ParkingSpot, error)
	updateFn   func(ctx context.Context, spotID uint, status string) error
	occupiedFn func(ctx context.Context, zoneID uint) (map[uint]struct{}, error)
	cancelFn   func(ctx context.Context, now time.Time) error
}

func (m *mockSpotStore) ListByZone(ctx context.Context, zoneID uint) ([]models.ParkingSpot, error) {
	return m.listFn(ctx, zoneID)
}

func (m *mockSpotStore) FindByIDInZone(ctx context.Context, zoneID, spotID uint) (*models.ParkingSpot, error) {
	return m.findFn(ctx, zoneID, spotID)
}

func (m *mockSpotStore) UpdateStatus(ctx context.Context, spotID uint, status string) error {
	return m.updateFn(ctx, spotID, status)
}

func (m *mockSpotStore) ActiveReservationSpotIDs(ctx context.Context, zoneID uint) (map[uint]struct{}, error) {
	return m.occupiedFn(ctx, zoneID)
}

func (m *mockSpotStore) CancelExpiredDemoReservations(ctx context.Context, now time.Time) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, now)
	}
	return nil
}

func TestSpotService_ListByZone(t *testing.T) {
	svc := service.NewSpotService(&mockSpotStore{
		listFn: func(_ context.Context, _ uint) ([]models.ParkingSpot, error) {
			return []models.ParkingSpot{{ID: 1, ZoneID: 1, Label: "A-01", Status: models.SpotStatusAvailable}}, nil
		},
		occupiedFn: func(_ context.Context, _ uint) (map[uint]struct{}, error) {
			return map[uint]struct{}{1: {}}, nil
		},
	}, &mockReservationStore{})

	out, err := svc.ListByZone(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || !out[0].Occupied {
		t.Fatalf("spots=%+v", out)
	}
}

func TestSpotService_BlockUnavailableWhenOccupied(t *testing.T) {
	svc := service.NewSpotService(&mockSpotStore{
		findFn: func(_ context.Context, _, _ uint) (*models.ParkingSpot, error) {
			return &models.ParkingSpot{ID: 5, ZoneID: 1, Status: models.SpotStatusAvailable}, nil
		},
	}, &mockReservationStore{
		hasActiveFn: func(_ context.Context, _ uint) (bool, error) { return true, nil },
	})

	_, err := svc.UpdateStatus(context.Background(), 1, 5, models.SpotStatusUnavailable)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}
