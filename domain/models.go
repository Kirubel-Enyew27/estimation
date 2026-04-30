package domain

import (
	"errors"
	"fmt"
)

type Material struct {
	ID             string  `json:"id,omitempty"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	DensityKgPerM3 float64 `json:"densityKgPerM3"`
	CoverageRateM2PerRonne float64 `json:"coverageRateM2PerRonne,omitempty"`
	Comment        string  `json:"comment,omitempty"`
}

type WallDimenstions struct {
	LengthM    float64 `json:"lengthM"`
	HeightM    float64 `json:"heightM"`
	ThicknessM float64 `json:"thicknessM"`
}

func (w WallDimenstions) SurfaceArea() float64 {
	return w.LengthM * w.HeightM
}

type Void struct {
	ID      string  `json:"id,omitempty"`
	WidthM  float64 `json:"widthM"`
	HeightM float64 `json:"heightM"`
}

func (v Void) Area() float64 { return v.WidthM * v.HeightM }

type CalcualtionRequest struct {
	ProjectID            string          `json:"projectId,omitempty"`
	MaterialCode         string          `json:"materialCode,omitempty"`
	Material             *Material       `json:"material,omitempty"`
	Wall                 WallDimenstions `json:"wall"`
	Voids                []Void          `json:"voids,omitempty"`
	JointWidthM          float64         `json:"jointWidthM,omitempty"`
	JointDepthM          float64         `json:"jointDepthM,omitempty"`
	WastePercent         float64         `json:"wastePercent,omitempty"`
	ComplexityMultiplier float64         `json:"complexityMultiplier,omitempty"`
	IncludeMortar        bool            `json:"includeMortar,omitempty"`
}

func (r *CalcualtionRequest) Validate() error {
	if r.Material == nil && r.MaterialCode == "" {
		return errors.New("material must be provided either by code or inline")
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
	if r.ComplexityMultiplier < 1.0 {
		return fmt.Errorf("complexityMultiplier must be >= 1.0")
	}
	if r.JointWidthM < 0 || r.JointDepthM < 0 {
		return fmt.Errorf("joint dimensions cannot be negative")
	}

	grossArea := r.Wall.SurfaceArea()
	for i, v := range r.Voids {
		if v.WidthM < 0 || v.HeightM < 0 {
			return fmt.Errorf("void[%d] dimensions can not be negative", i)
		}
		if v.Area() > grossArea {
			return fmt.Errorf("void[%d] area exceeds gross wall area", i)
		}
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
	BreakDown                   map[string]float64 `json:"breakDown,omitempty"`
}
