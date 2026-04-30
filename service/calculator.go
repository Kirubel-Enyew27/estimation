package service

import (
	"context"
	"errors"
	"fmt"

	"estimation/domain"
)

const MortarDensityKgPerM3 = 2160.0

type Calculator struct{}

type SurfaceAreaCalculation struct {
	TotalWallAreaM2 float64
	VoidAreaM2      float64
	NetAreaM2       float64
}

type StoneCalculation struct {
	VolumeM3     float64
	StoneMassKg  float64
	StoneTonnage float64
	WasteStoneKg float64
}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Estimate(ctx context.Context, req domain.CalcualtionRequest) (domain.CalculationResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.CalculationResult{}, err
	}
	if err := req.Validate(); err != nil {
		return domain.CalculationResult{}, err
	}
	if req.Material == nil {
		return domain.CalculationResult{}, errors.New("material lookup is not implemented for materialCode-only requests")
	}

	surface, err := CalculateSurfaceArea(req.Wall, req.Voids)
	if err != nil {
		return domain.CalculationResult{}, err
	}

	stone, err := CalculateStoneTonnage(surface.NetAreaM2, req.Wall.ThicknessM, *req.Material, req.WastePercent, req.ComplexityMultiplier)
	if err != nil {
		return domain.CalculationResult{}, err
	}

	result := domain.CalculationResult{
		SurfaceAreaM2:               surface.NetAreaM2,
		VolumeM3:                    stone.VolumeM3,
		StoneMassKg:                 stone.StoneMassKg,
		StoneTonnage:                stone.StoneTonnage,
		WasteStoneKg:                stone.WasteStoneKg,
		AppliedComplexityMultiplier: req.ComplexityMultiplier,
		BreakDown: map[string]float64{
			"totalWallAreaM2": surface.TotalWallAreaM2,
			"voidAreaM2":      surface.VoidAreaM2,
			"netWallAreaM2":   surface.NetAreaM2,
		},
	}

	if req.IncludeMortar {
		mortarVolume, mortarMass, err := CalculateMortar(surface.NetAreaM2, req.JointWidthM, req.JointDepthM)
		if err != nil {
			return domain.CalculationResult{}, err
		}
		result.MortarVolumeM3 = mortarVolume
		result.MortarMassKg = mortarMass
	}

	return result, nil
}

func CalculateSurfaceArea(wall domain.WallDimenstions, voids []domain.Void) (SurfaceAreaCalculation, error) {
	totalWallArea := wall.SurfaceArea()
	if totalWallArea <= 0 {
		return SurfaceAreaCalculation{}, fmt.Errorf("total wall area must be > 0, got %f", totalWallArea)
	}

	var voidArea float64
	for i, void := range voids {
		if void.WidthM < 0 || void.HeightM < 0 {
			return SurfaceAreaCalculation{}, fmt.Errorf("void[%d] dimensions cannot be negative", i)
		}
		voidArea += void.Area()
	}

	if voidArea > totalWallArea {
		return SurfaceAreaCalculation{}, fmt.Errorf("total void area %.2f exceeds wall area %.2f", voidArea, totalWallArea)
	}

	return SurfaceAreaCalculation{
		TotalWallAreaM2: totalWallArea,
		VoidAreaM2:      voidArea,
		NetAreaM2:       totalWallArea - voidArea,
	}, nil
}

func CalculateStoneTonnage(netAreaM2, wallThicknessM float64, material domain.Material, wastePercent, complexityMultiplier float64) (StoneCalculation, error) {
	if netAreaM2 < 0 {
		return StoneCalculation{}, fmt.Errorf("net area must be >= 0, got %f", netAreaM2)
	}
	if wallThicknessM < 0 {
		return StoneCalculation{}, fmt.Errorf("wall thickness must be >= 0, got %f", wallThicknessM)
	}
	if material.DensityKgPerM3 <= 0 {
		return StoneCalculation{}, fmt.Errorf("material density must be > 0, got %f", material.DensityKgPerM3)
	}
	if material.CoverageRateM2PerTonne < 0 {
		return StoneCalculation{}, fmt.Errorf("coverage rate cannot be negative")
	}
	if wastePercent < 0 {
		return StoneCalculation{}, fmt.Errorf("waste percent cannot be negative")
	}
	if complexityMultiplier < 1 {
		return StoneCalculation{}, fmt.Errorf("complexity multiplier must be >= 1.0")
	}

	volumeM3 := netAreaM2 * wallThicknessM
	densityMassKg := volumeM3 * material.DensityKgPerM3
	baseMassKg := densityMassKg

	if material.CoverageRateM2PerTonne > 0 {
		coverageMassKg := netAreaM2 / material.CoverageRateM2PerTonne * 1000
		if coverageMassKg > baseMassKg {
			baseMassKg = coverageMassKg
		}
	}

	wasteStoneKg := baseMassKg * wastePercent
	stoneMassKg := (baseMassKg + wasteStoneKg) * complexityMultiplier

	return StoneCalculation{
		VolumeM3:     volumeM3,
		StoneMassKg:  stoneMassKg,
		StoneTonnage: stoneMassKg / 1000,
		WasteStoneKg: wasteStoneKg,
	}, nil
}

func CalculateMortar(netAreaM2, jointWidthM, jointDepthM float64) (float64, float64, error) {
	if netAreaM2 < 0 {
		return 0, 0, fmt.Errorf("net area must be >= 0, got %f", netAreaM2)
	}
	if jointWidthM < 0 || jointDepthM < 0 {
		return 0, 0, fmt.Errorf("joint dimensions cannot be negative")
	}

	volumeM3 := netAreaM2 * jointWidthM * jointDepthM
	return volumeM3, volumeM3 * MortarDensityKgPerM3, nil
}
