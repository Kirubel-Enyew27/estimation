package service

import (
	"context"
	"errors"
	"estimation/domain"
)

type MaterialCatalogService struct {
	materials MaterialRepository
}

func NewMaterialCatalogService(materials MaterialRepository) *MaterialCatalogService {
	return &MaterialCatalogService{materials: materials}
}

func (s *MaterialCatalogService) ListMaterials(ctx context.Context) ([]domain.Material, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.materials == nil {
		return nil, errors.New("material repository is not configured")
	}
	return s.materials.List(ctx)
}
