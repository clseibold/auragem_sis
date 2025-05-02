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
	// Base terrain with multiple octaves to create natural variability
	// Using larger divisors to represent continental-scale features
	baseHeight := perlin.Noise2D(float64(x)/(MapWidth*0.7), float64(y)/(MapHeight*0.7)) * 0.5
	baseHeight += perlin.Noise2D(float64(x)/(MapWidth*0.3), float64(y)/(MapHeight*0.3)) * 0.3
	baseHeight += perlin.Noise2D(float64(x)/(MapWidth*0.1), float64(y)/(MapHeight*0.1)) * 0.2
	baseHeight += 0.2 // Baseline elevation adjustment

	// Mountain ranges - at regional scale we want elongated ranges, not isolated peaks
	finalHeight := baseHeight
	for _, peak := range peaks {
		peakX := peak.peakX
		peakY := peak.peakY

		// Calculate distance to the peak (mountain range center)
		distance := math.Sqrt(math.Pow(float64(x-peakX), 2) + math.Pow(float64(y-peakY), 2))

		// Create directional bias for the mountain range
		// This makes ranges extend in a specific direction rather than being circular
		dirX := float64(x - peakX)
		dirY := float64(y - peakY)
		angle := math.Atan2(dirY, dirX)

		// Generate a random dominant direction for this mountain range
		// Each range will extend in a different main direction
		rangeDirection := math.Pi * float64(int(seed+int64(peakX*peakY))%4) / 4.0

		// Adjust distance based on alignment with the range's main axis
		// Points along the mountain chain's axis will have effectively "shorter" distances
		directionAlignment := math.Abs(math.Cos(angle - rangeDirection))
		stretchFactor := 0.4 + 1.2*directionAlignment // Range from 0.4 to 1.6

		// Stretch the distance in the perpendicular direction to create elongated ranges
		modifiedDistance := distance * (1.0 / stretchFactor)

		// Regional mountains have gentler slopes
		sigma := 7.0 // Much wider influence for regional scale
		mountainHeight := 2.0 * math.Exp(-math.Pow(modifiedDistance, 1.8)/(2*math.Pow(sigma, 2)))

		// Apply elevation variability along the range
		rangeVariability := perlin.Noise2D(float64(x+peakX)/20, float64(y+peakY)/20) * 0.3
		mountainHeight *= (1.0 + rangeVariability)

		finalHeight += mountainHeight
	}

	return baseHeight, finalHeight
}
