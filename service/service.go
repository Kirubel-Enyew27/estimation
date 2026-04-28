package service

import (
	"context"

	"estimation/domain"
)

// EstimationService performs estimation use-cases.
// Keep this focused: a single method that consumes a domain request and
// produces a domain result. Concrete implementations live in the service
// package (not the interface file) to keep boundaries clear.
type EstimationService interface {
    Estimate(ctx context.Context, req domain.CalculationRequest) (domain.CalculationResult, error)
}
