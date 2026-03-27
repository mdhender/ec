// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/mdhender/ec/internal/app"
	deliveryhttp "github.com/mdhender/ec/internal/delivery/http"
	"github.com/mdhender/ec/internal/domain"
)

// stubParser is a local test double for app.OrderParser.
type stubParser struct {
	orders      []domain.Order
	diagnostics []app.ParseDiagnostic
	err         error
}

func (s *stubParser) Parse(_ string) ([]domain.Order, []app.ParseDiagnostic, error) {
	return s.orders, s.diagnostics, s.err
}

// stubOrder satisfies domain.Order for test purposes.
type stubOrder struct{}

func (stubOrder) Kind() domain.OrderKind { return domain.OrderKindMove }
func (stubOrder) TurnPhase() domain.Phase { return 1 }
func (stubOrder) Validate() error         { return nil }

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

func TestPostParseOrders_OK(t *testing.T) {
	parser := &stubParser{
		orders:      []domain.Order{stubOrder{}, stubOrder{}},
		diagnostics: []app.ParseDiagnostic{},
	}
	svc := &app.ParseOrdersService{Parser: parser}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/1/orders/parse", strings.NewReader("move order text"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := deliveryhttp.PostParseOrders(svc, 1024)
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
		t.Errorf("expected ok=true, got %v", body)
	}
	if count, exists := body["accepted_count"]; !exists || count != float64(2) {
		t.Errorf("expected accepted_count=2, got %v", body)
	}
	diags, exists := body["diagnostics"]
	if !exists {
		t.Errorf("expected diagnostics field, got %v", body)
	}
	if diagSlice, ok := diags.([]any); !ok || len(diagSlice) != 0 {
		t.Errorf("expected empty diagnostics array, got %v", diags)
	}
}

func TestPostParseOrders_PartialSuccess(t *testing.T) {
	parser := &stubParser{
		orders: []domain.Order{stubOrder{}},
		diagnostics: []app.ParseDiagnostic{
			{Line: 2, Code: "E001", Message: "unknown command"},
		},
	}
	svc := &app.ParseOrdersService{Parser: parser}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/1/orders/parse", strings.NewReader("some order text"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := deliveryhttp.PostParseOrders(svc, 1024)
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
	if ok, exists := body["ok"]; !exists || ok != false {
		t.Errorf("expected ok=false, got %v", body)
	}
	if count, exists := body["accepted_count"]; !exists || count != float64(1) {
		t.Errorf("expected accepted_count=1, got %v", body)
	}
	diags, exists := body["diagnostics"]
	if !exists {
		t.Errorf("expected diagnostics field, got %v", body)
	}
	if diagSlice, ok := diags.([]any); !ok || len(diagSlice) != 1 {
		t.Errorf("expected 1 diagnostic, got %v", diags)
	}
}

func TestPostParseOrders_TooLarge(t *testing.T) {
	svc := &app.ParseOrdersService{Parser: &stubParser{}}
	const maxBytes int64 = 16
	body := strings.Repeat("x", int(maxBytes)+1)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/1/orders/parse", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := deliveryhttp.PostParseOrders(svc, maxBytes)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", rec.Code)
	}
}

func TestPostParseOrders_InternalError(t *testing.T) {
	parser := &stubParser{err: errors.New("parser exploded")}
	svc := &app.ParseOrdersService{Parser: parser}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/1/orders/parse", strings.NewReader("any text"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := deliveryhttp.PostParseOrders(svc, 1024)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
