package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL")))
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	name := strings.TrimSpace(os.Getenv("SEED_ADMIN_NAME"))
	if name == "" {
		name = "Admin"
	}

	if email == "" || password == "" {
		return fmt.Errorf("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD are required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := platform.NewLogger(cfg.LogLevel)
	slog.SetDefault(log)

	db, err := platform.OpenPostgres(cfg, log)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	if err := platform.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath, log); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	zoneRepo := repository.NewZoneRepository(db)
	resRepo := repository.NewReservationRepository(db)

	admin, err := ensureUser(ctx, userRepo, cfg.BcryptCost, name, email, password, models.RoleAdmin, log)
	if err != nil {
		return err
	}

	driverAlice, err := ensureUser(ctx, userRepo, cfg.BcryptCost, "Alice Driver", "alice@spotsync.com", "DriverPass123!", models.RoleDriver, log)
	if err != nil {
		return err
	}

	driverBob, err := ensureUser(ctx, userRepo, cfg.BcryptCost, "Bob Driver", "bob@spotsync.com", "DriverPass123!", models.RoleDriver, log)
	if err != nil {
		return err
	}

	zones, err := ensureZones(ctx, zoneRepo, log)
	if err != nil {
		return err
	}

	if err := ensureReservations(ctx, resRepo, driverAlice, driverBob, zones, log); err != nil {
		return err
	}

	log.Info("seed complete",
		"admin_email", admin.Email,
		"driver_emails", []string{driverAlice.Email, driverBob.Email},
		"zones", len(zones),
	)
	return nil
}

func ensureUser(
	ctx context.Context,
	repo *repository.UserRepository,
	bcryptCost int,
	name, email, password, role string,
	log *slog.Logger,
) (*models.User, error) {
	existing, err := repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		log.Info("user already exists", "email", email, "role", existing.Role)
		return existing, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		Role:     role,
	}
	if err := repo.Create(ctx, user); err != nil {
		return nil, err
	}

	log.Info("user created", "email", email, "role", role, "id", user.ID)
	return user, nil
}

type demoZone struct {
	Name          string
	Type          string
	TotalCapacity int
	PricePerHour  float64
}

var demoZones = []demoZone{
	{Name: "Downtown Garage", Type: models.ZoneTypeGeneral, TotalCapacity: 50, PricePerHour: 4.50},
	{Name: "EV Charging Hub", Type: models.ZoneTypeEVCharging, TotalCapacity: 12, PricePerHour: 6.00},
	{Name: "Airport Covered Lot", Type: models.ZoneTypeCovered, TotalCapacity: 30, PricePerHour: 8.25},
}

func ensureZones(ctx context.Context, repo *repository.ZoneRepository, log *slog.Logger) ([]models.ParkingZone, error) {
	existing, err := repo.ListWithAvailability(ctx)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		log.Info("zones already exist", "count", len(existing))
		zones := make([]models.ParkingZone, len(existing))
		for i, row := range existing {
			zones[i] = row.ParkingZone
		}
		return zones, nil
	}

	zones := make([]models.ParkingZone, 0, len(demoZones))
	for _, spec := range demoZones {
		zone := &models.ParkingZone{
			Name:          spec.Name,
			Type:          spec.Type,
			TotalCapacity: spec.TotalCapacity,
			PricePerHour:  spec.PricePerHour,
		}
		if err := repo.Create(ctx, zone); err != nil {
			return nil, err
		}
		log.Info("zone created", "name", zone.Name, "id", zone.ID, "capacity", zone.TotalCapacity)
		zones = append(zones, *zone)
	}
	return zones, nil
}

type demoReservation struct {
	User         *models.User
	ZoneIndex    int
	LicensePlate string
}

func ensureReservations(
	ctx context.Context,
	repo *repository.ReservationRepository,
	alice, bob *models.User,
	zones []models.ParkingZone,
	log *slog.Logger,
) error {
	if len(zones) == 0 {
		return nil
	}

	aliceRes, err := repo.ListByUser(ctx, alice.ID)
	if err != nil {
		return err
	}
	if len(aliceRes) > 0 {
		log.Info("reservations already exist", "alice_count", len(aliceRes))
		return nil
	}

	plans := []demoReservation{
		{User: alice, ZoneIndex: 0, LicensePlate: "ABC-1234"},
		{User: alice, ZoneIndex: 1, LicensePlate: "ABC-1234"},
		{User: bob, ZoneIndex: 0, LicensePlate: "XYZ-9876"},
		{User: bob, ZoneIndex: 2, LicensePlate: "XYZ-9876"},
	}

	for _, plan := range plans {
		if plan.ZoneIndex >= len(zones) {
			continue
		}
		zone := zones[plan.ZoneIndex]
		res, err := repo.CreateActive(ctx, plan.User.ID, zone.ID, plan.LicensePlate)
		if err != nil {
			return fmt.Errorf("create reservation for %s in zone %q: %w", plan.User.Email, zone.Name, err)
		}
		log.Info("reservation created",
			"id", res.ID,
			"user", plan.User.Email,
			"zone", zone.Name,
			"plate", plan.LicensePlate,
		)
	}
	return nil
}
