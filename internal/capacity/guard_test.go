package capacity_test

import (
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/capacity"
)

func TestValidateStrategy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		strategy  string
		redisURL  string
		wantError bool
	}{
		{name: "row lock default", strategy: capacity.StrategyRowLock, redisURL: ""},
		{name: "optimistic ok", strategy: capacity.StrategyOptimistic, redisURL: ""},
		{name: "redis requires url", strategy: capacity.StrategyRedis, redisURL: "", wantError: true},
		{name: "redis with url", strategy: capacity.StrategyRedis, redisURL: "redis://localhost:6379/0"},
		{name: "unknown", strategy: "magic", redisURL: "", wantError: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := capacity.ValidateStrategy(tc.strategy, tc.redisURL)
			if tc.wantError && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewGuardDefaultsRowLock(t *testing.T) {
	guard, err := capacity.NewGuard("row_lock", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if guard == nil {
		t.Fatal("expected guard")
	}
}

func TestNewGuardRedisWithoutClient(t *testing.T) {
	_, err := capacity.NewGuard(capacity.StrategyRedis, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error when redis missing")
	}
}
