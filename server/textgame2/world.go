package textgame2

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/aquilax/go-perlin"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

const MapWidth = 50
const MapHeight = 50
const MapNumberOfMountainPeaks = 3

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
	for range MapNumberOfMountainPeaks - 1 {
		peakX := rand.Intn(MapWidth)
		peakY := rand.Intn(MapHeight)

		MapPeaks = append(MapPeaks, Peak{peakX, peakY})
	}

	for y := range MapHeight {
		for x := range MapWidth {
			perlinAltitude, altitude := generateHeight(MapPeaks, x, y, seed)
			Map[y][x] = Tile{altitude: altitude}
			MapPerlin[y][x] = Tile{altitude: perlinAltitude}
		}
	}
}

func PrintWorldMap(request *sis.Request) {
	request.Gemini("```\n")
	// Print the peaks
	for _, peak := range MapPeaks {
		request.PlainText("(%d, %d) ", peak.peakX, peak.peakY)
	}
	request.PlainText("\n\nJust Perlin Noise:\n")
	for y := range MapHeight {
		// Heading
		if y == 0 {
			request.PlainText("|")
			for x := range MapWidth {
				request.PlainText(fmt.Sprintf("%3d|", x))
			}
			request.PlainText("\n")
		}

		// Values
		request.PlainText("|%3d|", y)
		for x := range MapWidth {
			request.PlainText(fmt.Sprintf("%.2f|", MapPerlin[y][x].altitude))
		}
		request.PlainText("\n\n")
	}
	request.PlainText("\nWith Mountain Peaks:\n")
	for y := range MapHeight {
		request.PlainText("|")
		for x := range MapWidth {
			request.PlainText(fmt.Sprintf("%.2f|", Map[y][x].altitude))
		}
		request.PlainText("\n\n")
	}
	request.Gemini("```\n")
}

func generateHeight(peaks []Peak, x int, y int, seed int64) (float64, float64) {
	perlin := perlin.NewPerlin(2, 2, 3, seed)

	baseHeight := perlin.Noise2D(float64(x)/MapWidth, float64(y)/MapHeight)
	heightFactor := float64(2.0)
	height := baseHeight * heightFactor

	// Create a mountain peak effect
	finalHeight := height
	var sigma float64 = 1 // Width of peak, larger means mountains affect more points farther away from the peak
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
