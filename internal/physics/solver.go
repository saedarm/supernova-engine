// Package physics implements the Lane-Emden numerical solver and
// Hertzsprung-Russell diagram interpolation. This package has zero
// rendering dependencies — it is pure math.
package physics

import "math"

// ════════════════════════════════════════════════════════
// LANE-EMDEN NUMERICAL SOLVER
//
// The Lane-Emden equation describes hydrostatic equilibrium in a
// self-gravitating polytropic sphere:
//
//   d²θ/dξ² + (2/ξ)(dθ/dξ) + θⁿ = 0
//
// Boundary conditions: θ(0) = 1, θ'(0) = 0
// Rewritten as a system:
//   dθ/dξ = φ
//   dφ/dξ = -θⁿ - 2φ/ξ
//
// Solved via 4th-order Runge-Kutta.
// ════════════════════════════════════════════════════════

// DensityProfile holds the numerical solution to the Lane-Emden equation.
// Xi is the dimensionless radius, Theta is the dimensionless density (1 at
// center, 0 at surface), Phi is dθ/dξ, and SurfaceXi is ξ₁ where θ first
// reaches zero.
type DensityProfile struct {
	Xi        []float64
	Theta     []float64
	Phi       []float64
	SurfaceXi float64
}

// SolveLaneEmden numerically integrates the Lane-Emden equation for
// polytropic index n using 4th-order Runge-Kutta with the given number
// of integration steps (h=0.005 per step).
func SolveLaneEmden(n float64, steps int) *DensityProfile {
	h := 0.005
	xi := []float64{1e-6}
	theta := []float64{1.0}
	phi := []float64{0.0}

	for i := 0; i < steps; i++ {
		x, t, p := xi[i], theta[i], phi[i]
		if t <= 0 {
			break
		}

		// Stage 1
		f1t := p
		f1p := -math.Pow(math.Max(t, 0), n) - (2*p)/math.Max(x, 1e-10)

		// Stage 2
		x2 := x + 0.5*h
		t2, p2 := t+0.5*h*f1t, p+0.5*h*f1p
		f2t := p2
		f2p := -math.Pow(math.Max(t2, 0), n) - (2*p2)/math.Max(x2, 1e-10)

		// Stage 3
		t3, p3 := t+0.5*h*f2t, p+0.5*h*f2p
		f3t := p3
		f3p := -math.Pow(math.Max(t3, 0), n) - (2*p3)/math.Max(x2, 1e-10)

		// Stage 4
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
		phi = append(phi, newPhi)
	}

	return &DensityProfile{
		Xi:        xi,
		Theta:     theta,
		Phi:       phi,
		SurfaceXi: xi[len(xi)-1],
	}
}

// ════════════════════════════════════════════════════════
// HR DIAGRAM INTERPOLATION
// ════════════════════════════════════════════════════════

// HRPoint represents a waypoint on an evolutionary track:
// [0] = age fraction (0.0–1.2), [1] = log10(Teff), [2] = log10(L/Lsun).
// A point with LogT=0 and LogL=0 means the star has collapsed (black hole).
type HRPoint [3]float64

func (p HRPoint) AgeFrac() float64 { return p[0] }
func (p HRPoint) LogT() float64    { return p[1] }
func (p HRPoint) LogL() float64    { return p[2] }

// TrailPoint is a recorded position on the HR diagram during aging.
type TrailPoint struct {
	LogT, LogL float64
}

// InterpolateHR linearly interpolates along an evolutionary track to find
// the star's logT and logL at a given age fraction.
func InterpolateHR(track []HRPoint, ageFrac float64) (logT, logL float64) {
	if len(track) == 0 {
		return 3.76, 0 // Solar default
	}
	af := Clamp(ageFrac, 0, track[len(track)-1].AgeFrac())
	for i := 0; i < len(track)-1; i++ {
		if af >= track[i].AgeFrac() && af <= track[i+1].AgeFrac() {
			t := (af - track[i].AgeFrac()) / (track[i+1].AgeFrac() - track[i].AgeFrac())
			return track[i].LogT() + t*(track[i+1].LogT()-track[i].LogT()),
				track[i].LogL() + t*(track[i+1].LogL()-track[i].LogL())
		}
	}
	last := track[len(track)-1]
	return last.LogT(), last.LogL()
}

// ════════════════════════════════════════════════════════
// UTILITY
// ════════════════════════════════════════════════════════

// Clamp restricts v to the range [lo, hi].
func Clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
