package platform_test

import (
	"testing"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
)

func TestTokenManager_IssueAndVerify(t *testing.T) {
	tm := platform.NewTokenManager("secret", time.Hour)

	token, err := tm.Issue(42, models.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}

	id, role, err := tm.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
	if role != models.RoleAdmin {
		t.Errorf("role = %q", role)
	}
}

func TestTokenManager_VerifyInvalid(t *testing.T) {
	tm := platform.NewTokenManager("secret", time.Hour)

	_, _, err := tm.Verify("not-a-token")
	if err != domain.ErrUnauthorized {
		t.Fatalf("error = %v, want ErrUnauthorized", err)
	}
}
