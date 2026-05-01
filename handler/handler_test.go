package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"estimation/domain"
	"estimation/service"
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
	api := New(
		stubEstimator{err: service.ErrMaterialNotFound},
		stubMaterialService{materials: testMaterials()},
	)
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
	api := New(failingService{}, stubMaterialService{materials: testMaterials()})
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

func (failingService) Estimate(context.Context, domain.CalculationRequest) (domain.CalculationResult, error) {
	return domain.CalculationResult{}, errors.New("database unavailable")
}

func newTestHandler(t *testing.T) *Handler {
	t.Helper()

	return New(
		stubEstimator{
			result: domain.CalculationResult{
				SurfaceAreaM2: 10,
				VolumeM3:      2,
				StoneTonnage:  4.4,
			},
		},
		stubMaterialService{materials: testMaterials()},
	)
}

type stubEstimator struct {
	result domain.CalculationResult
	err    error
}

func (s stubEstimator) Estimate(context.Context, domain.CalculationRequest) (domain.CalculationResult, error) {
	if s.err != nil {
		return domain.CalculationResult{}, s.err
	}
	return s.result, nil
}

type stubMaterialService struct {
	materials []domain.Material
	err       error
}

func (s stubMaterialService) ListMaterials(context.Context) ([]domain.Material, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.materials, nil
}

func testMaterials() []domain.Material {
	return []domain.Material{
		{
			Type:                   "brick",
			DensityKgPerM3:         1800,
			CostPerTon:             65,
			CoverageRateM2PerRonne: 2.5,
		},
	}
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
