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

type mockSpotZoneManager struct {
	createBatchFn func(ctx context.Context, spots []models.ParkingSpot) error
	countFn       func(ctx context.Context, zoneID uint) (int64, error)
	listFn        func(ctx context.Context, zoneID uint) ([]models.ParkingSpot, error)
	deleteFn      func(ctx context.Context, ids []uint) error
	occupiedFn    func(ctx context.Context, zoneID uint) (map[uint]struct{}, error)
}

func (m *mockSpotZoneManager) CreateBatch(ctx context.Context, spots []models.ParkingSpot) error {
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, spots)
	}
	return nil
}

func (m *mockSpotZoneManager) CountByZone(ctx context.Context, zoneID uint) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx, zoneID)
	}
	return 0, nil
}

func (m *mockSpotZoneManager) ListByZone(ctx context.Context, zoneID uint) ([]models.ParkingSpot, error) {
	if m.listFn != nil {
		return m.listFn(ctx, zoneID)
	}
	return nil, nil
}

func (m *mockSpotZoneManager) DeleteByIDs(ctx context.Context, ids []uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, ids)
	}
	return nil
}

func (m *mockSpotZoneManager) ActiveReservationSpotIDs(ctx context.Context, zoneID uint) (map[uint]struct{}, error) {
	if m.occupiedFn != nil {
		return m.occupiedFn(ctx, zoneID)
	}
	return map[uint]struct{}{}, nil
}

type mockZoneActiveCounter struct {
	countFn func(ctx context.Context, zoneID uint) (int64, error)
}

func (m *mockZoneActiveCounter) CountActiveByZone(ctx context.Context, zoneID uint) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx, zoneID)
	}
	return 0, nil
}

type mockZoneStoreForZone struct {
	getFn    func(ctx context.Context, id uint) (*repository.ZoneAvailabilityRow, error)
	updateFn func(ctx context.Context, zone *models.ParkingZone) error
	deleteFn func(ctx context.Context, id uint) error
	listFn   func(ctx context.Context, f repository.ZoneListFilter) ([]repository.ZoneAvailabilityRow, error)
}

func (m *mockZoneStoreForZone) Create(context.Context, *models.ParkingZone) error { return nil }

func (m *mockZoneStoreForZone) ListWithAvailability(ctx context.Context) ([]repository.ZoneAvailabilityRow, error) {
	return m.ListWithAvailabilityFiltered(ctx, repository.ZoneListFilter{})
}

func (m *mockZoneStoreForZone) ListWithAvailabilityFiltered(ctx context.Context, f repository.ZoneListFilter) ([]repository.ZoneAvailabilityRow, error) {
	if m.listFn != nil {
		return m.listFn(ctx, f)
	}
	return nil, nil
}

func (m *mockZoneStoreForZone) GetByIDWithAvailability(ctx context.Context, id uint) (*repository.ZoneAvailabilityRow, error) {
	return m.getFn(ctx, id)
}

func (m *mockZoneStoreForZone) FindByID(ctx context.Context, id uint) (*models.ParkingZone, error) {
	row, err := m.getFn(ctx, id)
	if err != nil {
		return nil, err
	}
	z := row.ParkingZone
	return &z, nil
}

func (m *mockZoneStoreForZone) Update(ctx context.Context, zone *models.ParkingZone) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, zone)
	}
	return nil
}

func (m *mockZoneStoreForZone) Delete(ctx context.Context, id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestZoneService_UpdateCapacityBelowActive(t *testing.T) {
	svc := service.NewZoneService(&mockZoneStoreForZone{
		getFn: func(_ context.Context, _ uint) (*repository.ZoneAvailabilityRow, error) {
			return &repository.ZoneAvailabilityRow{
				ParkingZone: models.ParkingZone{ID: 1, TotalCapacity: 10},
			}, nil
		},
	}, &mockSpotZoneManager{}, &mockZoneActiveCounter{
		countFn: func(context.Context, uint) (int64, error) { return 5, nil },
	}, nil)

	_, err := svc.Update(context.Background(), 1, dto.UpdateZoneRequest{
		Name: "Lot", Type: models.ZoneTypeGeneral, TotalCapacity: 4, PricePerHour: 3,
	})
	if !errors.Is(err, domain.ErrCapacityBelowActive) {
		t.Fatalf("error = %v", err)
	}
}

func TestZoneService_DeleteDelegates(t *testing.T) {
	called := false
	svc := service.NewZoneService(&mockZoneStoreForZone{
		deleteFn: func(_ context.Context, id uint) error {
			called = true
			if id != 9 {
				t.Fatalf("id=%d", id)
			}
			return nil
		},
	}, &mockSpotZoneManager{}, &mockZoneActiveCounter{}, nil)

	if err := svc.Delete(context.Background(), 9); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected delete call")
	}
}

func TestZoneService_ListPassesFilters(t *testing.T) {
	svc := service.NewZoneService(&mockZoneStoreForZone{
		listFn: func(_ context.Context, f repository.ZoneListFilter) ([]repository.ZoneAvailabilityRow, error) {
			if f.Type != "ev_charging" || f.Query != "Terminal" || f.Sort != "name" || f.Order != "desc" {
				t.Fatalf("filter=%+v", f)
			}
			return []repository.ZoneAvailabilityRow{
				{ParkingZone: models.ParkingZone{ID: 1, Name: "Terminal"}, AvailableSpots: 3},
			}, nil
		},
	}, &mockSpotZoneManager{}, &mockZoneActiveCounter{}, nil)

	zones, err := svc.List(context.Background(), dto.ZoneListQuery{
		Type: "ev_charging", Q: "Terminal", Sort: "name", Order: "desc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(zones) != 1 || zones[0].Name != "Terminal" {
		t.Fatalf("zones=%+v", zones)
	}
}

func TestZoneService_GetByIDScopedRequiresIsDemoFlag(t *testing.T) {
	// Simulates SELECT omitting is_demo (zero value) — demo GET must not 404 for true demo rows.
	svc := service.NewZoneService(&mockZoneStoreForZone{
		getFn: func(_ context.Context, id uint) (*repository.ZoneAvailabilityRow, error) {
			return &repository.ZoneAvailabilityRow{
				ParkingZone:    models.ParkingZone{ID: id, Name: "Demo Lot", IsDemo: true},
				AvailableSpots: 4,
			}, nil
		},
	}, &mockSpotZoneManager{}, &mockZoneActiveCounter{}, nil)

	zone, err := svc.GetByIDScoped(context.Background(), 4, true, "sess-1")
	if err != nil {
		t.Fatalf("demo zone visible: %v", err)
	}
	if zone.ID != 4 {
		t.Fatalf("id=%d", zone.ID)
	}

	svcMissing := service.NewZoneService(&mockZoneStoreForZone{
		getFn: func(_ context.Context, id uint) (*repository.ZoneAvailabilityRow, error) {
			return &repository.ZoneAvailabilityRow{
				ParkingZone: models.ParkingZone{ID: id, Name: "Missing flag"},
			}, nil
		},
	}, &mockSpotZoneManager{}, &mockZoneActiveCounter{}, nil)
	_, err = svcMissing.GetByIDScoped(context.Background(), 4, true, "sess-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found without is_demo, got %v", err)
	}
}
