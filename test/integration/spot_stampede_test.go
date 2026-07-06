//go:build integration

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/domain/spots"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"github.com/rayeemomayeer/SpotSync/test/testutil"
)

func TestReservationRepository_concurrentSameSpotStampede(t *testing.T) {
	db, _ := testutil.SetupPostgres(t)
	ctx := context.Background()

	zoneRepo := repository.NewZoneRepository(db)
	spotRepo := repository.NewSpotRepository(db)
	resRepo := repository.NewReservationRepository(db)
	userRepo := repository.NewUserRepository(db)

	zone := &models.ParkingZone{Name: "Single Spot Lot", Type: models.ZoneTypeGeneral, TotalCapacity: 1, PricePerHour: 5}
	if err := zoneRepo.Create(ctx, zone); err != nil {
		t.Fatal(err)
	}
	spotBatch := spots.GridLayout(zone.ID, 1)
	if err := spotRepo.CreateBatch(ctx, spotBatch); err != nil {
		t.Fatal(err)
	}
	zone.TotalCapacity = 1
	if err := zoneRepo.Update(ctx, zone); err != nil {
		t.Fatal(err)
	}
	list, err := spotRepo.ListByZone(ctx, zone.ID)
	if err != nil || len(list) != 1 {
		t.Fatalf("spots: %v err=%v", list, err)
	}
	targetSpotID := list[0].ID

	const workers = 30
	users := make([]*models.User, workers)
	for i := range users {
		users[i] = mustCreateUser(t, ctx, userRepo, testutil.UniqueEmail("spot-race"))
	}

	var wg sync.WaitGroup
	var successCount atomic.Int32
	var takenCount atomic.Int32

	start := make(chan struct{})
	for i := range workers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			spotID := targetSpotID
			_, err := resRepo.CreateActiveWithOptions(ctx, repository.CreateReservationParams{
				UserID:       users[idx].ID,
				ZoneID:       zone.ID,
				LicensePlate: fmt.Sprintf("S-%03d", idx),
				SpotID:       &spotID,
			})
			switch {
			case err == nil:
				successCount.Add(1)
			case errors.Is(err, domain.ErrSpotTaken), errors.Is(err, domain.ErrZoneFull):
				takenCount.Add(1)
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
	if got := takenCount.Load(); got != workers-1 {
		t.Fatalf("rejected reservations = %d, want %d", got, workers-1)
	}

	active, err := resRepo.CountActiveByZone(ctx, zone.ID)
	if err != nil {
		t.Fatal(err)
	}
	if active != 1 {
		t.Fatalf("active in db = %d, want 1", active)
	}
}

func TestDemoReservation_lazyCleanupOnSpotList(t *testing.T) {
	db, _ := testutil.SetupPostgres(t)
	ctx := context.Background()

	zoneRepo := repository.NewZoneRepository(db)
	spotRepo := repository.NewSpotRepository(db)
	resRepo := repository.NewReservationRepository(db)
	userRepo := repository.NewUserRepository(db)

	zone := &models.ParkingZone{Name: "Demo TTL Lot", Type: models.ZoneTypeGeneral, TotalCapacity: 1, PricePerHour: 3}
	if err := zoneRepo.Create(ctx, zone); err != nil {
		t.Fatal(err)
	}
	if err := spotRepo.CreateBatch(ctx, spots.GridLayout(zone.ID, 1)); err != nil {
		t.Fatal(err)
	}
	spotList, err := spotRepo.ListByZone(ctx, zone.ID)
	if err != nil || len(spotList) != 1 {
		t.Fatal(err)
	}
	spotID := spotList[0].ID

	user := mustCreateUser(t, ctx, userRepo, testutil.UniqueEmail("demo-ttl"))
	expired := time.Now().Add(-time.Minute)
	_, err = resRepo.CreateActiveWithOptions(ctx, repository.CreateReservationParams{
		UserID:        user.ID,
		ZoneID:        zone.ID,
		LicensePlate:  "DEMO-TEST",
		SpotID:        &spotID,
		DemoExpiresAt: &expired,
	})
	if err != nil {
		t.Fatal(err)
	}

	occupied, err := spotRepo.ActiveReservationSpotIDs(ctx, zone.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := occupied[spotID]; !ok {
		t.Fatal("spot should be occupied before cleanup")
	}

	if err := spotRepo.CancelExpiredDemoReservations(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}

	occupied, err = spotRepo.ActiveReservationSpotIDs(ctx, zone.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := occupied[spotID]; ok {
		t.Fatal("spot should be free after demo TTL cleanup")
	}
}

func TestReservationRepository_autoAssignSpot(t *testing.T) {
	db, _ := testutil.SetupPostgres(t)
	ctx := context.Background()

	zoneRepo := repository.NewZoneRepository(db)
	spotRepo := repository.NewSpotRepository(db)
	resRepo := repository.NewReservationRepository(db)
	userRepo := repository.NewUserRepository(db)

	zone := &models.ParkingZone{Name: "Auto Assign", Type: models.ZoneTypeGeneral, TotalCapacity: 2, PricePerHour: 2}
	if err := zoneRepo.Create(ctx, zone); err != nil {
		t.Fatal(err)
	}
	if err := spotRepo.CreateBatch(ctx, spots.GridLayout(zone.ID, 2)); err != nil {
		t.Fatal(err)
	}

	user := mustCreateUser(t, ctx, userRepo, testutil.UniqueEmail("auto"))
	res, err := resRepo.CreateActive(ctx, user.ID, zone.ID, "AUTO-1")
	if err != nil {
		t.Fatal(err)
	}
	if res.SpotID == nil {
		t.Fatal("expected auto-assigned spot_id")
	}
	if res.Spot == nil || res.Spot.ID == 0 {
		t.Fatal("expected spot preload on create response")
	}
}
