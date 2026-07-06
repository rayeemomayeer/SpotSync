//go:build integration

package integration_test

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

func TestFrontendEnablement_zoneUpdateDeleteAndPagination(t *testing.T) {
	db, connStr := testutil.SetupPostgres(t)
	server := testutil.NewTestServer(t, db, connStr)
	defer server.Close()

	base := server.URL + "/api/v1"
	client := server.Client()

	adminEmail := testutil.UniqueEmail("admin")
	driverEmail := testutil.UniqueEmail("driver")

	status, _ := feDoJSON(client, http.MethodPost, base+"/auth/register", map[string]any{
		"name": "Admin", "email": adminEmail, "password": "password123", "role": models.RoleAdmin,
	}, "")
	if status != http.StatusCreated {
		t.Fatalf("register admin status=%d", status)
	}

	status, body := feDoJSON(client, http.MethodPost, base+"/auth/login", map[string]any{
		"email": adminEmail, "password": "password123",
	}, "")
	adminToken := feExtractToken(t, body)

	status, body = feDoJSON(client, http.MethodPost, base+"/zones", map[string]any{
		"name": "Filter Lot", "type": models.ZoneTypeGeneral, "total_capacity": 3, "price_per_hour": 2.5,
	}, adminToken)
	if status != http.StatusCreated {
		t.Fatalf("create zone status=%d", status)
	}
	zoneID := feExtractUint(t, body["data"], "id")

	status, body = feDoJSON(client, http.MethodPost, base+"/auth/register", map[string]any{
		"name": "Driver", "email": driverEmail, "password": "password123", "role": models.RoleDriver,
	}, "")
	if status != http.StatusCreated {
		t.Fatalf("register driver status=%d", status)
	}
	status, body = feDoJSON(client, http.MethodPost, base+"/auth/login", map[string]any{
		"email": driverEmail, "password": "password123",
	}, "")
	driverToken := feExtractToken(t, body)

	status, _ = feDoJSON(client, http.MethodPost, base+"/reservations", map[string]any{
		"zone_id": zoneID, "license_plate": "ABC-1",
	}, driverToken)
	if status != http.StatusCreated {
		t.Fatalf("reserve status=%d", status)
	}

	status, _ = feDoJSON(client, http.MethodPut, fmt.Sprintf("%s/zones/%d", base, zoneID), map[string]any{
		"name": "Filter Lot", "type": models.ZoneTypeGeneral, "total_capacity": 0, "price_per_hour": 2.5,
	}, adminToken)
	if status != http.StatusConflict {
		t.Fatalf("update below active status=%d want 409", status)
	}

	status, _ = feDoJSON(client, http.MethodDelete, fmt.Sprintf("%s/zones/%d", base, zoneID), nil, adminToken)
	if status != http.StatusConflict {
		t.Fatalf("delete with active status=%d want 409", status)
	}

	status, _ = feDoJSON(client, http.MethodGet, base+"/zones?type=general&q=Filter&sort=name&order=asc", nil, "")
	if status != http.StatusOK {
		t.Fatalf("filter zones status=%d", status)
	}

	status, _, headers := feDoJSONWithHeaders(client, http.MethodGet, base+"/reservations?page=1&limit=1", nil, adminToken)
	if status != http.StatusOK {
		t.Fatalf("paginated list status=%d", status)
	}
	if headers.Get("X-Total-Count") == "" {
		t.Fatal("missing X-Total-Count header")
	}
	if headers.Get("X-Page") != "1" || headers.Get("X-Limit") != "1" {
		t.Fatalf("headers page=%q limit=%q", headers.Get("X-Page"), headers.Get("X-Limit"))
	}
}

func feDoJSON(client *http.Client, method, url string, body any, token string) (int, map[string]any) {
	status, parsed, _ := feDoJSONWithHeaders(client, method, url, body, token)
	return status, parsed
}

func feDoJSONWithHeaders(client *http.Client, method, url string, body any, token string) (int, map[string]any, http.Header) {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		panic(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	_ = json.Unmarshal(raw, &parsed)
	return resp.StatusCode, parsed, resp.Header
}

func feExtractToken(t *testing.T, body map[string]any) string {
	t.Helper()
	data := feAsMap(t, body["data"])
	token, ok := data["token"].(string)
	if !ok || token == "" {
		t.Fatalf("token missing: %v", body)
	}
	return token
}

func feExtractUint(t *testing.T, data any, key string) uint {
	t.Helper()
	m := feAsMap(t, data)
	switch v := m[key].(type) {
	case float64:
		return uint(v)
	default:
		t.Fatalf("field %q = %T", key, m[key])
		return 0
	}
}

func feAsMap(t *testing.T, data any) map[string]any {
	t.Helper()
	m, ok := data.(map[string]any)
	if !ok {
		t.Fatalf("data = %T, want map", data)
	}
	return m
}
