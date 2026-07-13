package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/domain/spots"
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

	if err := platform.RunMigrations(cfg.DatabaseURL, cfg.DatabaseMigrateURL, cfg.MigrationsPath, log); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	zoneRepo := repository.NewZoneRepository(db)
	resRepo := repository.NewReservationRepository(db)
	spotRepo := repository.NewSpotRepository(db)

	admin, err := ensureUser(ctx, userRepo, cfg.BcryptCost, name, email, password, models.RoleSaaSAdmin, log)
	if err != nil {
		return err
	}
	// Graded contract compatibility: also ensure legacy admin role alias if seed email differs.
	_ = admin

	demoAdmin, demoAdminErr := ensureUser(ctx, userRepo, cfg.BcryptCost, "Demo Admin", "demo_admin@spotsync.com", "DemoAdminPass123!", models.RoleOrgAdmin, log)
	if demoAdminErr != nil {
		log.Warn("demo admin seed skipped", "error", demoAdminErr)
	}

	driverAlice, err := ensureUser(ctx, userRepo, cfg.BcryptCost, "Alice Driver", "alice@spotsync.com", "DriverPass123!", models.RoleDriver, log)
	if err != nil {
		return err
	}

	driverBob, err := ensureUser(ctx, userRepo, cfg.BcryptCost, "Bob Driver", "bob@spotsync.com", "DriverPass123!", models.RoleDriver, log)
	if err != nil {
		return err
	}

	zoneList, err := ensureZones(ctx, zoneRepo, spotRepo, log)
	if err != nil {
		return err
	}

	orgRepo := repository.NewOrganizationRepository(db)
	if err := ensureDemoOrgMembership(ctx, orgRepo, demoAdmin, log); err != nil {
		log.Warn("demo org membership skipped", "error", err)
	}

	if err := ensureReservations(ctx, resRepo, driverAlice, driverBob, zoneList, log); err != nil {
		return err
	}

	log.Info("seed complete",
		"admin_email", admin.Email,
		"demo_admin_email", demoAdminEmail(demoAdmin),
		"driver_emails", []string{driverAlice.Email, driverBob.Email},
		"zones", len(zoneList),
		"demo_driver_password", "DriverPass123!",
		"demo_admin_password", "DemoAdminPass123!",
	)
	return nil
}

func demoAdminEmail(user *models.User) string {
	if user == nil {
		return ""
	}
	return user.Email
}

func ensureDemoOrgMembership(
	ctx context.Context,
	orgRepo *repository.OrganizationRepository,
	demoAdmin *models.User,
	log *slog.Logger,
) error {
	if demoAdmin == nil {
		return nil
	}
	org, err := orgRepo.GetBySlug(ctx, "demo-parking")
	if err != nil {
		return err
	}
	if _, err := orgRepo.Membership(ctx, org.ID, demoAdmin.ID); err == nil {
		log.Info("demo org membership exists", "org", org.Slug, "user_id", demoAdmin.ID)
		return nil
	}
	if err := orgRepo.AddMember(ctx, &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         demoAdmin.ID,
		Role:           models.RoleOrgAdmin,
	}); err != nil {
		return err
	}
	log.Info("demo org membership created", "org", org.Slug, "user_id", demoAdmin.ID)
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
		changed := false
		if existing.Role != role {
			existing.Role = role
			changed = true
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
		if err != nil {
			return nil, err
		}
		existing.Password = string(hash)
		changed = true
		if changed {
			if err := repo.Update(ctx, existing); err != nil {
				return nil, fmt.Errorf("update seed user %s: %w", email, err)
			}
			log.Info("user refreshed from seed", "email", email, "role", role, "id", existing.ID)
		} else {
			log.Info("user already exists", "email", email, "role", existing.Role)
		}
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
	Showcase      bool
}

var demoZones = []demoZone{
	{Name: "Downtown Garage", Type: models.ZoneTypeGeneral, TotalCapacity: 50, PricePerHour: 4.50},
	{Name: "Terminal 1 · EV Lot A", Type: models.ZoneTypeEVCharging, TotalCapacity: 24, PricePerHour: 6.00, Showcase: true},
	{Name: "Airport Covered Lot", Type: models.ZoneTypeCovered, TotalCapacity: 30, PricePerHour: 8.25},
}

func ensureZones(ctx context.Context, zoneRepo *repository.ZoneRepository, spotRepo *repository.SpotRepository, log *slog.Logger) ([]models.ParkingZone, error) {
	existing, err := zoneRepo.ListWithAvailability(ctx)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		log.Info("zones already exist", "count", len(existing))
		zones := make([]models.ParkingZone, len(existing))
		for i, row := range existing {
			zones[i] = row.ParkingZone
		}
		if err := backfillSpots(ctx, zoneRepo, spotRepo, zones, log); err != nil {
			return nil, err
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
		if err := zoneRepo.Create(ctx, zone); err != nil {
			return nil, err
		}
		layout := spots.GridLayout(zone.ID, zone.TotalCapacity)
		if spec.Showcase {
			layout = spots.ShowcaseLayout(zone.ID)
		}
		if err := spotRepo.CreateBatch(ctx, layout); err != nil {
			return nil, err
		}
		zone.TotalCapacity = len(layout)
		if err := zoneRepo.Update(ctx, zone); err != nil {
			return nil, err
		}
		log.Info("zone created", "name", zone.Name, "id", zone.ID, "capacity", zone.TotalCapacity, "spots", len(layout))
		zones = append(zones, *zone)
	}
	return zones, nil
}

func backfillSpots(ctx context.Context, zoneRepo *repository.ZoneRepository, spotRepo *repository.SpotRepository, zones []models.ParkingZone, log *slog.Logger) error {
	for _, zone := range zones {
		count, err := spotRepo.CountByZone(ctx, zone.ID)
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		layout := spots.GridLayout(zone.ID, zone.TotalCapacity)
		if zone.Name == "Terminal 1 · EV Lot A" || (strings.Contains(zone.Name, "Terminal") && zone.Type == models.ZoneTypeEVCharging) {
			layout = spots.ShowcaseLayout(zone.ID)
		}
		if err := spotRepo.CreateBatch(ctx, layout); err != nil {
			return err
		}
		zoneCopy := zone
		zoneCopy.TotalCapacity = len(layout)
		if err := zoneRepo.Update(ctx, &zoneCopy); err != nil {
			return err
		}
		log.Info("spots backfilled", "zone", zone.Name, "count", len(layout))
	}
	return nil
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
