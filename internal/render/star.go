// Package render contains all Ebitengine drawing code for the star
// viewport, HR diagram, density plot, and UI panel.
package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/saedarm/supernova-engine/internal/physics"
	"github.com/saedarm/supernova-engine/internal/state"
)

// StarViewport draws the star simulation area including background stars,
// the star itself, and the HUD overlay.
func StarViewport(screen *ebiten.Image, g *state.Game, ox, oy, w, h float64) {
	st := g.ST()

	// Background stars
	for _, s := range g.BgStars {
		sx := s[0]/float64(state.ScreenWidth)*w + ox
		sy := s[1] / float64(state.ScreenHeight) * h
		if sx < ox || sx > ox+w {
			continue
		}
		a := uint8(s[2] * 180)
		ebitenutil.DrawRect(screen, sx, sy, 1.2, 1.2, color.RGBA{200, 210, 255, a})
	}

	cx, cy := ox+w/2, oy+h/2
	baseRadius := math.Min(w, h) * 0.22

	drawStar(screen, g, cx, cy, baseRadius, st)

	// HUD
	hudX := int(ox+w) - 180
	ebitenutil.DebugPrintAt(screen, st.Name, hudX, int(oy)+8)
	ebitenutil.DebugPrintAt(screen, g.Phase.String(), hudX, int(oy)+22)
	ebitenutil.DebugPrintAt(screen, fmtf("n=%.1f  %.1f M_sun", st.PolyIndex, st.Mass), hudX, int(oy)+36)
	ebitenutil.DebugPrintAt(screen, fmtf("Age: %.4f Gyr", g.Age), hudX, int(oy)+50)
	elStr := "Elements: "
	for _, el := range physics.ElementOrder {
		if g.Elements[el] {
			elStr += physics.Elements[el].Symbol + " "
		}
	}
	ebitenutil.DebugPrintAt(screen, elStr, hudX, int(oy)+64)
}

func drawStar(screen *ebiten.Image, g *state.Game, cx, cy, baseRadius float64, st *physics.StarType) {
	effectiveRadius := baseRadius
	coreColor := st.BaseColor
	coronaCol := st.CoronaColor
	pulseAmp := 0.03
	pulseSpeed := 2.0
	shellOp := 0.7

	switch g.Phase {
	case state.PhaseRedGiant:
		effectiveRadius = baseRadius * (1.5 + 0.8*g.AgeFrac())
		coreColor = [3]float64{1.0, 0.35, 0.1}
		coronaCol = [3]float64{1.0, 0.15, 0.0}
		pulseAmp = 0.06
		shellOp = 0.4
	case state.PhaseSupernova:
		t := g.CollapseProgress
		if t < 0.3 {
			effectiveRadius = baseRadius * (2.5 + t*8)
			coreColor = [3]float64{1.0, 1.0, 0.8}
		} else if t < 0.7 {
			effectiveRadius = baseRadius * (5.0 - (t-0.3)*10)
			coreColor = [3]float64{1.0, 0.7, 0.2}
		} else {
			effectiveRadius = baseRadius * math.Max(0.05, 0.3*(1-(t-0.7)/0.3))
			coreColor = [3]float64{0.5, 0.5, 1.0}
		}
	case state.PhaseNeutronStar:
		effectiveRadius = baseRadius * 0.12
		coreColor = [3]float64{0.6, 0.7, 1.0}
		coronaCol = [3]float64{0.3, 0.4, 1.0}
		pulseSpeed = 20
		pulseAmp = 0.15
	case state.PhaseBlackHole:
		drawBlackHole(screen, g, cx, cy, baseRadius)
		return
	case state.PhaseWhiteDwarfRemnant:
		effectiveRadius = baseRadius * 0.15
		coreColor = [3]float64{0.85, 0.9, 1.0}
		coronaCol = [3]float64{0.6, 0.7, 0.9}
	}

	effectiveRadius *= 1.0 + math.Sin(g.Time*pulseSpeed)*pulseAmp

	// Density shells from Lane-Emden profile
	if g.Profile != nil && g.Phase != state.PhaseSupernova {
		numShells := min(len(g.Profile.Theta), 60)
		step := max(1, len(g.Profile.Theta)/numShells)
		for i := numShells - 1; i >= 0; i-- {
			idx := min(i*step, len(g.Profile.Theta)-1)
			r := (g.Profile.Xi[idx] / g.Profile.SurfaceXi) * effectiveRadius
			density := math.Max(0, g.Profile.Theta[idx])
			cr := coreColor[0]*density + coronaCol[0]*(1-density)
			cg := coreColor[1]*density + coronaCol[1]*(1-density)
			cb := coreColor[2]*density + coronaCol[2]*(1-density)
			alpha := (0.02 + density*0.12) * shellOp * g.Brightness
			col := color.RGBA{uint8(cr * 255), uint8(cg * 255), uint8(cb * 255), uint8(physics.Clamp(alpha*255, 0, 255))}
			vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(r), col, true)
		}
	}

	// Core glow
	for i := 0; i < 6; i++ {
		r := effectiveRadius * float64(6-i) / 6
		a := uint8(physics.Clamp(float64(i)*35*g.Brightness, 0, 255))
		col := color.RGBA{uint8(coreColor[0] * 255), uint8(coreColor[1] * 255), uint8(coreColor[2] * 255), a}
		vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(r), col, true)
	}

	// Corona layers
	for i := 0; i < 4; i++ {
		r := effectiveRadius * (1.2 + float64(i)*0.3)
		a := uint8(physics.Clamp((20-float64(i)*5)*g.Brightness, 0, 255))
		col := color.RGBA{uint8(coronaCol[0] * 255), uint8(coronaCol[1] * 255), uint8(coronaCol[2] * 255), a}
		vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(r), col, true)
	}

	// Pulsar beams
	if g.Phase == state.PhaseNeutronStar {
		angle := g.Time * 8
		for b := 0; b < 2; b++ {
			a := angle + float64(b)*math.Pi
			for d := 0.0; d < baseRadius*3; d += 2 {
				bx := cx + math.Cos(a)*d
				by := cy + math.Sin(a)*d
				alpha := uint8(math.Max(0, 220-d*0.8))
				spread := float32(1.0 + d*0.015)
				vector.DrawFilledCircle(screen, float32(bx), float32(by), spread, color.RGBA{100, 150, 255, alpha}, true)
			}
		}
	}

	// Element shell indicators (main sequence only)
	if g.Phase == state.PhaseMainSequence {
		for _, elKey := range physics.ElementOrder {
			if !g.Elements[elKey] {
				continue
			}
			el := physics.Elements[elKey]
			r := el.ShellRadius * effectiveRadius
			c := el.Color
			a := uint8(30 + 20*math.Sin(g.Time*1.5+el.ShellRadius*10))
			col := color.RGBA{uint8(c[0] * 255), uint8(c[1] * 255), uint8(c[2] * 255), a}
			vector.StrokeCircle(screen, float32(cx), float32(cy), float32(r), 1.5, col, true)
			la := el.ShellRadius*5 + g.Time*0.2
			lx := cx + math.Cos(la)*r
			ly := cy + math.Sin(la)*r
			ebitenutil.DebugPrintAt(screen, el.Symbol, int(lx)+4, int(ly)-6)
		}
	}

	// Supernova ejection particles
	if g.Phase == state.PhaseSupernova {
		g.SnovaRng.Seed(int64(g.Time * 1000))
		for i := 0; i < 100; i++ {
			angle := float64(i)/100*math.Pi*2 + g.Time*0.5
			speed := 0.5 + (math.Sin(float64(i)*137.5)*0.5+0.5)*2
			dist := effectiveRadius * (0.8 + g.CollapseProgress*speed*3)
			px := cx + math.Cos(angle)*dist
			py := cy + math.Sin(angle)*dist
			alpha := uint8(physics.Clamp((0.8-g.CollapseProgress*0.5)*255, 0, 255))
			hue := uint8(200 + g.SnovaRng.Intn(55))
			vector.DrawFilledCircle(screen, float32(px), float32(py), float32(1+g.SnovaRng.Float64()*2), color.RGBA{255, hue, 50, alpha}, true)
		}
	}
}

func drawBlackHole(screen *ebiten.Image, g *state.Game, cx, cy, baseRadius float64) {
	// Accretion disk
	for ring := 0; ring < 50; ring++ {
		r := baseRadius*0.3 + float64(ring)/50*baseRadius
		a := uint8(math.Max(0, 200-float64(ring)*4))
		rr := uint8(min(255, 200+ring))
		gg := uint8(min(255, 100+ring*3))
		vector.StrokeCircle(screen, float32(cx), float32(cy), float32(r), 2, color.RGBA{rr, gg, 50, a}, true)
	}
	// Event horizon
	vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(baseRadius*0.25), color.RGBA{0, 0, 0, 255}, true)
	// Lensing ring
	a := uint8(80 + 50*math.Sin(g.Time*3))
	vector.StrokeCircle(screen, float32(cx), float32(cy), float32(baseRadius*0.3), 2.5, color.RGBA{180, 140, 255, a}, true)
	// Photon sphere
	a2 := uint8(40 + 30*math.Sin(g.Time*5))
	vector.StrokeCircle(screen, float32(cx), float32(cy), float32(baseRadius*0.22), 1.5, color.RGBA{255, 200, 100, a2}, true)
}
