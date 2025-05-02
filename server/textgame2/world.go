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
	rand := rand.New(rand.NewSource(1))

	MapPeaks = make([]Peak, 0)
	for range MapNumberOfMountainPeaks {
		peakX := rand.Intn(MapWidth)
		peakY := rand.Intn(MapHeight)

		peaks = append(peaks, Peak{peakX, peakY})
	}

	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			altitude := generateHeight(peaks, x, y, 1)
			Map[y][x] = Tile{altitude: altitude}
		}
	}
}

func PrintWorldMap(request *sis.Request) {
	request.Gemini("```\n")
	// Print the peaks
	for _, peak := range MapPeaks {
		request.PlainText("(%d, %d) ", peak.peakX, peak.peakY)
	}
	request.PlainText("\n\n")
	for y := 0; y < MapHeight; y++ {
		request.Gemini("|")
		for x := 0; x < MapWidth; x++ {
			request.PlainText(fmt.Sprintf("%.2f|", Map[y][x].altitude))
		}
		request.PlainText("\n")
	}
	request.Gemini("```\n")
}

func generateHeight(peaks []Peak, x int, y int, seed int64) float64 {
	perlin := perlin.NewPerlin(2, 2, 3, seed)

	baseHeight := (perlin.Noise2D(float64(x), float64(y)) + 1) * 0.5 // Scale to [0, 1]
	heightFactor := float64(1)
	height := baseHeight * heightFactor

	// Create a mountain peak effect
	finalHeight := height
	var sigma float64 = 0.5 // Width of peak
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

	return finalHeight
}
