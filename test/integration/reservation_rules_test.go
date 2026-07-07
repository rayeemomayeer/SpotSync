//go:build integration

package integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"github.com/rayeemomayeer/SpotSync/test/testutil"
	"golang.org/x/crypto/bcrypt"
)

func TestReservationRepository_capacityRule(t *testing.T) {
	db, _ := testutil.SetupPostgres(t)
	ctx := context.Background()

	zoneRepo := repository.NewZoneRepository(db)
	resRepo := repository.NewReservationRepository(db)
	userRepo := repository.NewUserRepository(db)

	zone := &models.ParkingZone{Name: "Single Spot", Type: models.ZoneTypeGeneral, TotalCapacity: 1, PricePerHour: 5}
	if err := zoneRepo.Create(ctx, zone); err != nil {
		t.Fatal(err)
	}

	userA := mustCreateUser(t, ctx, userRepo, "driver-a@test.com")
	userB := mustCreateUser(t, ctx, userRepo, "driver-b@test.com")

	tests := []struct {
		name      string
		userID    uint
		plate     string
		wantErr   error
		wantCount int64
	}{
		{
			name:      "first reservation succeeds",
			userID:    userA.ID,
			plate:     "AAA-1111",
			wantErr:   nil,
			wantCount: 1,
		},
		{
			name:      "second reservation rejected when full",
			userID:    userB.ID,
			plate:     "BBB-2222",
			wantErr:   domain.ErrZoneFull,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resRepo.CreateActive(ctx, tt.userID, zone.ID, tt.plate)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateActive() error = %v, want %v", err, tt.wantErr)
			}

			count, err := resRepo.CountActiveByZone(ctx, zone.ID)
			if err != nil {
				t.Fatal(err)
			}
			if count != tt.wantCount {
				t.Fatalf("active count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestReservationRepository_ownershipRule(t *testing.T) {
	db, _ := testutil.SetupPostgres(t)
	ctx := context.Background()

	zoneRepo := repository.NewZoneRepository(db)
	resRepo := repository.NewReservationRepository(db)
	userRepo := repository.NewUserRepository(db)

	zone := &models.ParkingZone{Name: "Owner Test", Type: models.ZoneTypeGeneral, TotalCapacity: 5, PricePerHour: 3}
	if err := zoneRepo.Create(ctx, zone); err != nil {
		t.Fatal(err)
	}

	owner := mustCreateUser(t, ctx, userRepo, "owner@test.com")
	other := mustCreateUser(t, ctx, userRepo, "other@test.com")

	res, err := resRepo.CreateActive(ctx, owner.ID, zone.ID, "OWN-1234")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		userID  uint
		wantErr error
	}{
		{name: "non-owner cannot cancel", userID: other.ID, wantErr: domain.ErrNotOwner},
		{name: "owner can cancel", userID: owner.ID, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resRepo.Cancel(ctx, res.ID, tt.userID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Cancel() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	var updated models.Reservation
	if err := db.First(&updated, res.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Status != models.ReservationStatusCancelled {
		t.Fatalf("status = %q, want cancelled", updated.Status)
	}
}

func mustCreateUser(t *testing.T, ctx context.Context, repo *repository.UserRepository, email string) *models.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	user := &models.User{
		Name:     "Test User",
		Email:    email,
		Password: string(hash),
		Role:     models.RoleDriver,
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}
	return user
}
