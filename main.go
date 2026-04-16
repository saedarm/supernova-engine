// Supernova Engine — Lane-Emden Stellar Structure Modeler
//
// A standalone Go/Ebitengine desktop application that numerically solves
// the Lane-Emden equation, visualizes stellar density profiles as layered
// shell renderings, simulates full stellar lifecycles, and plots the
// star's position on a Hertzsprung-Russell diagram in real time.
//
// Run:
//   go mod tidy && go run .
//
// Build:
//   go build -o supernova-engine .
package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/saedarm/supernova-engine/internal/physics"
	"github.com/saedarm/supernova-engine/internal/render"
	"github.com/saedarm/supernova-engine/internal/state"
)

// App implements ebiten.Game, wiring state updates to rendering.
type App struct {
	game   *state.Game
	mouseX int
	mouseY int
}

func (a *App) Update() error {
	a.mouseX, a.mouseY = ebiten.CursorPosition()
	g := a.game

	// ── Keyboard input ──
	for i, key := range physics.StarTypeOrder {
		if inpututil.IsKeyJustPressed(ebiten.Key1 + ebiten.Key(i)) {
			g.SelectStarType(key)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.IsAging = !g.IsAging
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.Reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.View = g.View.Next()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.ShowDensityPlot = !g.ShowDensityPlot
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		g.AgeSpeed = math.Min(g.AgeSpeed*2, 10)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		g.AgeSpeed = math.Max(g.AgeSpeed/2, 0.25)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
		g.Brightness = math.Min(g.Brightness+0.1, 2.0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
		g.Brightness = math.Max(g.Brightness-0.1, 0.1)
	}

	// Element toggles: Q W E A S D F
	elemKeys := []ebiten.Key{ebiten.KeyQ, ebiten.KeyW, ebiten.KeyE, ebiten.KeyA, ebiten.KeyS, ebiten.KeyD, ebiten.KeyF}
	for i, k := range elemKeys {
		if i < len(physics.ElementOrder) && inpututil.IsKeyJustPressed(k) {
			g.ToggleElement(physics.ElementOrder[i])
		}
	}

	// ── Mouse clicks on panel ──
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && a.mouseX < state.PanelWidth {
		a.handlePanelClick()
	}

	// ── Tick simulation ──
	g.Tick()

	return nil
}

func (a *App) handlePanelClick() {
	mx, my := a.mouseX, a.mouseY
	g := a.game

	// NOTE: These Y positions are approximate and correspond to the layout
	// in render.Panel(). If you change the panel layout, update these values.
	// A proper solution would be to track layout rects during drawing.

	// Star type rows (y starts ~82, each 18px)
	for i, key := range physics.StarTypeOrder {
		by := 82 + i*18
		if my >= by && my < by+16 && mx >= 8 && mx <= state.PanelWidth-8 {
			g.SelectStarType(key)
			return
		}
	}

	// Element buttons (y ~226, each 36px wide)
	for i, elKey := range physics.ElementOrder {
		bx := 10 + i*36
		if mx >= bx && mx < bx+32 && my >= 226 && my < 244 {
			g.ToggleElement(elKey)
			return
		}
	}

	// Play/Pause (~350)
	if my >= 350 && my < 368 {
		if mx >= 10 && mx < 80 {
			g.IsAging = !g.IsAging
			return
		}
		if mx >= 84 && mx < 140 {
			g.Reset()
			return
		}
	}

	// Speed buttons (~372)
	if my >= 372 && my < 390 {
		speeds := []float64{0.5, 1, 2, 5}
		for i, s := range speeds {
			bx := 10 + i*42
			if mx >= bx && mx < bx+38 {
				g.AgeSpeed = s
				return
			}
		}
	}

	// View mode buttons (~438)
	if my >= 438 && my < 456 {
		modes := []state.ViewMode{state.ViewSplit, state.ViewStarOnly, state.ViewHROnly}
		for i, m := range modes {
			bx := 10 + i*80
			if mx >= bx && mx < bx+76 {
				g.View = m
				return
			}
		}
	}
}

func (a *App) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{6, 6, 14, 255})
	g := a.game

	viewX := float64(state.PanelWidth)
	viewW := float64(state.ScreenWidth) - viewX
	viewH := float64(state.ScreenHeight)

	switch g.View {
	case state.ViewSplit:
		half := viewW / 2
		render.StarViewport(screen, g, viewX, 0, half, viewH)
		render.HRDiagram(screen, g, viewX+half, 0, half, viewH)
		ebitenutil.DrawRect(screen, viewX+half-0.5, 0, 1, viewH, color.RGBA{26, 26, 46, 255})
	case state.ViewStarOnly:
		render.StarViewport(screen, g, viewX, 0, viewW, viewH)
	case state.ViewHROnly:
		render.HRDiagram(screen, g, viewX, 0, viewW, viewH)
	}

	render.Panel(screen, g, a.mouseX, a.mouseY)
}

func (a *App) Layout(_, _ int) (int, int) {
	return state.ScreenWidth, state.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(state.ScreenWidth, state.ScreenHeight)
	ebiten.SetWindowTitle("Supernova Engine — Lane-Emden Stellar Structure Modeler")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)

	app := &App{game: state.New()}

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
