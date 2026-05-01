package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"estimation/domain"
	"estimation/service"
	"estimation/store"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListMaterialsReturnsCatalog(t *testing.T) {
	api := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/materials", nil)
	rec := httptest.NewRecorder()

	api.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200. body: %s", rec.Code, rec.Body.String())
	}

	var got []domain.Material
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d materials, want 1", len(got))
	}
	if got[0].Type != "brick" {
		t.Fatalf("got material type %q, want brick", got[0].Type)
	}
}

func TestCalculateProjectReturnsEstimate(t *testing.T) {
	api := newTestHandler(t)
	body := `{
		"materialCode":"brick",
		"wall":{"lengthM":5,"heightM":2,"thicknessM":0.2},
		"complexityMultiplier":1
	}`
	req := httptest.NewRequest(http.MethodPost, "/calculate/project", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	api.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200. body: %s", rec.Code, rec.Body.String())
	}

	var got domain.CalculationResult
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assertFloat(t, got.SurfaceAreaM2, 10)
	assertFloat(t, got.VolumeM3, 2)
	assertFloat(t, got.StoneTonnage, 4.4)
}

func TestCalculateProjectRejectsInvalidJSON(t *testing.T) {
	api := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/calculate/project", bytes.NewBufferString(`{"materialCode":`))
	rec := httptest.NewRecorder()

	api.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestCalculateProjectRejectsInvalidInput(t *testing.T) {
	api := newTestHandler(t)
	body := `{
		"materialCode":"brick",
		"wall":{"lengthM":0,"heightM":2,"thicknessM":0.2}
	}`
	req := httptest.NewRequest(http.MethodPost, "/calculate/project", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	api.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestCalculateProjectReturnsBadRequestForUnknownMaterial(t *testing.T) {
	api := newTestHandler(t)
	body := `{
		"materialCode":"granite",
		"wall":{"lengthM":5,"heightM":2,"thicknessM":0.2}
	}`
	req := httptest.NewRequest(http.MethodPost, "/calculate/project", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	api.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
}

func TestCalculateProjectReturnsInternalServerErrorForServiceFailure(t *testing.T) {
	catalog := newTestCatalog(t)
	api := New(failingService{}, catalog)
	body := `{
		"materialCode":"brick",
		"wall":{"lengthM":5,"heightM":2,"thicknessM":0.2}
	}`
	req := httptest.NewRequest(http.MethodPost, "/calculate/project", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	api.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want 500", rec.Code)
	}
}

type failingService struct{}

func (failingService) Estimate(context.Context, domain.CalcualtionRequest) (domain.CalculationResult, error) {
	return domain.CalculationResult{}, errors.New("database unavailable")
}

func newTestHandler(t *testing.T) *Handler {
	t.Helper()

	catalog := newTestCatalog(t)
	calculator := service.NewCalculatorWithMaterialStore(catalog)
	return New(calculator, catalog)
}

func newTestCatalog(t *testing.T) *store.MaterialCatalog {
	t.Helper()

	catalog, err := store.NewMaterialCatalog([]domain.Material{
		{
			Type:                   "brick",
			DensityKgPerM3:         1800,
			CostPerTon:             65,
			CoverageRateM2PerRonne: 2.5,
		},
	})
	if err != nil {
		t.Fatalf("NewMaterialCatalog returned error: %v", err)
	}
	return catalog
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
