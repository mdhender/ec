// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	deliveryhttp "github.com/mdhender/ec/internal/delivery/http"
)

func TestGetHealth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := deliveryhttp.GetHealth()
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if ok, exists := body["ok"]; !exists || ok != true {
		t.Errorf("expected body to have ok=true, got %v", body)
	}

	if _, exists := body["time"]; !exists {
		t.Errorf("expected body to have a 'time' field, got %v", body)
	}
}

func TestPostShutdownNoKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/shutdown/anything", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := deliveryhttp.PostShutdown("", nil)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotImplemented {
		t.Errorf("expected status 501, got %d", rec.Code)
	}
}

func TestPostLogout(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/logout", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := deliveryhttp.PostLogout()
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if ok, exists := body["ok"]; !exists || ok != true {
		t.Errorf("expected body to have ok=true, got %v", body)
	}
}
