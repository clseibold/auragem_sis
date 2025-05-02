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

	generateMapMountainPeaks(rand)

	for y := range MapHeight {
		for x := range MapWidth {
			perlinAltitude, altitude := generateHeight(MapPeaks, x, y, seed)
			Map[y][x] = Tile{altitude: altitude}
			MapPerlin[y][x] = Tile{altitude: perlinAltitude}
		}
	}
}

func generateMapMountainPeaks(rand *rand.Rand) {
	MapPeaks = make([]Peak, 0, 4)
	//MapPeaks = append(MapPeaks, Peak{peakX: 10, peakY: 0})

	// Keep mountains away from map edges to prevent them from being cut off
	edgeBuffer := 5

	// Create 3-4 mountain peaks with more spacing between them
	// First peak in upper left quadrant
	MapPeaks = append(MapPeaks, Peak{
		peakX: edgeBuffer + rand.Intn(MapWidth/4),
		peakY: edgeBuffer + rand.Intn(MapHeight/4),
	})

	// Second peak in lower right quadrant
	MapPeaks = append(MapPeaks, Peak{
		peakX: MapWidth/2 + rand.Intn(MapWidth/4),
		peakY: MapHeight/2 + rand.Intn(MapHeight/4),
	})

	// Third peak - ensure it's far enough from existing peaks
	for attempts := 0; attempts < 10; attempts++ {
		candidateX := edgeBuffer + rand.Intn(MapWidth-2*edgeBuffer)
		candidateY := edgeBuffer + rand.Intn(MapHeight-2*edgeBuffer)

		// Check distance to existing peaks
		tooClose := false
		for _, peak := range MapPeaks {
			dist := math.Sqrt(math.Pow(float64(candidateX-peak.peakX), 2) +
				math.Pow(float64(candidateY-peak.peakY), 2))
			if dist < 15 { // Ensure peaks are well-separated
				tooClose = true
				break
			}
		}

		if !tooClose {
			MapPeaks = append(MapPeaks, Peak{peakX: candidateX, peakY: candidateY})
			break
		}
	}

	// Occasionally add a fourth peak for variety (25% chance)
	if rand.Float64() < 0.25 {
		for attempts := 0; attempts < 10; attempts++ {
			candidateX := edgeBuffer + rand.Intn(MapWidth-2*edgeBuffer)
			candidateY := edgeBuffer + rand.Intn(MapHeight-2*edgeBuffer)

			// Check distance
			tooClose := false
			for _, peak := range MapPeaks {
				dist := math.Sqrt(math.Pow(float64(candidateX-peak.peakX), 2) +
					math.Pow(float64(candidateY-peak.peakY), 2))
				if dist < 18 {
					tooClose = true
					break
				}
			}

			if !tooClose {
				MapPeaks = append(MapPeaks, Peak{peakX: candidateX, peakY: candidateY})
				break
			}
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
	// Using larger divisors to create broader terrain features
	baseHeight := perlin.Noise2D(float64(x)/(MapWidth*0.4), float64(y)/(MapHeight*0.4)) * 0.6
	baseHeight += perlin.Noise2D(float64(x)/(MapWidth*0.1), float64(y)/(MapHeight*0.1)) * 0.15
	baseHeight += 0.2 // Offset for water level

	// Mountain ranges - at regional scale we want elongated ranges, not isolated peaks, but keep is more controlled
	finalHeight := baseHeight
	for _, peak := range peaks {
		peakX := peak.peakX
		peakY := peak.peakY

		// Calculate distance to the peak (mountain range center)
		distance := math.Sqrt(math.Pow(float64(x-peakX), 2) + math.Pow(float64(y-peakY), 2))

		// Only apply mountain effects within a reasonable radius.
		// This ensures mountains don't cover too much of the map, only about 10-12%
		maxMountainInfluence := 9.0 // Tiles from peak

		if distance < maxMountainInfluence {
			// Create directional bias for elongated ranges
			// Each peak gets a random direction for its range
			rangeDirection := (math.Mod(float64(peakX*peakY+int(seed)), 360)) * math.Pi / 180

			// Vector from peak to current point
			dirX := float64(x - peakX)
			dirY := float64(y - peakY)
			angle := math.Atan2(dirY, dirX)

			// Calculate how aligned this point is with the mountain range direction
			// Use cosine of angle difference: 1 = perfectly aligned, 0 = perpendicular
			angleAlignment := math.Abs(math.Cos(angle - rangeDirection))

			// Apply directional stretching - points along the range direction get reduced distance
			// Higher stretch values = narrower mountains perpendicular to range direction
			stretchFactor := 0.42 + 2.53*angleAlignment
			modifiedDistance := distance / stretchFactor

			// Create more compact mountain ranges, a steeper falloff, with sharper and narrower peaks
			sigma := 2.7 // Controls mountain width

			// Height falloff based on distance from peak
			// Switch exponent of modifiedDistance from 2.0 to 1.8 for more aggressive falloff for narrower mountains
			heightFactor := math.Exp(-math.Pow(modifiedDistance, 1.8) / (2 * math.Pow(sigma, 2)))

			// Scale height by distance from peak with some noise, but make it droppoff more quickly
			heightVariation := perlin.Noise2D(float64(x+peakX)/12, float64(y+peakY)/12) * 0.2

			// Height factor with sharp cutoff for more compact mountains
			// Only add significant height when heightFactor is substantial
			if heightFactor > 0.2 {
				mountainHeight := 1.7 * heightFactor * (1.0 + heightVariation)
				finalHeight += mountainHeight
			} else {
				// Add minimal height for foothills
				mountainHeight := 0.15 * heightFactor * (1.0 + heightVariation)
				finalHeight += mountainHeight
			}
		}
	}

	return baseHeight, finalHeight
}
