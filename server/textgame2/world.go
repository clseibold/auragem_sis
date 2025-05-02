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

// Each tile of the world map represents a 10 square kilometer region.
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

	generatePlateaus(seed)
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
	} else if query == "mountains" {
		debugMountainDimensions(request)
		return
	}

	request.Heading(1, "World Map")
	request.Gemini("\n")
	if !showValues {
		request.Link("/world-map?values", "Show Values")
		request.Link("/world-map?mountains", "Show Mountain Ranges")
	} else {
		request.Link("/world-map", "Show Terrain")
		request.Link("/world-map?mountains", "Show Mountain Ranges")
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

	request.PlainText("\nWith Full Terrain Generation:\n")
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
					request.PlainText(" ~|") // Water
				} else if altitude >= 1 {
					request.PlainText(" A|") // Mountain
				} else if altitude >= 0.5 { // Hills and plateaus
					if Map[y][x].landType == LandType_Plateaus {
						request.PlainText(" =|")
					} else {
						request.PlainText(" n|")
					}
				} else if altitude >= 0.3 { // Hills
					request.PlainText(" +|")
				} else {
					request.PlainText("  |") // Plains
				}
			}
		}
		request.PlainText("\n")
		if showValues {
			request.PlainText("\n")
		}
	}

	request.PlainText("\nBase Perlin Noise:\n")
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
					request.PlainText(" ~|") // Water
				} else if altitude >= 1 {
					request.PlainText(" A|") // Mountain
				} else if altitude >= 0.5 { // Hills?
					if Map[y][x].landType == LandType_Plateaus {
						request.PlainText(" =|")
					} else {
						request.PlainText(" n|")
					}
				} else if altitude >= 0.3 {
					request.PlainText(" +|")
				} else {
					request.PlainText("  |") // Plains
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

func debugMountainDimensions(request *sis.Request) {
	request.Heading(1, "Mountain Range Dimensions")
	request.Gemini("\n")
	request.Link("/world-map/", "Back to World Map")
	request.Gemini("\n")

	// Create a debug grid
	debugMap := make([][]string, MapHeight)
	for y := range debugMap {
		debugMap[y] = make([]string, MapWidth)
		for x := range debugMap[y] {
			debugMap[y][x] = " "
		}
	}

	// Mark mountain range areas
	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			if Map[y][x].altitude >= 1.0 {
				debugMap[y][x] = "M"
			}
		}
	}

	// Mark mountain peaks
	for _, peak := range MapPeaks {
		debugMap[peak.peakY][peak.peakX] = "P"
	}

	// Show theoretical stretch zones for one peak
	if len(MapPeaks) > 0 {
		peak := MapPeaks[0]
		rangeDirection := (math.Mod(float64(peak.peakX*peak.peakY+1239462936493264926), 360)) * math.Pi / 180

		// Draw direction line
		lineLength := 20
		for i := 0; i < lineLength; i++ {
			dx := int(math.Round(float64(i) * math.Cos(rangeDirection)))
			dy := int(math.Round(float64(i) * math.Sin(rangeDirection)))

			nx, ny := peak.peakX+dx, peak.peakY+dy
			if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
				if debugMap[ny][nx] == " " {
					debugMap[ny][nx] = "."
				}
			}
		}

		// Show stretch factor zones
		for y := 0; y < MapHeight; y++ {
			for x := 0; x < MapWidth; x++ {
				// Skip if already marked
				if debugMap[y][x] != " " {
					continue
				}

				// Vector from peak
				dirX := float64(x - peak.peakX)
				dirY := float64(y - peak.peakY)

				// Skip if too far
				dist := math.Sqrt(math.Pow(dirX, 2) + math.Pow(dirY, 2))
				if dist > 20 {
					continue
				}

				// Calculate angle alignment
				pointAngle := math.Atan2(dirY, dirX)
				angleAlignment := math.Abs(math.Cos(pointAngle - rangeDirection))

				// Show different stretch zones
				if angleAlignment > 0.9 {
					debugMap[y][x] = "S" // Strong stretch
				} else if angleAlignment > 0.7 {
					debugMap[y][x] = "s" // Medium stretch
				} else if angleAlignment > 0.3 {
					debugMap[y][x] = "w" // Weak stretch
				}
			}
		}

		// Show dimensional constraints
		maxLengthwise := 15.0
		maxCrosswise := 2.5

		for y := 0; y < MapHeight; y++ {
			for x := 0; x < MapWidth; x++ {
				// Vector from peak
				dirX := float64(x - peak.peakX)
				dirY := float64(y - peak.peakY)

				// Calculate rotated coordinates
				alignedX := dirX*math.Cos(-rangeDirection) + dirY*math.Sin(-rangeDirection)
				alignedY := -dirX*math.Sin(-rangeDirection) + dirY*math.Cos(-rangeDirection)

				// Use absolute values
				lengthwiseDistance := math.Abs(alignedX)
				crosswiseDistance := math.Abs(alignedY)

				// Mark boundary points
				if (math.Abs(lengthwiseDistance-maxLengthwise) < 0.5 && crosswiseDistance <= maxCrosswise) ||
					(math.Abs(crosswiseDistance-maxCrosswise) < 0.5 && lengthwiseDistance <= maxLengthwise) {
					debugMap[y][x] = "+"
				}
			}
		}
	}

	// Print the debug grid
	request.Gemini("```\n")
	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			request.PlainText(debugMap[y][x])
		}
		request.PlainText("\n")
	}
	request.Gemini("```\n")

	request.Gemini("Legend:\n")
	request.Gemini("- P: Mountain peak\n")
	request.Gemini("- M: Mountain terrain\n")
	request.Gemini("- .: Direction line\n")
	request.Gemini("- +: Range boundary\n")
	request.Gemini("- S/s/w: Strong/medium/weak stretch zones\n")
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
	// Base terrain with Perlin noise
	perlin := perlin.NewPerlin(2.0, 2.5, 3, seed)

	// Generate base terrain with gentle hills
	baseHeight := perlin.Noise2D(float64(x)/(MapWidth*0.4), float64(y)/(MapHeight*0.4)) * 0.6
	baseHeight += perlin.Noise2D(float64(x)/(MapWidth*0.1), float64(y)/(MapHeight*0.1)) * 0.15
	baseHeight += 0.2 // Baseline offset

	finalHeight := baseHeight

	// For each mountain peak, generate a highly elongated range
	for _, peak := range peaks {
		peakX := peak.peakX
		peakY := peak.peakY

		// Vector from peak to current point
		dirX := float64(x - peakX)
		dirY := float64(y - peakY)

		// Basic distance
		distance := math.Sqrt(math.Pow(dirX, 2) + math.Pow(dirY, 2))

		// Determine range direction (0 to 2π)
		rangeDirection := (math.Mod(float64(peakX*peakY+int(seed)), 360)) * math.Pi / 180

		// Calculate the angle of the current point relative to the peak
		pointAngle := math.Atan2(dirY, dirX)

		// Calculate how aligned this point is with the mountain range direction
		// 1 = perfectly aligned, 0 = perpendicular
		angleAlignment := math.Abs(math.Cos(pointAngle - rangeDirection))

		// Create extreme stretching factor with gentler transition
		stretchMinimum := 0.15 // Controls width (smaller = narrower)
		stretchMaximum := 8.0  // Controls length (larger = longer)

		// Calculate stretch factor with extreme bias for elongation
		// Using a gentler power function (squared instead of cubed)
		stretchFactor := stretchMinimum + (stretchMaximum-stretchMinimum)*math.Pow(angleAlignment, 2.5)

		// Apply the stretch factor to create a modified distance
		modifiedDistance := distance / stretchFactor

		// Calculate rotated coordinates aligned with range direction
		alignedX := dirX*math.Cos(-rangeDirection) + dirY*math.Sin(-rangeDirection)
		alignedY := -dirX*math.Sin(-rangeDirection) + dirY*math.Cos(-rangeDirection)

		// Use absolute values for dimension checking
		lengthwiseDistance := math.Abs(alignedX)
		crosswiseDistance := math.Abs(alignedY)

		// Extended maximum dimensions for smoother falloff
		// Inner bounds = hard constraints, outer bounds = falloff zone
		innerLengthwise := 8.5  // Core range length
		outerLengthwise := 13.5 // Extended falloff zone
		innerCrosswise := 2.0   // Core range half-width
		outerCrosswise := 4.0   // Extended falloff zone

		// Only process points within the extended range boundaries
		if lengthwiseDistance <= outerLengthwise && crosswiseDistance <= outerCrosswise {
			// Distance-based falloff with moderate steepness
			// Increase the exponent for steeper falloff
			// Decrease the denominator for steeper falloff
			distanceFactor := math.Exp(-math.Pow(modifiedDistance, 2.0) / 8.0)

			// Dimension-based falloff - calculate based on position relative to inner/outer bounds
			var widthFactor, lengthFactor float64

			// Width falloff calculation
			if crosswiseDistance <= innerCrosswise {
				// Inside the core width - moderate internal falloff
				widthFactor = 1.0 - 0.2*(crosswiseDistance/innerCrosswise)
			} else {
				// In the extended width falloff zone
				widthPosition := (crosswiseDistance - innerCrosswise) / (outerCrosswise - innerCrosswise)
				// Use a gentler falloff function (square root for less steep decline)
				//widthFactor = 0.8 * (1.0 - math.Sqrt(widthPosition))

				// Linear falloff in extension zone
				widthFactor = 0.8 * (1.0 - widthPosition)
			}

			// Length falloff calculation
			if lengthwiseDistance <= innerLengthwise {
				// Inside the core length - very minimal falloff
				lengthFactor = 1.0 - 0.3*(lengthwiseDistance/innerLengthwise)
			} else {
				// In the extended length falloff zone
				lengthPosition := (lengthwiseDistance - innerLengthwise) / (outerLengthwise - innerLengthwise)
				// Use a gentler falloff function
				// lengthFactor = 0.7 * (1.0 - math.Pow(lengthPosition, 0.7))

				// Linear falloff in extension zone
				lengthFactor = 0.7 * (1.0 - lengthPosition)
			}

			// Combine all factors with emphasis on maintaining height
			// Use a weighted average that prioritizes the highest values
			// heightFactor := math.Max(distanceFactor, 0.7*widthFactor*lengthFactor)

			// Combine all factors
			heightFactor := distanceFactor * widthFactor * lengthFactor

			// Apply some noise along the range for varied peaks
			heightVariation := perlin.Noise2D(float64(x+peakX)/10, float64(y+peakY)/10) * 0.2

			// Ensure mountain height is substantial with gentler threshold
			baseHeight := 1.5 // NOTE: Base Mountain Height! Lower this if peaks are too high!
			mountainHeight := baseHeight * heightFactor * (1.0 + heightVariation)

			// More gradual cutoff for adding height
			// Lower threshold to extend mountain influence
			/*heightContributionThreshold := 0.04 // Higher for sharper cutoff and steeper transition
			if heightFactor > heightContributionThreshold {
				// Apply a smoothstep-like function for gradual addition near edges
				// The denominator of the blendFactor determines where mountains "end". Higher values = more distinct mountain boundaries.
				blendFactor := math.Min(1.0, (heightFactor-heightContributionThreshold)/0.10)
				finalHeight += mountainHeight * blendFactor
				}*/

			// Allow much smaller contribution to be visible
			// No need for threshold cutoff - let the falloff be naturally visible
			finalHeight += mountainHeight
		}
	}

	return baseHeight, finalHeight
}

func generatePlateaus(seed int64) {
	// Create a separate Perlin noise generator for plateau locations
	plateauNoise := perlin.NewPerlin(1.8, 3.0, 2, seed+42)

	// Parameters for plateau generation
	plateauThreshold := 0.58       // Higher value = fewer plateaus
	plateauHeightBase := 0.65      // Base elevation for plateaus (higher than hills)
	plateauHeightVariation := 0.15 // How much elevation varies between plateaus
	plateauFlatness := 0.85        // How flat plateaus are (higher = flatter)

	// First pass - identify potential plateau regions
	potentialPlateaus := 0
	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			// Skip areas that are too low (water) or mountains
			// Also skip areas that are already too high (near mountains)
			if Map[y][x].altitude <= 0.25 || Map[y][x].altitude >= 0.9 {
				continue
			}

			// Use noise to determine plateau locations
			plateauValue := plateauNoise.Noise2D(float64(x)/(MapWidth*0.2), float64(y)/(MapHeight*0.2))

			if plateauValue > plateauThreshold {
				potentialPlateaus++
			}
		}
	}

	// If we have enough potential plateau regions, create them
	if potentialPlateaus > 0 {
		// Each plateau region gets a slightly different target height
		heightNoise := perlin.NewPerlin(2.5, 2.0, 2, seed+84)

		// Second pass - apply plateau heights
		for y := 0; y < MapHeight; y++ {
			for x := 0; x < MapWidth; x++ {
				// Skip areas that are too low (water) or near mountains
				if Map[y][x].altitude <= 0.25 || Map[y][x].altitude >= 0.9 {
					continue
				}

				// Use the same noise function to find plateau regions
				plateauValue := plateauNoise.Noise2D(float64(x)/(MapWidth*0.2), float64(y)/(MapHeight*0.2))

				if plateauValue > plateauThreshold {
					// Determine the target height for this plateau region
					regionHeight := heightNoise.Noise2D(float64(x)/(MapWidth*0.6), float64(y)/(MapHeight*0.6))

					// Calculate plateau height - varying between plateaus but flat within each
					// Ensure plateaus are higher than hills (0.5-0.8 range)
					plateauHeight := plateauHeightBase + regionHeight*plateauHeightVariation

					// Blend between original height and plateau height
					blendStrength := (plateauValue - plateauThreshold) * 3.0
					blendStrength = math.Min(blendStrength, plateauFlatness)

					// Calculate the new height as a blend between original and plateau
					newHeight := Map[y][x].altitude*(1-blendStrength) + plateauHeight*blendStrength

					// Ensure plateau remains in proper range
					if newHeight > 0.9 {
						newHeight = 0.9 // Cap plateau height below mountains
					}

					// Apply the new height
					Map[y][x].altitude = newHeight
					Map[y][x].landType = LandType_Plateaus
				}
			}
		}

		// Smooth plateau edges (keeping the same code from before)
		var tempMap [MapHeight][MapWidth]float64
		for y := 0; y < MapHeight; y++ {
			for x := 0; x < MapWidth; x++ {
				tempMap[y][x] = Map[y][x].altitude
			}
		}

		// Apply edge smoothing
		for y := 1; y < MapHeight-1; y++ {
			for x := 1; x < MapWidth-1; x++ {
				plateauValue := plateauNoise.Noise2D(float64(x)/(MapWidth*0.2), float64(y)/(MapHeight*0.2))
				if math.Abs(plateauValue-plateauThreshold) > 0.1 {
					continue
				}

				// Calculate average height of neighbors
				sum := 0.0
				count := 0

				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}

						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
							sum += tempMap[ny][nx]
							count++
						}
					}
				}

				if count > 0 {
					avgHeight := sum / float64(count)

					// Blend between current height and average height at plateau edges
					edgeBlend := 1.0 - math.Abs(plateauValue-plateauThreshold)*10.0
					edgeBlend = math.Max(0.0, math.Min(0.5, edgeBlend))

					Map[y][x].altitude = tempMap[y][x]*(1-edgeBlend) + avgHeight*edgeBlend
				}
			}
		}
	}
}
