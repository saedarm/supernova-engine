// Package state manages the simulation state: current star selection,
// age, lifecycle phase, element composition, and HR trail history.
// This package has no rendering dependencies.
package state

import (
	"math/rand"

	"github.com/saedarm/supernova-engine/internal/physics"
)

// ════════════════════════════════════════════════════════
// PHASE
// ════════════════════════════════════════════════════════

type Phase int

const (
	PhaseMainSequence Phase = iota
	PhaseRedGiant
	PhaseSupernova
	PhaseNeutronStar
	PhaseBlackHole
	PhaseWhiteDwarfRemnant
)

var phaseNames = [...]string{
	"Main Sequence", "Red Giant", "Supernova!",
	"Neutron Star / Pulsar", "Black Hole", "White Dwarf Remnant",
}

var phaseIcons = [...]string{"*", "O", "!", "~", "@", "."}

func (p Phase) String() string {
	if int(p) < len(phaseNames) {
		return phaseNames[p]
	}
	return "Unknown"
}

func (p Phase) Icon() string {
	if int(p) < len(phaseIcons) {
		return phaseIcons[p]
	}
	return "?"
}

// ════════════════════════════════════════════════════════
// VIEW MODE
// ════════════════════════════════════════════════════════

type ViewMode int

const (
	ViewSplit ViewMode = iota
	ViewStarOnly
	ViewHROnly
)

func (v ViewMode) Next() ViewMode { return (v + 1) % 3 }

// ════════════════════════════════════════════════════════
// GAME STATE
// ════════════════════════════════════════════════════════

// Game holds all mutable simulation state.
type Game struct {
	StarTypeKey      string
	Elements         map[string]bool
	Age              float64
	IsAging          bool
	AgeSpeed         float64
	Brightness       float64
	Phase            Phase
	CollapseProgress float64
	Time             float64
	Profile          *physics.DensityProfile
	View             ViewMode
	ShowDensityPlot  bool

	// HR trail history
	HRTrail     []physics.TrailPoint
	MaxTrailLen int

	// Background stars (static, deterministic)
	BgStars [][3]float64

	// Supernova RNG
	SnovaRng *rand.Rand
}

const (
	ScreenWidth  = 1400
	ScreenHeight = 850
	PanelWidth   = 290
)

// New creates a new Game with default state (sun-like star).
func New() *Game {
	g := &Game{
		StarTypeKey: "sun_like",
		Elements:    make(map[string]bool),
		AgeSpeed:    1.0,
		Brightness:  1.0,
		View:        ViewSplit,
		MaxTrailLen: 300,
		SnovaRng:    rand.New(rand.NewSource(99)),
	}
	for _, el := range physics.StarTypes["sun_like"].DefaultElements {
		g.Elements[el] = true
	}
	// Deterministic background starfield
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 400; i++ {
		g.BgStars = append(g.BgStars, [3]float64{
			r.Float64() * float64(ScreenWidth),
			r.Float64() * float64(ScreenHeight),
			r.Float64()*0.6 + 0.1,
		})
	}
	g.RecalcProfile()
	return g
}

// ST returns the current StarType.
func (g *Game) ST() *physics.StarType {
	return physics.StarTypes[g.StarTypeKey]
}

// AgeFrac returns the current age as a fraction of max age.
func (g *Game) AgeFrac() float64 {
	return g.Age / g.ST().MaxAge
}

// RecalcProfile re-solves the Lane-Emden equation for the current star type.
func (g *Game) RecalcProfile() {
	g.Profile = physics.SolveLaneEmden(g.ST().PolyIndex, 2000)
}

// SelectStarType switches to a new star type, resetting age and elements.
func (g *Game) SelectStarType(key string) {
	g.StarTypeKey = key
	g.Elements = make(map[string]bool)
	for _, el := range physics.StarTypes[key].DefaultElements {
		g.Elements[el] = true
	}
	g.Reset()
	g.RecalcProfile()
}

// Reset sets age back to zero and clears the HR trail.
func (g *Game) Reset() {
	g.Age = 0
	g.IsAging = false
	g.Phase = PhaseMainSequence
	g.CollapseProgress = 0
	g.HRTrail = nil
}

// ToggleElement flips an element on or off.
func (g *Game) ToggleElement(key string) {
	g.Elements[key] = !g.Elements[key]
}

// Tick advances the simulation by one frame (called at 60fps).
func (g *Game) Tick() {
	g.Time += 1.0 / 60.0
	st := g.ST()

	// Aging
	if g.IsAging {
		g.Age += st.MaxAge * 0.002 * g.AgeSpeed * (1.0 / 60.0)
		if g.Age > st.MaxAge*1.2 {
			g.Age = st.MaxAge * 1.2
			g.IsAging = false
		}
		// Record trail point
		logT, logL := physics.InterpolateHR(st.HRTrack, g.AgeFrac())
		g.HRTrail = append(g.HRTrail, physics.TrailPoint{LogT: logT, LogL: logL})
		if len(g.HRTrail) > g.MaxTrailLen {
			g.HRTrail = g.HRTrail[len(g.HRTrail)-g.MaxTrailLen:]
		}
	}

	// Determine lifecycle phase
	af := g.AgeFrac()
	switch {
	case af < 0.7:
		g.Phase = PhaseMainSequence
		g.CollapseProgress = 0
	case af < 0.9:
		g.Phase = PhaseRedGiant
		g.CollapseProgress = 0
	case af < 1.0:
		g.Phase = PhaseSupernova
		g.CollapseProgress = (af - 0.9) / 0.1
	default:
		g.CollapseProgress = 1
		switch st.Fate {
		case "black_hole":
			g.Phase = PhaseBlackHole
		case "neutron_star":
			g.Phase = PhaseNeutronStar
		case "white_dwarf":
			g.Phase = PhaseWhiteDwarfRemnant
		}
	}
}

// RecordTrailManual adds a trail point (used when scrubbing the age slider).
func (g *Game) RecordTrailManual() {
	logT, logL := physics.InterpolateHR(g.ST().HRTrack, g.AgeFrac())
	g.HRTrail = append(g.HRTrail, physics.TrailPoint{LogT: logT, LogL: logL})
	if len(g.HRTrail) > g.MaxTrailLen {
		g.HRTrail = g.HRTrail[len(g.HRTrail)-g.MaxTrailLen:]
	}
}

// HRPosition returns the current HR diagram coordinates.
func (g *Game) HRPosition() (logT, logL float64) {
	return physics.InterpolateHR(g.ST().HRTrack, g.AgeFrac())
}
