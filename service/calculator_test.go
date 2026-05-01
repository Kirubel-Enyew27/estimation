package service

import (
	"context"
	"errors"
	"estimation/domain"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestCalculatesurfaceAreaSubtractsVoid(t *testing.T) {
	got, err := CalculateSurfaceArea(domain.WallDimensions{
		LengthM: 10,
		HeightM: 3,
	}, []domain.Void{
		{WidthM: 1.2, HeightM: 1.5},
		{WidthM: 0.9, HeightM: 2.1},
	})
	if err != nil {
		t.Fatalf("CalculateSurfaceArea returned error: %v", err)
	}

	assertFloat(t, got.TotalWallAreaM2, 30)
	assertFloat(t, got.VoidAreaM2, 3.69)
	assertFloat(t, got.NetAreaM2, 26.31)
}

func TestCalculateSurfaceAreaRejectsTotalVoidsOverWallArea(t *testing.T) {
	_, err := CalculateSurfaceArea(domain.WallDimensions{
		LengthM: 2,
		HeightM: 2,
	}, []domain.Void{
		{WidthM: 2, HeightM: 2},
		{WidthM: 1, HeightM: 1},
	})
	if err == nil {
		t.Fatal("expected error when total void area exceeds wall area")
	}
}

func TestCalculateSurfaceAreaRejectsZeroDimensions(t *testing.T) {
	cases := []struct {
		name string
		wall domain.WallDimensions
	}{
		{
			name: "zero length",
			wall: domain.WallDimensions{LengthM: 0, HeightM: 3},
		},
		{
			name: "zero height",
			wall: domain.WallDimensions{LengthM: 3, HeightM: 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := CalculateSurfaceArea(tc.wall, nil)
			if err == nil {
				t.Fatal("expected error for zero wall dimension")
			}
		})
	}
}

func TestCalculateSurfaceAreaRejectsInvalidVoidSizes(t *testing.T) {
	cases := []struct {
		name  string
		voids []domain.Void
	}{
		{
			name:  "negative width",
			voids: []domain.Void{{WidthM: -1, HeightM: 1}},
		},
		{
			name:  "negative height",
			voids: []domain.Void{{WidthM: 1, HeightM: -1}},
		},
		{
			name: "combined void area exceeds wall",
			voids: []domain.Void{
				{WidthM: 2, HeightM: 2},
				{WidthM: 1, HeightM: 1},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := CalculateSurfaceArea(domain.WallDimensions{
				LengthM: 2,
				HeightM: 2,
			}, tc.voids)
			if err == nil {
				t.Fatal("expected error for invalid void sizes")
			}
		})
	}
}

func TestCalculateStoneTonnageUsesDensityCoverageWasteAndComplexity(t *testing.T) {
	got, err := CalculateStoneTonnage(20, 0.2, domain.Material{
		DensityKgPerM3:         2400,
		CoverageRateM2PerRonne: 1.5,
	}, 0.10, 1.2)
	if err != nil {
		t.Fatalf("CalculateStoneTonnage returned error: %v", err)
	}

	assertFloat(t, got.VolumeM3, 4)
	assertFloat(t, got.WasteStoneKg, 1333.3333333333335)
	assertFloat(t, got.StoneMassKg, 17600)
	assertFloat(t, got.StoneTonnage, 17.6)
}

func TestApplyMultipliersUsesMaterialTypeAndPatternDefaults(t *testing.T) {
	calculator := NewCalculator()

	got := calculator.ApplyMultipliers(domain.CalculationRequest{
		Material: &domain.Material{
			Type: "Brick",
		},
		Pattern: "Herringbone",
	})

	assertFloat(t, got.WastePercent, 0.10)
	assertFloat(t, got.ComplexityMultiplier, 1.20)
}

func TestApplyMultipliersCanBeConfigured(t *testing.T) {
	calculator := NewCalculatorWithConfig(MultiplierConfig{
		MaterialWastePercent: map[string]float64{
			"limestone": 0.18,
		},
		PatternComplexityMultipliers: map[string]float64{
			"coursed random": 1.35,
		},
		DefaultWastePercent:         0.05,
		DefaultComplexityMultiplier: 1.05,
	})

	got := calculator.ApplyMultipliers(domain.CalculationRequest{
		Material: &domain.Material{
			Type: "Limestone",
		},
		Pattern: "Coursed Random",
	})

	assertFloat(t, got.WastePercent, 0.18)
	assertFloat(t, got.ComplexityMultiplier, 1.35)
}

func TestApplyMultipliersLetsExplicitRequestValuesOverrideConfig(t *testing.T) {
	calculator := NewCalculator()

	got := calculator.ApplyMultipliers(domain.CalculationRequest{
		Material: &domain.Material{
			Type: "Fieldstone",
		},
		Pattern:              "Herringbone",
		WastePercent:         0.08,
		ComplexityMultiplier: 1.05,
	})

	assertFloat(t, got.WastePercent, 0.08)
	assertFloat(t, got.ComplexityMultiplier, 1.05)
}

func TestApplyMultipliersUsesDefaultsForUnknownMaterialAndPattern(t *testing.T) {
	calculator := NewCalculatorWithConfig(MultiplierConfig{
		MaterialWastePercent: map[string]float64{
			"brick": 0.10,
		},
		PatternComplexityMultipliers: map[string]float64{
			"herringbone": 1.20,
		},
		DefaultWastePercent:         0.03,
		DefaultComplexityMultiplier: 1.05,
	})

	got := calculator.ApplyMultipliers(domain.CalculationRequest{
		Material: &domain.Material{
			Type: "granite",
		},
		Pattern: "unknown-pattern",
	})

	assertFloat(t, got.WastePercent, 0.03)
	assertFloat(t, got.ComplexityMultiplier, 1.05)
}

func TestApplyMultipliersUsesMaterialCodeWhenInlineMaterialDoesNotMatch(t *testing.T) {
	calculator := NewCalculatorWithConfig(MultiplierConfig{
		MaterialWastePercent: map[string]float64{
			"brick": 0.10,
		},
		DefaultComplexityMultiplier: 1,
	})

	got := calculator.ApplyMultipliers(domain.CalculationRequest{
		MaterialCode: "brick",
		Material: &domain.Material{
			Type: "unknown",
		},
	})

	assertFloat(t, got.WastePercent, 0.10)
	assertFloat(t, got.ComplexityMultiplier, 1)
}

func TestCalculateStoneTonnageFallsBackToDensityWhenCoverageIsUnset(t *testing.T) {
	got, err := CalculateStoneTonnage(12, 0.25, domain.Material{
		DensityKgPerM3: 2200,
	}, 0, 1)
	if err != nil {
		t.Fatalf("CalculateStoneTonnage returned error: %v", err)
	}

	assertFloat(t, got.VolumeM3, 3)
	assertFloat(t, got.StoneMassKg, 6600)
	assertFloat(t, got.StoneTonnage, 6.6)
}

func TestCalculateMortarUsesJointWidthAndDepth(t *testing.T) {
	volume, mass, err := CalculateMortar(25, 0.012, 0.02)
	if err != nil {
		t.Fatalf("CalculateMortar returned error: %v", err)
	}

	assertFloat(t, volume, 0.006)
	assertFloat(t, mass, 12.96)
}

func TestCalculateMortarHandlesZeroJointDimensions(t *testing.T) {
	volume, mass, err := CalculateMortar(25, 0, 0.02)
	if err != nil {
		t.Fatalf("CalculateMortar returned error: %v", err)
	}

	assertFloat(t, volume, 0)
	assertFloat(t, mass, 0)
}

func TestCalculateMortarRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		name       string
		netAreaM2  float64
		jointWidth float64
		jointDepth float64
	}{
		{name: "negative area", netAreaM2: -1, jointWidth: 0.01, jointDepth: 0.02},
		{name: "negative width", netAreaM2: 10, jointWidth: -0.01, jointDepth: 0.02},
		{name: "negative depth", netAreaM2: 10, jointWidth: 0.01, jointDepth: -0.02},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := CalculateMortar(tc.netAreaM2, tc.jointWidth, tc.jointDepth)
			if err == nil {
				t.Fatal("expected error for invalid mortar inputs")
			}
		})
	}
}

func TestCalculatorEstimateAppliesConfiguredMultipliersToFinalMaterialsEstimate(t *testing.T) {
	calculator := NewCalculator()

	got, err := calculator.Estimate(context.Background(), domain.CalculationRequest{
		Material: &domain.Material{
			Type:           "Fieldstone",
			DensityKgPerM3: 2000,
		},
		Wall: domain.WallDimensions{
			LengthM:    5,
			HeightM:    2,
			ThicknessM: 0.2,
		},
		Pattern: "Herringbone",
	})
	if err != nil {
		t.Fatalf("Estimate returned error: %v", err)
	}

	assertFloat(t, got.SurfaceAreaM2, 10)
	assertFloat(t, got.VolumeM3, 2)
	assertFloat(t, got.WasteStoneKg, 1000)
	assertFloat(t, got.StoneMassKg, 6000)
	assertFloat(t, got.StoneTonnage, 6)
	assertFloat(t, got.AppliedComplexityMultiplier, 1.2)
	assertFloat(t, got.Breakdown["wastePercent"], 0.25)
}

func TestCalculatorEstimate(t *testing.T) {
	calculator := NewCalculator()

	got, err := calculator.Estimate(context.Background(), domain.CalculationRequest{
		Material: &domain.Material{
			DensityKgPerM3:         2500,
			CoverageRateM2PerRonne: 2,
		},
		Wall: domain.WallDimensions{
			LengthM:    6,
			HeightM:    2.5,
			ThicknessM: 0.2,
		},
		Voids: []domain.Void{
			{WidthM: 1, HeightM: 2},
		},
		JointWidthM:          0.01,
		JointDepthM:          0.03,
		ComplexityMultiplier: 1,
		IncludeMortar:        true,
	})
	if err != nil {
		t.Fatalf("Estimate returned error: %v", err)
	}

	assertFloat(t, got.SurfaceAreaM2, 13)
	assertFloat(t, got.VolumeM3, 2.6)
	assertFloat(t, got.StoneMassKg, 6500)
	assertFloat(t, got.StoneTonnage, 6.5)
	assertFloat(t, got.MortarVolumeM3, 0.0039)
	assertFloat(t, got.MortarMassKg, 8.424)

}

func TestCalculatorEstimateLookUpMaterialByType(t *testing.T) {
	catalog := fakeMaterialRepository{
		"limestone": {
			Type:                   "limestone",
			DensityKgPerM3:         2300,
			CostPerTon:             110,
			CoverageRateM2PerRonne: 1.7,
		},
	}

	calculator := NewCalculatorWithMaterialRepository(catalog)
	got, err := calculator.Estimate(context.Background(), domain.CalculationRequest{
		MaterialCode: "Limestone",
		Wall: domain.WallDimensions{
			LengthM:    6,
			HeightM:    2,
			ThicknessM: 0.2,
		},
		ComplexityMultiplier: 1,
	})
	if err != nil {
		t.Fatalf("Estimate returned error: %v", err)
	}

	assertFloat(t, got.SurfaceAreaM2, 12)
	assertFloat(t, got.VolumeM3, 2.4)
	assertFloat(t, got.StoneMassKg, 7058.823529411765)
	assertFloat(t, got.StoneTonnage, 7.058823529411765)
}

func TestCalculatorEstimateRequiresConfiguredCatalogForMaterialCode(t *testing.T) {
	calculator := NewCalculator()

	_, err := calculator.Estimate(context.Background(), domain.CalculationRequest{
		MaterialCode: "brick",
		Wall: domain.WallDimensions{
			LengthM:    1,
			HeightM:    1,
			ThicknessM: 0.2,
		},
		ComplexityMultiplier: 1,
	})
	if err == nil {
		t.Fatal("expected error when material catalog is not configured")
	}
}

func TestCalculatorEstimateReturnsNotFoundForUnknownMaterial(t *testing.T) {
	catalog := fakeMaterialRepository{
		"brick": {
			Type:                   "brick",
			DensityKgPerM3:         1800,
			CostPerTon:             65,
			CoverageRateM2PerRonne: 2.2,
		},
	}

	calculator := NewCalculatorWithMaterialRepository(catalog)
	_, err := calculator.Estimate(context.Background(), domain.CalculationRequest{
		MaterialCode: "granite",
		Wall: domain.WallDimensions{
			LengthM:    1,
			HeightM:    1,
			ThicknessM: 0.2,
		},
		ComplexityMultiplier: 1,
	})
	if !errors.Is(err, ErrMaterialNotFound) {
		t.Fatalf("got error %v, want ErrMaterialNotFound", err)
	}
	if !strings.Contains(err.Error(), `material type "granite" does not exist`) {
		t.Fatalf("got error %q, want clear unknown material message", err.Error())
	}
}

type fakeMaterialRepository map[string]domain.Material

func (r fakeMaterialRepository) GetByType(ctx context.Context, materialType string) (*domain.Material, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	material, ok := r[strings.ToLower(strings.TrimSpace(materialType))]
	if !ok {
		return nil, fmt.Errorf("%w: material type %q", ErrMaterialNotFound, materialType)
	}

	copied := material
	return &copied, nil
}

func (r fakeMaterialRepository) List(ctx context.Context) ([]domain.Material, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	materials := make([]domain.Material, 0, len(r))
	for _, material := range r {
		materials = append(materials, material)
	}
	return materials, nil
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("got %f, want %f", got, want)
	}
}
