package app

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/handler"
	appmw "github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"github.com/rayeemomayeer/SpotSync/internal/service"
	"gorm.io/gorm"
)

const defaultAuthRateLimitPerMinute = 20

type Options struct {
	AuthRateLimitPerMinute int
	EnableRequestLogger    bool
}

func NewEcho(cfg *config.Config, db *gorm.DB, log *slog.Logger, opts Options) *echo.Echo {
	if opts.AuthRateLimitPerMinute < 1 {
		opts.AuthRateLimitPerMinute = defaultAuthRateLimitPerMinute
	}

	userRepo := repository.NewUserRepository(db)
	zoneRepo := repository.NewZoneRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	spotRepo := repository.NewSpotRepository(db)

	tokenManager := platform.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiry)
	authSvc := service.NewAuthService(userRepo, tokenManager, cfg.BcryptCost, cfg.AllowSelfAdminRegistration)
	zoneSvc := service.NewZoneService(zoneRepo, spotRepo)
	reservationSvc := service.NewReservationService(reservationRepo, zoneRepo, cfg.DemoReservationTTL)
	spotSvc := service.NewSpotService(spotRepo, reservationRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	zoneHandler := handler.NewZoneHandler(zoneSvc)
	reservationHandler := handler.NewReservationHandler(reservationSvc)
	spotHandler := handler.NewSpotHandler(spotSvc)

	readiness := &handler.DBReadinessChecker{
		PingFn: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.PingContext(ctx)
		},
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = handler.NewValidator()
	e.HTTPErrorHandler = handler.HTTPErrorHandler

	e.Use(echomw.Recover())
	e.Use(appmw.RequestID())
	if opts.EnableRequestLogger && log != nil {
		e.Use(appmw.RequestLogger(log))
	}
	e.Use(appmw.CORS(cfg.CORSAllowedOrigins))

	health := handler.NewHealthHandler(readiness)
	e.GET("/healthz", health.Healthz)
	e.GET("/readyz", health.Readyz)

	jwtAuth := appmw.JWTAuth(tokenManager)
	requireAdmin := appmw.RequireAdmin()
	authRateLimit := appmw.IPRateLimit(opts.AuthRateLimitPerMinute)

	v1 := e.Group("/api/v1")

	auth := v1.Group("/auth", authRateLimit)
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.GET("/me", authHandler.Me, jwtAuth)

	zones := v1.Group("/zones")
	zones.GET("", zoneHandler.List)
	zones.GET("/:id", zoneHandler.GetByID)
	zones.GET("/:id/spots", spotHandler.ListByZone)
	zones.POST("", zoneHandler.Create, jwtAuth, requireAdmin)
	zones.PUT("/:id/spots/:spotId", spotHandler.UpdateStatus, jwtAuth, requireAdmin)

	reservations := v1.Group("/reservations", jwtAuth)
	reservations.POST("", reservationHandler.Create)
	reservations.GET("/my-reservations", reservationHandler.ListMine)
	reservations.DELETE("/:id", reservationHandler.Cancel)
	reservations.GET("", reservationHandler.ListAll, requireAdmin)

	return e
}
