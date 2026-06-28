package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/dto"
)

func TestSuccessEnvelope(t *testing.T) {
	resp := dto.Success("User registered", map[string]int{"id": 1})
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded["success"] != true {
		t.Errorf("success = %v, want true", decoded["success"])
	}
	if decoded["message"] != "User registered" {
		t.Errorf("message = %v", decoded["message"])
	}
	if decoded["data"] == nil {
		t.Error("data should be present")
	}
}

func TestErrorEnvelope(t *testing.T) {
	resp := dto.Error("Validation failed", map[string]string{"email": "Invalid email"})
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded["success"] != false {
		t.Errorf("success = %v, want false", decoded["success"])
	}

	errors, ok := decoded["errors"].(map[string]any)
	if !ok {
		t.Fatalf("errors = %T, want map", decoded["errors"])
	}
	if errors["email"] != "Invalid email" {
		t.Errorf("errors[email] = %v", errors["email"])
	}
}

func TestErrorEnvelopeNilErrorsBecomesEmptyObject(t *testing.T) {
	resp := dto.Error("Server error", nil)
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	if string(raw) == "" {
		t.Fatal("empty marshal")
	}

	var decoded dto.ErrorResponse
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Errors == nil {
		t.Error("errors should be non-nil empty map")
	}
}
