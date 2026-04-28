package domain

import "testing"

func TestCalculationRequestValidate_Valid(t *testing.T) {
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

func TestCalculationRequestValidate_Invalid(t *testing.T) {
    cases := []struct {
        name    string
        req     CalculationRequest
        wantErr bool
    }{
        {"missing material", CalculationRequest{Wall: WallDimensions{LengthM: 1, HeightM: 1, ThicknessM: 0.2}, ComplexityMultiplier: 1.0}, true},
        {"zero length", CalculationRequest{MaterialCode: "X", Wall: WallDimensions{LengthM: 0, HeightM: 1, ThicknessM: 0.2}, ComplexityMultiplier: 1.0}, true},
        {"negative waste", CalculationRequest{MaterialCode: "X", Wall: WallDimensions{LengthM: 1, HeightM: 1, ThicknessM: 0.2}, WastePercent: -0.1, ComplexityMultiplier: 1.0}, true},
        {"complexity < 1", CalculationRequest{MaterialCode: "X", Wall: WallDimensions{LengthM: 1, HeightM: 1, ThicknessM: 0.2}, ComplexityMultiplier: 0.5}, true},
        {"void too large", CalculationRequest{MaterialCode: "X", Wall: WallDimensions{LengthM: 2, HeightM: 2, ThicknessM: 0.2}, Voids: []Void{{WidthM: 10, HeightM: 10}}, ComplexityMultiplier: 1.0}, true},
    }

    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            err := c.req.Validate()
            if (err != nil) != c.wantErr {
                t.Fatalf("case %q: expected wantErr=%v, got err=%v", c.name, c.wantErr, err)
            }
        })
    }
}
