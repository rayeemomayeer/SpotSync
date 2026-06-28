package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

type mockReservationStore struct {
	createFn func(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error)
	cancelFn func(ctx context.Context, reservationID, userID uint) error
	listUser func(ctx context.Context, userID uint) ([]models.Reservation, error)
	listAll  func(ctx context.Context, page, limit int) ([]models.Reservation, error)
}

func (m *mockReservationStore) CreateActive(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error) {
	return m.createFn(ctx, userID, zoneID, licensePlate)
}

func (m *mockReservationStore) Cancel(ctx context.Context, reservationID, userID uint) error {
	return m.cancelFn(ctx, reservationID, userID)
}

func (m *mockReservationStore) ListByUser(ctx context.Context, userID uint) ([]models.Reservation, error) {
	return m.listUser(ctx, userID)
}

func (m *mockReservationStore) ListAll(ctx context.Context, page, limit int) ([]models.Reservation, error) {
	return m.listAll(ctx, page, limit)
}

type mockZoneStore struct {
	getFn func(ctx context.Context, id uint) (*repository.ZoneAvailabilityRow, error)
}

func (m *mockZoneStore) Create(context.Context, *models.ParkingZone) error { return nil }

func (m *mockZoneStore) ListWithAvailability(context.Context) ([]repository.ZoneAvailabilityRow, error) {
	return nil, nil
}

func (m *mockZoneStore) GetByIDWithAvailability(ctx context.Context, id uint) (*repository.ZoneAvailabilityRow, error) {
	return m.getFn(ctx, id)
}

func TestReservationService_CancelNotOwner(t *testing.T) {
	svc := service.NewReservationService(&mockReservationStore{
		cancelFn: func(_ context.Context, _, _ uint) error {
			return domain.ErrNotOwner
		},
	}, &mockZoneStore{})

	err := svc.Cancel(context.Background(), 1, 99)
	if !errors.Is(err, domain.ErrNotOwner) {
		t.Fatalf("error = %v", err)
	}
}

func TestReservationService_CreateZoneFull(t *testing.T) {
	svc := service.NewReservationService(&mockReservationStore{
		createFn: func(_ context.Context, _, _ uint, _ string) (*models.Reservation, error) {
			return nil, domain.ErrZoneFull
		},
	}, &mockZoneStore{})

	_, err := svc.Create(context.Background(), 1, dto.CreateReservationRequest{ZoneID: 1, LicensePlate: "ABC"})
	if !errors.Is(err, domain.ErrZoneFull) {
		t.Fatalf("error = %v", err)
	}
}

func TestReservationService_ListAllNoPagination(t *testing.T) {
	svc := service.NewReservationService(&mockReservationStore{
		listAll: func(_ context.Context, page, limit int) ([]models.Reservation, error) {
			if page != 0 || limit != 0 {
				t.Fatalf("page=%d limit=%d, want both 0", page, limit)
			}
			return []models.Reservation{}, nil
		},
	}, &mockZoneStore{})

	_, err := svc.ListAll(context.Background(), dto.PaginationQuery{})
	if err != nil {
		t.Fatal(err)
	}
}
