package service

import (
	"errors"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"gorm.io/gorm"
)

func TestMapOrgWriteErr_duplicateSlug(t *testing.T) {
	err := mapOrgWriteErr(gorm.ErrDuplicatedKey)
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want ValidationError, got %T %v", err, err)
	}
	if ve.Fields["slug"] == "" {
		t.Fatalf("fields = %#v", ve.Fields)
	}
}

func TestMapOrgWriteErr_passthrough(t *testing.T) {
	in := errors.New("boom")
	if got := mapOrgWriteErr(in); !errors.Is(got, in) {
		t.Fatalf("got %v", got)
	}
}
