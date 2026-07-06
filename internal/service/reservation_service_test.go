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
	createFn     func(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error)
	createOptsFn func(ctx context.Context, p repository.CreateReservationParams) (*models.Reservation, error)
	cancelFn     func(ctx context.Context, reservationID, userID uint) error
	listUser     func(ctx context.Context, userID uint) ([]models.Reservation, error)
	listAll      func(ctx context.Context, page, limit int) ([]models.Reservation, error)
	countAll     func(ctx context.Context) (int64, error)
	hasActiveFn  func(ctx context.Context, spotID uint) (bool, error)
}

func (m *mockReservationStore) CreateActive(ctx context.Context, userID, zoneID uint, licensePlate string) (*models.Reservation, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, zoneID, licensePlate)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReservationStore) CreateActiveWithOptions(ctx context.Context, p repository.CreateReservationParams) (*models.Reservation, error) {
	if m.createOptsFn != nil {
		return m.createOptsFn(ctx, p)
	}
	if m.createFn != nil {
		return m.createFn(ctx, p.UserID, p.ZoneID, p.LicensePlate)
	}
	return nil, errors.New("not implemented")
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

func (m *mockReservationStore) CountAll(ctx context.Context) (int64, error) {
	if m.countAll != nil {
		return m.countAll(ctx)
	}
	return 0, nil
}

func (m *mockReservationStore) HasActiveOnSpot(ctx context.Context, spotID uint) (bool, error) {
	if m.hasActiveFn != nil {
		return m.hasActiveFn(ctx, spotID)
	}
	return false, nil
}

type mockZoneStore struct {
	getFn    func(ctx context.Context, id uint) (*repository.ZoneAvailabilityRow, error)
	listFn   func(ctx context.Context, f repository.ZoneListFilter) ([]repository.ZoneAvailabilityRow, error)
	updateFn func(ctx context.Context, zone *models.ParkingZone) error
	deleteFn func(ctx context.Context, id uint) error
}

func (m *mockZoneStore) Create(context.Context, *models.ParkingZone) error { return nil }

func (m *mockZoneStore) ListWithAvailability(ctx context.Context) ([]repository.ZoneAvailabilityRow, error) {
	return m.ListWithAvailabilityFiltered(ctx, repository.ZoneListFilter{})
}

func (m *mockZoneStore) ListWithAvailabilityFiltered(ctx context.Context, f repository.ZoneListFilter) ([]repository.ZoneAvailabilityRow, error) {
	if m.listFn != nil {
		return m.listFn(ctx, f)
	}
	return nil, nil
}

func (m *mockZoneStore) GetByIDWithAvailability(ctx context.Context, id uint) (*repository.ZoneAvailabilityRow, error) {
	return m.getFn(ctx, id)
}

func (m *mockZoneStore) Update(ctx context.Context, zone *models.ParkingZone) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, zone)
	}
	return nil
}

func (m *mockZoneStore) Delete(ctx context.Context, id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestReservationService_CancelNotOwner(t *testing.T) {
	svc := service.NewReservationService(&mockReservationStore{
		cancelFn: func(_ context.Context, _, _ uint) error {
			return domain.ErrNotOwner
		},
	}, &mockZoneStore{}, 0)

	err := svc.Cancel(context.Background(), 1, 99)
	if !errors.Is(err, domain.ErrNotOwner) {
		t.Fatalf("error = %v", err)
	}
}

func TestReservationService_CreateZoneFull(t *testing.T) {
	svc := service.NewReservationService(&mockReservationStore{
		createOptsFn: func(_ context.Context, _ repository.CreateReservationParams) (*models.Reservation, error) {
			return nil, domain.ErrZoneFull
		},
	}, &mockZoneStore{}, 0)

	_, err := svc.Create(context.Background(), 1, dto.CreateReservationRequest{ZoneID: 1, LicensePlate: "ABC"}, service.CreateReservationOptions{})
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
	}, &mockZoneStore{}, 0)

	_, err := svc.ListAll(context.Background(), dto.PaginationQuery{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReservationService_ListAllWithPagination(t *testing.T) {
	svc := service.NewReservationService(&mockReservationStore{
		listAll: func(_ context.Context, page, limit int) ([]models.Reservation, error) {
			if page != 2 || limit != 10 {
				t.Fatalf("page=%d limit=%d, want page=2 limit=10", page, limit)
			}
			return []models.Reservation{{ID: 1}}, nil
		},
		countAll: func(context.Context) (int64, error) {
			return 25, nil
		},
	}, &mockZoneStore{}, 0)

	result, err := svc.ListAll(context.Background(), dto.PaginationQuery{Page: 2, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Paginated {
		t.Fatal("expected paginated result")
	}
	if result.Total != 25 {
		t.Fatalf("total=%d, want 25", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items=%d, want 1", len(result.Items))
	}
}
