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

	MapPeaks = generateMapMountainPeaks(rand)

	for y := range MapHeight {
		for x := range MapWidth {
			perlinAltitude, altitude := generateHeight(MapPeaks, x, y, seed)
			Map[y][x] = Tile{altitude: altitude}
			MapPerlin[y][x] = Tile{altitude: perlinAltitude}
		}
	}
}

func generateMapMountainPeaks(rand *rand.Rand) []Peak {
	peaks := make([]Peak, 0, 4)
	//MapPeaks = append(MapPeaks, Peak{peakX: 10, peakY: 0})

	// Keep mountains away from map edges to prevent them from being cut off
	edgeBuffer := 8

	// Calculate the usable area for peak placement
	minX, maxX := edgeBuffer, MapWidth-edgeBuffer
	minY, maxY := edgeBuffer, MapHeight-edgeBuffer

	// Create 3-4 mountain peaks that will form ranges

	// Place first peak
	firstX := minX + rand.Intn(maxX-minX)
	firstY := minY + rand.Intn(maxY-minY)
	peaks = append(peaks, Peak{peakX: firstX, peakY: firstY})

	// Place remaining peaks ensuring they have enough separation
	for i := 1; i < 4; i++ { // Try to place 3 more peaks
		// Make multiple attempts to find a suitable position
		for attempt := 0; attempt < 20; attempt++ {
			candidateX := minX + rand.Intn(maxX-minX)
			candidateY := minY + rand.Intn(maxY-minY)

			// Check minimum distance from existing peaks
			// Peaks need to be at least 20 tiles apart to prevent ranges from overlapping
			tooClose := false
			for _, peak := range peaks {
				dist := math.Sqrt(math.Pow(float64(candidateX-peak.peakX), 2) +
					math.Pow(float64(candidateY-peak.peakY), 2))

				// Minimum distance depends on range length
				minDistance := 20.0 // With 15-tile ranges, 20-tile separation prevents overlap
				if dist < minDistance {
					tooClose = true
					break
				}
			}

			if !tooClose {
				peaks = append(peaks, Peak{peakX: candidateX, peakY: candidateY})
				break
			}
		}
	}

	return peaks
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
	//perlin := perlin.NewPerlin(1.5, 4, 3, seed)
	perlin := perlin.NewPerlin(2.0, 2.5, 3, seed)

	// perlin.Noise2D generates 2-dimensional perlin noise value given an x and y.
	// Base terrain with multiple octaves to create natural variability
	// Using larger divisors to create broader terrain features
	baseHeight := perlin.Noise2D(float64(x)/(MapWidth*0.4), float64(y)/(MapHeight*0.4)) * 0.6
	baseHeight += perlin.Noise2D(float64(x)/(MapWidth*0.1), float64(y)/(MapHeight*0.1)) * 0.15
	baseHeight += 0.2 // Offset for water level

	// Mountain ranges - at regional scale we want elongated ranges, not isolated peaks, but keep is more controlled
	finalHeight := baseHeight
	for _, peak := range peaks {
		peakX := peak.peakX
		peakY := peak.peakY

		// Create directional bias for elongated ranges
		// Each peak gets a random direction for its range
		rangeDirection := (math.Mod(float64(peakX*peakY+int(seed)), 360)) * math.Pi / 180

		// Vector from peak to current point
		dirX := float64(x - peakX)
		dirY := float64(y - peakY)
		pointAngle := math.Atan2(dirY, dirX)

		// Calculate how aligned this point is with the mountain range direction
		// Use cosine of angle difference: 1 = perfectly aligned, 0 = perpendicular
		angleAlignment := math.Abs(math.Cos(pointAngle - rangeDirection))

		// Distance from current point to peak
		distance := math.Sqrt(math.Pow(dirX, 2) + math.Pow(dirY, 2))

		// Create extreme stretching factor:
		// - Along the range direction: very little distance penalty
		// - Perpendicular to range: very high distance penalty
		// This creates very narrow but long ranges

		// Start with a severe disparity between along-range and cross-range scaling
		// 0.2 = extremely narrow perpendicular to range
		// 6.0 = extends far along the range axis
		stretchMinimum := 0.2 // Controls width (smaller = narrower)
		stretchMaximum := 5.0 // Controls length (larger = longer)

		// Apply directional stretching - points along the range direction get reduced distance
		// Higher stretch values = narrower mountains perpendicular to range direction
		stretchFactor := stretchMinimum + (stretchMaximum-stretchMinimum)*math.Pow(angleAlignment, 2)

		// Apply stretch factor to create a modified distance
		// Points along the range direction will have effectively much shorter distances
		// Points perpendicular to range will have effectively much longer distances
		modifiedDistance := distance / stretchFactor

		// Hard distance cutoff for mountains - if beyond the range's influence, contribute nothing
		// This ensures ranges are exactly as long as we want them
		maxLengthwise := 15.0 // Maximum tiles along range direction
		maxCrosswise := 4.0   // Maximum tiles perpendicular to range direction

		// Calculate distance components along and perpendicular to range direction
		// This allows us to precisely control length and width
		alongRangeComponent := math.Abs(distance * math.Cos(pointAngle-rangeDirection))
		perpRangeComponent := math.Abs(distance * math.Sin(pointAngle-rangeDirection))

		// Only contribute to height if within our desired length and width bounds
		if alongRangeComponent <= maxLengthwise && perpRangeComponent <= maxCrosswise {
			// Very steep falloff perpendicular to range direction
			// Much gentler falloff along range direction
			sigmaPerpendicular := 1.2 // Controls width falloff (smaller = steeper sides)
			sigmaParallel := 6.0      // Controls length falloff (larger = more gradual along range)

			// Calculate height contribution with different falloff rates in different directions
			perpFactor := math.Exp(-math.Pow(perpRangeComponent, 2) / (2 * math.Pow(sigmaPerpendicular, 2)))
			parallelFactor := math.Exp(-math.Pow(alongRangeComponent, 2) / (2 * math.Pow(sigmaParallel, 2)))

			// Combine factors - both need to be high for maximum height
			heightFactor := perpFactor * parallelFactor

			// Apply some noise along the range for varied peaks
			heightVariation := perlin.Noise2D(float64(x+peakX)/10, float64(y+peakY)/10) * 0.3
			mountainHeight := 1.8 * heightFactor * (1.0 + heightVariation)

			// Apply height to the final terrain
			finalHeight += mountainHeight
		}
	}

	return baseHeight, finalHeight
}
