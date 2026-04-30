package service

import (
	"context"
	"estimation/domain"
)

type EstimationService interface {
	Estimate(ctx context.Context, req domain.CalcualtionRequest) (domain.CalculationResult, error)
}
