package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort              = "8080"
	defaultJWTExpiry         = 24 * time.Hour
	defaultBcryptCost        = 12
	defaultLogLevel          = "info"
	defaultDBMaxOpenConns    = 25
	defaultDBMaxIdleConns    = 5
	defaultDBConnMaxLifetime = 5 * time.Minute
	defaultMigrateOnStartup  = true
	defaultDemoReservationTTL = 10 * time.Minute
)

// Config holds validated application settings loaded from the environment.
type Config struct {
	Port                       string
	DatabaseURL                string
	DatabaseMigrateURL         string
	JWTSecret                  string
	JWTExpiry                  time.Duration
	BcryptCost                 int
	AllowSelfAdminRegistration bool
	CORSAllowedOrigins         []string
	LogLevel                   string
	DBMaxOpenConns             int
	DBMaxIdleConns             int
	DBConnMaxLifetime          time.Duration
	MigrateOnStartup           bool
	MigrationsPath             string
	DemoReservationTTL         time.Duration
	RedisURL                   string
	CapacityStrategy           string
}

// Load reads configuration from the environment and fails fast on missing required values.
// A local .env file is loaded when present (ignored in production).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:              envOrDefault("PORT", defaultPort),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:         strings.TrimSpace(os.Getenv("JWT_SECRET")),
		BcryptCost:        defaultBcryptCost,
		LogLevel:          envOrDefault("LOG_LEVEL", defaultLogLevel),
		DBMaxOpenConns:    defaultDBMaxOpenConns,
		DBMaxIdleConns:    defaultDBMaxIdleConns,
		DBConnMaxLifetime: defaultDBConnMaxLifetime,
		MigrateOnStartup:  defaultMigrateOnStartup,
		MigrationsPath:    envOrDefault("MIGRATIONS_PATH", "migrations"),
	}

	if err := cfg.parseOptionalFields(); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) parseOptionalFields() error {
	jwtExpiryRaw := strings.TrimSpace(os.Getenv("JWT_EXPIRY"))
	if jwtExpiryRaw == "" {
		c.JWTExpiry = defaultJWTExpiry
	} else {
		d, err := time.ParseDuration(jwtExpiryRaw)
		if err != nil {
			return fmt.Errorf("JWT_EXPIRY: invalid duration %q: %w", jwtExpiryRaw, err)
		}
		c.JWTExpiry = d
	}

	if raw := strings.TrimSpace(os.Getenv("BCRYPT_COST")); raw != "" {
		cost, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("BCRYPT_COST: must be an integer: %w", err)
		}
		c.BcryptCost = cost
	}

	c.AllowSelfAdminRegistration = parseBoolEnv("ALLOW_SELF_ADMIN_REGISTRATION", true)

	if raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); raw != "" {
		for _, origin := range strings.Split(raw, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				c.CORSAllowedOrigins = append(c.CORSAllowedOrigins, origin)
			}
		}
	}

	if raw := strings.TrimSpace(os.Getenv("DB_MAX_OPEN_CONNS")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("DB_MAX_OPEN_CONNS: must be an integer: %w", err)
		}
		c.DBMaxOpenConns = n
	}

	if raw := strings.TrimSpace(os.Getenv("DB_MAX_IDLE_CONNS")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("DB_MAX_IDLE_CONNS: must be an integer: %w", err)
		}
		c.DBMaxIdleConns = n
	}

	if raw := strings.TrimSpace(os.Getenv("DB_CONN_MAX_LIFETIME")); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("DB_CONN_MAX_LIFETIME: invalid duration %q: %w", raw, err)
		}
		c.DBConnMaxLifetime = d
	}

	c.MigrateOnStartup = parseBoolEnv("MIGRATE_ON_STARTUP", defaultMigrateOnStartup)

	if raw := strings.TrimSpace(os.Getenv("DEMO_RESERVATION_TTL")); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("DEMO_RESERVATION_TTL: invalid duration %q: %w", raw, err)
		}
		c.DemoReservationTTL = d
	} else {
		c.DemoReservationTTL = defaultDemoReservationTTL
	}

	c.RedisURL = strings.TrimSpace(os.Getenv("REDIS_URL"))
	c.DatabaseMigrateURL = strings.TrimSpace(os.Getenv("DATABASE_MIGRATE_URL"))
	c.CapacityStrategy = envOrDefault("CAPACITY_STRATEGY", "row_lock")

	return nil
}

func (c *Config) validate() error {
	var missing []string

	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}

	if c.BcryptCost < 10 || c.BcryptCost > 14 {
		return errors.New("BCRYPT_COST: must be between 10 and 14")
	}
	if c.DBMaxOpenConns < 1 {
		return errors.New("DB_MAX_OPEN_CONNS: must be at least 1")
	}
	if c.DBMaxIdleConns < 0 {
		return errors.New("DB_MAX_IDLE_CONNS: must be non-negative")
	}
	if c.DBMaxIdleConns > c.DBMaxOpenConns {
		return errors.New("DB_MAX_IDLE_CONNS: must not exceed DB_MAX_OPEN_CONNS")
	}

	return nil
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func parseBoolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}
