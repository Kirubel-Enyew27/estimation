package service

import (
	"context"
	"estimation/domain"
	"math"
	"testing"
)

func TestCalculatesurfaceAreaSubtractsVoid(t *testing.T) {
	got, err := CalculateSurfaceArea(domain.WallDimenstions{
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
	_, err := CalculateSurfaceArea(domain.WallDimenstions{
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

	got := calculator.ApplyMultipliers(domain.CalcualtionRequest{
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
		DefaultWastePercent: 0.05,
		DefaultComplexityMultiplier: 1.05,
	})

	got := calculator.ApplyMultipliers(domain.CalcualtionRequest{
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

	got := calculator.ApplyMultipliers(domain.CalcualtionRequest{
		Material: &domain.Material{
			Type: "Fieldstone",
		},
		Pattern: "Herringbone",
		WastePercent: 0.08,
		ComplexityMultiplier: 1.05,
	})

	assertFloat(t, got.WastePercent, 0.08)
	assertFloat(t, got.ComplexityMultiplier, 1.05)
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

func TestCalculatorEstimateAppliesConfiguredMultipliersToFinalMaterialsEstimate(t *testing.T) {
	Calcualator := NewCalculator()

	got, err := Calcualator.Estimate(context.Background(), domain.CalcualtionRequest{
		Material: &domain.Material{
			Type: "Fieldstone",
			DensityKgPerM3: 2000,
		},
		Wall: domain.WallDimenstions{
			LengthM: 5,
			HeightM: 2,
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
	assertFloat(t, got.BreakDown["wastePercent"], 0.25)
}

func TestCalculatorEstimate(t *testing.T) {
	calculator := NewCalculator()

	got, err := calculator.Estimate(context.Background(), domain.CalcualtionRequest{
		Material: &domain.Material{
			DensityKgPerM3:         2500,
			CoverageRateM2PerRonne: 2,
		},
		Wall: domain.WallDimenstions{
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

func TestCalculatorEstimateAppliesConfiguredMultipliersToFinalMaterialEstimate(t *testing.T) {
	calculator := NewCalculator()

	got, err := calculator.Estimate(context.Background(), domain.CalcualtionRequest{
		Material: &domain.Material{
			Type:           "Fieldstone",
			DensityKgPerM3: 2000,
		},
		Wall: domain.WallDimenstions{
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
	assertFloat(t, got.BreakDown["wastePercent"], 0.25)
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("got %f, want %f", got, want)
	}
}
