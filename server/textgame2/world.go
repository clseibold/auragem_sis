package textgame2

import (
	"math"
	"math/rand"
	"sort"

	"github.com/aquilax/go-perlin"
)

// | Terrain Type | Altitude Range | Display |
// |--------------|----------------|---------|
// | Water        | ≤ 0.0          | ~ |
// | Plains       | 0.0 - 0.3      | (space) |
// | Hills        | 0.3 - 0.5      | + |
// | Plateaus     | 0.5 - 0.8      | = |
// | Rough High   | 0.8 - 1.0      | n |
// | Mountains    | ≥ 1.0          | A |

// TODO: Canyons, Gorges, Cliffs, Waterfalls, Escarpments, Islands, Caves and Caverns?, Rock formations

// TODO: Assign biomes to each tile given its land type, adjacent biomes, and bodies of water

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

	// Water features
	hasStream bool // Contains a small stream/creek within the tile
	hasPond   bool // Contains a small pond within the tile
	hasSpring bool // Contains a natural spring (water source)
	hasMarsh  bool // Contains a marshy area (soggy ground)

	// Plains features
	hasGrove     bool // Contains a small grove of trees
	hasMeadow    bool // Contains a flower-rich meadow
	hasScrub     bool // Contains scrubland with brush
	hasRocks     bool // Contains small rock outcroppings
	hasGameTrail bool // Contains animal paths/trails
	hasFloodArea bool // Contains area that seasonally floods
	hasSaltFlat  bool // Contains a small salt flat or mineral deposit
}

type Peak struct {
	peakX int
	peakY int
}

func generateWorldMap() {
	var seed int64 = 1239462936493264926
	rand := rand.New(rand.NewSource(seed))

	// Generate mountain peaks
	MapPeaks = generateMapMountainPeaks(rand)

	// Generate base terrain with mountains
	for y := range MapHeight {
		for x := range MapWidth {
			perlinAltitude, altitude := generateHeight(MapPeaks, x, y, seed)
			Map[y][x] = Tile{altitude: altitude}
			MapPerlin[y][x] = Tile{altitude: perlinAltitude}
		}
	}

	// Create additional water bodies
	createWaterBodies(seed)

	// Assign basic land types based on altitude
	assignLandTypes()

	// Generate plateaus (this will set LandType_Plateaus)
	generatePlateaus(seed)

	// Generate rivers flowing from high to low elevation
	generateRivers(seed)

	// Generate small-scale water features (ponds, streams, springs, and marshes)
	generateSmallWaterFeatures(seed)

	// Generate plains-specific features to add variety
	generatePlainsFeatures(seed)

	// Identify valleys
	identifyValleys()

	// Identify coastal areas
	identifyCoastalAreas()
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
		for range 20 {
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
	baseHeight := perlin.Noise2D(float64(x)/(MapWidth*0.7), float64(y)/(MapHeight*0.7)) * 0.45
	secondaryNoise := perlin.Noise2D(float64(x)/(MapWidth*0.18), float64(y)/(MapHeight*0.18)) * 0.1
	tertiaryNoise := perlin.Noise2D(float64(x+50)/(MapWidth*0.3), float64(y+50)/(MapHeight*0.3)) * 0.15
	baseHeight += secondaryNoise + tertiaryNoise + 0.2 // With added baseline offset

	// Adjust mid-range heights to create more distinct plains/hills separation
	// This will help spread out hills more evenly
	if baseHeight > 0.3 && baseHeight < 0.4 {
		// Create a steeper transition between plains and hills
		// This makes hills more distinct and better distributed
		transitionFactor := (baseHeight - 0.3) / 0.1
		baseHeight = 0.3 + transitionFactor*0.15
	}

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
		outerLengthwise := 10.5 // Extended falloff zone
		innerCrosswise := 1.75  // Core range half-width
		outerCrosswise := 3.75  // Extended falloff zone

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

// Do this before generating plateaus and other terrain, but after generating base terrain with mountains.
func assignLandTypes() {
	// Assign land types based on altitude and other characteristics
	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			altitude := Map[y][x].altitude

			// First, assign basic land types based on altitude
			if altitude <= 0.0 {
				// Water features
				Map[y][x].landType = LandType_Water
			} else if altitude >= 1.0 {
				// Mountain terrain
				Map[y][x].landType = LandType_Mountains
			} else if altitude >= 0.8 && altitude < 1.0 {
				// High terrain / foothills - usually near mountains
				Map[y][x].landType = LandType_Hills
			} else if altitude >= 0.5 && altitude < 0.8 {
				// Default to hills for mid-high elevation
				// (Will be overridden if it's a plateau)
				Map[y][x].landType = LandType_Hills
			} else if altitude >= 0.3 && altitude < 0.5 {
				// Regular hills
				Map[y][x].landType = LandType_Hills
			} else {
				// Plains for low elevation
				Map[y][x].landType = LandType_Plains
			}

			// Additional terrain analysis (checking for valleys, etc.)
			// could be added here
		}
	}
}

func generatePlateaus(seed int64) {
	// Create a separate Perlin noise generator for plateau locations
	plateauNoise := perlin.NewPerlin(1.8, 3.0, 2, seed+42)

	// Parameters for plateau generation
	plateauThreshold := 0.54       // Higher value = fewer plateaus
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
					Map[y][x].landType = LandType_Plateaus
				}
			}
		}
	}
}

func identifyValleys() {
	// Create temporary array to store gradients
	gradientMap := make([][]float64, MapHeight)
	for i := range gradientMap {
		gradientMap[i] = make([]float64, MapWidth)
	}

	// Calculate local gradients - how quickly altitude changes
	for y := 1; y < MapHeight-1; y++ {
		for x := 1; x < MapWidth-1; x++ {
			// Skip water
			if Map[y][x].altitude <= 0 {
				continue
			}

			// Calculate average height difference with neighbors
			totalDiff := 0.0
			count := 0

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
						heightDiff := Map[y][x].altitude - Map[ny][nx].altitude
						totalDiff += heightDiff
						count++
					}
				}
			}

			// Average gradient
			if count > 0 {
				gradientMap[y][x] = totalDiff / float64(count)
			}
		}
	}

	// Identify valleys - areas lower than surroundings
	for y := 1; y < MapHeight-1; y++ {
		for x := 1; x < MapWidth-1; x++ {
			// Skip water
			if Map[y][x].altitude <= 0 {
				continue
			}

			// If we're lower than average surroundings and not too high
			if gradientMap[y][x] < -0.05 && Map[y][x].altitude < 0.7 {
				// Avoid marking plateaus or mountains as valleys
				if Map[y][x].landType != LandType_Plateaus &&
					Map[y][x].landType != LandType_Mountains {
					Map[y][x].landType = LandType_Valleys
				}
			}
		}
	}
}

func identifyCoastalAreas() {
	// Mark tiles near water as coastal
	for y := range MapHeight {
		for x := range MapWidth {
			// Skip water tiles
			if Map[y][x].altitude <= 0 {
				continue
			}

			// Check if any neighbor is water
			hasWaterNeighbor := false
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
						if Map[ny][nx].altitude <= 0 {
							hasWaterNeighbor = true
							break
						}
					}
				}
				if hasWaterNeighbor {
					break
				}
			}

			// If next to water and not a mountain or plateau, mark as coastal
			// Preserve valleys that are next to water - these are river valleys
			if hasWaterNeighbor && Map[y][x].altitude < 1.0 {
				// Don't overwrite valleys or plateaus
				if Map[y][x].landType != LandType_Valleys &&
					Map[y][x].landType != LandType_Plateaus {
					Map[y][x].landType = LandType_Coastal
				}
			}
		}
	}
}

// Add this function to create additional water bodies
func createWaterBodies(seed int64) {
	// Generate water bodies using separate noise
	waterNoise := perlin.NewPerlin(2.2, 2.0, 2, seed+789)

	// Use a second noise field for more varied water patterns
	secondaryWaterNoise := perlin.NewPerlin(1.8, 2.5, 2, seed+367)

	// Parameters for water body generation
	mainWaterThreshold := -0.55      // less negative = more water
	secondaryWaterThreshold := -0.60 // Threshold for secondary water features
	maxElevationForWater := 0.35

	for y := range MapHeight {
		for x := range MapWidth {
			// Skip existing water and mountains
			if Map[y][x].altitude <= 0 || Map[y][x].altitude >= 0.9 {
				continue
			}

			// Generate water body noise
			waterValue := waterNoise.Noise2D(float64(x)/(MapWidth*0.25), float64(y)/(MapHeight*0.25))

			// Create water where the noise is strongly negative and terrain is low
			if waterValue < mainWaterThreshold && Map[y][x].altitude < maxElevationForWater {
				// Depress terrain below water level
				// Depress terrain below water level - deeper water for stronger noise values
				waterDepth := math.Min(-0.1, waterValue*0.15)
				Map[y][x].altitude = waterDepth
				Map[y][x].landType = LandType_Water
			}
		}
	}

	// Second pass - add secondary water features
	for y := range MapHeight {
		for x := range MapWidth {
			// Skip existing water and mountains
			if Map[y][x].altitude <= 0 || Map[y][x].altitude >= 0.9 {
				continue
			}

			// Generate secondary water noise
			secondaryWater := secondaryWaterNoise.Noise2D(float64(x)/(MapWidth*0.15), float64(y)/(MapHeight*0.15))

			// Add small lakes and ponds where secondary noise is very negative and terrain is very low
			if secondaryWater < secondaryWaterThreshold && Map[y][x].altitude < 0.25 {
				Map[y][x].altitude = -0.05
				Map[y][x].landType = LandType_Water
			}
		}
	}

	// Third pass - expand existing water bodies slightly to create more natural shapes
	var waterExpansion [MapHeight][MapWidth]bool

	// Identify tiles adjacent to water that could become water
	for y := 1; y < MapHeight-1; y++ {
		for x := 1; x < MapWidth-1; x++ {
			// Skip existing water and higher terrain
			if Map[y][x].altitude <= 0 || Map[y][x].altitude >= 0.3 {
				continue
			}

			// Count adjacent water tiles
			waterNeighbors := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight && Map[ny][nx].altitude <= 0 {
						waterNeighbors++
					}
				}
			}

			// Mark low-lying tiles with water neighbors for potential expansion
			if waterNeighbors >= 2 && Map[y][x].altitude < 0.2 {
				// Use a noise value to make expansion more natural and varied
				noiseVal := waterNoise.Noise2D(float64(x+30)/(MapWidth*0.1), float64(y+30)/(MapHeight*0.1))
				if noiseVal < 0.2 {
					waterExpansion[y][x] = true
				}
			}
		}
	}

	// Apply the water expansion
	for y := range MapHeight {
		for x := range MapWidth {
			if waterExpansion[y][x] {
				Map[y][x].altitude = -0.05
				Map[y][x].landType = LandType_Water
			}
		}
	}
}

func generateRivers(seed int64) {
	// Initialize random source for river generation
	rng := rand.New(rand.NewSource(seed + 12345))

	// Parameters for river generation
	numberOfRivers := 4 + rng.Intn(3) // 4-6 rivers
	minRiverLength := 5               // Minimum tiles a river should span
	maxRiverLength := 25              // Maximum river length
	minElevationStart := 0.6          // Rivers start in higher elevations

	// Store river paths for debug visualization if needed
	riverPaths := make([][]struct{ x, y int }, 0, numberOfRivers)

	// Track tiles that already have rivers to avoid overlaps
	var riverTiles [MapHeight][MapWidth]bool

	// Track "river influence zone" - areas near rivers where new rivers shouldn't start
	var riverInfluence [MapHeight][MapWidth]bool

	// Find all potential river source points
	type potentialSource struct {
		x, y  int
		score float64 // Score for how good this source point is
	}

	potentialSources := make([]potentialSource, 0, 100)

	// Scan the entire map for potential river sources
	for y := 1; y < MapHeight-1; y++ {
		for x := 1; x < MapWidth-1; x++ {
			// Check if this point meets our criteria for a river source
			if Map[y][x].altitude >= minElevationStart &&
				Map[y][x].altitude < 0.95 &&
				!riverTiles[y][x] {

				// Check for downhill flow potential
				hasLowerNeighbor := false
				steepestDrop := 0.0

				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}

						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
							Map[ny][nx].altitude < Map[y][x].altitude {
							hasLowerNeighbor = true
							drop := Map[y][x].altitude - Map[ny][nx].altitude
							if drop > steepestDrop {
								steepestDrop = drop
							}
						}
					}
				}

				// If we can flow downhill, add to potential sources
				if hasLowerNeighbor {
					// Score based on elevation and steepness of descent
					// Higher elevations and steeper initial descents make better sources
					sourceScore := Map[y][x].altitude*0.7 + steepestDrop*0.3

					potentialSources = append(potentialSources, potentialSource{
						x:     x,
						y:     y,
						score: sourceScore,
					})
				}
			}
		}
	}

	// Sort potential sources by score (best sources first)
	sort.Slice(potentialSources, func(i, j int) bool {
		return potentialSources[i].score > potentialSources[j].score
	})

	// Keep track of how many rivers we've successfully created
	riversCreated := 0

	// Try to create rivers starting from the best source points
	for i := 0; i < len(potentialSources) && riversCreated < numberOfRivers; i++ {
		source := potentialSources[i]

		// Skip if this source is already part of a river
		if riverTiles[source.y][source.x] {
			continue
		}

		// NEW: Skip if this source is too close to an existing river
		if riverInfluence[source.y][source.x] {
			continue
		}

		// Trace river path from this source
		river := traceRiverPath(source.x, source.y, rng, riverTiles, riverInfluence, minRiverLength, maxRiverLength)

		// Only apply rivers that meet the minimum length requirement
		if len(river) >= minRiverLength {
			riverPaths = append(riverPaths, river)
			riversCreated++

			// Apply the river to the map
			for _, point := range river {
				x, y := point.x, point.y

				// Mark as river tile
				riverTiles[y][x] = true

				// Make this point water
				Map[y][x].altitude = -0.05
				Map[y][x].landType = LandType_Water

				// NEW: Mark river influence zone - area around the river where new rivers shouldn't go
				for dy := -2; dy <= 2; dy++ {
					for dx := -2; dx <= 2; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
							// Distance-based influence (stronger near the river)
							distance := math.Sqrt(float64(dx*dx + dy*dy))
							if distance <= 2.0 {
								riverInfluence[ny][nx] = true
							}
						}
					}
				}

				// Create river valleys by slightly lowering adjacent terrain
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}

						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
							Map[ny][nx].altitude > 0 && Map[ny][nx].altitude < 0.9 {
							// Create subtle river valley
							Map[ny][nx].altitude -= 0.05
							if Map[ny][nx].altitude < 0.05 {
								Map[ny][nx].altitude = 0.05
							}
						}
					}
				}
			}
		}
	}
}

func traceRiverPath(startX, startY int, rng *rand.Rand, riverTiles [MapHeight][MapWidth]bool, riverInfluence [MapHeight][MapWidth]bool, minLength, maxLength int) []struct{ x, y int } {
	// River path
	path := make([]struct{ x, y int }, 0, maxLength)
	path = append(path, struct{ x, y int }{startX, startY})

	// Current position
	x, y := startX, startY

	// Noise for adding natural meandering to river flow
	flowNoise := perlin.NewPerlin(1.5, 2.0, 2, rng.Int63())

	// Keep flowing downhill until we reach water or can't flow further
	for len(path) < maxLength {
		// Determine possible flow directions
		type flowOption struct {
			x, y           int
			elevation      float64
			distance       float64 // Distance from ideal flow direction
			riverProximity float64 // NEW: Penalty for being near existing rivers
		}

		options := make([]flowOption, 0, 8)

		// Current elevation
		currentElevation := Map[y][x].altitude

		// Calculate flow direction based on overall slope and existing path
		flowDirX, flowDirY := 0.0, 0.0

		// Look at the last few points in the path to determine trend
		pathLength := len(path)
		lookback := 5
		if pathLength > lookback {
			for i := 1; i <= lookback; i++ {
				prevPoint := path[pathLength-i]
				flowDirX += float64(x - prevPoint.x)
				flowDirY += float64(y - prevPoint.y)
			}

			// Normalize the flow direction
			magnitude := math.Sqrt(flowDirX*flowDirX + flowDirY*flowDirY)
			if magnitude > 0 {
				flowDirX /= magnitude
				flowDirY /= magnitude
			}
		}

		// Check all 8 neighbors
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := x+dx, y+dy

				// Skip if out of bounds
				if nx < 0 || nx >= MapWidth || ny < 0 || ny >= MapHeight {
					continue
				}

				// Skip if already a river (to avoid loops) unless it's a water body
				if riverTiles[ny][nx] && Map[ny][nx].altitude > -0.1 {
					continue
				}

				// Check elevation - must be lower or water
				neighborElevation := Map[ny][nx].altitude
				if neighborElevation < currentElevation || neighborElevation <= 0 {
					// Calculate how well this direction aligns with the current flow trend
					alignment := 1.0
					if pathLength > lookback {
						dotProduct := flowDirX*float64(dx) + flowDirY*float64(dy)
						alignment = (dotProduct + 1.0) / 2.0 // Scale from [-1,1] to [0,1]
					}

					// Add noise to make the flow more natural
					noiseValue := flowNoise.Noise2D(float64(nx)/10.0, float64(ny)/10.0)

					// Add a penalty for flowing near existing rivers
					// This discourages rivers from running parallel to each other
					riverProximityPenalty := 0.0
					if riverInfluence[ny][nx] {
						// Strong penalty for getting too close to existing rivers
						riverProximityPenalty = 0.5
					}

					// Calculate elevation difference including noise and flow alignment
					elevationDiff := currentElevation - neighborElevation
					flowScore := elevationDiff + noiseValue*0.1 + alignment*0.2 - riverProximityPenalty

					options = append(options, flowOption{
						x:              nx,
						y:              ny,
						elevation:      neighborElevation,
						distance:       flowScore,
						riverProximity: riverProximityPenalty,
					})
				}
			}
		}

		// If no downhill options, we've reached a local minimum
		if len(options) == 0 {
			break
		}

		// Choose the best option, favoring steeper descent and flow alignment
		// But avoiding proximity to other rivers
		bestOption := options[0]
		for _, option := range options {
			if option.distance > bestOption.distance {
				bestOption = option
			}
		}

		// Move to the next point
		x, y = bestOption.x, bestOption.y
		path = append(path, struct{ x, y int }{x, y})

		// If we've reached a water body or existing river, we're done
		if Map[y][x].altitude <= 0 {
			// We reached water, the river is complete
			if len(path) >= minLength {
				return path
			}
			break
		}
	}

	// Only return the path if it meets the minimum length requirement
	if len(path) >= minLength {
		return path
	}

	// Return an empty path if it's too short
	return []struct{ x, y int }{}
}

func generateSmallWaterFeatures(seed int64) {
	rng := rand.New(rand.NewSource(seed + 5552))

	// Parameters for small water features
	springCount := 8 + rng.Intn(5)      // 8-12 springs
	marshCount := 12 + rng.Intn(8)      // 12-19 marshes
	smallRiverCount := 10 + rng.Intn(8) // 10-17 small rivers
	smallPondCount := 10 + rng.Intn(5)  // 10-14 small ponds

	// Track where we've already placed water features
	var waterFeaturePlaced [MapHeight][MapWidth]bool

	// Mark existing water and adjacent tiles as unavailable
	for y := range MapHeight {
		for x := range MapWidth {
			if Map[y][x].altitude <= 0 { // Water tiles
				waterFeaturePlaced[y][x] = true

				// Mark adjacent tiles as unavailable too
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
							waterFeaturePlaced[ny][nx] = true
						}
					}
				}
			}
		}
	}

	// 1. Generate springs first (they can be sources for other features)
	springsGenerated := 0
	var springLocations []struct{ x, y int }

	for attempts := 0; attempts < 200 && springsGenerated < springCount; attempts++ {
		// Springs often form at specific geological interfaces
		// Typically at hillsides, mountain bases, or where permeable rock meets impermeable layers

		// Try to find a location at the base of higher elevation
		x := rng.Intn(MapWidth-2) + 1
		y := rng.Intn(MapHeight-2) + 1

		// Good spring locations: hillsides, mountain bases, or plateau edges
		isGoodSpringLocation := false
		hasHigherNeighbor := false
		baseElevation := Map[y][x].altitude

		// Check if we have higher terrain nearby (spring source)
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
					// Springs tend to form where there's a significant elevation change
					elevationDiff := Map[ny][nx].altitude - baseElevation
					if elevationDiff > 0.25 {
						hasHigherNeighbor = true
						break
					}
				}
			}
			if hasHigherNeighbor {
				break
			}
		}

		// Check if this is a suitable location for a spring
		if !waterFeaturePlaced[y][x] &&
			baseElevation > 0.25 && baseElevation < 0.85 &&
			hasHigherNeighbor &&
			Map[y][x].landType != LandType_Mountains &&
			Map[y][x].landType != LandType_Water {

			isGoodSpringLocation = true

			// Extra check: favor locations at the edge of plateaus or hills
			if Map[y][x].landType == LandType_Hills ||
				Map[y][x].landType == LandType_Plateaus {
				// Higher chance to place springs here
				if rng.Float64() < 0.8 {
					isGoodSpringLocation = true
				}
			}

			// Check if we're near (but not at) the foot of a mountain
			nearMountain := false
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
						if Map[ny][nx].landType == LandType_Mountains {
							nearMountain = true
							break
						}
					}
				}
				if nearMountain {
					break
				}
			}

			// Higher chance to place springs near mountains
			if nearMountain {
				isGoodSpringLocation = rng.Float64() < 0.7
			}
		}

		if isGoodSpringLocation {
			// Set the spring flag
			Map[y][x].hasSpring = true

			// Mark as placed to avoid overlaps
			waterFeaturePlaced[y][x] = true

			// Save location for potential use as source of streams/ponds
			springLocations = append(springLocations, struct{ x, y int }{x, y})

			springsGenerated++
		}
	}

	// 2. Generate marshes (soggy areas)
	marshesGenerated := 0

	for attempts := 0; attempts < 200 && marshesGenerated < marshCount; attempts++ {
		// Marshes typically form in low-lying areas with poor drainage
		// Or areas with high water tables (near rivers/streams)

		// Try to find a location for a marsh
		x := rng.Intn(MapWidth-2) + 1
		y := rng.Intn(MapHeight-2) + 1

		// Good marsh locations: low-lying areas, near water, flat terrain
		isGoodMarshLocation := false
		elevation := Map[y][x].altitude

		// Check if we're near water or in a low-lying area
		nearWater := false
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
					if Map[ny][nx].altitude <= 0 { // Water nearby
						nearWater = true
						break
					}
				}
			}
			if nearWater {
				break
			}
		}

		// Calculate how flat the terrain is
		isFlat := true
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
					if math.Abs(Map[ny][nx].altitude-elevation) > 0.1 {
						isFlat = false
						break
					}
				}
			}
			if !isFlat {
				break
			}
		}

		// Check if this is a suitable location for a marsh
		if !waterFeaturePlaced[y][x] &&
			elevation > 0.05 && elevation < 0.4 && // Low-lying areas
			Map[y][x].landType != LandType_Mountains &&
			Map[y][x].landType != LandType_Plateaus {

			// Higher chance in flat and low areas, especially near water
			if isFlat {
				isGoodMarshLocation = true

				if nearWater {
					// Much higher chance near water
					isGoodMarshLocation = rng.Float64() < 0.8
				} else {
					// Lower chance away from water
					isGoodMarshLocation = rng.Float64() < 0.4
				}
			}

			// Special case: near a spring
			nearSpring := false
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
						if Map[ny][nx].hasSpring {
							nearSpring = true
							break
						}
					}
				}
				if nearSpring {
					break
				}
			}

			// Higher chance to place marshes near springs
			if nearSpring {
				isGoodMarshLocation = rng.Float64() < 0.7
			}
		}

		if isGoodMarshLocation {
			// Set the marsh flag
			Map[y][x].hasMarsh = true

			// Mark as placed to avoid overlaps
			waterFeaturePlaced[y][x] = true

			marshesGenerated++
		}
	}

	// 3. Generate small ponds (some from springs)
	pondsGenerated := 0

	// First try to place some ponds at springs
	if len(springLocations) > 0 && smallPondCount > 0 {
		// Shuffle spring locations
		for i := len(springLocations) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			springLocations[i], springLocations[j] = springLocations[j], springLocations[i]
		}

		// Try to create ponds at some springs
		maxSpringPonds := min(len(springLocations), smallPondCount/2)
		for i := range maxSpringPonds {
			springX, springY := springLocations[i].x, springLocations[i].y

			// Find a suitable nearby location for the pond
			pondPlaced := false
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					// Skip the spring tile itself
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := springX+dx, springY+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
						if !waterFeaturePlaced[ny][nx] &&
							Map[ny][nx].altitude > 0.05 && Map[ny][nx].altitude < 0.6 &&
							Map[ny][nx].landType != LandType_Mountains &&
							Map[ny][nx].landType != LandType_Plateaus {

							// Place a pond here
							Map[ny][nx].hasPond = true
							waterFeaturePlaced[ny][nx] = true
							pondPlaced = true
							pondsGenerated++
							break
						}
					}
				}
				if pondPlaced {
					break
				}
			}
		}
	}

	// Generate remaining ponds in suitable locations
	for attempts := 0; attempts < 200 && pondsGenerated < smallPondCount; attempts++ {
		// Choose a random location
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable location for a pond
		if !waterFeaturePlaced[y][x] &&
			Map[y][x].altitude > 0.05 && Map[y][x].altitude < 0.5 &&
			Map[y][x].landType != LandType_Mountains &&
			Map[y][x].landType != LandType_Plateaus {

			// Higher chance in valleys or near marshes
			placePond := false

			if Map[y][x].landType == LandType_Valleys {
				placePond = rng.Float64() < 0.7 // High chance in valleys
			} else {
				// Check if near a marsh
				nearMarsh := false
				for dy := -2; dy <= 2; dy++ {
					for dx := -2; dx <= 2; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
							if Map[ny][nx].hasMarsh {
								nearMarsh = true
								break
							}
						}
					}
					if nearMarsh {
						break
					}
				}

				if nearMarsh {
					placePond = rng.Float64() < 0.6 // Good chance near marshes
				} else {
					placePond = rng.Float64() < 0.3 // Lower chance elsewhere
				}
			}

			if placePond {
				// Set the pond flag
				Map[y][x].hasPond = true

				// Mark as placed to avoid overlaps
				waterFeaturePlaced[y][x] = true
				pondsGenerated++
			}
		}
	}

	// 4. Generate small rivers (streams)
	streamsGenerated := 0

	// First try to place some streams starting from springs
	for _, spring := range springLocations {
		// Limit the number of streams
		if streamsGenerated >= smallRiverCount {
			break
		}

		// Only some springs form streams
		if rng.Float64() < 0.7 {
			// First, mark the spring tile as having a stream too
			Map[spring.y][spring.x].hasStream = true

			// Trace a path downhill from the spring
			streamPath := traceSmallStreamPath(spring.x, spring.y, rng, waterFeaturePlaced)

			// If we found a valid path of appropriate length
			if len(streamPath) >= 2 && len(streamPath) <= 5 {
				// Apply the stream to the map
				for i, point := range streamPath {
					// Skip the first point since we already marked it
					if i == 0 {
						continue
					}

					sx, sy := point.x, point.y

					// Set the stream flag
					Map[sy][sx].hasStream = true

					// Mark as placed to avoid overlaps
					waterFeaturePlaced[sy][sx] = true
				}

				streamsGenerated++
			}
		}
	}

	// Generate remaining streams in suitable locations
	for attempts := 0; attempts < 200 && streamsGenerated < smallRiverCount; attempts++ {
		// Choose a random location for the stream source
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable location for a stream source
		if !waterFeaturePlaced[y][x] &&
			Map[y][x].altitude > 0.3 && Map[y][x].altitude < 0.8 &&
			Map[y][x].landType != LandType_Mountains &&
			Map[y][x].landType != LandType_Plateaus {

			// Trace a short path downhill
			streamPath := traceSmallStreamPath(x, y, rng, waterFeaturePlaced)

			// If we found a valid path of appropriate length
			if len(streamPath) >= 2 && len(streamPath) <= 5 {
				// Apply the stream to the map
				for _, point := range streamPath {
					sx, sy := point.x, point.y

					// Set the stream flag
					Map[sy][sx].hasStream = true

					// Mark as placed to avoid overlaps
					waterFeaturePlaced[sy][sx] = true
				}

				streamsGenerated++
			}
		}
	}
}

// Helper function to trace a small stream path
func traceSmallStreamPath(startX, startY int, rng *rand.Rand, occupied [MapHeight][MapWidth]bool) []struct{ x, y int } {
	path := make([]struct{ x, y int }, 0, 5)
	path = append(path, struct{ x, y int }{startX, startY})

	x, y := startX, startY
	currentAltitude := Map[y][x].altitude

	// Maximum length for small streams
	maxLength := 5

	// Trace a short path downhill
	for len(path) < maxLength {
		// Find the lowest unoccupied neighbor
		lowestX, lowestY := -1, -1
		lowestAlt := currentAltitude

		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := x+dx, y+dy

				// Check bounds
				if nx < 0 || nx >= MapWidth || ny < 0 || ny >= MapHeight {
					continue
				}

				// For streams, we should allow flowing through spring tiles
				// but not through tiles already occupied by other features
				if occupied[ny][nx] && !Map[ny][nx].hasSpring {
					continue
				}

				// Skip if too high
				if Map[ny][nx].altitude >= currentAltitude {
					continue
				}

				// Check if this is the lowest neighbor so far
				if Map[ny][nx].altitude < lowestAlt {
					lowestAlt = Map[ny][nx].altitude
					lowestX = nx
					lowestY = ny
				}
			}
		}

		// If we couldn't find a lower neighbor, stop
		if lowestX == -1 {
			break
		}

		// Move to the lowest neighbor
		x, y = lowestX, lowestY
		currentAltitude = Map[y][x].altitude

		// Add to path
		path = append(path, struct{ x, y int }{x, y})

		// If we reached water, stop
		if Map[y][x].altitude <= 0 {
			break
		}

		// Random chance to end stream early (creates springs, seeps, etc.)
		if rng.Float64() < 0.2 {
			break
		}
	}

	return path
}

func generatePlainsFeatures(seed int64) {
	rng := rand.New(rand.NewSource(seed + 7890))

	// Feature quantity parameters
	groveCount := 20 + rng.Intn(10)   // 20-29 tree groves
	meadowCount := 15 + rng.Intn(10)  // 15-24 flower meadows
	scrubCount := 25 + rng.Intn(15)   // 25-39 scrubland patches
	rockCount := 10 + rng.Intn(8)     // 10-17 rock outcroppings
	gameTrailCount := 8 + rng.Intn(5) // 8-12 game trails
	floodAreaCount := 6 + rng.Intn(5) // 6-10 seasonal flood areas
	saltFlatCount := 3 + rng.Intn(3)  // 3-5 salt flats

	// Track places where we've already placed features
	var featurePlaced [MapHeight][MapWidth]bool

	// Mark existing water and special features as unavailable
	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			// Skip tiles that already have features
			if Map[y][x].altitude <= 0 || // Water
				Map[y][x].hasStream ||
				Map[y][x].hasPond ||
				Map[y][x].hasSpring ||
				Map[y][x].hasMarsh {
				featurePlaced[y][x] = true
			}
		}
	}

	// 1. Generate groves (small clusters of trees)
	grovesGenerated := 0

	for attempts := 0; attempts < 100 && grovesGenerated < groveCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable spot for a grove
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.1 && Map[y][x].altitude < 0.7 {

			// More likely near water sources
			placeGrove := false

			// Check if near water
			nearWater := false
			for dy := -3; dy <= 3; dy++ {
				for dx := -3; dx <= 3; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
						if Map[ny][nx].altitude <= 0 || Map[ny][nx].hasStream ||
							Map[ny][nx].hasPond || Map[ny][nx].hasSpring {
							nearWater = true
							break
						}
					}
				}
				if nearWater {
					break
				}
			}

			if nearWater {
				placeGrove = rng.Float64() < 0.7 // Higher chance near water
			} else {
				placeGrove = rng.Float64() < 0.3 // Lower chance away from water
			}

			if placeGrove {
				Map[y][x].hasGrove = true
				featurePlaced[y][x] = true

				// Some groves form small clusters
				if rng.Float64() < 0.4 {
					// Try to add 1-3 adjacent grove tiles
					extraGroves := 1 + rng.Intn(3)
					for e := 0; e < extraGroves; e++ {
						// Pick a random direction
						dx := rng.Intn(3) - 1
						dy := rng.Intn(3) - 1

						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
							!featurePlaced[ny][nx] &&
							Map[ny][nx].landType == LandType_Plains {
							Map[ny][nx].hasGrove = true
							featurePlaced[ny][nx] = true
						}
					}
				}

				grovesGenerated++
			}
		}
	}

	// 2. Generate meadows (flower-rich areas)
	meadowsGenerated := 0

	for attempts := 0; attempts < 100 && meadowsGenerated < meadowCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable spot for a meadow
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.1 && Map[y][x].altitude < 0.6 {

			// Meadows are more likely in wetter areas, but not too wet
			placeMeadow := false

			// Meadows often form in valleys or near water
			if Map[y][x].landType == LandType_Valleys {
				placeMeadow = rng.Float64() < 0.6
			} else {
				// Check if near water
				nearWater := false
				for dy := -3; dy <= 3; dy++ {
					for dx := -3; dx <= 3; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
							if Map[ny][nx].altitude <= 0 || Map[ny][nx].hasStream ||
								Map[ny][nx].hasPond || Map[ny][nx].hasSpring {
								nearWater = true
								break
							}
						}
					}
					if nearWater {
						break
					}
				}

				if nearWater {
					placeMeadow = rng.Float64() < 0.5
				} else {
					placeMeadow = rng.Float64() < 0.2
				}
			}

			if placeMeadow {
				Map[y][x].hasMeadow = true
				featurePlaced[y][x] = true
				meadowsGenerated++
			}
		}
	}

	// 3. Generate scrubland (areas with brush and small woody plants)
	scrubGenerated := 0

	for attempts := 0; attempts < 150 && scrubGenerated < scrubCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable spot for scrubland
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.2 && Map[y][x].altitude < 0.7 {

			// Scrubland is common in slightly drier areas, but not desert-dry
			placeScrub := rng.Float64() < 0.5 // Base chance is high, plains often have scrub

			if placeScrub {
				Map[y][x].hasScrub = true
				featurePlaced[y][x] = true

				// Scrubland often forms larger patches
				if rng.Float64() < 0.7 {
					// Try to add 2-5 adjacent scrub tiles
					extraScrub := 2 + rng.Intn(4)
					for e := 0; e < extraScrub; e++ {
						// Pick a random direction
						dx := rng.Intn(3) - 1
						dy := rng.Intn(3) - 1

						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
							!featurePlaced[ny][nx] &&
							Map[ny][nx].landType == LandType_Plains {
							Map[ny][nx].hasScrub = true
							featurePlaced[ny][nx] = true
						}
					}
				}

				scrubGenerated++
			}
		}
	}

	// 4. Generate rock outcroppings
	rocksGenerated := 0

	for attempts := 0; attempts < 100 && rocksGenerated < rockCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable spot for exposed rocks
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.3 && Map[y][x].altitude < 0.8 {

			// Rocks are more common at higher elevations
			placeRocks := false
			if Map[y][x].altitude > 0.6 {
				placeRocks = rng.Float64() < 0.6
			} else {
				placeRocks = rng.Float64() < 0.3
			}

			if placeRocks {
				Map[y][x].hasRocks = true
				featurePlaced[y][x] = true
				rocksGenerated++
			}
		}
	}

	// 5. Generate game trails
	trailsGenerated := 0

	for attempts := 0; attempts < 50 && trailsGenerated < gameTrailCount; attempts++ {
		// Game trails typically start at water sources and go to other features

		// Find a water source to start from
		waterSources := make([]struct{ x, y int }, 0)

		for y := 0; y < MapHeight; y++ {
			for x := 0; x < MapWidth; x++ {
				if Map[y][x].altitude <= 0 || Map[y][x].hasPond || Map[y][x].hasStream {
					waterSources = append(waterSources, struct{ x, y int }{x, y})
				}
			}
		}

		if len(waterSources) == 0 {
			break // No water sources to start from
		}

		// Pick a random water source
		source := waterSources[rng.Intn(len(waterSources))]

		// Create a trail from this source
		x, y := source.x, source.y
		trailLength := 3 + rng.Intn(5) // Trails are 3-7 tiles long

		// Pick a random direction that isn't into water
		var dx, dy int
		for {
			dx = rng.Intn(3) - 1
			dy = rng.Intn(3) - 1
			if dx != 0 || dy != 0 {
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
					Map[ny][nx].altitude > 0 {
					break
				}
			}
		}

		// Form the trail
		validTrail := true
		for i := 0; i < trailLength; i++ {
			x += dx
			y += dy

			// Stay on map
			if x < 0 || x >= MapWidth || y < 0 || y >= MapHeight {
				validTrail = false
				break
			}

			// Don't put trails in water
			if Map[y][x].altitude <= 0 {
				validTrail = false
				break
			}

			// Mark as game trail (ok to overlap with other features)
			Map[y][x].hasGameTrail = true

			// Occasionally change direction slightly
			if rng.Float64() < 0.3 {
				// Small direction change
				dx += rng.Intn(3) - 1
				dy += rng.Intn(3) - 1

				// Ensure we're still moving
				if dx == 0 && dy == 0 {
					if rng.Float64() < 0.5 {
						dx = 1
					} else {
						dy = 1
					}
				}

				// Limit max speed
				if dx > 1 {
					dx = 1
				} else if dx < -1 {
					dx = -1
				}
				if dy > 1 {
					dy = 1
				} else if dy < -1 {
					dy = -1
				}
			}
		}

		if validTrail {
			trailsGenerated++
		}
	}

	// 6. Generate seasonal flood areas
	floodAreasGenerated := 0

	for attempts := 0; attempts < 100 && floodAreasGenerated < floodAreaCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable spot for a flood area
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.05 && Map[y][x].altitude < 0.3 {

			// Flood areas are typically near water and in low spots
			placeFloodArea := false

			// Check if near water
			nearWater := false
			for dy := -4; dy <= 4; dy++ {
				for dx := -4; dx <= 4; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
						if Map[ny][nx].altitude <= 0 || Map[ny][nx].hasStream ||
							Map[ny][nx].hasPond {
							nearWater = true
							break
						}
					}
				}
				if nearWater {
					break
				}
			}

			// Flood areas typically form in valleys or low areas near water
			if Map[y][x].altitude < 0.2 && nearWater {
				placeFloodArea = rng.Float64() < 0.7
			} else if Map[y][x].landType == LandType_Valleys {
				placeFloodArea = rng.Float64() < 0.5
			} else if nearWater {
				placeFloodArea = rng.Float64() < 0.3
			} else {
				placeFloodArea = rng.Float64() < 0.1
			}

			if placeFloodArea {
				Map[y][x].hasFloodArea = true
				featurePlaced[y][x] = true

				// Flood areas often extend
				if rng.Float64() < 0.6 {
					// Try to add 1-3 adjacent flood area tiles
					extraFlood := 1 + rng.Intn(3)
					for e := 0; e < extraFlood; e++ {
						// Pick a random direction
						dx := rng.Intn(3) - 1
						dy := rng.Intn(3) - 1

						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
							!featurePlaced[ny][nx] &&
							Map[ny][nx].landType == LandType_Plains &&
							Map[ny][nx].altitude < 0.3 {
							Map[ny][nx].hasFloodArea = true
							featurePlaced[ny][nx] = true
						}
					}
				}

				floodAreasGenerated++
			}
		}
	}

	// 7. Generate salt flats
	saltFlatsGenerated := 0

	for attempts := 0; attempts < 50 && saltFlatsGenerated < saltFlatCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable spot for a salt flat
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.15 && Map[y][x].altitude < 0.4 {

			// Salt flats typically form in dry basins
			placeSaltFlat := rng.Float64() < 0.3

			if placeSaltFlat {
				Map[y][x].hasSaltFlat = true
				featurePlaced[y][x] = true

				// Salt flats can form small patches
				if rng.Float64() < 0.5 {
					// Try to add 1-2 adjacent salt flat tiles
					extraSalt := 1 + rng.Intn(2)
					for e := 0; e < extraSalt; e++ {
						// Pick a random direction
						dx := rng.Intn(3) - 1
						dy := rng.Intn(3) - 1

						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
							!featurePlaced[ny][nx] &&
							Map[ny][nx].landType == LandType_Plains &&
							Map[ny][nx].altitude < 0.4 {
							Map[ny][nx].hasSaltFlat = true
							featurePlaced[ny][nx] = true
						}
					}
				}

				saltFlatsGenerated++
			}
		}
	}
}
