package service

import (
	"context"
	"estimation/domain"
)

var ErrMaterialNotFound = domain.ErrMaterialNotFound

type EstimationService interface {
	Estimate(ctx context.Context, req domain.CalculationRequest) (domain.CalculationResult, error)
}

type MaterialRepository interface {
	GetByType(ctx context.Context, materialType string) (*domain.Material, error)
	List(ctx context.Context) ([]domain.Material, error)
}

type MaterialService interface {
	ListMaterials(ctx context.Context) ([]domain.Material, error)
}
