package domain

import (
	"strings"
	"testing"
)

func TestCalculationRequestValidate(t *testing.T) {
	req := CalculationRequest{
		MaterialCode: "GRAN-01",
		Wall: WallDimensions{
			LengthM:    5.0,
			HeightM:    2.5,
			ThicknessM: 0.25,
		},
		WastePercent:         0.05,
		ComplexityMultiplier: 1.0,
		JointWidthM:          0.01,
		JointDepthM:          0.02,
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}
}

func TestCalculatonRequestValidate_Invalid(t *testing.T) {
	cases := []struct {
		name    string
		req     CalculationRequest
		wantErr string
	}{
		{"missing material", CalculationRequest{
			Wall: WallDimensions{
				LengthM:    1,
				HeightM:    1,
				ThicknessM: 0.2},
			ComplexityMultiplier: 1.0,
		}, "material must be provided"},
		{"zero length", CalculationRequest{
			MaterialCode: "X",
			Wall: WallDimensions{
				LengthM:    0,
				HeightM:    1,
				ThicknessM: 0.2},
			ComplexityMultiplier: 1.0,
		}, "wall.lengthM must be > 0"},
		{"negative thickness", CalculationRequest{
			MaterialCode: "X",
			Wall: WallDimensions{
				LengthM:    1,
				HeightM:    1,
				ThicknessM: -0.2},
			ComplexityMultiplier: 1.0,
		}, "wall.thicknessM must be >= 0"},
		{"negative waste", CalculationRequest{
			MaterialCode: "X",
			Wall: WallDimensions{
				LengthM:    1,
				HeightM:    1,
				ThicknessM: 0.2},
			WastePercent:         -0.1,
			ComplexityMultiplier: 1.0,
		}, "wastePercent cannot be negative"},
		{"complexity < 1", CalculationRequest{
			MaterialCode: "X",
			Wall: WallDimensions{
				LengthM:    1,
				HeightM:    1,
				ThicknessM: 0.2},
			ComplexityMultiplier: 0.5,
		}, "complexityMultiplier must be >= 1.0"},
		{"void negative width", CalculationRequest{
			MaterialCode: "X",
			Wall: WallDimensions{
				LengthM:    2,
				HeightM:    2,
				ThicknessM: 0.2},
			Voids: []Void{
				{WidthM: -1, HeightM: 1},
			},
			ComplexityMultiplier: 1.0,
		}, "void[0].widthM must be >= 0"},
		{"void width exceeds wall", CalculationRequest{
			MaterialCode: "X",
			Wall: WallDimensions{
				LengthM:    2,
				HeightM:    2,
				ThicknessM: 0.2},
			Voids: []Void{
				{WidthM: 3, HeightM: 1},
			},
			ComplexityMultiplier: 1.0,
		}, "void[0].widthM must be <= wall.lengthM"},
		{"void height exceeds wall", CalculationRequest{
			MaterialCode: "X",
			Wall: WallDimensions{
				LengthM:    2,
				HeightM:    2,
				ThicknessM: 0.2},
			Voids: []Void{
				{WidthM: 1, HeightM: 3},
			},
			ComplexityMultiplier: 1.0,
		}, "void[0].heightM must be <= wall.heightM"},
		{"total void area too large", CalculationRequest{
			MaterialCode: "X",
			Wall: WallDimensions{
				LengthM:    2,
				HeightM:    2,
				ThicknessM: 0.2},
			Voids: []Void{
				{WidthM: 2, HeightM: 2},
				{WidthM: 1, HeightM: 1},
			},
			ComplexityMultiplier: 1.0,
		}, "total void area"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.req.Validate()
			if err == nil {
				t.Fatalf("expected error containing %q", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("got error %q, want it to contain %q", err.Error(), c.wantErr)
			}
		})
	}
}
