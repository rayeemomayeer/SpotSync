//go:build integration

package contract_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/test/testutil"
)

func TestContract_gradedAPI(t *testing.T) {
	db, connStr := testutil.SetupPostgres(t)
	server := testutil.NewTestServer(t, db, connStr)
	defer server.Close()

	base := server.URL + "/api/v1"
	client := server.Client()

	adminEmail := testutil.UniqueEmail("admin")
	driverEmail := testutil.UniqueEmail("driver")

	// POST /auth/register (admin)
	regAdmin := map[string]any{
		"name": "Admin User", "email": adminEmail, "password": "password123", "role": models.RoleAdmin,
	}
	status, body := doJSON(t, client, http.MethodPost, base+"/auth/register", regAdmin, "")
	assertStatus(t, status, http.StatusCreated)
	assertSuccessEnvelope(t, body)
	assertUserShape(t, body["data"])

	// POST /auth/register (driver)
	regDriver := map[string]any{
		"name": "Driver User", "email": driverEmail, "password": "password123", "role": models.RoleDriver,
	}
	status, body = doJSON(t, client, http.MethodPost, base+"/auth/register", regDriver, "")
	assertStatus(t, status, http.StatusCreated)
	assertSuccessEnvelope(t, body)

	// POST /auth/login
	status, body = doJSON(t, client, http.MethodPost, base+"/auth/login", map[string]any{
		"email": adminEmail, "password": "password123",
	}, "")
	assertStatus(t, status, http.StatusOK)
	assertSuccessEnvelope(t, body)
	adminToken := extractToken(t, body["data"])
	assertUserShape(t, asMap(t, body["data"])["user"])

	status, body = doJSON(t, client, http.MethodPost, base+"/auth/login", map[string]any{
		"email": driverEmail, "password": "password123",
	}, "")
	assertStatus(t, status, http.StatusOK)
	driverToken := extractToken(t, body["data"])

	// POST /zones
	status, body = doJSON(t, client, http.MethodPost, base+"/zones", map[string]any{
		"name": "EV Lot", "type": models.ZoneTypeEVCharging, "total_capacity": 1, "price_per_hour": 5.5,
	}, adminToken)
	assertStatus(t, status, http.StatusCreated)
	assertSuccessEnvelope(t, body)
	zoneID := extractUint(t, body["data"], "id")
	assertZoneShape(t, body["data"])

	// GET /zones
	status, body = doJSON(t, client, http.MethodGet, base+"/zones", nil, "")
	assertStatus(t, status, http.StatusOK)
	assertSuccessEnvelope(t, body)
	zones := asSlice(t, body["data"])
	if len(zones) != 1 {
		t.Fatalf("zones count = %d, want 1", len(zones))
	}
	assertZoneShape(t, zones[0])

	// GET /zones/:id
	status, body = doJSON(t, client, http.MethodGet, fmt.Sprintf("%s/zones/%d", base, zoneID), nil, "")
	assertStatus(t, status, http.StatusOK)
	assertZoneShape(t, body["data"])

	// POST /reservations (success)
	status, body = doJSON(t, client, http.MethodPost, base+"/reservations", map[string]any{
		"zone_id": zoneID, "license_plate": "ABC-1234",
	}, driverToken)
	assertStatus(t, status, http.StatusCreated)
	assertSuccessEnvelope(t, body)
	resID := extractUint(t, body["data"], "id")
	assertReservationShape(t, body["data"])

	// POST /reservations (zone full -> 409)
	status, body = doJSON(t, client, http.MethodPost, base+"/reservations", map[string]any{
		"zone_id": zoneID, "license_plate": "XYZ-9999",
	}, driverToken)
	assertStatus(t, status, http.StatusConflict)
	assertErrorEnvelope(t, body)

	// GET /reservations/my-reservations
	status, body = doJSON(t, client, http.MethodGet, base+"/reservations/my-reservations", nil, driverToken)
	assertStatus(t, status, http.StatusOK)
	mine := asSlice(t, body["data"])
	if len(mine) != 1 {
		t.Fatalf("my reservations = %d, want 1", len(mine))
	}

	// DELETE /reservations/:id forbidden for admin (not owner)
	status, body = doJSON(t, client, http.MethodDelete, fmt.Sprintf("%s/reservations/%d", base, resID), nil, adminToken)
	assertStatus(t, status, http.StatusForbidden)
	assertErrorEnvelope(t, body)

	// DELETE /reservations/:id success for owner
	status, body = doJSON(t, client, http.MethodDelete, fmt.Sprintf("%s/reservations/%d", base, resID), nil, driverToken)
	assertStatus(t, status, http.StatusOK)
	assertSuccessEnvelope(t, body)

	// GET /reservations (admin)
	status, body = doJSON(t, client, http.MethodGet, base+"/reservations", nil, adminToken)
	assertStatus(t, status, http.StatusOK)
	all := asSlice(t, body["data"])
	if len(all) != 1 {
		t.Fatalf("admin reservations = %d, want 1", len(all))
	}

	// Unauthorized without token
	status, body = doJSON(t, client, http.MethodPost, base+"/reservations", map[string]any{
		"zone_id": zoneID, "license_plate": "NO-AUTH",
	}, "")
	assertStatus(t, status, http.StatusUnauthorized)
	assertErrorEnvelope(t, body)
}

func doJSON(t *testing.T, client *http.Client, method, url string, payload any, token string) (int, map[string]any) {
	t.Helper()

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("decode response: %v body=%s", err, string(raw))
		}
	}

	return resp.StatusCode, decoded
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

func assertSuccessEnvelope(t *testing.T, body map[string]any) {
	t.Helper()
	if body["success"] != true {
		t.Fatalf("success = %v, want true", body["success"])
	}
	if _, ok := body["message"].(string); !ok {
		t.Fatalf("message missing or not string: %v", body["message"])
	}
	if _, ok := body["data"]; !ok {
		t.Fatal("data key should be present on success")
	}
}

func assertErrorEnvelope(t *testing.T, body map[string]any) {
	t.Helper()
	if body["success"] != false {
		t.Fatalf("success = %v, want false", body["success"])
	}
	if _, ok := body["message"].(string); !ok {
		t.Fatalf("message missing or not string: %v", body["message"])
	}
	if body["errors"] == nil {
		t.Fatal("errors should be present on error")
	}
}

func assertUserShape(t *testing.T, data any) {
	t.Helper()
	m := asMap(t, data)
	for _, key := range []string{"id", "name", "email", "role", "created_at", "updated_at"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("user missing field %q", key)
		}
	}
	if _, ok := m["password"]; ok {
		t.Fatal("password must not appear in user response")
	}
}

func assertZoneShape(t *testing.T, data any) {
	t.Helper()
	m := asMap(t, data)
	for _, key := range []string{"id", "name", "type", "total_capacity", "price_per_hour", "available_spots", "created_at", "updated_at"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("zone missing field %q", key)
		}
	}
}

func assertReservationShape(t *testing.T, data any) {
	t.Helper()
	m := asMap(t, data)
	for _, key := range []string{"id", "user_id", "zone_id", "license_plate", "status", "created_at", "updated_at"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("reservation missing field %q", key)
		}
	}
}

func extractToken(t *testing.T, data any) string {
	t.Helper()
	m := asMap(t, data)
	token, ok := m["token"].(string)
	if !ok || token == "" {
		t.Fatalf("token missing in login data: %v", data)
	}
	return token
}

func extractUint(t *testing.T, data any, key string) uint {
	t.Helper()
	m := asMap(t, data)
	switch v := m[key].(type) {
	case float64:
		return uint(v)
	default:
		t.Fatalf("field %q = %T, want number", key, m[key])
		return 0
	}
}

func asMap(t *testing.T, data any) map[string]any {
	t.Helper()
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("data = %T, want map", data)
	}
	return m
}

func asSlice(t *testing.T, data any) []any {
	t.Helper()
	s, ok := data.([]any)
	if !ok {
		t.Fatalf("data = %T, want slice", data)
	}
	return s
}
