package service

import (
	"context"
	"estimation/domain"
	"testing"
)

func BenchmarkCalculateSurfaceAreaLargeInput(b *testing.B) {
	voids := makeBenchmarkVoids(100_000)
	wall := domain.WallDimensions{
		LengthM: 10_000,
		HeightM: 1_000,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := CalculateSurfaceArea(wall, voids); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEstimateLargeInput(b *testing.B) {
	voids := makeBenchmarkVoids(100_000)
	calculator := NewCalculator()
	req := domain.CalculationRequest{
		Material: &domain.Material{
			Type:                   "brick",
			DensityKgPerM3:         1800,
			CoverageRateM2PerRonne: 2.5,
		},
		Wall: domain.WallDimensions{
			LengthM:    10_000,
			HeightM:    1_000,
			ThicknessM: 0.2,
		},
		Voids:                voids,
		Pattern:              "herringbone",
		ComplexityMultiplier: 1,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := calculator.Estimate(context.Background(), req); err != nil {
			b.Fatal(err)
		}
	}
}

func makeBenchmarkVoids(count int) []domain.Void {
	voids := make([]domain.Void, count)
	for i := range voids {
		voids[i] = domain.Void{
			WidthM:  0.5,
			HeightM: 0.5,
		}
	}
	return voids
}
