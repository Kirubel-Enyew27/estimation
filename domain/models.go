package domain

import (
	"errors"
	"fmt"
)

// Material represents an entry in the material catalog.
// Density is expected in kg/m³.
type Material struct {
    ID             string  `json:"id,omitempty"`
    Code           string  `json:"code"`
    Name           string  `json:"name"`
    DensityKgPerM3 float64 `json:"densityKgPerM3"`
    Comment        string  `json:"comment,omitempty"`
}

// WallDimensions describes the gross wall geometry in meters.
type WallDimensions struct {
    LengthM    float64 `json:"lengthM"`
    HeightM    float64 `json:"heightM"`
    ThicknessM float64 `json:"thicknessM"`
}

// SurfaceArea returns the gross planar area (length * height).
func (w WallDimensions) SurfaceArea() float64 { return w.LengthM * w.HeightM }

// Void represents an opening (window/door) to subtract from area.
type Void struct {
    ID      string  `json:"id,omitempty"`
    WidthM  float64 `json:"widthM"`
    HeightM float64 `json:"heightM"`
}

// Area returns the area of the void.
func (v Void) Area() float64 { return v.WidthM * v.HeightM }

// CalculationRequest is the input DTO for an estimate.
type CalculationRequest struct {
    ProjectID            string         `json:"projectId,omitempty"`
    MaterialCode         string         `json:"materialCode,omitempty"`
    Material             *Material      `json:"material,omitempty"`
    Wall                 WallDimensions `json:"wall"`
    Voids                []Void         `json:"voids,omitempty"`
    JointWidthM          float64        `json:"jointWidthM,omitempty"`
    JointDepthM          float64        `json:"jointDepthM,omitempty"`
    WastePercent         float64        `json:"wastePercent,omitempty"`
    ComplexityMultiplier float64        `json:"complexityMultiplier,omitempty"`
    IncludeMortar        bool           `json:"includeMortar,omitempty"`
}

// Validate performs basic validation on the request fields.
// It enforces presence of either MaterialCode or an inline Material and
// ensures geometric and percentage fields are within sensible ranges.
func (r *CalculationRequest) Validate() error {
    if r.Material == nil && r.MaterialCode == "" {
        return errors.New("material must be provided either by code or inline")
    }
    if r.Wall.LengthM <= 0 {
        return fmt.Errorf("wall.lengthM must be > 0")
    }
    if r.Wall.HeightM <= 0 {
        return fmt.Errorf("wall.heightM must be > 0")
    }
    if r.Wall.ThicknessM <= 0 {
        return fmt.Errorf("wall.thicknessM must be > 0")
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
            return fmt.Errorf("void[%d] dimensions cannot be negative", i)
        }
        if v.Area() > grossArea {
            return fmt.Errorf("void[%d] area exceeds gross wall area", i)
        }
    }

    return nil
}

// CalculationResult is the output DTO returned by the estimator.
// Units: areas in m², volumes in m³, mass in kg, tonnage in metric tons.
type CalculationResult struct {
    SurfaceAreaM2               float64            `json:"surfaceAreaM2"`
    VolumeM3                    float64            `json:"volumeM3"`
    StoneMassKg                 float64            `json:"stoneMassKg"`
    StoneTonnage                float64            `json:"stoneTonnage"`
    MortarVolumeM3              float64            `json:"mortarVolumeM3,omitempty"`
    MortarMassKg                float64            `json:"mortarMassKg,omitempty"`
    WasteStoneKg                float64            `json:"wasteStoneKg,omitempty"`
    AppliedComplexityMultiplier float64            `json:"appliedComplexityMultiplier"`
    Breakdown                   map[string]float64 `json:"breakdown,omitempty"`
}
