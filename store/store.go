package store

import (
	"context"
	"encoding/json"
	"errors"
	"estimation/domain"
	"fmt"
	"os"
	"strings"
)

type MaterialCatalog struct {
	byType    map[string]*domain.Material
	materials []domain.Material
}

func LoadMaterialCatalog(path string) (*MaterialCatalog, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open material catalog: %w", err)
	}
	defer file.Close()

	var materials []domain.Material
	if err := json.NewDecoder(file).Decode(&materials); err != nil {
		return nil, fmt.Errorf("decode material catalog: %w", err)
	}

	return NewMaterialCatalog(materials)
}

func NewMaterialCatalog(materials []domain.Material) (*MaterialCatalog, error) {
	if len(materials) == 0 {
		return nil, errors.New("material catalog must contain at least one material")
	}

	catalog := &MaterialCatalog{
		byType:    make(map[string]*domain.Material, len(materials)),
		materials: make([]domain.Material, 0, len(materials)),
	}

	for i := range materials {
		material := materials[i]
		if err := validateMaterial(material, i); err != nil {
			return nil, err
		}

		key := normalizeLookupKey(material.Type)
		if _, exists := catalog.byType[key]; exists {
			return nil, fmt.Errorf("material catalog contains duplicate type %q", material.Type)
		}

		copied := material
		catalog.byType[key] = &copied
		catalog.materials = append(catalog.materials, copied)
	}
	return catalog, nil
}

func (c *MaterialCatalog) GetByType(ctx context.Context, materialType string) (*domain.Material, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("material catalog is nil")
	}

	key := normalizeLookupKey(materialType)
	if key == "" {
		return nil, errors.New("material type is required")
	}

	material, ok := c.byType[key]
	if !ok {
		return nil, fmt.Errorf("%w: material type %q", domain.ErrMaterialNotFound, materialType)
	}

	copied := *material
	return &copied, nil
}

func (c *MaterialCatalog) List(ctx context.Context) ([]domain.Material, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("material catalog is nil")
	}

	materials := make([]domain.Material, len(c.materials))
	copy(materials, c.materials)
	return materials, nil
}

func validateMaterial(material domain.Material, index int) error {
	if strings.TrimSpace(material.Type) == "" {
		return fmt.Errorf("material[%d].type is required", index)
	}
	if material.DensityKgPerM3 <= 0 {
		return fmt.Errorf("material[%d].density must be > 0, got %f", index, material.DensityKgPerM3)
	}
	if material.CostPerTon < 0 {
		return fmt.Errorf("material[%d].cost_per_ton must be >= 0, got %f", index, material.CostPerTon)
	}
	if material.CoverageRateM2PerRonne <= 0 {
		return fmt.Errorf("material[%d].coverage_rate must be > 0, got %f", index, material.CoverageRateM2PerRonne)
	}
	return nil
}

func normalizeLookupKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
