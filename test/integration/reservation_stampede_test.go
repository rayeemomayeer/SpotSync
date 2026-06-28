//go:build integration

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"github.com/rayeemomayeer/SpotSync/test/testutil"
)

func TestReservationRepository_concurrentStampede(t *testing.T) {
	db, _ := testutil.SetupPostgres(t)
	ctx := context.Background()

	zoneRepo := repository.NewZoneRepository(db)
	resRepo := repository.NewReservationRepository(db)
	userRepo := repository.NewUserRepository(db)

	zone := &models.ParkingZone{Name: "Stampede Zone", Type: models.ZoneTypeEVCharging, TotalCapacity: 1, PricePerHour: 10}
	if err := zoneRepo.Create(ctx, zone); err != nil {
		t.Fatal(err)
	}

	const workers = 50
	users := make([]*models.User, workers)
	for i := range users {
		email := testutil.UniqueEmail("stampede")
		users[i] = mustCreateUser(t, ctx, userRepo, email)
	}

	var wg sync.WaitGroup
	var successCount atomic.Int32
	var fullCount atomic.Int32

	start := make(chan struct{})
	for i := range workers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			_, err := resRepo.CreateActive(ctx, users[idx].ID, zone.ID, fmt.Sprintf("PLT-%03d", idx))
			switch {
			case err == nil:
				successCount.Add(1)
			case errors.Is(err, domain.ErrZoneFull):
				fullCount.Add(1)
			default:
				t.Errorf("worker %d: unexpected error: %v", idx, err)
			}
		}(i)
	}

	close(start)
	wg.Wait()

	if got := successCount.Load(); got != 1 {
		t.Fatalf("successful reservations = %d, want exactly 1", got)
	}

	count, err := resRepo.CountActiveByZone(ctx, zone.ID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("active reservations in db = %d, want 1", count)
	}

	if fullCount.Load() != workers-1 {
		t.Fatalf("zone full responses = %d, want %d", fullCount.Load(), workers-1)
	}
}
