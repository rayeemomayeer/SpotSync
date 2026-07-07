package capacity

import (
	"context"

	"github.com/rayeemomayeer/SpotSync/internal/models"
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

func NewGuard(strategy string, repo *repository.ReservationRepository) Guard {
	switch strategy {
	case StrategyOptimistic, StrategyRedis:
// Future strategies delegate to row lock until fully wired.
// See deploy/runbook.md — optimistic/redis_counter are intentionally deferred.
		fallthrough
	default:
		return NewRowLockGuard(repo)
	}
}
