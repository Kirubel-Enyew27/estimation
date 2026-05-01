package store

import (
	"context"
	"errors"
	"estimation/service"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMaterialCatalogLooksUpByMaterialType(t *testing.T) {
	path := writeCatalog(t, `[
		{"type":"Fieldstone","density":2400,"cost_per_ton":95,"coverage_rate":1.5}
	]`)

	catalog, err := LoadMaterialCatalog(path)
	if err != nil {
		t.Fatalf("LoadMaterialCatalog returned error: %v", err)
	}

	material, err := catalog.GetByType(context.Background(), " fieldstone ")
	if err != nil {
		t.Fatalf("GetByType returned error: %v", err)
	}

	if material.Type != "fieldstone" {
		t.Fatalf("got type %q, want fieldstone", material.Type)
	}
	if material.DensityKgPerM3 != 2400 {
		t.Fatalf("got cost per ton %f, want 2400", material.DensityKgPerM3)
	}
	if material.CostPerTon != 95 {
		t.Fatalf("got cost per ton %f, want 95", material.CostPerTon)
	}
	if material.CoverageRateM2PerRonne != 1.5 {
		t.Fatalf("got coverage rate %f, want 1.5", material.CoverageRateM2PerRonne)
	}
}

func TestMaterialCatalogReturnsCopies(t *testing.T) {
	catalog, err := NewMaterialCatalogFromJSON(t, `[
		{"type":"brick","density":1800,"cost_per_ton":65,"coverage_rate":2.2}
	]`)
	if err != nil {
		t.Fatalf("NewMaterialCatalog returned error: %v", err)
	}

	material, err := catalog.GetByType(context.Background(), "brick")
	if err != nil {
		t.Fatalf("GetByType returned error: %v", err)
	}
	material.DensityKgPerM3 = 1

	material, err = catalog.GetByType(context.Background(), "brick")
	if err != nil {
		t.Fatalf("GetByType returned error: %v", err)
	}
	if material.DensityKgPerM3 != 1800 {
		t.Fatalf("catalog material was mutated, got density %f", material.DensityKgPerM3)
	}
}

func TestLoadMaterialCatalogRejectsInvalidMaterial(t *testing.T) {
	path := writeCatalog(t, `[
		{"type":"brick","density":0,"cost_per_ton":65,"coverage_rate":2.2}
	]`)

	_, err := LoadMaterialCatalog(path)
	if err == nil {
		t.Fatal("expected invalid material error")
	}
}

func TestMaterialCatalogReturnsNotFoundForUnknownType(t *testing.T) {
	catalog, err := NewMaterialCatalogFromJSON(t, `[
		{"type":"brick","density":1800,"cost_per_ton":65,"coverage_rate":2.2}
	]`)
	if err != nil {
		t.Fatalf("NewMaterialCatalog returned error: %v", err)
	}

	_, err = catalog.GetByType(context.Background(), "granite")
	if !errors.Is(err, service.ErrMaterialNotFound) {
		t.Fatalf("got err %v, want ErrMaterialNotFound", err)
	}
}

func NewMaterialCatalogFromJSON(t *testing.T, content string) (*MaterialCatalog, error) {
	t.Helper()
	return LoadMaterialCatalog(writeCatalog(t, content))
}

func writeCatalog(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "materials.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write catalog: %v", err)
	}
	return path
}
