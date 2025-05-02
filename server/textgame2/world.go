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
				request.PlainText("|   |")
			}
			for x := range MapWidth {
				if showValues {
					request.PlainText(fmt.Sprintf("%5d|", x))
				} else {
					request.PlainText(fmt.Sprintf("%3d|", x))
				}
			}
			request.PlainText("\n\n")
		}

		// Values
		if showValues {
			request.PlainText("|%5d|", y)
		} else {
			request.PlainText("|%3d|", y)
		}
		for x := range MapWidth {
			if showValues {
				request.PlainText(fmt.Sprintf("%+.2f|", MapPerlin[y][x].altitude))
			} else {
				altitude := MapPerlin[y][x].altitude
				if altitude <= 0 {
					request.PlainText(" o |") // Water
				} else if altitude >= 1 {
					request.PlainText(" M |") // Mountain
				} else {
					request.PlainText(" I |") // Plains and hills
				}
			}
		}
		request.PlainText("\n\n")
	}

	request.PlainText("\nWith Mountain Peaks:\n")
	for y := range MapHeight {
		// Heading
		if y == 0 {
			if showValues {
				request.PlainText("|     |")
			} else {
				request.PlainText("|   |")
			}
			for x := range MapWidth {
				if showValues {
					request.PlainText("%5d|", x)
				} else {
					request.PlainText("%3d|", x)
				}
			}
			request.PlainText("\n\n")
		}

		if showValues {
			request.PlainText("|%5d|", y)
		} else {
			request.PlainText("|%3d|", y)
		}
		for x := range MapWidth {
			if showValues {
				request.PlainText(fmt.Sprintf("%+.2f|", Map[y][x].altitude))
			} else {
				altitude := Map[y][x].altitude
				if altitude <= 0 {
					request.PlainText(" o |") // Water
				} else if altitude >= 1 {
					request.PlainText(" M |") // Mountain
				} else {
					request.PlainText(" I |") // Plains and hills
				}
			}
		}
		request.PlainText("\n\n")
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
	perlin := perlin.NewPerlin(2, 2, 3, seed)

	baseHeight := perlin.Noise2D(float64(x)/MapWidth, float64(y)/MapHeight) + 0.04
	heightFactor := float64(2.0)
	height := baseHeight * heightFactor

	// Create a mountain peak effect
	finalHeight := height
	var sigma float64 = 0.5 // Width of peak, larger means mountains affect more points farther away from the peak
	for _, peak := range peaks {
		peakX := peak.peakX
		peakY := peak.peakY

		// Calculate the distance from the current point to the peak
		distance := math.Sqrt(math.Pow(float64(x-peakX), 2) + math.Pow(float64(y-peakY), 2))

		// Calculate the mountain contribution using a Gaussian function
		mountainHeight := math.Exp(-math.Pow(distance, 2) / (2 * math.Pow(sigma, 2)))

		// Add the mountain contribution to the final height
		finalHeight += mountainHeight
	}

	return baseHeight, finalHeight
}
