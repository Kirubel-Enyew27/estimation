package domain

import (
	"errors"
	"fmt"
)

var ErrMaterialNotFound = errors.New("material not found")

type Material struct {
	ID                     string  `json:"id,omitempty"`
	Code                   string  `json:"code"`
	Name                   string  `json:"name"`
	DensityKgPerM3         float64 `json:"density"`
	CostPerTon             float64 `json:"cost_per_ton"`
	Type                   string  `json:"type,omitempty"`
	CoverageRateM2PerRonne float64 `json:"coverage_rate,omitempty"`
	Comment                string  `json:"comment,omitempty"`
}

type WallDimensions struct {
	LengthM    float64 `json:"lengthM"`
	HeightM    float64 `json:"heightM"`
	ThicknessM float64 `json:"thicknessM"`
}

func (w WallDimensions) SurfaceArea() float64 {
	return w.LengthM * w.HeightM
}

type Void struct {
	ID      string  `json:"id,omitempty"`
	WidthM  float64 `json:"widthM"`
	HeightM float64 `json:"heightM"`
}

func (v Void) Area() float64 { return v.WidthM * v.HeightM }

type CalculationRequest struct {
	ProjectID            string         `json:"projectId,omitempty"`
	MaterialCode         string         `json:"materialCode,omitempty"`
	Material             *Material      `json:"material,omitempty"`
	Wall                 WallDimensions `json:"wall"`
	Voids                []Void         `json:"voids,omitempty"`
	Pattern              string         `json:"pattern,omitempty"`
	JointWidthM          float64        `json:"jointWidthM,omitempty"`
	JointDepthM          float64        `json:"jointDepthM,omitempty"`
	WastePercent         float64        `json:"wastePercent,omitempty"`
	ComplexityMultiplier float64        `json:"complexityMultiplier,omitempty"`
	IncludeMortar        bool           `json:"includeMortar,omitempty"`
}

func (r *CalculationRequest) Validate() error {
	if r.Material == nil && r.MaterialCode == "" {
		return errors.New("material must be provided either by code or inline")
	}
	if r.Material != nil {
		if r.Material.DensityKgPerM3 <= 0 {
			return fmt.Errorf("material.density must be > 0, got %f", r.Material.DensityKgPerM3)
		}
		if r.Material.CostPerTon < 0 {
			return fmt.Errorf("material.cost_per_ton must be >= 0, got %f", r.Material.CostPerTon)
		}
		if r.Material.CoverageRateM2PerRonne < 0 {
			return fmt.Errorf("material.coverage_rate must be >= 0, got %f", r.Material.CoverageRateM2PerRonne)
		}
	}
	if r.Wall.LengthM <= 0 {
		return fmt.Errorf("wall.lengthM must be > 0, got %f", r.Wall.LengthM)
	}
	if r.Wall.HeightM <= 0 {
		return fmt.Errorf("wall.heightM must be > 0, got %f", r.Wall.HeightM)
	}
	if r.Wall.ThicknessM < 0 {
		return fmt.Errorf("wall.thicknessM must be >= 0, got %f", r.Wall.ThicknessM)
	}
	if r.WastePercent < 0 {
		return fmt.Errorf("wastePercent cannot be negative")
	}
	if r.ComplexityMultiplier != 0 && r.ComplexityMultiplier < 1.0 {
		return fmt.Errorf("complexityMultiplier must be >= 1.0")
	}
	if r.JointWidthM < 0 || r.JointDepthM < 0 {
		return fmt.Errorf("joint dimensions cannot be negative")
	}

	grossArea := r.Wall.SurfaceArea()
	var totalVoidArea float64
	for i := range r.Voids {
		width := r.Voids[i].WidthM
		height := r.Voids[i].HeightM
		if width < 0 {
			return fmt.Errorf("void[%d].widthM must be >= 0, got %f", i, width)
		}
		if height < 0 {
			return fmt.Errorf("void[%d].heightM must be >= 0, got %f", i, height)
		}
		if width > r.Wall.LengthM {
			return fmt.Errorf("void[%d].widthM must be <= wall.lengthM (%f), got %f", i, r.Wall.LengthM, width)
		}
		if height > r.Wall.HeightM {
			return fmt.Errorf("void[%d].heightM must be <= wall.heightM (%f), got %f", i, r.Wall.HeightM, height)
		}
		area := width * height
		if area > grossArea {
			return fmt.Errorf("void[%d] area exceeds gross wall area", i)
		}
		totalVoidArea += area
	}
	if totalVoidArea > grossArea {
		return fmt.Errorf("total void area %.2f exceeds gross wall area %.2f", totalVoidArea, grossArea)
	}
	return nil
}

type CalculationResult struct {
	SurfaceAreaM2               float64            `json:"surfaceAreaM2"`
	VolumeM3                    float64            `json:"volumeM3"`
	StoneMassKg                 float64            `json:"stoneMassKg"`
	StoneTonnage                float64            `json:"stoneTonnage"`
	MortarVolumeM3              float64            `json:"mortarVolumeM3,omitempty"`
	MortarMassKg                float64            `json:"mortarMassKg,omitempty"`
	WasteStoneKg                float64            `json:"wasteStoneKg,omitempty"`
	AppliedComplexityMultiplier float64            `json:"appliedComplexityMultiplier"`
	Breakdown                   map[string]float64 `json:"breakDown,omitempty"`
}
