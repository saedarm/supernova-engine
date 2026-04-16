package physics

// ════════════════════════════════════════════════════════
// STAR TYPES
// ════════════════════════════════════════════════════════

// StarType defines a category of star with its physical properties,
// visual colors, and HR evolutionary track.
type StarType struct {
	Name, Description, Fate string
	Mass, PolyIndex         float64
	BaseTemp, MaxAge        float64
	BaseColor, CoronaColor  [3]float64
	DefaultElements         []string
	HRTrack                 []HRPoint
}

// StarTypes is the catalog of available star types.
var StarTypes = map[string]*StarType{
	"red_dwarf": {
		Name: "Red Dwarf (M-type)", Mass: 0.3, PolyIndex: 1.5, BaseTemp: 3200,
		BaseColor: [3]float64{1.0, 0.3, 0.1}, CoronaColor: [3]float64{0.8, 0.15, 0.05},
		Description: "Small, cool, long-lived. Fades to white dwarf.", Fate: "white_dwarf", MaxAge: 100,
		DefaultElements: []string{"hydrogen", "helium"},
		HRTrack: []HRPoint{
			{0, 3.505, -1.5}, {0.3, 3.51, -1.45}, {0.6, 3.52, -1.3},
			{0.7, 3.50, -1.2}, {0.85, 3.48, -0.8}, {0.95, 3.45, -0.5},
			{1.0, 3.55, -2.0}, {1.1, 3.60, -2.8}, {1.2, 3.55, -3.5},
		},
	},
	"sun_like": {
		Name: "Sun-like (G-type)", Mass: 1.0, PolyIndex: 3.0, BaseTemp: 5778,
		BaseColor: [3]float64{1.0, 0.95, 0.6}, CoronaColor: [3]float64{1.0, 0.7, 0.2},
		Description: "Medium star. Red giant -> white dwarf.", Fate: "white_dwarf", MaxAge: 10,
		DefaultElements: []string{"hydrogen", "helium", "carbon"},
		HRTrack: []HRPoint{
			{0, 3.76, 0.0}, {0.2, 3.76, 0.1}, {0.5, 3.77, 0.2},
			{0.65, 3.75, 0.4}, {0.7, 3.70, 0.8}, {0.75, 3.65, 1.5},
			{0.8, 3.58, 2.5}, {0.85, 3.55, 3.2}, {0.88, 3.56, 3.5},
			{0.9, 3.65, 2.0}, {0.92, 3.70, 1.8}, {0.95, 3.58, 3.6},
			{0.98, 3.65, 2.5}, {1.0, 4.0, -1.0}, {1.1, 3.95, -2.0}, {1.2, 3.85, -3.0},
		},
	},
	"blue_giant": {
		Name: "Blue Giant (O/B-type)", Mass: 15, PolyIndex: 3.0, BaseTemp: 25000,
		BaseColor: [3]float64{0.5, 0.7, 1.0}, CoronaColor: [3]float64{0.3, 0.4, 1.0},
		Description: "Massive, hot, short-lived. Supernova!", Fate: "neutron_star", MaxAge: 0.05,
		DefaultElements: []string{"hydrogen", "helium", "carbon", "oxygen", "silicon"},
		HRTrack: []HRPoint{
			{0, 4.40, 4.5}, {0.3, 4.38, 4.6}, {0.5, 4.35, 4.65},
			{0.65, 4.30, 4.7}, {0.7, 4.20, 4.8}, {0.78, 4.0, 5.0},
			{0.82, 3.80, 5.2}, {0.86, 3.65, 5.3}, {0.9, 3.75, 5.8},
			{0.95, 4.5, 7.0}, {0.98, 4.0, 5.0}, {1.0, 4.5, -1.0},
		},
	},
	"supergiant": {
		Name: "Red Supergiant", Mass: 20, PolyIndex: 3.0, BaseTemp: 3500,
		BaseColor: [3]float64{1.0, 0.4, 0.15}, CoronaColor: [3]float64{1.0, 0.2, 0.0},
		Description: "Iron core collapses -> black hole.", Fate: "black_hole", MaxAge: 0.03,
		DefaultElements: []string{"hydrogen", "helium", "carbon", "oxygen", "silicon", "iron"},
		HRTrack: []HRPoint{
			{0, 4.55, 5.0}, {0.2, 4.50, 5.1}, {0.4, 4.40, 5.15},
			{0.55, 4.20, 5.2}, {0.65, 3.90, 5.3}, {0.7, 3.65, 5.2},
			{0.75, 3.55, 5.15}, {0.8, 3.55, 5.3}, {0.85, 3.56, 5.4},
			{0.9, 3.58, 5.5}, {0.95, 4.5, 7.5}, {1.0, 0, 0},
		},
	},
	"white_dwarf": {
		Name: "White Dwarf", Mass: 0.6, PolyIndex: 1.5, BaseTemp: 12000,
		BaseColor: [3]float64{0.85, 0.9, 1.0}, CoronaColor: [3]float64{0.6, 0.7, 1.0},
		Description: "Stellar remnant. Cooling forever.", Fate: "black_dwarf", MaxAge: 1000,
		DefaultElements: []string{"carbon", "oxygen"},
		HRTrack: []HRPoint{
			{0, 4.10, -1.0}, {0.1, 4.05, -1.3}, {0.3, 3.95, -1.8},
			{0.5, 3.85, -2.3}, {0.7, 3.75, -2.8}, {0.9, 3.65, -3.3},
			{1.0, 3.55, -3.8}, {1.2, 3.45, -4.5},
		},
	},
	"wolf_rayet": {
		Name: "Wolf-Rayet Star", Mass: 25, PolyIndex: 3.0, BaseTemp: 50000,
		BaseColor: [3]float64{0.4, 0.5, 1.0}, CoronaColor: [3]float64{0.7, 0.3, 1.0},
		Description: "Extreme mass-loss. Doomed.", Fate: "black_hole", MaxAge: 0.01,
		DefaultElements: []string{"helium", "carbon", "nitrogen", "oxygen"},
		HRTrack: []HRPoint{
			{0, 4.70, 5.5}, {0.2, 4.65, 5.4}, {0.4, 4.60, 5.3},
			{0.6, 4.55, 5.2}, {0.7, 4.50, 5.3}, {0.8, 4.45, 5.5},
			{0.85, 4.40, 5.6}, {0.9, 4.50, 6.5}, {0.95, 4.6, 7.0}, {1.0, 0, 0},
		},
	},
}

// StarTypeOrder defines the display/key-binding order for star types.
var StarTypeOrder = []string{"red_dwarf", "sun_like", "blue_giant", "supergiant", "white_dwarf", "wolf_rayet"}

// ════════════════════════════════════════════════════════
// ELEMENTS
// ════════════════════════════════════════════════════════

// Element represents a chemical element present in stellar composition.
type Element struct {
	Symbol      string
	Color       [3]float64
	ShellRadius float64 // Fractional radius where this element's fusion shell sits
}

// Elements is the catalog of stellar composition elements.
var Elements = map[string]*Element{
	"hydrogen": {"H", [3]float64{0.2, 0.6, 1.0}, 0.9},
	"helium":   {"He", [3]float64{1.0, 0.95, 0.4}, 0.8},
	"carbon":   {"C", [3]float64{0.4, 0.4, 0.4}, 0.6},
	"nitrogen": {"N", [3]float64{0.2, 0.8, 0.4}, 0.55},
	"oxygen":   {"O", [3]float64{0.3, 0.9, 0.9}, 0.5},
	"silicon":  {"Si", [3]float64{0.8, 0.6, 0.3}, 0.3},
	"iron":     {"Fe", [3]float64{0.7, 0.2, 0.2}, 0.15},
}

// ElementOrder defines the display/key-binding order for elements.
var ElementOrder = []string{"hydrogen", "helium", "carbon", "nitrogen", "oxygen", "silicon", "iron"}

// ════════════════════════════════════════════════════════
// HR DIAGRAM REFERENCE DATA
// ════════════════════════════════════════════════════════

// RefStar is a well-known star plotted on the HR diagram for reference.
type RefStar struct {
	Name       string
	LogT, LogL float64
}

// ReferenceStars are plotted as background markers on the HR diagram.
var ReferenceStars = []RefStar{
	{"Rigel", 4.08, 4.8}, {"Betelgeuse", 3.55, 4.9}, {"Sirius A", 3.98, 1.4},
	{"Sun", 3.76, 0.0}, {"Proxima Cen", 3.49, -2.5}, {"Sirius B", 4.40, -1.7},
	{"Vega", 3.98, 1.7}, {"Aldebaran", 3.59, 2.1}, {"Deneb", 3.93, 5.1},
	{"Polaris", 3.76, 3.5}, {"Antares", 3.55, 4.5}, {"Spica", 4.35, 4.0},
}

// SpectralClass represents a spectral classification label on the HR x-axis.
type SpectralClass struct {
	Label string
	LogT  float64
}

// SpectralClasses are the standard Morgan-Keenan spectral types.
var SpectralClasses = []SpectralClass{
	{"O", 4.55}, {"B", 4.25}, {"A", 3.95}, {"F", 3.84}, {"G", 3.76}, {"K", 3.65}, {"M", 3.50},
}

// MainSequenceBand defines the center line of the main sequence for plotting.
var MainSequenceBand = [][2]float64{
	{4.65, 5.5}, {4.50, 4.5}, {4.30, 3.5}, {4.00, 2.0},
	{3.80, 0.5}, {3.76, 0.0}, {3.60, -1.0}, {3.50, -1.8}, {3.40, -2.5},
}
