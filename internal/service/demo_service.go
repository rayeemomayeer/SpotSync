package service

import (
	"context"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/repository"
)

type DemoService struct {
	demo *repository.DemoRepository
}

func NewDemoService(demo *repository.DemoRepository) *DemoService {
	return &DemoService{demo: demo}
}

func (s *DemoService) ResetSession(ctx context.Context, sessionID string) (repository.DemoResetStats, error) {
	return s.demo.ResetSession(ctx, sessionID)
}

func (s *DemoService) PurgeStaleSessions(ctx context.Context, inactiveBefore time.Time) (int, repository.DemoResetStats, error) {
	return s.demo.PurgeStaleSessions(ctx, inactiveBefore)
}
