package textgame2

import (
	"fmt"
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

// TODO: Canyons, Gorges, Cliffs, Waterfalls, Escarpments, Islands, Caves and Caverns?
// TODO: Aquifers should be generated *before* rainfall&temp. Rainfall/temp + aquifers should determine the locations of springs.
// TODO: Assign biomes to each tile given its land type, adjacent biomes, and bodies of water
// TODO: For streams, every tile of the stream should store the uphill tile location (or -1 if it is the source/start) and downhill tile location (or -1 if it's the end)

const MapWidth = 50
const MapHeight = 50

// const MapNumberOfMountainPeaks = 3

var MapPeaks []Peak
var Map [MapHeight][MapWidth]Tile
var MapPerlin [MapHeight][MapWidth]Tile

type TileLocation struct {
	X int
	Y int
}

// Each tile of the world map represents a 10 square kilometer region.
type Tile struct {
	altitude float64
	biome    Biome
	landType LandType
	isDesert bool

	// Climate factors (0.0 to 1.0 scale)
	rainfall    float64 // 0.0 = extremely dry, 1.0 = extremely wet
	temperature float64 // 0.0 = very cold, 1.0 = very hot

	// Climate factors (0.0 to 1.0 scale)
	climate Climate

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

// Helper methods to get temperature for a specific season
func (t *Tile) GetTemperature(season Season) float64 {
	switch season {
	case Winter:
		return t.climate.winterTemp
	case Spring:
		return t.climate.springTemp
	case Summer:
		return t.climate.summerTemp
	case Fall:
		return t.climate.fallTemp
	default:
		return t.climate.avgTemp
	}
}

// Helper methods to get rainfall for a specific season
func (t *Tile) GetRainfall(season Season) float64 {
	switch season {
	case Winter:
		return t.climate.winterRain
	case Spring:
		return t.climate.springRain
	case Summer:
		return t.climate.summerRain
	case Fall:
		return t.climate.fallRain
	default:
		return t.climate.avgRain
	}
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

	// Generate climate data (temperature and rainfall)
	generateClimate(seed)

	// Generate small-scale water features (ponds, streams, springs, and marshes)
	generateSmallWaterFeatures(seed)

	// Generate plains-specific features to add variety
	generatePlainsFeatures(seed)

	// Generate desert regions (after other features so we can adjust them)
	generateDeserts(seed)

	// Identify valleys
	identifyValleys()

	// Identify coastal areas
	identifyCoastalAreas()

	assignBiomes()

	generateGameTrails(seed)
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
	for y := range MapHeight {
		for x := range MapWidth {
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
	for y := range MapHeight {
		for x := range MapWidth {
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
		for y := range MapHeight {
			for x := range MapWidth {
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

					// Calculate the new height as a blend between original and plateau, capping at 0.9
					newHeight := min(Map[y][x].altitude*(1-blendStrength)+plateauHeight*blendStrength, 0.9)

					// Apply the new height
					Map[y][x].altitude = newHeight
					Map[y][x].landType = LandType_Plateaus
				}
			}
		}

		// Smooth plateau edges (keeping the same code from before)
		var tempMap [MapHeight][MapWidth]float64
		for y := range MapHeight {
			for x := range MapWidth {
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

func createWaterBodies(seed int64) {
	rng := rand.New(rand.NewSource(seed))

	// Generate water bodies using separate noise
	waterNoise := perlin.NewPerlin(2.2, 2.0, 2, seed+789)
	detailNoise := perlin.NewPerlin(3.0, 1.5, 2, seed+921)

	// Parameters for water generation
	largeWaterThreshold := -0.53  // -0.60
	mediumWaterThreshold := -0.47 // -0.53
	maxElevationForWater := 0.4   // 0.35

	// Track water bodies to maintain proper spacing
	var waterTiles [MapHeight][MapWidth]bool

	// First pass - create smaller, more numerous water bodies
	for y := range MapHeight {
		for x := range MapWidth {
			// Skip existing water and mountains
			if Map[y][x].altitude <= 0 || Map[y][x].altitude >= 0.9 {
				continue
			}

			// Generate water body noise
			waterValue := waterNoise.Noise2D(float64(x)/(MapWidth*0.25), float64(y)/(MapHeight*0.25))

			// Detail noise to break up water edges
			detailValue := detailNoise.Noise2D(float64(x)/(MapWidth*0.05), float64(y)/(MapHeight*0.05)) * 0.1

			// Different sizes of water bodies
			if waterValue < largeWaterThreshold && Map[y][x].altitude < maxElevationForWater {
				// Check spacing from existing water bodies
				tooClose := false
				searchRadius := 3 // Spacing for large bodies

				// Skip this check for the first few water bodies
				if countWaterTiles(waterTiles) > 12 {
					for dy := -searchRadius; dy <= searchRadius; dy++ {
						for dx := -searchRadius; dx <= searchRadius; dx++ {
							nx, ny := x+dx, y+dy
							if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
								waterTiles[ny][nx] {
								tooClose = true
								break
							}
						}
						if tooClose {
							break
						}
					}
				}

				if !tooClose {
					// Large water body
					waterDepth := math.Min(-0.1, waterValue*0.15)
					Map[y][x].altitude = waterDepth
					Map[y][x].landType = LandType_Water
					waterTiles[y][x] = true

					// Generate a small cluster of water around large bodies
					// But limit the size with a strict radius check
					maxRadius := 1 + rng.Intn(2) // 1-2 tile radius max

					for dy := -maxRadius; dy <= maxRadius; dy++ {
						for dx := -maxRadius; dx <= maxRadius; dx++ {
							// Skip center tile
							if dx == 0 && dy == 0 {
								continue
							}

							// Calculate distance for circular lakes
							dist := math.Sqrt(float64(dx*dx + dy*dy))
							if dist > float64(maxRadius) {
								continue // Outside the radius
							}

							nx, ny := x+dx, y+dy
							if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
								Map[ny][nx].altitude > 0 && // Not already water
								Map[ny][nx].altitude < maxElevationForWater {

								// Add detail noise for natural shorelines
								edgeNoise := detailNoise.Noise2D(float64(nx)/(MapWidth*0.08),
									float64(ny)/(MapHeight*0.08)) * 0.15

								// Higher chance to become water closer to center
								waterChance := 0.85 - (dist/float64(maxRadius))*0.4 + edgeNoise

								if rng.Float64() < waterChance {
									// Slightly shallower at edges
									edgeDepth := waterDepth * (1.0 - dist/float64(maxRadius+1))
									Map[ny][nx].altitude = edgeDepth
									Map[ny][nx].landType = LandType_Water
									waterTiles[ny][nx] = true
								}
							}
						}
					}
				}
			} else if waterValue < mediumWaterThreshold && Map[y][x].altitude < maxElevationForWater {
				// Medium water bodies (ponds)
				// Check spacing
				tooClose := false
				searchRadius := 2 // Smaller spacing for medium bodies

				for dy := -searchRadius; dy <= searchRadius; dy++ {
					for dx := -searchRadius; dx <= searchRadius; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
							waterTiles[ny][nx] {
							tooClose = true
							break
						}
					}
					if tooClose {
						break
					}
				}

				if !tooClose {
					// Medium water body - just a single tile
					Map[y][x].altitude = -0.08 + detailValue
					Map[y][x].landType = LandType_Water
					waterTiles[y][x] = true
				}
			}
		}
	}
}

func countWaterTiles(waterTiles [MapHeight][MapWidth]bool) int {
	count := 0
	for y := range MapHeight {
		for x := range MapWidth {
			if waterTiles[y][x] {
				count++
			}
		}
	}
	return count
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
	springCount := 8 + rng.Intn(5)     // 8-12 springs
	marshCount := 12 + rng.Intn(8)     // 12-19 marshes
	smallPondCount := 10 + rng.Intn(5) // 10-14 small ponds

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
		// Springs often form at specific geological interfaces,
		// typically at hillsides, mountain bases, or where permeable
		// rock meets impermeable layers. They also require adequate
		// rainfall over the year.

		// Try to find a location at the base of higher elevation
		x := rng.Intn(MapWidth-2) + 1
		y := rng.Intn(MapHeight-2) + 1

		// Check annual and seasonal rainfall
		avgRainfall := Map[y][x].climate.avgRain
		winterRainfall := Map[y][x].climate.winterRain

		// Calculate the minimum seasonal rainfall (sources need year-round water)
		minSeasonalRain := math.Min(
			math.Min(Map[y][x].climate.winterRain, Map[y][x].climate.springRain),
			math.Min(Map[y][x].climate.summerRain, Map[y][x].climate.fallRain),
		)

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
			avgRainfall > 0.3 && // Need moderate annual rainfall
			minSeasonalRain > 0.2 && // Need some rain even in dry season
			Map[y][x].landType != LandType_Mountains &&
			Map[y][x].landType != LandType_Water {

			// Higher rainfall increases spring chance
			// Snow in winter can feed springs in spring (snowmelt)
			springChance := 0.4 + avgRainfall*0.4 // 0.4 to 0.8 based on rainfall

			// Winter precipitation as snow (cold + wet winters) helps springs
			if winterRainfall > 0.4 && Map[y][x].climate.winterTemp < 0.3 {
				springChance += 0.2 // Snowmelt bonus
			}

			isGoodSpringLocation = rng.Float64() < springChance

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

	// 2. Generate marshes (soggy areas) - heavily influenced by seasonal rainfall
	marshesGenerated := 0

	for attempts := 0; attempts < 200 && marshesGenerated < marshCount; attempts++ {
		// Marshes typically form in low-lying areas with poor drainage and high rainfall,
		// Or areas with high water tables (near rivers/streams).

		// Try to find a location for a marsh
		x := rng.Intn(MapWidth-2) + 1
		y := rng.Intn(MapHeight-2) + 1

		// Examine the seasonal rainfall pattern
		springRain := Map[y][x].climate.springRain
		summerRain := Map[y][x].climate.summerRain
		fallRain := Map[y][x].climate.fallRain
		avgRain := Map[y][x].climate.avgRain

		// Adjust marsh count based on overall map rainfall
		if avgRain > 0.7 {
			// More marshes in very wet regions
			marshCount += 2
		} else if avgRain < 0.4 {
			// Fewer marshes in dry regions
			marshCount = max(5, marshCount-2)
		}

		// Good marsh locations: low-lying areas, near water, flat terrain, high rainfall
		isGoodMarshLocation := false
		elevation := Map[y][x].altitude

		// Marshes need significant rainfall for at least part of the year
		if avgRain < 0.4 && math.Max(springRain, math.Max(summerRain, fallRain)) < 0.6 {
			continue // Too dry for a marsh year-round
		}

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
			avgRain > 0.4 && // Sufficient average rainfall
			Map[y][x].landType != LandType_Mountains &&
			Map[y][x].landType != LandType_Plateaus {

			// Seasonal marshes form in areas with wet and dry seasons
			// Permanent marshes need consistently high rainfall

			// Calculate seasonal variation in rainfall
			rainVariation := math.Max(
				math.Max(springRain, math.Max(summerRain, fallRain))-
					math.Min(Map[y][x].climate.winterRain, math.Min(springRain, math.Min(summerRain, fallRain))),
				0.1, // Minimum variation to avoid division by zero
			)

			// Base chance influenced by rainfall and seasonality
			marshChance := 0.0

			// Permanent marshes (year-round wet)
			if avgRain > 0.65 && rainVariation < 0.3 {
				marshChance = 0.7 // High chance for permanent marsh
			} else if avgRain > 0.5 {
				// Seasonal marshes (wet season/dry season)
				marshChance = 0.5
			} else if math.Max(springRain, math.Max(summerRain, fallRain)) > 0.7 {
				// Very seasonal marsh (only during wet season)
				marshChance = 0.4
			} else {
				marshChance = 0.2 // Low baseline chance
			}

			// Bonus for being near water
			if nearWater {
				marshChance += 0.2
			}

			// Bonus for flat terrain
			if isFlat {
				marshChance += 0.2
			}

			// Cap the chance
			marshChance = math.Min(marshChance, 0.9)

			isGoodMarshLocation = rng.Float64() < marshChance

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
					if dx == 0 && dy == 0 { // TODO: I might want to remove this.
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

		// Get climate data
		avgRain := Map[y][x].climate.avgRain

		// Seasonal rainfall variation affects pond persistence
		winterRain := Map[y][x].climate.winterRain
		summerRain := Map[y][x].climate.summerRain
		rainVariation := math.Abs(winterRain - summerRain)

		// Check if this is a suitable location for a pond
		if !waterFeaturePlaced[y][x] &&
			Map[y][x].altitude > 0.05 && Map[y][x].altitude < 0.5 &&
			avgRain > 0.4 && // Need reasonable rainfall
			Map[y][x].landType != LandType_Mountains &&
			Map[y][x].landType != LandType_Plateaus {

			// Determine if pond is permanent or seasonal // TODO
			// isPermanent := avgRain > 0.6 && rainVariation < 0.3

			// Base chance for pond formation
			pondChance := 0.0

			// Higher chance in valleys or near marshes
			if Map[y][x].landType == LandType_Valleys {
				pondChance = 0.6 // High chance in valleys
			} else {
				pondChance = 0.3 // Base chance
			}

			// Adjust based on rainfall
			pondChance += avgRain * 0.2

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
				pondChance += 0.2 // Good chance near marshes
			}

			// Apply seasonal factor - highly seasonal rainfall makes ponds less likely
			if rainVariation > 0.5 {
				pondChance -= 0.2 // High variation reduces pond chance
			}

			// Cap the chance
			pondChance = math.Max(0.1, math.Min(pondChance, 0.8))

			if rng.Float64() < pondChance {
				// Set the pond flag
				Map[y][x].hasPond = true

				// Mark as placed to avoid overlaps
				waterFeaturePlaced[y][x] = true
				pondsGenerated++
			}
		}
	}

	// 4. Generate small rivers (streams)
	// 4. Generate seasonal streams based on rainfall patterns
	generateSeasonalStreams(seed, springLocations, waterFeaturePlaced)
}

func generateSeasonalStreams(seed int64, springLocations []struct{ x, y int }, waterFeaturePlaced [MapHeight][MapWidth]bool) {
	rng := rand.New(rand.NewSource(seed + 8888))

	// Base parameters
	baseSmallRiverCount := 10 + rng.Intn(8) // 10-17 small rivers

	// Track permanent and seasonal streams separately
	streamsGenerated := 0
	seasonalStreamsGenerated := 0

	// First try to place some streams starting from springs
	for _, spring := range springLocations {
		// Limit the number of streams
		if streamsGenerated >= baseSmallRiverCount {
			break
		}

		// Check rainfall at spring
		avgRain := Map[spring.y][spring.x].climate.avgRain
		/*minSeasonalRain := math.Min(
		math.Min(Map[spring.y][spring.x].climate.winterRain, Map[spring.y][spring.x].climate.springRain),
		math.Min(Map[spring.y][spring.x].climate.summerRain, Map[spring.y][spring.x].climate.fallRain),
		)*/

		// Determine if this will be a permanent or seasonal stream // TODO
		// isPermanent := avgRain > 0.55 && minSeasonalRain > 0.3

		// Only some springs form streams
		streamChance := 0.5 + avgRain*0.4
		if rng.Float64() < streamChance {
			// First, mark the spring tile as having a stream too
			Map[spring.y][spring.x].hasStream = true

			// Trace a path downhill from the spring
			streamPath := traceSmallStreamPath(spring.x, spring.y, rng, waterFeaturePlaced)

			// If we found a valid path of appropriate length
			if len(streamPath) >= 2 && len(streamPath) <= 8 {
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

	// Calculate how many additional seasonal streams to generate based on map conditions
	mapWetness := 0.0
	mapCount := 0

	for y := range MapHeight {
		for x := range MapWidth {
			if Map[y][x].altitude > 0 {
				mapWetness += Map[y][x].climate.avgRain
				mapCount++
			}
		}
	}

	averageMapWetness := 0.5
	if mapCount > 0 {
		averageMapWetness = mapWetness / float64(mapCount)
	}

	// More seasonal streams in wetter maps
	seasonalStreamTarget := 5 + int(averageMapWetness*15)

	// Generate seasonal streams in suitable locations
	for attempts := 0; attempts < 200 && seasonalStreamsGenerated < seasonalStreamTarget; attempts++ {
		// Choose a random location for the stream source
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Examine seasonal rainfall
		springRain := Map[y][x].climate.springRain
		summerRain := Map[y][x].climate.summerRain
		fallRain := Map[y][x].climate.fallRain
		winterRain := Map[y][x].climate.winterRain

		// Calculate peak seasonal rainfall
		peakRainfall := math.Max(springRain, math.Max(summerRain, math.Max(fallRain, winterRain)))

		// Check if this is a suitable location for a seasonal stream source
		if !waterFeaturePlaced[y][x] &&
			Map[y][x].altitude > 0.3 && Map[y][x].altitude < 0.8 &&
			peakRainfall > 0.6 && // Need high rainfall in at least one season
			Map[y][x].landType != LandType_Mountains &&
			Map[y][x].landType != LandType_Plateaus {

			// Trace a short path downhill
			streamPath := traceSmallStreamPath(x, y, rng, waterFeaturePlaced)

			// If we found a valid path of appropriate length
			if len(streamPath) >= 3 && len(streamPath) <= 8 {
				// Apply the stream to the map
				for _, point := range streamPath {
					sx, sy := point.x, point.y

					// Set the stream flag - these will be seasonal
					Map[sy][sx].hasStream = true

					// Mark as placed to avoid overlaps
					waterFeaturePlaced[sy][sx] = true
				}

				seasonalStreamsGenerated++
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
	groveCount := 20 + rng.Intn(10)  // 20-29 tree groves
	meadowCount := 15 + rng.Intn(10) // 15-24 flower meadows
	scrubCount := 25 + rng.Intn(15)  // 25-39 scrubland patches
	rockCount := 10 + rng.Intn(8)    // 10-17 rock outcroppings
	saltFlatCount := 3 + rng.Intn(3) // 3-5 salt flats

	// Track places where we've already placed features
	var featurePlaced [MapHeight][MapWidth]bool

	// Mark existing water and special features as unavailable
	for y := range MapHeight {
		for x := range MapWidth {
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

		// Get climate data
		avgTemp := Map[y][x].climate.avgTemp
		avgRain := Map[y][x].climate.avgRain

		// Look at seasonal variations
		tempRange := math.Max(
			Map[y][x].climate.summerTemp-Map[y][x].climate.winterTemp,
			0.1, // Avoid division by zero
		)

		// Calculate biome suitability for trees
		// Trees need adequate rainfall and moderate temperatures
		treeSuitability := 0.0

		if avgRain > 0.4 && avgRain < 0.9 && avgTemp > 0.3 && avgTemp < 0.8 {
			// Ideal conditions: moderate rainfall and temperature
			treeSuitability = 0.8
		} else if avgRain > 0.6 {
			// High rainfall can support trees even at temperature extremes
			treeSuitability = 0.6
		} else if avgRain > 0.3 && avgTemp > 0.4 && avgTemp < 0.7 {
			// Marginal conditions
			treeSuitability = 0.4
		} else {
			// Poor conditions
			treeSuitability = 0.2
		}

		// Extreme seasonal temperature variation makes trees less likely
		if tempRange > 0.5 {
			treeSuitability *= 0.8
		}

		// Check if this is a suitable spot for a grove
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.1 && Map[y][x].altitude < 0.7 {

			// Calculate grove chance based on climate suitability
			groveChance := treeSuitability

			// Check if near water, since they are more likely near water sources.
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
				groveChance += 0.2 // Higher chance near water
			}

			// Cap the chance
			groveChance = math.Min(groveChance, 0.9)

			if rng.Float64() < groveChance {
				Map[y][x].hasGrove = true
				featurePlaced[y][x] = true

				// Some groves form small clusters - larger in fertile areas
				clusterSize := 1
				if avgRain > 0.6 && avgTemp > 0.4 && avgTemp < 0.7 {
					// More fertile conditions produce larger groves
					clusterSize = 1 + rng.Intn(3)
				} else {
					// Less optimal conditions produce smaller groves
					clusterSize = 1 + rng.Intn(2)
				}

				// Try to add additional grove tiles
				for range clusterSize {
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

				grovesGenerated++
			}
		}
	}

	// 2. Generate meadows (flower-rich areas) - affected by climate
	meadowsGenerated := 0

	for attempts := 0; attempts < 100 && meadowsGenerated < meadowCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Get climate data
		// avgTemp := Map[y][x].climate.avgTemp
		springTemp := Map[y][x].climate.springTemp
		summerTemp := Map[y][x].climate.summerTemp
		springRain := Map[y][x].climate.springRain
		summerRain := Map[y][x].climate.summerRain

		// Meadows flourish in spring and early summer with moderate temps and good rainfall
		meadowSuitability := 0.0

		// Ideal conditions for meadows: warm spring/summer with adequate rainfall
		if springTemp > 0.3 && springTemp < 0.7 &&
			summerTemp > 0.4 && summerTemp < 0.8 &&
			springRain > 0.4 && summerRain > 0.3 {
			meadowSuitability = 0.8
		} else if springRain > 0.5 || summerRain > 0.5 {
			// Good rainfall can support meadows in less optimal temperatures
			meadowSuitability = 0.5
		} else {
			// Marginal conditions
			meadowSuitability = 0.3
		}

		// Check if this is a suitable spot for a meadow
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.1 && Map[y][x].altitude < 0.6 {

			// Calculate meadow chance based on climate suitability
			meadowChance := meadowSuitability

			// Meadows often form in valleys or near water
			if Map[y][x].landType == LandType_Valleys {
				meadowChance += 0.2
			}

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
				meadowChance += 0.15
			}

			// Cap the chance
			meadowChance = math.Min(meadowChance, 0.9)

			if rng.Float64() < meadowChance {
				Map[y][x].hasMeadow = true
				featurePlaced[y][x] = true
				meadowsGenerated++
			}
		}
	}

	// 3. Generate scrubland (areas with brush and small woody plants) - climate-affected
	scrubGenerated := 0

	for attempts := 0; attempts < 150 && scrubGenerated < scrubCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Get climate data
		avgTemp := Map[y][x].climate.avgTemp
		avgRain := Map[y][x].climate.avgRain
		summerTemp := Map[y][x].climate.summerTemp

		// Scrubland occurs in drier, often warmer conditions than forest
		// But still needs some rainfall
		scrubSuitability := 0.0

		// Ideal conditions for scrubland: warm and somewhat dry
		if avgRain > 0.25 && avgRain < 0.55 && avgTemp > 0.45 {
			scrubSuitability = 0.9
		} else if avgRain > 0.2 && avgRain < 0.7 {
			// Moderate conditions
			scrubSuitability = 0.6
		} else {
			// Marginal conditions
			scrubSuitability = 0.3
		}

		// Very hot summers favor scrubland over forest
		if summerTemp > 0.7 {
			scrubSuitability += 0.1
		}

		// Check if this is a suitable spot for scrubland
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.2 && Map[y][x].altitude < 0.7 {

			// Calculate scrub chance based on climate suitability
			scrubChance := scrubSuitability

			// Cap the chance
			scrubChance = math.Min(scrubChance, 0.9)

			if rng.Float64() < scrubChance {
				Map[y][x].hasScrub = true
				featurePlaced[y][x] = true

				// Scrubland often forms larger patches
				// Size depends on climate - larger patches in optimal conditions
				patchSize := 1
				if avgRain > 0.3 && avgRain < 0.5 && avgTemp > 0.5 {
					// Ideal conditions - larger patches
					patchSize = 2 + rng.Intn(4)
				} else {
					// Less optimal - smaller patches
					patchSize = 1 + rng.Intn(3)
				}

				// Try to add additional scrub tiles
				for range patchSize {
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

				scrubGenerated++
			}
		}
	}

	// 4. Generate rock outcroppings - less affected by climate, more by geology
	rocksGenerated := 0

	for attempts := 0; attempts < 100 && rocksGenerated < rockCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Check if this is a suitable spot for exposed rocks
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.3 && Map[y][x].altitude < 0.8 {

			// Rocks are more common at higher elevations and in drier areas
			rockChance := 0.3 // Base chance

			// Higher elevation increases chance
			if Map[y][x].altitude > 0.6 {
				rockChance += 0.2
			}

			// Drier climates have more exposed rock
			if Map[y][x].climate.avgRain < 0.4 {
				rockChance += 0.2
			}

			// Extremes of temperature (freeze/thaw cycles) can create more exposed rock
			tempRange := math.Abs(Map[y][x].climate.summerTemp - Map[y][x].climate.winterTemp)
			if tempRange > 0.4 {
				rockChance += 0.1
			}

			// Areas with seasonal flooding tend to expose rocks
			if Map[y][x].hasFloodArea {
				rockChance += 0.1
			}

			// Cap the chance
			rockChance = math.Min(rockChance, 0.8)

			if rng.Float64() < rockChance {
				Map[y][x].hasRocks = true
				featurePlaced[y][x] = true
				rocksGenerated++
			}
		}
	}

	// 5. Generate seasonal flood areas
	generateFloodAreas(seed)

	// 6. Generate salt flats - highly climate dependent (and dependent on seasonal flood areas)
	saltFlatsGenerated := 0

	for attempts := 0; attempts < 50 && saltFlatsGenerated < saltFlatCount; attempts++ {
		x := rng.Intn(MapWidth)
		y := rng.Intn(MapHeight)

		// Salt flats need hot, dry conditions and are often seasonal
		avgTemp := Map[y][x].climate.avgTemp
		avgRain := Map[y][x].climate.avgRain
		summerTemp := Map[y][x].climate.summerTemp

		// Salt flats need high evaporation (hot) and low rainfall
		saltFlatSuitability := 0.0

		if avgRain < 0.3 && summerTemp > 0.7 {
			// Ideal conditions: very hot and dry
			saltFlatSuitability = 0.8
		} else if avgRain < 0.4 && avgTemp > 0.6 {
			// Moderate conditions
			saltFlatSuitability = 0.4
		} else {
			// Poor conditions
			saltFlatSuitability = 0.1
		}

		// Check if this is a suitable spot for a salt flat
		if !featurePlaced[y][x] &&
			Map[y][x].landType == LandType_Plains &&
			Map[y][x].altitude > 0.15 && Map[y][x].altitude < 0.4 {

			// Calculate salt flat chance
			saltFlatChance := saltFlatSuitability

			// Depression or basin increases chance
			if Map[y][x].landType == LandType_Valleys && Map[y][x].altitude < 0.3 {
				saltFlatChance += 0.2
			}

			// Areas with seasonal flooding but high evaporation develop salt flats
			if Map[y][x].hasFloodArea && summerTemp > 0.7 {
				saltFlatChance += 0.3
			}

			// Cap the chance
			saltFlatChance = math.Min(saltFlatChance, 0.8)

			if rng.Float64() < saltFlatChance {
				Map[y][x].hasSaltFlat = true
				featurePlaced[y][x] = true

				// Salt flats can form small patches
				if rng.Float64() < 0.5 {
					// Try to add 1-2 adjacent salt flat tiles
					extraSalt := 1 + rng.Intn(2)
					for range extraSalt {
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

func generateFloodAreas(seed int64) {
	rng := rand.New(rand.NewSource(seed + 3333))

	// Parameters for flood areas
	floodAreaCount := 2 + rng.Intn(4) // 2-5 flood regions based on climate

	// Track all water sources (rivers, lakes, streams)
	var waterSources []struct{ x, y int }

	for y := range MapHeight {
		for x := range MapWidth {
			// Only include major water bodies (no streams)
			// Seasonal floods typically come from larger bodies of water
			if Map[y][x].altitude <= 0 { // Only major water bodies
				waterSources = append(waterSources, struct{ x, y int }{x, y})
			}
		}
	}

	// If we don't have any water sources, we can't have floods
	if len(waterSources) == 0 {
		return
	}

	// Shuffle water sources so we don't always start floods from the same places
	for i := len(waterSources) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		waterSources[i], waterSources[j] = waterSources[j], waterSources[i]
	}

	// Areas we've already checked for flooding potential
	var checkedTiles [MapHeight][MapWidth]bool

	// Generate each flood region starting from a suitable water source
	floodAreasGenerated := 0
	for i := 0; i < len(waterSources) && floodAreasGenerated < floodAreaCount; i++ {
		source := waterSources[i]
		x, y := source.x, source.y

		// Ensure this water source hasn't been checked before
		if checkedTiles[y][x] {
			continue
		}

		checkedTiles[y][x] = true

		// Skip if this water tile doesn't have enough adjacent water tiles
		// This ensures floods only come from substantial water bodies
		adjacentWaterCount := 0
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
					Map[ny][nx].altitude <= 0 { // Adjacent water
					adjacentWaterCount++
				}
			}
		}

		// Require at least 3 adjacent water tiles
		// This means floods only start from water "edges" not single water tiles
		if adjacentWaterCount < 3 {
			continue // Not a substantial enough water body
		}

		// Check for seasonal rainfall variation in nearby land
		// We need an area with significant seasonal variation in rainfall
		// to create natural flood cycles

		highestSeasonalRain := 0.0
		lowestSeasonalRain := 1.0
		neighborCount := 0

		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
					Map[ny][nx].altitude > 0 {

					// Check seasonal rainfall
					winterRain := Map[ny][nx].climate.winterRain
					springRain := Map[ny][nx].climate.springRain
					summerRain := Map[ny][nx].climate.summerRain
					fallRain := Map[ny][nx].climate.fallRain

					// Find highest and lowest seasonal rainfall
					seasonalMax := math.Max(winterRain, math.Max(springRain, math.Max(summerRain, fallRain)))
					seasonalMin := math.Min(winterRain, math.Min(springRain, math.Min(summerRain, fallRain)))

					if seasonalMax > highestSeasonalRain {
						highestSeasonalRain = seasonalMax
					}

					if seasonalMin < lowestSeasonalRain {
						lowestSeasonalRain = seasonalMin
					}

					neighborCount++
				}
			}
		}

		// If we didn't find neighboring land, skip this source
		if neighborCount == 0 {
			continue
		}

		// Calculate seasonal variation
		seasonalVariation := highestSeasonalRain - lowestSeasonalRain

		// Flood regions typically develop in areas with significant seasonal variation
		// or very high rainfall in at least one season
		if seasonalVariation < 0.25 && highestSeasonalRain < 0.7 {
			continue // Not enough seasonal variation for flooding
		}

		// Flood regions can only form from water sources near low-lying land
		// hasLowLand := false
		hasVeryLowLand := false

		// Check if this water source has adjacent low-lying land
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				nx, ny := x+dx, y+dy

				if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
					Map[ny][nx].altitude > 0 && Map[ny][nx].altitude < 0.3 &&
					Map[ny][nx].landType != LandType_Mountains &&
					Map[ny][nx].landType != LandType_Plateaus {
					// hasLowLand = true

					if Map[ny][nx].altitude < 0.15 {
						hasVeryLowLand = true
					}
				}
			}
		}

		if !hasVeryLowLand {
			continue // This water source isn't suitable for flooding
		}

		// This water source is good for flooding. Generate a connected flood region
		// Seasonal rain data is included in the flood region calculation
		floodTiles := generateConnectedFloodRegion(x, y, rng)

		// If we found enough flood tiles, mark it as a flood region
		if len(floodTiles) >= 3 {
			// Apply flood area to the map
			for _, tile := range floodTiles {
				Map[tile.y][tile.x].hasFloodArea = true

				// Mark a wide radius as checked to avoid too-close flood regions
				for dy := -4; dy <= 4; dy++ {
					for dx := -4; dx <= 4; dx++ {
						nx, ny := tile.x+dx, tile.y+dy
						if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
							checkedTiles[ny][nx] = true
						}
					}
				}
			}

			floodAreasGenerated++
		}
	}
}

// Helper function to generate a connected flood region from a water source
func generateConnectedFloodRegion(waterX, waterY int, rng *rand.Rand) []struct{ x, y int } {
	// Define a flood fill algorithm that prioritizes lower elevation
	var floodTiles []struct{ x, y int }
	var processed [MapHeight][MapWidth]bool

	// Queue for flood fill algorithm
	queue := []struct {
		x, y     int
		distance int // Distance from water source
	}{
		{x: waterX, y: waterY, distance: 0},
	}

	processed[waterY][waterX] = true

	// Base max distance
	baseMaxDistance := 5 + rng.Intn(7) // 5-11 tiles maximum flood distance

	// Adjust max distance based on seasonal rainfall
	// Find the season with highest rainfall near the water source
	highestSeasonalRain := 0.0

	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			nx, ny := waterX+dx, waterY+dy
			if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight && Map[ny][nx].altitude > 0 {
				winterRain := Map[ny][nx].climate.winterRain
				springRain := Map[ny][nx].climate.springRain
				summerRain := Map[ny][nx].climate.summerRain
				fallRain := Map[ny][nx].climate.fallRain

				seasonalMax := math.Max(winterRain, math.Max(springRain, math.Max(summerRain, fallRain)))
				if seasonalMax > highestSeasonalRain {
					highestSeasonalRain = seasonalMax
				}
			}
		}
	}

	// Adjust flood distance based on seasonal rainfall
	maxDistance := baseMaxDistance
	if highestSeasonalRain > 0.7 {
		// Higher rainfall seasons create more extensive floods
		maxDistance += 3
	} else if highestSeasonalRain < 0.4 {
		// Low rainfall creates smaller flood zones
		maxDistance -= 2
	}

	// Ensure minimum size
	maxDistance = max(maxDistance, 4)

	// Process queue
	for len(queue) > 0 {
		// Get next tile
		current := queue[0]
		queue = queue[1:]

		// Skip water tiles (we're looking for land to flood)
		if Map[current.y][current.x].altitude <= 0 {
			// But water is part of the flood system, so check its neighbors
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := current.x+dx, current.y+dy

					// Check bounds
					if nx < 0 || nx >= MapWidth || ny < 0 || ny >= MapHeight {
						continue
					}

					// Skip processed tiles
					if processed[ny][nx] {
						continue
					}

					// Add to queue
					queue = append(queue, struct{ x, y, distance int }{
						x:        nx,
						y:        ny,
						distance: current.distance + 1,
					})

					processed[ny][nx] = true
				}
			}
			continue
		}

		// We've found a land tile
		// Check if it's suitable for flooding and not too far from water
		if current.distance <= maxDistance &&
			Map[current.y][current.x].altitude < 0.3 &&
			Map[current.y][current.x].landType != LandType_Mountains &&
			Map[current.y][current.x].landType != LandType_Plateaus {
			/*!Map[current.y][current.x].hasGrove*/ // TODO: Do trees typically don't grow in flood zones? Certain species can, afaik

			// Get local rainfall information
			winterRain := Map[current.y][current.x].climate.winterRain
			springRain := Map[current.y][current.x].climate.springRain
			summerRain := Map[current.y][current.x].climate.summerRain
			fallRain := Map[current.y][current.x].climate.fallRain

			// Find highest seasonal rainfall
			highestRain := math.Max(winterRain, math.Max(springRain, math.Max(summerRain, fallRain)))

			// The likelihood of flooding decreases with elevation and distance
			baseChance := 0.75
			elevationEffect := 0.9
			distanceEffect := 0.5

			// Adjust flood chance based on seasonal rainfall
			rainfallEffect := 0.5 + highestRain*0.5 // 0.5 to 1.0 based on rainfall

			floodChance := baseChance - (Map[current.y][current.x].altitude/0.3)*elevationEffect - (float64(current.distance)/float64(maxDistance))*distanceEffect
			floodChance *= rainfallEffect // Scale by rainfall

			// Additional factor: valleys are more likely to flood
			if Map[current.y][current.x].landType == LandType_Valleys {
				floodChance += 0.2
			}

			// Steep falloff at the edges
			if float64(current.distance) > float64(maxDistance)*0.7 {
				floodChance *= 0.5 // Flood probability drops sharply at edges
			}

			// Apply randomness
			if rng.Float64() < floodChance {
				// This tile gets flooded
				floodTiles = append(floodTiles, struct{ x, y int }{
					x: current.x,
					y: current.y,
				})

				// Add neighbors to queue to continue the flood
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := current.x+dx, current.y+dy

						// Check bounds
						if nx < 0 || nx >= MapWidth || ny < 0 || ny >= MapHeight {
							continue
						}

						// Skip processed tiles
						if processed[ny][nx] {
							continue
						}

						// Add to queue with increased distance
						queue = append(queue, struct{ x, y, distance int }{
							x:        nx,
							y:        ny,
							distance: current.distance + 1,
						})

						processed[ny][nx] = true
					}
				}
			}
		}
	}

	return floodTiles
}

func generateGameTrails(seed int64) {
	rng := rand.New(rand.NewSource(seed + 8765))

	// Parameters
	gameTrailCount := 8 + rng.Intn(5) // 8-12 game trails

	// Game trails can cross any terrain except water
	trailsGenerated := 0

	for attempts := 0; attempts < 100 && trailsGenerated < gameTrailCount; attempts++ {
		// Game trails typically:
		// 1. Connect water sources to food sources (meadows, groves)
		// 2. Connect different biomes/terrain types
		// 3. Follow paths of least resistance (valleys, passes)
		// 4. May be seasonal depending on climate (migration routes)

		// Find a good starting point - typically water, meadows, or groves
		var startX, startY int
		foundStart := false

		// Try to find water first (most common starting point)
		waterSources := make([]struct{ x, y int }, 0)
		foodSources := make([]struct{ x, y int }, 0)

		for y := range MapHeight {
			for x := range MapWidth {
				// Water sources
				if Map[y][x].altitude <= 0 || Map[y][x].hasPond ||
					Map[y][x].hasStream || Map[y][x].hasSpring {
					waterSources = append(waterSources, struct{ x, y int }{x, y})
				}

				// Food sources
				if Map[y][x].hasMeadow || Map[y][x].hasGrove {
					foodSources = append(foodSources, struct{ x, y int }{x, y})
				}
			}
		}

		// Determine if this is a seasonal migration route or regular game trail
		isMigrationRoute := rng.Float64() < 0.3 // 30% chance for seasonal migration routes

		// 75% of trails start at water, 25% at food sources
		if len(waterSources) > 0 && rng.Float64() < 0.75 {
			source := waterSources[rng.Intn(len(waterSources))]
			startX, startY = source.x, source.y
			foundStart = true
		} else if len(foodSources) > 0 {
			source := foodSources[rng.Intn(len(foodSources))]
			startX, startY = source.x, source.y
			foundStart = true
		}

		if !foundStart {
			// If no water or food sources, just pick a random non-water tile
			for range 20 {
				x := rng.Intn(MapWidth)
				y := rng.Intn(MapHeight)
				if Map[y][x].altitude > 0 {
					startX, startY = x, y
					foundStart = true
					break
				}
			}
		}

		if !foundStart {
			continue // Couldn't find a start point
		}

		// Now find a destination - typically a different type of area than the start
		// If we started at water, aim for food. If we started at food, aim for water.
		var targetX, targetY int
		foundTarget := false

		if isMigrationRoute {
			// For migration routes, we want to connect areas with different seasonal conditions
			// Animals migrate to find better conditions as seasons change

			// Get climate at start location
			startWinterTemp := Map[startY][startX].climate.winterTemp
			startSummerTemp := Map[startY][startX].climate.summerTemp

			// Find a suitable migration destination with different seasonal patterns
			for range 30 {
				x := rng.Intn(MapWidth)
				y := rng.Intn(MapHeight)

				// Skip water and points too close to start
				if Map[y][x].altitude <= 0 {
					continue
				}

				// Calculate distance
				dist := math.Sqrt(math.Pow(float64(x-startX), 2) + math.Pow(float64(y-startY), 2))
				if dist < 10 || dist > 30 {
					continue // Too close or too far
				}

				// Check for different seasonal patterns
				targetWinterTemp := Map[y][x].climate.winterTemp
				targetSummerTemp := Map[y][x].climate.summerTemp

				// Look for contrasting conditions (warmer winters or cooler summers)
				tempDiff := math.Abs(targetWinterTemp-startWinterTemp) +
					math.Abs(targetSummerTemp-startSummerTemp)

				if tempDiff > 0.3 {
					// Found a location with sufficiently different seasonal conditions
					targetX, targetY = x, y
					foundTarget = true
					break
				}
			}
		} else if len(waterSources) > 0 && len(foodSources) > 0 {
			// Regular game trail - connect water to food or different terrain

			// If we started at water, look for food
			isWaterStart := false
			for _, source := range waterSources {
				if source.x == startX && source.y == startY {
					isWaterStart = true
					break
				}
			}

			if isWaterStart && len(foodSources) > 0 {
				// Select a food source that's reasonably distant (8-20 tiles)
				validTargets := make([]struct{ x, y int }, 0)
				for _, target := range foodSources {
					dist := math.Sqrt(math.Pow(float64(target.x-startX), 2) +
						math.Pow(float64(target.y-startY), 2))
					if dist >= 8 && dist <= 25 {
						validTargets = append(validTargets, target)
					}
				}

				if len(validTargets) > 0 {
					target := validTargets[rng.Intn(len(validTargets))]
					targetX, targetY = target.x, target.y
					foundTarget = true
				}
			} else if !isWaterStart && len(waterSources) > 0 {
				// Select a water source that's reasonably distant
				validTargets := make([]struct{ x, y int }, 0)
				for _, target := range waterSources {
					dist := math.Sqrt(math.Pow(float64(target.x-startX), 2) +
						math.Pow(float64(target.y-startY), 2))
					if dist >= 8 && dist <= 25 {
						validTargets = append(validTargets, target)
					}
				}

				if len(validTargets) > 0 {
					target := validTargets[rng.Intn(len(validTargets))]
					targetX, targetY = target.x, target.y
					foundTarget = true
				}
			}
		}

		// If we couldn't find a specific target, just pick a different terrain type
		if !foundTarget {
			startTerrainType := Map[startY][startX].landType

			// Try to find a different terrain type
			for range 20 {
				x := rng.Intn(MapWidth)
				y := rng.Intn(MapHeight)

				// Ensure it's not water and different from start
				if Map[y][x].altitude > 0 && Map[y][x].landType != startTerrainType {
					dist := math.Sqrt(math.Pow(float64(x-startX), 2) +
						math.Pow(float64(y-startY), 2))
					if dist >= 8 && dist <= 25 {
						targetX, targetY = x, y
						foundTarget = true
						break
					}
				}
			}
		}

		if !foundTarget {
			continue // Couldn't find a target
		}

		// Now trace a path from start to target
		// Animals will follow the path of least resistance
		path := findGameTrailPath(startX, startY, targetX, targetY, rng)

		if len(path) >= 5 {
			// Apply the trail to the map
			for _, point := range path {
				x, y := point.x, point.y

				// Game trails can go through other features
				Map[y][x].hasGameTrail = true
			}

			trailsGenerated++
		}
	}
}

// Helper function to find a game trail path (uses A* with terrain costs and considers seasonal climate)
func findGameTrailPath(startX, startY, targetX, targetY int, rng *rand.Rand) []struct{ x, y int } {
	// Define terrain cost factors, including seasonal considerations
	// Animals will prefer easier paths (valleys) and avoid difficult ones (steep mountains)
	getTerrainCost := func(x, y int) float64 {
		// Water is impassable
		if Map[y][x].altitude <= 0 {
			return math.Inf(1)
		}

		// Get climate information
		avgTemp := Map[y][x].climate.avgTemp
		avgRain := Map[y][x].climate.avgRain

		// Base costs by terrain type
		var baseCost float64
		switch Map[y][x].landType {
		case LandType_Plains:
			baseCost = 1.0 // Easiest to traverse
		case LandType_Valleys:
			baseCost = 0.8 // Even easier (animals prefer valleys)
		case LandType_Hills:
			baseCost = 2.0 // Moderately difficult
		case LandType_Plateaus:
			baseCost = 1.5 // Somewhat difficult
		case LandType_Mountains:
			// Mountains are hard but not impossible
			if Map[y][x].altitude > 1.3 {
				baseCost = 10.0 // Very difficult (high peaks)
			} else {
				baseCost = 5.0 // Difficult but passable (lower mountains)
			}
		case LandType_Coastal:
			baseCost = 1.1 // Slightly harder than plains
		case LandType_SandDunes:
			baseCost = 2.5 // Difficult to traverse
		default:
			baseCost = 1.0
		}

		// Modify costs for specific features animals might prefer or avoid
		// Animals prefer areas with food and water
		if Map[y][x].hasGrove {
			baseCost *= 0.8 // Animals like cover
		}
		if Map[y][x].hasMeadow {
			baseCost *= 0.7 // Animals like food
		}
		if Map[y][x].hasStream || Map[y][x].hasPond || Map[y][x].hasSpring {
			baseCost *= 0.6 // Animals strongly prefer paths near water
		}

		// Animals avoid certain features
		if Map[y][x].hasRocks {
			baseCost *= 1.3 // Animals avoid rocky areas
		}
		if Map[y][x].isDesert && avgRain < 0.3 {
			baseCost *= 1.5 // Animals avoid very dry areas unless necessary
		}

		// Climate considerations
		// Animals prefer moderate temperatures
		if avgTemp < 0.2 || avgTemp > 0.8 {
			baseCost *= 1.2 // Avoid temperature extremes
		}

		// Extremely wet areas can be difficult to traverse
		if avgRain > 0.8 {
			baseCost *= 1.2
		}

		// Add some randomness (animals don't always take perfect paths)
		baseCost *= 0.9 + rng.Float64()*0.2

		return baseCost
	}

	// A* pathfinding algorithm
	type Node struct {
		x, y int
		g, f float64 // g = cost from start, f = g + heuristic
	}

	openSet := make(map[string]Node)
	closedSet := make(map[string]bool)

	// Add start node
	startNode := Node{
		x: startX,
		y: startY,
		g: 0,
		f: heuristic(startX, startY, targetX, targetY),
	}
	openSet[nodeKey(startX, startY)] = startNode

	// Define previous nodes to reconstruct path
	cameFrom := make(map[string]struct{ x, y int })

	// A* main loop
	for len(openSet) > 0 {
		// Find node with lowest f in open set
		var current Node
		lowestF := math.Inf(1)
		for _, node := range openSet {
			if node.f < lowestF {
				lowestF = node.f
				current = node
			}
		}

		// Remove current from open set
		delete(openSet, nodeKey(current.x, current.y))

		// Check if we reached the target
		if current.x == targetX && current.y == targetY {
			// Reconstruct path
			path := make([]struct{ x, y int }, 0)
			x, y := current.x, current.y
			for {
				path = append(path, struct{ x, y int }{x, y})
				prev, exists := cameFrom[nodeKey(x, y)]
				if !exists {
					break
				}
				x, y = prev.x, prev.y
			}

			// Reverse path (from start to target)
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}

			return path
		}

		// Mark current as processed
		closedSet[nodeKey(current.x, current.y)] = true

		// Check neighbors
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}

				nx, ny := current.x+dx, current.y+dy

				// Skip if out of bounds
				if nx < 0 || nx >= MapWidth || ny < 0 || ny >= MapHeight {
					continue
				}

				// Skip if already processed
				if closedSet[nodeKey(nx, ny)] {
					continue
				}

				// Get terrain cost
				terrainCost := getTerrainCost(nx, ny)
				if math.IsInf(terrainCost, 1) {
					continue // Impassable
				}

				// Calculate g score (cost from start)
				// Diagonal movement costs more
				moveCost := terrainCost
				if dx != 0 && dy != 0 {
					moveCost *= 1.414 // sqrt(2)
				}

				tentativeG := current.g + moveCost

				neighbor, exists := openSet[nodeKey(nx, ny)]
				if !exists {
					// New node
					neighbor = Node{
						x: nx,
						y: ny,
						g: tentativeG,
						f: tentativeG + heuristic(nx, ny, targetX, targetY),
					}
					openSet[nodeKey(nx, ny)] = neighbor
					cameFrom[nodeKey(nx, ny)] = struct{ x, y int }{current.x, current.y}
				} else if tentativeG < neighbor.g {
					// Better path found
					neighbor.g = tentativeG
					neighbor.f = tentativeG + heuristic(nx, ny, targetX, targetY)
					openSet[nodeKey(nx, ny)] = neighbor
					cameFrom[nodeKey(nx, ny)] = struct{ x, y int }{current.x, current.y}
				}
			}
		}
	}

	// No path found
	return nil
}

// Helper functions for A* pathfinding
func nodeKey(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

func heuristic(x1, y1, x2, y2 int) float64 {
	// Manhattan distance
	return float64(abs(x1-x2) + abs(y1-y2))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func generateDeserts(seed int64) {
	rng := rand.New(rand.NewSource(seed + 9876))

	// Create desert noise patterns
	// We'll use a combination of noises to create realistic desert distributions
	// desertNoise := perlin.NewPerlin(2.5, 3.0, 2, seed+444)
	// rainfallNoise := perlin.NewPerlin(3.0, 2.5, 2, seed+555)
	// temperatureNoise := perlin.NewPerlin(4.0, 2.0, 3, seed+666)

	// Parameters for desert generation
	//equator := MapHeight / 2        // The vertical center of the map is the equator
	//equatorWidth := MapHeight * 0.4 // Deserts typically form in bands north and south of equator

	// Process all map tiles to determine desert regions
	for y := range MapHeight {
		for x := range MapWidth {
			// Skip water tiles - deserts don't form on water
			if Map[y][x].altitude <= 0 {
				continue
			}

			// Desert determination is now direct based on temperature and rainfall
			// Get seasonal climate data
			avgTemp := Map[y][x].climate.avgTemp
			avgRain := Map[y][x].climate.avgRain

			// Get the hottest season's temperature and driest season's rainfall
			hottestTemp := math.Max(
				math.Max(Map[y][x].climate.winterTemp, Map[y][x].climate.springTemp),
				math.Max(Map[y][x].climate.summerTemp, Map[y][x].climate.fallTemp),
			)

			driestRain := math.Min(
				math.Min(Map[y][x].climate.winterRain, Map[y][x].climate.springRain),
				math.Min(Map[y][x].climate.summerRain, Map[y][x].climate.fallRain),
			)

			// Calculate seasonal variation
			tempRange := math.Abs(Map[y][x].climate.summerTemp - Map[y][x].climate.winterTemp)
			rainRange := math.Max(
				math.Max(Map[y][x].climate.winterRain, Map[y][x].climate.springRain),
				math.Max(Map[y][x].climate.summerRain, Map[y][x].climate.fallRain),
			) - driestRain

			// Desert criteria based on climate patterns
			// True deserts are hot and dry year-round
			isExtremeTrueDesert := avgTemp > 0.7 && avgRain < 0.2 && driestRain < 0.15 && hottestTemp > 0.8

			// Hot deserts have hot summers but can have cooler winters
			isHotDesert := hottestTemp > 0.75 && avgRain < 0.25 && driestRain < 0.2

			// Semi-desert/arid regions: somewhat hot and dry with seasonal variation
			isSemiDesert := avgTemp > 0.6 && avgRain < 0.35 && driestRain < 0.25

			// Cold deserts exist with extreme temperature ranges but consistent dryness
			isColdDesert := tempRange > 0.5 && avgTemp < 0.6 && avgRain < 0.3 && driestRain < 0.2

			if isExtremeTrueDesert || isHotDesert || isSemiDesert || isColdDesert {
				Map[y][x].isDesert = true

				// Determine desert intensity for feature adjustment
				desertIntensity := 0.0

				if isExtremeTrueDesert {
					desertIntensity = 0.9 // Extreme true desert (very strong effects)
				} else if isHotDesert {
					desertIntensity = 0.7 // Hot desert (strong effects)
				} else if isColdDesert {
					desertIntensity = 0.6 // Cold desert (moderate effects)
				} else if isSemiDesert {
					desertIntensity = 0.4 // Semi-desert (mild effects)
				}

				// Apply feature modifications based on desert intensity

				// 1. Water features (rarer in more intense deserts)
				// Streams can be seasonal in deserts
				if Map[y][x].hasStream {
					// Most streams in hot deserts are ephemeral (seasonal)
					if desertIntensity > 0.7 && rng.Float64() < 0.8 {
						Map[y][x].hasStream = false
					} else if desertIntensity > 0.5 && rng.Float64() < 0.6 {
						Map[y][x].hasStream = false
					} else if desertIntensity > 0.3 && rng.Float64() < 0.4 {
						Map[y][x].hasStream = false
					}
				}

				// Ponds are very rare in hot deserts, can exist in cold deserts
				if Map[y][x].hasPond {
					if isHotDesert && rng.Float64() < 0.9 {
						Map[y][x].hasPond = false
					} else if desertIntensity > 0.5 && rng.Float64() < 0.7 {
						Map[y][x].hasPond = false
					} else if desertIntensity > 0.3 && rng.Float64() < 0.5 {
						Map[y][x].hasPond = false
					}
				}

				// Springs can exist in deserts but are rare
				if Map[y][x].hasSpring {
					if desertIntensity > 0.7 && rng.Float64() < 0.7 {
						Map[y][x].hasSpring = false
					} else if desertIntensity > 0.5 && rng.Float64() < 0.5 {
						Map[y][x].hasSpring = false
					} else if desertIntensity > 0.3 && rng.Float64() < 0.3 {
						Map[y][x].hasSpring = false
					}
				}

				// Marshes are extremely rare in hot deserts
				if Map[y][x].hasMarsh {
					if desertIntensity > 0.7 && rng.Float64() < 0.95 {
						Map[y][x].hasMarsh = false
					} else if desertIntensity > 0.5 && rng.Float64() < 0.8 {
						Map[y][x].hasMarsh = false
					} else if desertIntensity > 0.3 && rng.Float64() < 0.6 {
						Map[y][x].hasMarsh = false
					}
				}

				// Flood areas can exist seasonally in deserts
				if Map[y][x].hasFloodArea {
					// Desert floods are typically flash floods
					// More likely if there's high seasonal rainfall variation
					if rainRange < 0.3 || (desertIntensity > 0.7 && rng.Float64() < 0.7) {
						Map[y][x].hasFloodArea = false
					} else if desertIntensity > 0.5 && rng.Float64() < 0.5 {
						Map[y][x].hasFloodArea = false
					}
				}

				// 2. Vegetation features
				// Groves are extremely rare in true deserts, possible in semi-arid regions
				if Map[y][x].hasGrove {
					if desertIntensity > 0.7 && rng.Float64() < 0.95 {
						Map[y][x].hasGrove = false
					} else if desertIntensity > 0.5 && rng.Float64() < 0.8 {
						Map[y][x].hasGrove = false
					} else if desertIntensity > 0.3 && rng.Float64() < 0.6 {
						Map[y][x].hasGrove = false
					}
				}

				// Meadows are extremely rare in true deserts
				if Map[y][x].hasMeadow {
					// Deserts can have spring wildflower blooms after rain
					// But persistent meadows are very rare
					if desertIntensity > 0.7 && rng.Float64() < 0.95 {
						Map[y][x].hasMeadow = false
					} else if desertIntensity > 0.5 && rng.Float64() < 0.85 {
						Map[y][x].hasMeadow = false
					} else if desertIntensity > 0.3 && rng.Float64() < 0.7 {
						Map[y][x].hasMeadow = false
					}
				}

				// 3. Add desert-specific features
				// Increase chance of scrubland and rocks in deserts
				if !Map[y][x].hasScrub && !Map[y][x].hasRocks && !Map[y][x].hasSaltFlat {
					// Desert vegetation pattern depends on desert type
					if isHotDesert || isExtremeTrueDesert {
						// Hot deserts have more scrub and rocks
						if rng.Float64() < 0.5 {
							Map[y][x].hasScrub = true
						} else if rng.Float64() < 0.4 {
							Map[y][x].hasRocks = true
						}
					} else if isColdDesert {
						// Cold deserts have more rocks and less vegetation
						if rng.Float64() < 0.6 {
							Map[y][x].hasRocks = true
						} else if rng.Float64() < 0.3 {
							Map[y][x].hasScrub = true
						}
					} else {
						// Semi-deserts have more vegetation
						if rng.Float64() < 0.6 {
							Map[y][x].hasScrub = true
						} else if rng.Float64() < 0.3 {
							Map[y][x].hasRocks = true
						}
					}

					// Salt flats in appropriate desert basins
					// More common in hot deserts with seasonal water
					if Map[y][x].altitude < 0.4 &&
						(isHotDesert || isExtremeTrueDesert) &&
						rng.Float64() < 0.2 {
						Map[y][x].hasSaltFlat = true
					}
				}

				// 4. Generate sand dunes in very dry, hot areas with appropriate terrain
				if (isHotDesert || isExtremeTrueDesert) &&
					Map[y][x].landType == LandType_Plains &&
					avgRain < 0.2 &&
					hottestTemp > 0.75 &&
					rng.Float64() < 0.4 {
					Map[y][x].landType = LandType_SandDunes
				}
			}
		}
	}
}

// generateClimate calculates realistic temperature and rainfall patterns for all seasons
func generateClimate(seed int64) {
	// rng := rand.New(rand.NewSource(seed + 54321))

	// Create noise patterns for base climate variation
	rainfallNoise := perlin.NewPerlin(2.5, 2.0, 3, seed+111)
	temperatureNoise := perlin.NewPerlin(3.0, 2.5, 2, seed+222)
	localVariationNoise := perlin.NewPerlin(5.0, 2.0, 2, seed+333)

	// Additional noise for seasonal variations
	winterNoise := perlin.NewPerlin(2.8, 2.0, 2, seed+444)
	summerNoise := perlin.NewPerlin(2.8, 2.0, 2, seed+555)
	springNoise := perlin.NewPerlin(2.8, 2.0, 2, seed+666)
	fallNoise := perlin.NewPerlin(2.8, 2.0, 2, seed+777)

	// Define the equator position
	equator := MapHeight / 2

	// Track all water bodies for rainfall calculations
	var waterBodies []struct{ x, y int }
	var rivers []struct{ x, y int }

	// Find all water bodies and rivers for rainfall influence
	for y := range MapHeight {
		for x := range MapWidth {
			if Map[y][x].altitude <= 0 {
				// Major water body (lake, ocean)
				waterBodies = append(waterBodies, struct{ x, y int }{x, y})
			} else if Map[y][x].landType == LandType_Water && Map[y][x].altitude <= 0 {
				// River
				rivers = append(rivers, struct{ x, y int }{x, y})
			}
		}
	}

	// Calculate climate factors for each tile
	for y := range MapHeight {
		for x := range MapWidth {
			// Determine which hemisphere this tile is in
			isNorthernHemisphere := y < equator

			// 1. BASE ANNUAL TEMPERATURE CALCULATION

			// Calculate latitude impact on temperature (equator is hottest)
			distanceFromEquator := math.Abs(float64(y-equator)) / float64(MapHeight/2)
			// Temperature decreases as you move away from equator (0.0-1.0)
			latitudeTemp := 1.0 - math.Pow(distanceFromEquator, 0.8)

			// Altitude impact on temperature (higher = colder)
			altitude := Map[y][x].altitude
			altitudeTemp := 0.0
			if altitude <= 0 {
				// Water bodies maintain more stable temperatures
				altitudeTemp = 0.6
			} else {
				// Temperature decreases with elevation
				altitudeTemp = math.Max(0, 1.0-altitude*0.8)
			}

			// Global temperature variation from noise
			tempVariation := (temperatureNoise.Noise2D(float64(x)/(MapWidth*0.4), float64(y)/(MapHeight*0.4)) + 1) / 2

			// Local temperature variation
			localTemp := (localVariationNoise.Noise2D(float64(x)/(MapWidth*0.1), float64(y)/(MapHeight*0.1)) + 1) / 6

			// Calculate base annual temperature (0.0 to 1.0 scale)
			baseTemp := (latitudeTemp*0.8 + altitudeTemp*0.2 + tempVariation*0.2) / 1.2 // Test out a new calculation for this
			// baseTemp := (latitudeTemp*0.6 + altitudeTemp*0.3 + tempVariation*0.3) / 1.2
			baseTemp += localTemp - 0.08 // Adjust range slightly

			// 2. BASE ANNUAL RAINFALL CALCULATION

			// Basic rainfall pattern from noise (continental precipitation patterns)
			continentalRainfall := (rainfallNoise.Noise2D(float64(x)/(MapWidth*0.3), float64(y)/(MapHeight*0.3)) + 1) / 2

			// Calculate proximity to water bodies (affects rainfall)
			proximityToWater := calculateWaterProximity(x, y, waterBodies, rivers)

			// Mountains increase rainfall on windward sides (simplified)
			orographicEffect := 0.0
			if Map[y][x].altitude > 0.7 && Map[y][x].altitude < 1.3 {
				// Higher elevations capture more moisture
				orographicEffect = (Map[y][x].altitude - 0.7) * 0.5
			}

			// Rain shadow effect - areas behind mountains get less rain
			rainShadowEffect := calculateRainShadow(x, y)

			// Calculate base rainfall (0.0 to 1.0 scale)
			baseRainfall := (continentalRainfall*0.3 + proximityToWater*0.4 + orographicEffect*0.2) / 0.9
			baseRainfall -= rainShadowEffect * 0.3

			// Adjust rainfall based on latitude (tropical and temperate regions get more rain)
			if distanceFromEquator < 0.3 || (distanceFromEquator > 0.6 && distanceFromEquator < 0.8) {
				baseRainfall *= 1.3
			}

			// Local rainfall variation
			localRain := (localVariationNoise.Noise2D(float64(x+50)/(MapWidth*0.05), float64(y+50)/(MapHeight*0.05)) + 1) / 10
			baseRainfall += localRain - 0.05

			// 3. CALCULATE SEASONAL VARIATIONS

			// Initialize climate structure
			var climate Climate

			// Seasonal temperature variations based on hemisphere
			// First, calculate raw seasonal temperatures (before hemisphere adjustment)

			// Seasonal temperature modifiers
			// These create the temperature difference between seasons
			summerModifier := 0.25  // Hotter in summer
			winterModifier := -0.25 // Colder in winter

			// Moderate spring/fall modifiers
			springModifier := 0.0 // Neutral in spring
			fallModifier := 0.0   // Neutral in fall

			// Apply local seasonal variations from noise
			winterLocalVariation := (winterNoise.Noise2D(float64(x)/(MapWidth*0.3), float64(y)/(MapHeight*0.3)) + 1) / 10
			summerLocalVariation := (summerNoise.Noise2D(float64(x)/(MapWidth*0.3), float64(y)/(MapHeight*0.3)) + 1) / 10
			springLocalVariation := (springNoise.Noise2D(float64(x)/(MapWidth*0.3), float64(y)/(MapHeight*0.3)) + 1) / 10
			fallLocalVariation := (fallNoise.Noise2D(float64(x)/(MapWidth*0.3), float64(y)/(MapHeight*0.3)) + 1) / 10

			// Calculate seasonal modifiers accounting for latitude
			// Seasonal variation increases with distance from equator
			seasonalRange := 0.2 + distanceFromEquator*0.4 // 0.2 at equator, up to 0.6 at poles
			winterMod := winterModifier*seasonalRange + winterLocalVariation - 0.05
			summerMod := summerModifier*seasonalRange + summerLocalVariation - 0.05
			springMod := springModifier*seasonalRange + springLocalVariation - 0.05
			fallMod := fallModifier*seasonalRange + fallLocalVariation - 0.05

			// Calculate raw seasonal temperatures (before hemisphere adjustment)
			rawWinterTemp := math.Max(0.0, math.Min(1.0, baseTemp+winterMod))
			rawSummerTemp := math.Max(0.0, math.Min(1.0, baseTemp+summerMod))
			rawSpringTemp := math.Max(0.0, math.Min(1.0, baseTemp+springMod))
			rawFallTemp := math.Max(0.0, math.Min(1.0, baseTemp+fallMod))

			// Now apply hemisphere-specific assignments
			if isNorthernHemisphere {
				// Northern hemisphere seasons are aligned with their names
				climate.winterTemp = rawWinterTemp
				climate.springTemp = rawSpringTemp
				climate.summerTemp = rawSummerTemp
				climate.fallTemp = rawFallTemp
			} else {
				// Southern hemisphere has opposite seasons
				climate.winterTemp = rawSummerTemp // Winter in S.H. = Summer in N.H.
				climate.springTemp = rawFallTemp   // Spring in S.H. = Fall in N.H.
				climate.summerTemp = rawWinterTemp // Summer in S.H. = Winter in N.H.
				climate.fallTemp = rawSpringTemp   // Fall in S.H. = Spring in N.H.
			}

			// 4. CALCULATE SEASONAL RAINFALL

			// Seasonal rainfall patterns
			// In most climates, there's a rainy season and a dry season
			// The timing varies by latitude and climate type

			// Apply water body influence more strongly to certain seasons
			// Water bodies moderate nearby temperature and increase rainfall
			waterInfluence := proximityToWater * 0.3

			// Base seasonal rainfall values
			baseWinterRain := baseRainfall
			baseSpringRain := baseRainfall
			baseSummerRain := baseRainfall
			baseFallRain := baseRainfall

			// Apply seasonal rainfall patterns based on latitude
			if distanceFromEquator < 0.25 {
				// Tropical pattern - wet and dry seasons, less seasonal variation
				// Typically wet summer, drier winter in tropics
				baseSummerRain += 0.2
				baseWinterRain -= 0.1
			} else if distanceFromEquator < 0.5 {
				// Subtropical pattern - typically dry summers, wet winters
				baseSummerRain -= 0.2
				baseWinterRain += 0.15
				baseSpringRain += 0.1
				baseFallRain += 0.05
			} else {
				// Temperate/Polar pattern - more evenly distributed, slightly wetter in spring/fall
				baseSpringRain += 0.1
				baseFallRain += 0.1
				baseSummerRain += 0.05
			}

			// Apply local seasonal rainfall variations
			winterRainVar := (winterNoise.Noise2D(float64(x+100)/(MapWidth*0.2), float64(y+100)/(MapHeight*0.2)) + 1) / 8
			summerRainVar := (summerNoise.Noise2D(float64(x+200)/(MapWidth*0.2), float64(y+200)/(MapHeight*0.2)) + 1) / 8
			springRainVar := (springNoise.Noise2D(float64(x+300)/(MapWidth*0.2), float64(y+300)/(MapHeight*0.2)) + 1) / 8
			fallRainVar := (fallNoise.Noise2D(float64(x+400)/(MapWidth*0.2), float64(y+400)/(MapHeight*0.2)) + 1) / 8

			// Add variation and water influence to seasonal rainfall
			rawWinterRain := math.Max(0.0, math.Min(1.0, baseWinterRain+winterRainVar-0.06+waterInfluence))
			rawSummerRain := math.Max(0.0, math.Min(1.0, baseSummerRain+summerRainVar-0.06+waterInfluence))
			rawSpringRain := math.Max(0.0, math.Min(1.0, baseSpringRain+springRainVar-0.06+waterInfluence))
			rawFallRain := math.Max(0.0, math.Min(1.0, baseFallRain+fallRainVar-0.06+waterInfluence))

			// Apply hemisphere-specific seasonal rainfall
			if isNorthernHemisphere {
				climate.winterRain = rawWinterRain
				climate.springRain = rawSpringRain
				climate.summerRain = rawSummerRain
				climate.fallRain = rawFallRain
			} else {
				// Southern hemisphere has opposite seasons
				climate.winterRain = rawSummerRain
				climate.springRain = rawFallRain
				climate.summerRain = rawWinterRain
				climate.fallRain = rawSpringRain
			}

			// 5. CALCULATE ANNUAL AVERAGES
			climate.avgTemp = (climate.winterTemp + climate.springTemp + climate.summerTemp + climate.fallTemp) / 4.0
			climate.avgRain = (climate.winterRain + climate.springRain + climate.summerRain + climate.fallRain) / 4.0

			// 6. APPLY WATER BODY ADJUSTMENTS
			// Water bodies have more stable temperatures year-round
			if Map[y][x].altitude <= 0 {
				// Moderate the seasonal extremes for water
				avgTemp := climate.avgTemp
				seasonalRange := 0.15 // Reduced seasonal variation for water

				// Water warms/cools more slowly than land
				if isNorthernHemisphere {
					climate.winterTemp = avgTemp - seasonalRange
					climate.summerTemp = avgTemp + seasonalRange
					// Spring is warming, fall is cooling - offset peaks
					climate.springTemp = avgTemp - seasonalRange*0.3
					climate.fallTemp = avgTemp + seasonalRange*0.3
				} else {
					climate.summerTemp = avgTemp - seasonalRange
					climate.winterTemp = avgTemp + seasonalRange
					// Fall is warming, spring is cooling in southern hemisphere
					climate.fallTemp = avgTemp - seasonalRange*0.3
					climate.springTemp = avgTemp + seasonalRange*0.3
				}

				// Water doesn't need rainfall values (set to moderate for coastline influence)
				moderate := 0.6
				climate.winterRain = moderate
				climate.springRain = moderate
				climate.summerRain = moderate
				climate.fallRain = moderate
				climate.avgRain = moderate
			}

			// Save the climate data to the tile
			Map[y][x].climate = climate
		}
	}

	// Apply additional climate adjustments
	refineClimate()
}

// Calculate how close a tile is to water bodies (affects rainfall and temperature)
func calculateWaterProximity(x, y int, waterBodies, rivers []struct{ x, y int }) float64 {
	// Start with no water influence
	proximity := 0.0

	// Check proximity to major water bodies (lakes, oceans)
	// Water bodies have stronger influence than rivers
	for _, wb := range waterBodies {
		dist := math.Sqrt(math.Pow(float64(x-wb.x), 2) + math.Pow(float64(y-wb.y), 2))

		// Water influence drops with distance
		if dist < 12 {
			// Closer water has stronger influence
			influence := math.Max(0, 1.0-dist/12.0)
			proximity = math.Max(proximity, influence)
		}
	}

	// Check proximity to rivers (less influence than major water bodies)
	for _, r := range rivers {
		dist := math.Sqrt(math.Pow(float64(x-r.x), 2) + math.Pow(float64(y-r.y), 2))

		// River influence drops with distance but is weaker than lakes/oceans
		if dist < 5 {
			influence := math.Max(0, (1.0-dist/5.0)*0.7) // 70% as effective as major water
			proximity = math.Max(proximity, influence)
		}
	}

	// Also consider small water features
	if Map[y][x].hasStream || Map[y][x].hasPond || Map[y][x].hasSpring {
		proximity = math.Max(proximity, 0.4) // Local water sources have modest influence
	}

	return proximity
}

// Calculate rain shadow effect caused by mountains
func calculateRainShadow(x, y int) float64 {
	rainShadow := 0.0

	// Simplified approach: check if there are mountains to the "west" (left)
	// In a more realistic model, you would consider prevailing wind direction
	mountainCount := 0
	maxCheck := 8 // How far to look for mountains

	for dist := 1; dist <= maxCheck; dist++ {
		checkX := x - dist // Look "west"

		if checkX >= 0 && checkX < MapWidth {
			if Map[y][checkX].altitude >= 1.0 {
				// Found a mountain - stronger effect if closer
				mountainEffect := (float64(maxCheck) - float64(dist)) / float64(maxCheck)
				mountainCount++
				rainShadow += mountainEffect * 0.1
			}
		}
	}

	// Cap the rain shadow effect
	return math.Min(rainShadow, 0.6)
}

// Apply final adjustments to climate based on terrain features and land types
func refineClimate() {
	// Adjustments based on specific terrain features
	for y := range MapHeight {
		for x := range MapWidth {
			// Skip water tiles - they already have special handling
			if Map[y][x].altitude <= 0 {
				continue
			}

			// Valleys tend to collect moisture
			if Map[y][x].landType == LandType_Valleys {
				Map[y][x].climate.winterRain = math.Min(1.0, Map[y][x].climate.winterRain+0.1)
				Map[y][x].climate.springRain = math.Min(1.0, Map[y][x].climate.springRain+0.1)
				Map[y][x].climate.summerRain = math.Min(1.0, Map[y][x].climate.summerRain+0.1)
				Map[y][x].climate.fallRain = math.Min(1.0, Map[y][x].climate.fallRain+0.1)
				Map[y][x].climate.avgRain = math.Min(1.0, Map[y][x].climate.avgRain+0.1)

				// Valleys can be colder in winter (cold air sinks)
				Map[y][x].climate.winterTemp = math.Max(0.0, Map[y][x].climate.winterTemp-0.1)
			}

			// Plateaus tend to be drier than surrounding areas
			if Map[y][x].landType == LandType_Plateaus {
				Map[y][x].climate.winterRain = math.Max(0.0, Map[y][x].climate.winterRain-0.15)
				Map[y][x].climate.springRain = math.Max(0.0, Map[y][x].climate.springRain-0.15)
				Map[y][x].climate.summerRain = math.Max(0.0, Map[y][x].climate.summerRain-0.15)
				Map[y][x].climate.fallRain = math.Max(0.0, Map[y][x].climate.fallRain-0.15)
				Map[y][x].climate.avgRain = math.Max(0.0, Map[y][x].climate.avgRain-0.15)

				// Plateaus have more extreme temperatures (colder winters, hotter summers)
				Map[y][x].climate.winterTemp = math.Max(0.0, Map[y][x].climate.winterTemp-0.1)
				Map[y][x].climate.summerTemp = math.Min(1.0, Map[y][x].climate.summerTemp+0.1)
			}

			// Coastal areas have moderated temperatures and higher humidity
			if Map[y][x].landType == LandType_Coastal {
				// Moderate extreme temperatures (move toward middle)
				avgTemp := Map[y][x].climate.avgTemp

				// Reduce the difference between seasonal extremes
				Map[y][x].climate.winterTemp = Map[y][x].climate.winterTemp*0.7 + avgTemp*0.3
				Map[y][x].climate.summerTemp = Map[y][x].climate.summerTemp*0.7 + avgTemp*0.3
				Map[y][x].climate.springTemp = Map[y][x].climate.springTemp*0.8 + avgTemp*0.2
				Map[y][x].climate.fallTemp = Map[y][x].climate.fallTemp*0.8 + avgTemp*0.2

				// Increase rainfall across all seasons
				Map[y][x].climate.winterRain = math.Min(1.0, Map[y][x].climate.winterRain+0.15)
				Map[y][x].climate.springRain = math.Min(1.0, Map[y][x].climate.springRain+0.15)
				Map[y][x].climate.summerRain = math.Min(1.0, Map[y][x].climate.summerRain+0.15)
				Map[y][x].climate.fallRain = math.Min(1.0, Map[y][x].climate.fallRain+0.15)
				Map[y][x].climate.avgRain = math.Min(1.0, Map[y][x].climate.avgRain+0.15)
			}

			// Water bodies have moderated temperatures
			if Map[y][x].altitude <= 0 {
				// Water temperature changes more slowly than land
				// Moderate extreme temperatures (move toward middle)
				avgTemp := Map[y][x].climate.avgTemp

				// Reduce the difference between seasonal extremes
				Map[y][x].climate.winterTemp = Map[y][x].climate.winterTemp*0.7 + avgTemp*0.3
				Map[y][x].climate.summerTemp = Map[y][x].climate.summerTemp*0.7 + avgTemp*0.3
				Map[y][x].climate.springTemp = Map[y][x].climate.springTemp*0.8 + avgTemp*0.2
				Map[y][x].climate.fallTemp = Map[y][x].climate.fallTemp*0.8 + avgTemp*0.2

				// Increase rainfall across all seasons
				Map[y][x].climate.winterRain = math.Min(1.0, Map[y][x].climate.winterRain+0.15)
				Map[y][x].climate.springRain = math.Min(1.0, Map[y][x].climate.springRain+0.15)
				Map[y][x].climate.summerRain = math.Min(1.0, Map[y][x].climate.summerRain+0.15)
				Map[y][x].climate.fallRain = math.Min(1.0, Map[y][x].climate.fallRain+0.15)
				Map[y][x].climate.avgRain = math.Min(1.0, Map[y][x].climate.avgRain+0.15)
			}

			// Marshes are wetter
			if Map[y][x].hasMarsh {
				Map[y][x].climate.winterRain = math.Min(1.0, Map[y][x].climate.winterRain+0.25)
				Map[y][x].climate.springRain = math.Min(1.0, Map[y][x].climate.springRain+0.25)
				Map[y][x].climate.summerRain = math.Min(1.0, Map[y][x].climate.summerRain+0.25)
				Map[y][x].climate.fallRain = math.Min(1.0, Map[y][x].climate.fallRain+0.25)
				Map[y][x].climate.avgRain = math.Min(1.0, Map[y][x].climate.avgRain+0.25)
			}

			// SandDunes and desert areas are drier
			if Map[y][x].landType == LandType_SandDunes || Map[y][x].isDesert {
				Map[y][x].climate.winterRain = math.Max(0.0, Map[y][x].climate.winterRain-0.3)
				Map[y][x].climate.springRain = math.Max(0.0, Map[y][x].climate.springRain-0.3)
				Map[y][x].climate.summerRain = math.Max(0.0, Map[y][x].climate.summerRain-0.3)
				Map[y][x].climate.fallRain = math.Max(0.0, Map[y][x].climate.fallRain-0.3)
				Map[y][x].climate.avgRain = math.Max(0.0, Map[y][x].climate.avgRain-0.3)

				// Deserts have more extreme temperatures (very hot days, cold nights)
				// This is simplified to seasonal scale, but represents the high variation
				Map[y][x].climate.summerTemp = math.Min(1.0, Map[y][x].climate.summerTemp+0.2)
				Map[y][x].climate.winterTemp = math.Max(0.0, Map[y][x].climate.winterTemp-0.1)
			}
		}
	}

	// Apply smoothing to prevent unrealistic sharp transitions
	smoothClimate()
}

// Apply a smoothing pass to climate data to prevent unrealistic sharp transitions
func smoothClimate() {
	// Create temporary arrays for smoothed values (for each season)
	var tempWinter [MapHeight][MapWidth]float64
	var tempSpring [MapHeight][MapWidth]float64
	var tempSummer [MapHeight][MapWidth]float64
	var tempFall [MapHeight][MapWidth]float64

	var rainWinter [MapHeight][MapWidth]float64
	var rainSpring [MapHeight][MapWidth]float64
	var rainSummer [MapHeight][MapWidth]float64
	var rainFall [MapHeight][MapWidth]float64

	// Copy existing values
	for y := range MapHeight {
		for x := range MapWidth {
			tempWinter[y][x] = Map[y][x].climate.winterTemp
			tempSpring[y][x] = Map[y][x].climate.springTemp
			tempSummer[y][x] = Map[y][x].climate.summerTemp
			tempFall[y][x] = Map[y][x].climate.fallTemp

			rainWinter[y][x] = Map[y][x].climate.winterRain
			rainSpring[y][x] = Map[y][x].climate.springRain
			rainSummer[y][x] = Map[y][x].climate.summerRain
			rainFall[y][x] = Map[y][x].climate.fallRain
		}
	}

	// Apply simple averaging filter
	for y := 1; y < MapHeight-1; y++ {
		for x := 1; x < MapWidth-1; x++ {
			// Skip deep water tiles for complete smoothing
			if Map[y][x].altitude <= -0.3 {
				continue
			}

			// Calculate average of surrounding tiles for each season
			wTempSum, spTempSum, suTempSum, fTempSum := 0.0, 0.0, 0.0, 0.0
			wRainSum, spRainSum, suRainSum, fRainSum := 0.0, 0.0, 0.0, 0.0
			count := 0

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
						// Temperature smoothing
						wTempSum += Map[ny][nx].climate.winterTemp
						spTempSum += Map[ny][nx].climate.springTemp
						suTempSum += Map[ny][nx].climate.summerTemp
						fTempSum += Map[ny][nx].climate.fallTemp

						// Only smooth rainfall for land tiles
						if Map[y][x].altitude > 0 {
							wRainSum += Map[ny][nx].climate.winterRain
							spRainSum += Map[ny][nx].climate.springRain
							suRainSum += Map[ny][nx].climate.summerRain
							fRainSum += Map[ny][nx].climate.fallRain
						}

						count++
					}
				}
			}

			if count > 0 {
				// Blend original value with average (80% original, 20% average for gentle smoothing)
				blendFactor := 0.8

				// Temperature smoothing
				tempWinter[y][x] = Map[y][x].climate.winterTemp*blendFactor + (wTempSum/float64(count))*(1-blendFactor)
				tempSpring[y][x] = Map[y][x].climate.springTemp*blendFactor + (spTempSum/float64(count))*(1-blendFactor)
				tempSummer[y][x] = Map[y][x].climate.summerTemp*blendFactor + (suTempSum/float64(count))*(1-blendFactor)
				tempFall[y][x] = Map[y][x].climate.fallTemp*blendFactor + (fTempSum/float64(count))*(1-blendFactor)

				// Rainfall smoothing (only for land)
				if Map[y][x].altitude > 0 {
					rainWinter[y][x] = Map[y][x].climate.winterRain*blendFactor + (wRainSum/float64(count))*(1-blendFactor)
					rainSpring[y][x] = Map[y][x].climate.springRain*blendFactor + (spRainSum/float64(count))*(1-blendFactor)
					rainSummer[y][x] = Map[y][x].climate.summerRain*blendFactor + (suRainSum/float64(count))*(1-blendFactor)
					rainFall[y][x] = Map[y][x].climate.fallRain*blendFactor + (fRainSum/float64(count))*(1-blendFactor)
				}
			}
		}
	}

	// Apply the smoothed values
	for y := range MapHeight {
		for x := range MapWidth {
			// Apply temperature smoothing
			Map[y][x].climate.winterTemp = tempWinter[y][x]
			Map[y][x].climate.springTemp = tempSpring[y][x]
			Map[y][x].climate.summerTemp = tempSummer[y][x]
			Map[y][x].climate.fallTemp = tempFall[y][x]

			// Apply rainfall smoothing (only for land)
			if Map[y][x].altitude > 0 {
				Map[y][x].climate.winterRain = rainWinter[y][x]
				Map[y][x].climate.springRain = rainSpring[y][x]
				Map[y][x].climate.summerRain = rainSummer[y][x]
				Map[y][x].climate.fallRain = rainFall[y][x]
			}

			// Recalculate averages
			Map[y][x].climate.avgTemp = (Map[y][x].climate.winterTemp +
				Map[y][x].climate.springTemp +
				Map[y][x].climate.summerTemp +
				Map[y][x].climate.fallTemp) / 4.0

			Map[y][x].climate.avgRain = (Map[y][x].climate.winterRain +
				Map[y][x].climate.springRain +
				Map[y][x].climate.summerRain +
				Map[y][x].climate.fallRain) / 4.0
		}
	}
}

// LatLongResult holds the latitude and longitude values
type LatLongResult struct {
	Latitude  float64 // Degrees North/South of equator (-90 to 90)
	Longitude float64 // Degrees East/West (-180 to 180)
}

// GetLatitudeLongitude converts map coordinates to Earth-like latitude and longitude
func GetLatitudeLongitude(x, y int) LatLongResult {
	// The equator is at MapHeight / 2
	equator := MapHeight / 2

	// Calculate latitude: ranges from -90° (South Pole) to 90° (North Pole)
	// with 0° at the equator
	normalizedY := float64(equator-y) / float64(MapHeight/2)

	// Apply the same transformation used in climate generation
	// In generateClimate(), we use distanceFromEquator with a power function:
	// distanceFromEquator := math.Abs(float64(y - equator)) / float64(MapHeight/2)
	// latitudeTemp := 1.0 - math.Pow(distanceFromEquator, 0.8)

	// latitude = 90° * normalizedY gives us a direct linear mapping
	// We'll keep the sign to differentiate between N and S hemispheres
	latitude := 90.0 * normalizedY

	// For longitude, we'll map the x coordinate from 0-MapWidth to -180° to 180°
	normalizedX := (float64(x)/float64(MapWidth))*2.0 - 1.0
	longitude := 180.0 * normalizedX

	return LatLongResult{
		Latitude:  math.Round(latitude*100) / 100,  // Round to 2 decimal places
		Longitude: math.Round(longitude*100) / 100, // Round to 2 decimal places
	}
}

// GetLatLongDescription returns a formatted latitude/longitude string
func GetLatLongDescription(x, y int) string {
	result := GetLatitudeLongitude(x, y)

	// Format direction indicators
	var latDir, longDir string

	if result.Latitude >= 0 {
		latDir = "N"
	} else {
		latDir = "S"
	}

	if result.Longitude >= 0 {
		longDir = "E"
	} else {
		longDir = "W"
	}

	// Format with absolute values and direction indicators
	return fmt.Sprintf("%.2f°%s, %.2f°%s",
		math.Abs(result.Latitude), latDir,
		math.Abs(result.Longitude), longDir)
}
