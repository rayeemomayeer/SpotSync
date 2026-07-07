package capacity

import (
	"context"
	"errors"
	"fmt"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
)

const (
	StrategyRowLock    = "row_lock"
	StrategyOptimistic = "optimistic"
	StrategyRedis      = "redis_counter"
)

// Guard atomically checks capacity and records a reservation.
type Guard interface {
	Reserve(ctx context.Context, p repository.CreateReservationParams) (*models.Reservation, error)
	ReleaseCapacity(ctx context.Context, zoneID uint) error
}

type RowLockGuard struct {
	repo *repository.ReservationRepository
}

func NewRowLockGuard(repo *repository.ReservationRepository) RowLockGuard {
	return RowLockGuard{repo: repo}
}

func (g RowLockGuard) Reserve(ctx context.Context, p repository.CreateReservationParams) (*models.Reservation, error) {
	return g.repo.CreateActiveWithOptions(ctx, p)
}

func (g RowLockGuard) ReleaseCapacity(context.Context, uint) error {
	return nil
}

type OptimisticGuard struct {
	repo *repository.ReservationRepository
}

func NewOptimisticGuard(repo *repository.ReservationRepository) OptimisticGuard {
	return OptimisticGuard{repo: repo}
}

func (g OptimisticGuard) Reserve(ctx context.Context, p repository.CreateReservationParams) (*models.Reservation, error) {
	return g.repo.CreateActiveOptimistic(ctx, p)
}

func (g OptimisticGuard) ReleaseCapacity(context.Context, uint) error {
	return nil
}

type RedisCounterGuard struct {
	repo  *repository.ReservationRepository
	zones *repository.ZoneRepository
	redis *platform.RedisClient
}

func NewRedisCounterGuard(
	repo *repository.ReservationRepository,
	zones *repository.ZoneRepository,
	redis *platform.RedisClient,
) (RedisCounterGuard, error) {
	if redis == nil {
		return RedisCounterGuard{}, fmt.Errorf("redis_counter strategy requires REDIS_URL")
	}
	return RedisCounterGuard{repo: repo, zones: zones, redis: redis}, nil
}

func (g RedisCounterGuard) Reserve(ctx context.Context, p repository.CreateReservationParams) (*models.Reservation, error) {
	if err := g.ensureCounter(ctx, p.ZoneID); err != nil {
		return nil, err
	}

	n, err := g.redis.DecrIfPositive(ctx, platform.ZoneCapacityKey(p.ZoneID))
	if err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, domain.ErrZoneFull
	}

	res, err := g.repo.CreateActiveWithOptions(ctx, p)
	if err != nil {
		_ = g.redis.Incr(ctx, platform.ZoneCapacityKey(p.ZoneID))
		return nil, err
	}
	return res, nil
}

func (g RedisCounterGuard) ReleaseCapacity(ctx context.Context, zoneID uint) error {
	return g.redis.Incr(ctx, platform.ZoneCapacityKey(zoneID))
}

func (g RedisCounterGuard) ensureCounter(ctx context.Context, zoneID uint) error {
	key := platform.ZoneCapacityKey(zoneID)
	exists, err := g.redis.Exists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	zone, err := g.zones.FindByID(ctx, zoneID)
	if err != nil {
		return err
	}
	active, err := g.repo.CountActiveByZone(ctx, zoneID)
	if err != nil {
		return err
	}
	remaining := int64(zone.TotalCapacity) - active
	if remaining < 0 {
		remaining = 0
	}
	return g.redis.SetNX(ctx, key, remaining)
}

func NewGuard(strategy string, repo *repository.ReservationRepository, zones *repository.ZoneRepository, redis *platform.RedisClient) (Guard, error) {
	switch strategy {
	case StrategyOptimistic:
		return NewOptimisticGuard(repo), nil
	case StrategyRedis:
		return NewRedisCounterGuard(repo, zones, redis)
	default:
		return NewRowLockGuard(repo), nil
	}
}

// ValidateStrategy ensures env/config combinations are usable.
func ValidateStrategy(strategy, redisURL string) error {
	if strategy == StrategyRedis && redisURL == "" {
		return errors.New("CAPACITY_STRATEGY=redis_counter requires REDIS_URL")
	}
	switch strategy {
	case StrategyRowLock, StrategyOptimistic, StrategyRedis, "":
		return nil
	default:
		return fmt.Errorf("CAPACITY_STRATEGY: unknown value %q", strategy)
	}
}
