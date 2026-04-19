// Package render draws the star, density plot, and control panel
// using Ebitengine's 2D primitives.
package render

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/saedarm/supernova-engine/internal/physics"
)

// ════════════════════════════════════════════════════════
// STAR VIEWPORT
// ════════════════════════════════════════════════════════

// Star draws the Lane-Emden density shell visualization.
func Star(screen *ebiten.Image, profile *physics.DensityProfile, st *physics.StarType,
	bgStars [][3]float64, brightness, time, ox, oy, w, h float64) {

	// Background stars
	for _, s := range bgStars {
		sx := s[0]/1400*w + ox
		sy := s[1] / 850 * h
		if sx < ox || sx > ox+w {
			continue
		}
		ebitenutil.DrawRect(screen, sx, sy, 1.2, 1.2,
			color.RGBA{200, 210, 255, uint8(s[2] * 180)})
	}

	cx, cy := ox+w/2, oy+h/2
	baseRadius := math.Min(w, h) * 0.28

	// Breathing pulse
	pulse := 1.0 + math.Sin(time*2)*0.03
	effectiveRadius := baseRadius * pulse

	// ── Density shells from Lane-Emden solution ──
	if profile != nil && len(profile.Theta) > 0 {
		numShells := min(len(profile.Theta), 70)
		step := max(1, len(profile.Theta)/numShells)

		for i := numShells - 1; i >= 0; i-- {
			idx := min(i*step, len(profile.Theta)-1)
			r := (profile.Xi[idx] / profile.SurfaceXi) * effectiveRadius
			density := math.Max(0, profile.Theta[idx])

			cr := st.BaseColor[0]*density + st.CoronaColor[0]*(1-density)
			cg := st.BaseColor[1]*density + st.CoronaColor[1]*(1-density)
			cb := st.BaseColor[2]*density + st.CoronaColor[2]*(1-density)
			alpha := (0.02 + density*0.12) * 0.7 * brightness

			col := color.RGBA{
				uint8(cr * 255), uint8(cg * 255), uint8(cb * 255),
				uint8(physics.Clamp(alpha*255, 0, 255)),
			}
			vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(r), col, true)
		}
	}

	// ── Core glow ──
	for i := 0; i < 6; i++ {
		r := effectiveRadius * float64(6-i) / 6
		a := uint8(physics.Clamp(float64(i)*35*brightness, 0, 255))
		col := color.RGBA{
			uint8(st.BaseColor[0] * 255), uint8(st.BaseColor[1] * 255),
			uint8(st.BaseColor[2] * 255), a,
		}
		vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(r), col, true)
	}

	// ── Corona ──
	for i := 0; i < 4; i++ {
		r := effectiveRadius * (1.2 + float64(i)*0.3)
		a := uint8(physics.Clamp((20-float64(i)*5)*brightness, 0, 255))
		col := color.RGBA{
			uint8(st.CoronaColor[0] * 255), uint8(st.CoronaColor[1] * 255),
			uint8(st.CoronaColor[2] * 255), a,
		}
		vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(r), col, true)
	}
}

// ════════════════════════════════════════════════════════
// DENSITY PLOT — θ(ξ)
// ════════════════════════════════════════════════════════

// DensityPlot draws the Lane-Emden solution curve.
func DensityPlot(screen *ebiten.Image, profile *physics.DensityProfile, st *physics.StarType,
	x, y, w, h float64) {

	// Background
	ebitenutil.DrawRect(screen, x, y, w, h, color.RGBA{8, 8, 20, 255})

	if profile == nil || len(profile.Theta) == 0 {
		return
	}

	padL, padB, padT, padR := 36.0, 20.0, 14.0, 10.0
	plotW := w - padL - padR
	plotH := h - padT - padB
	axX := x + padL
	axY := y + h - padB

	// Axes
	axisCol := color.RGBA{50, 50, 70, 255}
	vector.StrokeLine(screen, float32(axX), float32(y+padT), float32(axX), float32(axY), 1, axisCol, true)
	vector.StrokeLine(screen, float32(axX), float32(axY), float32(x+w-padR), float32(axY), 1, axisCol, true)

	// Axis labels
	ebitenutil.DebugPrintAt(screen, "theta(xi)", int(x)+2, int(y)+2)
	ebitenutil.DebugPrintAt(screen, "xi", int(x+w)-18, int(axY)+4)

	// Y-axis ticks
	for _, tv := range []float64{0, 0.25, 0.5, 0.75, 1.0} {
		ty := axY - tv*plotH
		vector.StrokeLine(screen, float32(axX-3), float32(ty), float32(axX), float32(ty), 1, axisCol, true)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.1f", tv), int(x)+2, int(ty)-4)
		// Grid line
		if tv > 0 && tv < 1 {
			vector.StrokeLine(screen, float32(axX), float32(ty), float32(x+w-padR), float32(ty),
				0.5, color.RGBA{255, 255, 255, 8}, true)
		}
	}

	// X-axis ticks (ξ values)
	xiMax := profile.SurfaceXi
	for frac := 0.25; frac <= 1.0; frac += 0.25 {
		tx := axX + frac*plotW
		vector.StrokeLine(screen, float32(tx), float32(axY), float32(tx), float32(axY+3), 1, axisCol, true)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.1f", frac*xiMax), int(tx)-8, int(axY)+4)
		// Grid line
		vector.StrokeLine(screen, float32(tx), float32(y+padT), float32(tx), float32(axY),
			0.5, color.RGBA{255, 255, 255, 8}, true)
	}

	// ── Plot the curve ──
	var prevPx, prevPy float32
	for i := 0; i < len(profile.Theta); i += 2 {
		frac := profile.Xi[i] / profile.SurfaceXi
		px := float32(axX + frac*plotW)
		py := float32(axY - profile.Theta[i]*plotH)
		if i > 0 {
			vector.StrokeLine(screen, prevPx, prevPy, px, py, 2,
				color.RGBA{68, 170, 255, 220}, true)
		}
		prevPx, prevPy = px, py
	}

	// ── Fill area under curve (subtle) ──
	for i := 0; i < len(profile.Theta); i += 4 {
		frac := profile.Xi[i] / profile.SurfaceXi
		px := float32(axX + frac*plotW)
		py := float32(axY - profile.Theta[i]*plotH)
		lineH := float32(axY) - py
		if lineH > 0 {
			vector.DrawFilledRect(screen, px, py, 2, lineH,
				color.RGBA{68, 170, 255, 12}, true)
		}
	}

	// Info box
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("n = %.2f   xi_1 = %.4f   %s", st.PolyIndex, profile.SurfaceXi, st.Name),
		int(axX)+8, int(y)+4)
}

// ════════════════════════════════════════════════════════
// CONTROL PANEL
// ════════════════════════════════════════════════════════

const PanelWidth = 280

// Panel draws the left sidebar.
func Panel(screen *ebiten.Image, starTypeKey string, customN, brightness float64,
	profile *physics.DensityProfile, mouseX, mouseY int) {

	st := physics.StarTypes[starTypeKey]
	pw := float64(PanelWidth)
	sh := float64(screen.Bounds().Dy())

	// Background
	ebitenutil.DrawRect(screen, 0, 0, pw, sh, color.RGBA{10, 10, 26, 245})
	ebitenutil.DrawRect(screen, pw-1, 0, 1, sh, color.RGBA{26, 26, 46, 255})

	y := 10.0

	// Title
	ebitenutil.DebugPrintAt(screen, "=== SUPERNOVA ENGINE ===", 40, int(y))
	y += 14
	ebitenutil.DebugPrintAt(screen, "Lane-Emden Equation Solver", 32, int(y))
	y += 22

	// Star type selector
	sectionHeader(screen, "POLYTROPE [1-6] (click)", 8, int(y))
	y += 14
	for i, key := range physics.StarTypeOrder {
		t := physics.StarTypes[key]
		sel := key == starTypeKey
		if sel {
			ebitenutil.DrawRect(screen, 8, y, pw-16, 16, color.RGBA{26, 26, 58, 255})
			ebitenutil.DrawRect(screen, 8, y, 2, 16, color.RGBA{68, 170, 255, 255})
		}
		hovered := mouseX >= 8 && mouseX <= int(pw)-8 && mouseY >= int(y) && mouseY < int(y)+16
		if hovered && !sel {
			ebitenutil.DrawRect(screen, 8, y, pw-16, 16, color.RGBA{20, 20, 40, 255})
		}
		marker := "  "
		if sel {
			marker = "> "
		}
		nStr := fmt.Sprintf("%.1f", t.PolyIndex)
		if key == "custom" {
			nStr = fmt.Sprintf("%.2f", customN)
		}
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("%s%d. n=%s %s", marker, i+1, nStr, t.Name), 10, int(y)+2)
		y += 18
	}
	y += 6

	// Description
	ebitenutil.DebugPrintAt(screen, st.Description, 10, int(y))
	y += 14
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Mass: %.1f M_sun", st.Mass), 10, int(y))
	y += 18

	// Custom n slider info
	if starTypeKey == "custom" {
		sectionHeader(screen, "CUSTOM INDEX [+/-]", 8, int(y))
		y += 14
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("n = %.2f  (range 0.5 - 4.9)", customN), 10, int(y))
		y += 14

		// Visual bar
		barW := pw - 24
		ebitenutil.DrawRect(screen, 10, y, barW, 8, color.RGBA{20, 20, 35, 255})
		frac := (customN - 0.5) / 4.4
		ebitenutil.DrawRect(screen, 10, y, frac*barW, 8, color.RGBA{68, 170, 255, 180})
		y += 16
	} else {
		y += 4
	}

	// Solution readout
	sectionHeader(screen, "SOLUTION", 8, int(y))
	y += 14
	if profile != nil {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("Polytropic index n = %.2f", st.PolyIndex), 10, int(y))
		if starTypeKey == "custom" {
			// Show custom n instead
			ebitenutil.DrawRect(screen, 10, y-1, pw-20, 12, color.RGBA{10, 10, 26, 255})
			ebitenutil.DebugPrintAt(screen,
				fmt.Sprintf("Polytropic index n = %.2f", customN), 10, int(y))
		}
		y += 14
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("Surface xi_1 = %.4f", profile.SurfaceXi), 10, int(y))
		y += 14
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("Integration steps: %d", len(profile.Theta)), 10, int(y))
		y += 14

		// Central density ratio
		if len(profile.Theta) > 10 {
			// ξ₁²/3 gives ρ_c/ρ_avg for some polytropes
			ratio := math.Pow(profile.SurfaceXi, 2) / 3
			ebitenutil.DebugPrintAt(screen,
				fmt.Sprintf("rho_c / rho_avg ~ %.1f", ratio), 10, int(y))
			y += 14
		}
	}
	y += 10

	// Brightness
	sectionHeader(screen, "BRIGHTNESS [ ] ]", 8, int(y))
	y += 14
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("%.0f%%", brightness*100), 10, int(y))
	y += 18

	// Controls
	sectionHeader(screen, "CONTROLS", 8, int(y))
	y += 14
	for _, l := range []string{
		"1-6  Select polytrope",
		"+/-  Adjust custom n",
		"[ ]  Brightness",
	} {
		ebitenutil.DebugPrintAt(screen, l, 10, int(y))
		y += 12
	}
	y += 8

	// Physics info
	sectionHeader(screen, "EQUATION", 8, int(y))
	y += 14
	for _, l := range []string{
		"d2theta/dxi2 + (2/xi)",
		"  * (dtheta/dxi) + theta^n = 0",
		"",
		"Boundary conditions:",
		"  theta(0) = 1, theta'(0) = 0",
		"",
		"Solver: RK4, h=0.005",
	} {
		ebitenutil.DebugPrintAt(screen, l, 10, int(y))
		y += 12
	}
	y += 8

	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("FPS: %.0f", ebiten.ActualFPS()), 10, int(y))
}

func sectionHeader(screen *ebiten.Image, title string, x, y int) {
	ebitenutil.DrawRect(screen, float64(x), float64(y),
		float64(PanelWidth)-16, 12, color.RGBA{15, 15, 28, 255})
	ebitenutil.DebugPrintAt(screen, strings.ToUpper(title), x+2, y+1)
}
