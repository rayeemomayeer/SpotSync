package service

import (
	"context"

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
