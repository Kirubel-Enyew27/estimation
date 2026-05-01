package service

import (
	"context"
	"errors"
	"estimation/domain"
	"estimation/store"
	"fmt"
	"strings"
)

const MortarDensityKgPerM3 = 2160

type Calcualator struct {
	multipliers MultiplierConfig
	materials   store.MaterialStore
}

type MultiplierConfig struct {
	MaterialWastePercent         map[string]float64
	PatternComplexityMultipliers map[string]float64
	DefaultWastePercent          float64
	DefaultComplexityMultiplier  float64
}

type AppliedMultipliers struct {
	WastePercent         float64
	ComplexityMultiplier float64
}
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

func NewCalculator() *Calcualator {
	return NewCalculatorWithConfig(DefaultMultiplierConfig())
}

func NewCalculatorWithConfig(config MultiplierConfig) *Calcualator {
	config = normalizeMultiplierConfig(config)
	return &Calcualator{multipliers: config}
}

func NewCalculatorWithMaterialStore(materials store.MaterialStore) *Calcualator {
	return NewCalculatorWithConfigAndMaterialStore(DefaultMultiplierConfig(), materials)
}

func NewCalculatorWithConfigAndMaterialStore(config MultiplierConfig, materials store.MaterialStore) *Calcualator {
	config = normalizeMultiplierConfig(config)
	return &Calcualator{
		multipliers: config,
		materials:   materials,
	}
}

func DefaultMultiplierConfig() MultiplierConfig {
	return MultiplierConfig{
		MaterialWastePercent: map[string]float64{
			"brick":      0.10,
			"fieldstone": 0.25,
		},
		PatternComplexityMultipliers: map[string]float64{
			"running-bond": 1.00,
			"stacked":      1.00,
			"ashlar":       1.10,
			"herringbone":  1.20,
		},
		DefaultWastePercent:         0,
		DefaultComplexityMultiplier: 1,
	}
}

func (c *Calcualator) Estimate(ctx context.Context, req domain.CalcualtionRequest) (domain.CalculationResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.CalculationResult{}, err
	}
	if err := req.Validate(); err != nil {
		return domain.CalculationResult{}, err
	}
	if req.Material == nil {
		if c.materials == nil {
			return domain.CalculationResult{}, errors.New("material catalog is not configured")
		}

		material, err := c.materials.GetByType(ctx, req.MaterialCode)
		if err != nil {
			return domain.CalculationResult{}, fmt.Errorf("lookup material type %q: %w", req.MaterialCode, err)
		}
		req.Material = material
	}

	surface, err := CalculateSurfaceArea(req.Wall, req.Voids)
	if err != nil {
		return domain.CalculationResult{}, err
	}

	multipliers := c.ApplyMultipliers(req)
	stone, err := CalculateStoneTonnage(surface.NetAreaM2, req.Wall.ThicknessM, *req.Material,
		multipliers.WastePercent, multipliers.ComplexityMultiplier)
	if err != nil {
		return domain.CalculationResult{}, err
	}

	result := domain.CalculationResult{
		SurfaceAreaM2:               surface.NetAreaM2,
		VolumeM3:                    surface.NetAreaM2 * req.Wall.ThicknessM,
		StoneMassKg:                 stone.StoneMassKg,
		StoneTonnage:                stone.StoneTonnage,
		WasteStoneKg:                stone.WasteStoneKg,
		AppliedComplexityMultiplier: multipliers.ComplexityMultiplier,
		BreakDown: map[string]float64{
			"totalWallAreaM2":      surface.TotalWallAreaM2,
			"voidAreaM2":           surface.VoidAreaM2,
			"netWallAreaM2":        surface.NetAreaM2,
			"wastePercent":         multipliers.WastePercent,
			"complexityMultiplier": multipliers.ComplexityMultiplier,
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

func (c *Calcualator) ApplyMultipliers(req domain.CalcualtionRequest) AppliedMultipliers {
	config := c.multipliers
	if config.DefaultComplexityMultiplier == 0 {
		config.DefaultComplexityMultiplier = 1
	}

	wastePercent := config.DefaultWastePercent
	if req.Material != nil {
		for _, key := range []string{req.Material.Type, req.Material.Code, req.Material.Name} {
			if configuredWaste, ok := config.MaterialWastePercent[normalizeMultiplierKey(key)]; ok {
				wastePercent = configuredWaste
				break
			}
		}
	}
	if configuredWaste, ok := config.MaterialWastePercent[normalizeMultiplierKey(req.MaterialCode)]; ok {
		wastePercent = configuredWaste
	}
	if req.WastePercent > 0 {
		wastePercent = req.WastePercent
	}

	complexityMultiplier := config.DefaultComplexityMultiplier
	if configuredComplexity, ok := config.PatternComplexityMultipliers[normalizeMultiplierKey(req.Pattern)]; ok {
		complexityMultiplier = configuredComplexity
	}
	if req.ComplexityMultiplier > 0 {
		complexityMultiplier = req.ComplexityMultiplier
	}

	return AppliedMultipliers{
		WastePercent:         wastePercent,
		ComplexityMultiplier: complexityMultiplier,
	}
}

func CalculateSurfaceArea(wall domain.WallDimenstions, voids []domain.Void) (SurfaceAreaCalculation, error) {
	totalWallArea := wall.SurfaceArea()
	if totalWallArea <= 0 {
		return SurfaceAreaCalculation{}, fmt.Errorf("total wall area must be > 0, got %f", totalWallArea)
	}

	var voidArea float64
	for i, void := range voids {
		if void.WidthM < 0 || void.HeightM < 0 {
			return SurfaceAreaCalculation{}, fmt.Errorf("void[%d] dimensions can not be negative", i)
		}
		voidArea += void.Area()
	}

	if voidArea > totalWallArea {
		return SurfaceAreaCalculation{}, fmt.Errorf("total void area %.2f exceeds total wall area %.2f", voidArea, totalWallArea)
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
	if material.CoverageRateM2PerRonne < 0 {
		return StoneCalculation{}, fmt.Errorf("coverage rate can not be negative")
	}
	if wastePercent < 0 {
		return StoneCalculation{}, fmt.Errorf("waste percent can not be negative")
	}
	if complexityMultiplier < 1 {
		return StoneCalculation{}, fmt.Errorf("complexity multiplier must be >= 1.0, got %f", complexityMultiplier)
	}

	volumeM3 := netAreaM2 * wallThicknessM
	densityMassKg := volumeM3 * material.DensityKgPerM3
	baseMassKg := densityMassKg

	if material.CoverageRateM2PerRonne > 0 {
		coverageMassKg := netAreaM2 / material.CoverageRateM2PerRonne * 1000
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
		return 0, 0, fmt.Errorf("joint dimensions can not be negative")
	}

	volumeM3 := netAreaM2 * jointWidthM * jointDepthM
	return volumeM3, volumeM3 * MortarDensityKgPerM3, nil
}

func normalizeMultiplierConfig(config MultiplierConfig) MultiplierConfig {
	if config.MaterialWastePercent == nil {
		config.MaterialWastePercent = map[string]float64{}
	}
	if config.PatternComplexityMultipliers == nil {
		config.PatternComplexityMultipliers = map[string]float64{}
	}
	if config.DefaultComplexityMultiplier == 0 {
		config.DefaultComplexityMultiplier = 1
	}

	config.MaterialWastePercent = normalizeFloatMapKeys(config.MaterialWastePercent)
	config.PatternComplexityMultipliers = normalizeFloatMapKeys(config.PatternComplexityMultipliers)
	return config
}

func normalizeFloatMapKeys(values map[string]float64) map[string]float64 {
	normalized := make(map[string]float64, len(values))
	for key, value := range values {
		normalized[normalizeMultiplierKey(key)] = value
	}
	return normalized
}

func normalizeMultiplierKey(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", "-"))
}
