package textgame2

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/aquilax/go-perlin"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

// TODO: Generate Valleys, Plateaus, and Rivers
// TODO: Assign land types to each tile. Then assign biomes to each tile given its land type, adjacent biomes, and bodies of water

const MapWidth = 50
const MapHeight = 50

// const MapNumberOfMountainPeaks = 3

var MapPeaks []Peak
var Map [MapHeight][MapWidth]Tile
var MapPerlin [MapHeight][MapWidth]Tile

type Tile struct {
	altitude float64
	biome    Biome
	landType LandType
}

type Peak struct {
	peakX int
	peakY int
}

func generateWorldMap() {
	var seed int64 = 1239462936493264926
	rand := rand.New(rand.NewSource(seed))

	MapPeaks = make([]Peak, 0, 4)
	MapPeaks = append(MapPeaks, Peak{peakX: 10, peakY: 0})

	// Generate random peaks in four quadrants of map

	MapPeaks = append(MapPeaks, Peak{rand.Intn(MapWidth / 2), rand.Intn(MapHeight / 2)})
	MapPeaks = append(MapPeaks, Peak{rand.Intn(MapWidth/2) - 1 + MapWidth/2, rand.Intn(MapHeight / 2)})
	MapPeaks = append(MapPeaks, Peak{rand.Intn(MapWidth / 2), rand.Intn(MapHeight/2) - 1 + MapHeight/2})
	MapPeaks = append(MapPeaks, Peak{rand.Intn(MapWidth/2) - 1 + MapWidth/2, rand.Intn(MapHeight/2) - 1 + MapHeight/2})

	for y := range MapHeight {
		for x := range MapWidth {
			perlinAltitude, altitude := generateHeight(MapPeaks, x, y, seed)
			Map[y][x] = Tile{altitude: altitude}
			MapPerlin[y][x] = Tile{altitude: perlinAltitude}
		}
	}
}

func PrintWorldMap(request *sis.Request) {
	showValues := false
	query, _ := request.Query()
	if query == "values" {
		showValues = true
	}

	request.Heading(1, "World Map")
	request.Gemini("\n")
	if !showValues {
		request.Link("/world-map?values", "Show Values")
	} else {
		request.Link("/world-map", "Show Terrain")
	}
	request.Gemini("\n")

	request.Gemini("```\n")
	// Print the peaks
	request.PlainText("Peaks: ")
	for _, peak := range MapPeaks {
		request.PlainText("(%d, %d) ", peak.peakX, peak.peakY)
	}
	request.PlainText("\n")

	// Print the lowest and highest tiles
	request.PlainText("\nLowest and Highest Altitudes on Map with Mountain Peaks:\n")
	lowest, highest := getMapLowestAndHighestPoints()
	request.PlainText("Lowest Tile Altitude: %+.2f\n", lowest.altitude)
	request.PlainText("Highest Tile Altitude: %+.2f\n", highest.altitude)

	request.PlainText("\nJust Perlin Noise:\n")
	for y := range MapHeight {
		// Heading
		if y == 0 {
			if showValues {
				request.PlainText("|     |")
			} else {
				request.PlainText("|  |")
			}
			for x := range MapWidth {
				if showValues {
					request.PlainText(fmt.Sprintf("%5d|", x))
				} else {
					request.PlainText(fmt.Sprintf("%2d|", x))
				}
			}
			request.PlainText("\n")
			if showValues {
				request.PlainText("\n")
			}
		}

		// Values
		if showValues {
			request.PlainText("|%5d|", y)
		} else {
			request.PlainText("|%2d|", y)
		}
		for x := range MapWidth {
			if showValues {
				request.PlainText(fmt.Sprintf("%+.2f|", MapPerlin[y][x].altitude))
			} else {
				altitude := MapPerlin[y][x].altitude
				if altitude <= 0 {
					request.PlainText(" o|") // Water
				} else if altitude >= 1 {
					request.PlainText(" 5|") // Mountain
				} else {
					request.PlainText(" 1|") // Plains and hills
				}
			}
		}
		request.PlainText("\n")
		if showValues {
			request.PlainText("\n")
		}
	}

	request.PlainText("\nWith Mountain Peaks:\n")
	for y := range MapHeight {
		// Heading
		if y == 0 {
			if showValues {
				request.PlainText("|     |")
			} else {
				request.PlainText("|  |")
			}
			for x := range MapWidth {
				if showValues {
					request.PlainText("%5d|", x)
				} else {
					request.PlainText("%2d|", x)
				}
			}
			request.PlainText("\n")
			if showValues {
				request.PlainText("\n")
			}
		}

		if showValues {
			request.PlainText("|%5d|", y)
		} else {
			request.PlainText("|%2d|", y)
		}
		for x := range MapWidth {
			if showValues {
				request.PlainText(fmt.Sprintf("%+.2f|", Map[y][x].altitude))
			} else {
				altitude := Map[y][x].altitude
				if altitude <= 0 {
					request.PlainText(" o|") // Water
				} else if altitude >= 1 {
					request.PlainText(" 5|") // Mountain
				} else {
					request.PlainText(" 1|") // Plains and hills
				}
			}
		}
		request.PlainText("\n")
		if showValues {
			request.PlainText("\n")
		}
	}
	request.Gemini("```\n")
}

func getMapLowestAndHighestPoints() (Tile, Tile) {
	var lowest Tile
	var highest Tile
	for y := range MapHeight {
		for x := range MapWidth {
			if Map[y][x].altitude < lowest.altitude {
				lowest = Map[y][x]
			}
			if Map[y][x].altitude > highest.altitude {
				highest = Map[y][x]
			}
		}
	}

	return lowest, highest
}

func generateHeight(peaks []Peak, x int, y int, seed int64) (float64, float64) {
	// perlin.NewPerlin(alpha, beta, n, seed)
	// NewPerlin creates new Perlin noise generator In what follows “alpha” is the weight when the sum is formed. Typically it is 2, As this approaches 1 the function is noisier. “beta” is the harmonic scaling/spacing, typically 2, n is the number of iterations and seed is the math.rand seed value to use.
	perlin := perlin.NewPerlin(1.5, 4, 3, seed)

	// perlin.Noise2D generates 2-dimensional perlin noise value given an x and y.
	baseHeight := perlin.Noise2D(float64(x)/MapWidth, float64(y)/MapHeight) * 0.6
	baseHeight += perlin.Noise2D(float64(x)/(MapWidth/2), float64(y)/(MapHeight/2)) * 0.3
	baseHeight += perlin.Noise2D(float64(x)/(MapWidth/4), float64(y)/(MapHeight/4)) * 0.1
	baseHeight += 0.2 // Offset to control water level

	// Create a mountain peak effect
	finalHeight := baseHeight
	for _, peak := range peaks {
		peakX := peak.peakX
		peakY := peak.peakY

		// Calculate the distance from the current point to the peak
		distance := math.Sqrt(math.Pow(float64(x-peakX), 2) + math.Pow(float64(y-peakY), 2))

		// Create ridge effects by applying directional bias
		// Create more elongated mountain ranges instead of just circular peaks
		dirX := float64(x - peakX)
		dirY := float64(y - peakY)
		angle := math.Atan2(dirY, dirX)

		// Use ridge noise to create elongated mountain chains
		ridgeEffect := math.Sin(angle*2) * 0.3 // Controls the direction of ridges
		modifiedDistance := distance * (1.0 - ridgeEffect)

		// Use a more dramatic height formula for mountains
		sigma := 0.75 // Base width
		mountainHeight := 1.8 * math.Exp(-math.Pow(modifiedDistance, 2)/(2*math.Pow(sigma, 2)))

		// Apply some noise to the mountain to make it less uniform
		mountainNoise := perlin.Noise2D(float64(x+peakX)/30, float64(y+peakY)/30) * 0.2
		mountainHeight *= (1.0 + mountainNoise)

		finalHeight += mountainHeight
	}

	return baseHeight, finalHeight
}
