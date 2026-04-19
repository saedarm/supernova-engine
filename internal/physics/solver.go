// Package physics implements the Lane-Emden numerical solver.
// Zero rendering dependencies — pure math.
//
// The Lane-Emden equation describes hydrostatic equilibrium in a
// self-gravitating polytropic sphere:
//
//	d²θ/dξ² + (2/ξ)(dθ/dξ) + θⁿ = 0
//
// Boundary conditions: θ(0) = 1, θ'(0) = 0
// Solved via 4th-order Runge-Kutta.
package physics

import "math"

// DensityProfile holds the numerical solution to the Lane-Emden equation.
type DensityProfile struct {
	Xi        []float64 // Dimensionless radius
	Theta     []float64 // Dimensionless density (1 at center, 0 at surface)
	SurfaceXi float64   // ξ₁ — the radius where θ reaches zero
}

// SolveLaneEmden numerically integrates the Lane-Emden equation for
// polytropic index n using 4th-order Runge-Kutta (h=0.005, up to 2000 steps).
func SolveLaneEmden(n float64, steps int) *DensityProfile {
	h := 0.005
	xi := []float64{1e-6}
	theta := []float64{1.0}
	phi := 0.0 // dθ/dξ — we only need the current value, not the history

	for i := 0; i < steps; i++ {
		x, t, p := xi[i], theta[i], phi
		if t <= 0 {
			break
		}

		// RK4 stages
		f1t := p
		f1p := -math.Pow(math.Max(t, 0), n) - (2*p)/math.Max(x, 1e-10)

		x2 := x + 0.5*h
		t2, p2 := t+0.5*h*f1t, p+0.5*h*f1p
		f2t := p2
		f2p := -math.Pow(math.Max(t2, 0), n) - (2*p2)/math.Max(x2, 1e-10)

		t3, p3 := t+0.5*h*f2t, p+0.5*h*f2p
		f3t := p3
		f3p := -math.Pow(math.Max(t3, 0), n) - (2*p3)/math.Max(x2, 1e-10)

		x4 := x + h
		t4, p4 := t+h*f3t, p+h*f3p
		f4t := p4
		f4p := -math.Pow(math.Max(t4, 0), n) - (2*p4)/math.Max(x4, 1e-10)

		newTheta := t + (h/6)*(f1t+2*f2t+2*f3t+f4t)
		newPhi := p + (h/6)*(f1p+2*f2p+2*f3p+f4p)

		if newTheta <= 0 || math.IsNaN(newTheta) {
			break
		}

		xi = append(xi, x4)
		theta = append(theta, newTheta)
		phi = newPhi
	}

	return &DensityProfile{
		Xi:        xi,
		Theta:     theta,
		SurfaceXi: xi[len(xi)-1],
	}
}

// StarType defines a category of star with its polytropic properties and colors.
type StarType struct {
	Name        string
	Mass        float64    // Solar masses
	PolyIndex   float64    // n in the Lane-Emden equation
	Description string     // One-line physics description
	BaseColor   [3]float64 // Core RGB [0-1]
	CoronaColor [3]float64 // Surface RGB [0-1]
}

// StarTypes is the catalog of available star types.
var StarTypes = map[string]*StarType{
	"red_dwarf": {
		Name: "Red Dwarf (M-type)", Mass: 0.3, PolyIndex: 1.5,
		Description: "Convective interior. n=1.5 polytrope.",
		BaseColor: [3]float64{1.0, 0.3, 0.1}, CoronaColor: [3]float64{0.8, 0.15, 0.05},
	},
	"sun_like": {
		Name: "Sun-like (G-type)", Mass: 1.0, PolyIndex: 3.0,
		Description: "Radiative equilibrium. Eddington standard model. n=3.0.",
		BaseColor: [3]float64{1.0, 0.95, 0.6}, CoronaColor: [3]float64{1.0, 0.7, 0.2},
	},
	"blue_giant": {
		Name: "Blue Giant (O/B-type)", Mass: 15, PolyIndex: 3.0,
		Description: "Massive radiative star. Same n=3.0, higher mass.",
		BaseColor: [3]float64{0.5, 0.7, 1.0}, CoronaColor: [3]float64{0.3, 0.4, 1.0},
	},
	"white_dwarf": {
		Name: "White Dwarf", Mass: 0.6, PolyIndex: 1.5,
		Description: "Electron-degenerate. n=1.5 (non-relativistic).",
		BaseColor: [3]float64{0.85, 0.9, 1.0}, CoronaColor: [3]float64{0.6, 0.7, 1.0},
	},
	"neutron_star_rel": {
		Name: "Neutron Star (rel.)", Mass: 1.4, PolyIndex: 3.0,
		Description: "Relativistic degeneracy. n=3.0 polytrope.",
		BaseColor: [3]float64{0.6, 0.7, 1.0}, CoronaColor: [3]float64{0.3, 0.4, 0.8},
	},
	"custom": {
		Name: "Custom (n=2.0)", Mass: 1.0, PolyIndex: 2.0,
		Description: "Intermediate polytrope. Adjust with +/-.",
		BaseColor: [3]float64{0.8, 0.6, 1.0}, CoronaColor: [3]float64{0.5, 0.3, 0.7},
	},
}

// StarTypeOrder defines the display/key-binding order.
var StarTypeOrder = []string{
	"red_dwarf", "sun_like", "blue_giant", "white_dwarf", "neutron_star_rel", "custom",
}

// Clamp restricts v to [lo, hi].
func Clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
