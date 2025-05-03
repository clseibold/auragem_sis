package textgame2

import "math"

type Biome uint8

const (
	// Warm
	Biome_TropicalRainforest      Biome = iota // Specific type of tropical forest
	Biome_TropicalSeasonalForest               // Specific type of tropical forest
	Biome_TropicalMontaneForest                // Specific type of tropical forest
	Biome_TropicalMoistForest                  // Specific type of moist forest
	Biome_TropicalEvergreenForest              // Specific type of evergreen forest
	Biome_Savanna                              // Warm (tropical) grassland
	Biome_TropicalSwampForest                  // Forested wetland area
	Biome_TropicalSwamp                        // General wetland area
	Biome_Mangrove

	// Temperate
	Biome_TemperateDeciduousForest  // Specific type of temperate forest
	Biome_TemperateMixedForest      // Specific type of temperate forest
	Biome_TemperateRainforest       // Specific type of temperate forest
	Biome_TemperateConiferousForest // Specific type of temperate forest
	Biome_TemperateSwamp
	Biome_CypressSwamp
	Biome_MangroveSwamp
	Biome_Pampas       // Temperate grassland
	Biome_Veld         // Temperate grassland
	Biome_Prairie      // Temperate grassland
	Biome_TemperateFen // Temperate wetland

	// Biome_Marsh - dominated by herbaceous (non-woody) plants like grasses and reeds.

	// Cold
	Biome_BorealForest // aka. Taiga
	Biome_Alpine
	Biome_Tundra
	Biome_Steppe  // Cold grassland, but can also be considered semi-arid grassland
	Biome_ColdBog // Cold wetland
	Biome_ColdFen // Cold wetland
	Biome_ColdDesert
	Biome_IceSheet
	Biome_SeaIce // Hopeless

	// Semi-Arid
	Biome_SagebrushSteppe
	Biome_Matorral // Semi-arid shrubland

	// Hot
	Biome_MediterraneanShrubland // Hot Shrubland
	Biome_Fynbos                 // Hot Shrubland
	Biome_DesertShrubland        // Hot Shurbland
	Biome_HotDesert
	Biome_ExtremeDesert

	// Uncategorized
	Biome_Ocean
	Biome_Lake
	Biome_River
	Biome_Max
)

type LandType uint8

// All of these have variants that encompass or are next to bodies of water (e.g., floodplains that are flooded by rivers)
const (
	LandType_Hills LandType = iota // If altitude is >= 0.8, then they are foothills (next to mountains)
	// LandType_Foothills          // Near mountains
	LandType_Mountains
	LandType_Plains   // Plains that are next to rivers (floodplains) have most fertile soil and are where civilizations often started.
	LandType_Valleys  // Valleys between higher altitudes, near mountains, and river valleys.
	LandType_Plateaus // Rivers can cut through plateaus to create canyons and gorges. Plateaus can also be formed by volcanic activity.
	LandType_Coastal  // TODO: Implies next to water/sea?
	LandType_Water
	LandType_SandDunes

	// Rocky Outcrops?

	LandType_Max
)

// BiomeLandTypes maps each biome to the land types it can exist on
var BiomeLandTypes = [Biome_Max][]LandType{
	// Warm
	Biome_TropicalRainforest:      {LandType_Plains, LandType_Valleys, LandType_Hills, LandType_Coastal},
	Biome_TropicalSeasonalForest:  {LandType_Plains, LandType_Valleys, LandType_Hills},
	Biome_TropicalMontaneForest:   {LandType_Hills, LandType_Mountains},
	Biome_TropicalMoistForest:     {LandType_Plains, LandType_Valleys, LandType_Hills},
	Biome_TropicalEvergreenForest: {LandType_Plains, LandType_Valleys, LandType_Hills},
	Biome_Savanna:                 {LandType_Plains}, // Can be adjacent to Hills
	Biome_TropicalSwampForest:     {LandType_Plains, LandType_Valleys, LandType_Coastal},
	Biome_TropicalSwamp:           {LandType_Plains, LandType_Valleys, LandType_Coastal},
	Biome_Mangrove:                {LandType_Coastal},

	// Temperate
	Biome_TemperateDeciduousForest:  {LandType_Plains, LandType_Valleys, LandType_Hills},
	Biome_TemperateMixedForest:      {LandType_Plains, LandType_Valleys, LandType_Hills},
	Biome_TemperateRainforest:       {LandType_Plains, LandType_Valleys, LandType_Coastal},
	Biome_TemperateConiferousForest: {LandType_Hills, LandType_Mountains, LandType_Plateaus}, // Can be on mountains at lower to mid elevation. They transition to Boreal Forests and/or Alpine forests at higher altitudes.
	Biome_TemperateSwamp:            {LandType_Plains, LandType_Valleys, LandType_Coastal},
	Biome_CypressSwamp:              {LandType_Plains, LandType_Valleys, LandType_Coastal},
	Biome_MangroveSwamp:             {LandType_Coastal},
	Biome_Pampas:                    {LandType_Plains},
	Biome_Veld:                      {LandType_Plains, LandType_Plateaus},
	Biome_Prairie:                   {LandType_Plains},
	Biome_TemperateFen:              {LandType_Plains, LandType_Valleys},

	// Cold
	Biome_BorealForest: {LandType_Plains, LandType_Valleys, LandType_Hills, LandType_Mountains}, // Can be on mountains if altitude is at lower elevation where coniferous trees are present. The mountain peaks transition to Apline.
	Biome_Alpine:       {LandType_Mountains},
	Biome_Tundra:       {LandType_Plains, LandType_Valleys, LandType_Plateaus},
	Biome_Steppe:       {LandType_Plains, LandType_Plateaus},
	Biome_ColdBog:      {LandType_Plains, LandType_Valleys},
	Biome_ColdFen:      {LandType_Plains, LandType_Valleys},
	Biome_ColdDesert:   {LandType_Plains, LandType_Plateaus},
	Biome_IceSheet:     {LandType_Water},
	Biome_SeaIce:       {LandType_Water},

	// Semi-Arid
	Biome_SagebrushSteppe: {LandType_Plains, LandType_Plateaus},
	Biome_Matorral:        {LandType_Plains, LandType_Hills},

	// Hot
	Biome_MediterraneanShrubland: {LandType_Plains, LandType_Hills, LandType_Coastal},
	Biome_Fynbos:                 {LandType_Plains, LandType_Hills, LandType_Coastal},
	Biome_DesertShrubland:        {LandType_Plains, LandType_Plateaus},
	Biome_HotDesert:              {LandType_Plains, LandType_Plateaus, LandType_SandDunes},
	Biome_ExtremeDesert:          {LandType_Plains, LandType_Plateaus, LandType_SandDunes},

	// Uncategorized
	Biome_Ocean: {LandType_Water},
	Biome_Lake:  {LandType_Water},
	Biome_River: {LandType_Water, LandType_Valleys},
}

// Adjacent biomes - which biomes can border each other
var AdjacentBiomes = [Biome_Max][]Biome{
	// Warm tropical forests can be adjacent to other tropical forests, swamps, savannas
	Biome_TropicalRainforest: {
		Biome_TropicalSeasonalForest, Biome_TropicalMontaneForest, Biome_TropicalMoistForest,
		Biome_TropicalEvergreenForest, Biome_TropicalSwampForest, Biome_Savanna,
		Biome_Mangrove, Biome_TropicalSwamp,
	},
	Biome_TropicalSeasonalForest: {
		Biome_TropicalRainforest, Biome_TropicalMontaneForest, Biome_TropicalMoistForest,
		Biome_TropicalEvergreenForest, Biome_Savanna, Biome_TropicalSwampForest,
		Biome_DesertShrubland, Biome_Mangrove,
	},
	Biome_TropicalMontaneForest: {
		Biome_TropicalRainforest, Biome_TropicalSeasonalForest, Biome_TropicalMoistForest,
		Biome_TropicalEvergreenForest, Biome_Savanna, Biome_Alpine,
	},
	Biome_TropicalMoistForest: {
		Biome_TropicalRainforest, Biome_TropicalSeasonalForest, Biome_TropicalMontaneForest,
		Biome_TropicalEvergreenForest, Biome_TropicalSwampForest, Biome_Savanna,
	},
	Biome_TropicalEvergreenForest: {
		Biome_TropicalRainforest, Biome_TropicalSeasonalForest, Biome_TropicalMontaneForest,
		Biome_TropicalMoistForest, Biome_TropicalSwampForest, Biome_Savanna,
	},
	Biome_Savanna: {
		Biome_TropicalRainforest, Biome_TropicalSeasonalForest, Biome_TropicalMontaneForest,
		Biome_TropicalMoistForest, Biome_TropicalEvergreenForest, Biome_DesertShrubland,
		Biome_HotDesert,
	},
	Biome_TropicalSwampForest: {
		Biome_TropicalRainforest, Biome_TropicalSeasonalForest, Biome_TropicalMoistForest,
		Biome_TropicalEvergreenForest, Biome_TropicalSwamp, Biome_Mangrove,
	},
	Biome_TropicalSwamp: {
		Biome_TropicalRainforest, Biome_TropicalSwampForest, Biome_Mangrove,
		Biome_MangroveSwamp,
	},
	Biome_Mangrove: {
		Biome_TropicalRainforest, Biome_TropicalSeasonalForest, Biome_TropicalSwampForest,
		Biome_TropicalSwamp, Biome_MangroveSwamp, Biome_Ocean, Biome_Lake,
	},

	// Temperate forests and their adjacent biomes
	Biome_TemperateDeciduousForest: {
		Biome_TemperateMixedForest, Biome_TemperateRainforest, Biome_TemperateConiferousForest,
		Biome_TemperateSwamp, Biome_Prairie, Biome_Pampas, Biome_Veld,
		Biome_MediterraneanShrubland, Biome_BorealForest,
	},
	Biome_TemperateMixedForest: {
		Biome_TemperateDeciduousForest, Biome_TemperateRainforest, Biome_TemperateConiferousForest,
		Biome_BorealForest, Biome_TemperateSwamp, Biome_Prairie, Biome_Pampas,
		Biome_MediterraneanShrubland,
	},
	Biome_TemperateRainforest: {
		Biome_TemperateDeciduousForest, Biome_TemperateMixedForest, Biome_TemperateConiferousForest,
		Biome_TemperateSwamp, Biome_CypressSwamp, Biome_Ocean,
	},
	Biome_TemperateConiferousForest: {
		Biome_TemperateDeciduousForest, Biome_TemperateMixedForest, Biome_TemperateRainforest,
		Biome_BorealForest, Biome_Alpine, Biome_Tundra, Biome_Prairie,
	},
	Biome_TemperateSwamp: {
		Biome_TemperateDeciduousForest, Biome_TemperateMixedForest, Biome_TemperateRainforest,
		Biome_CypressSwamp, Biome_TemperateFen, Biome_MangroveSwamp,
	},
	Biome_CypressSwamp: {
		Biome_TemperateSwamp, Biome_TemperateRainforest, Biome_MangroveSwamp,
		Biome_TemperateFen,
	},
	Biome_MangroveSwamp: {
		Biome_TemperateSwamp, Biome_CypressSwamp, Biome_Mangrove, Biome_TropicalSwamp,
		Biome_Ocean, Biome_Lake,
	},
	Biome_Pampas: {
		Biome_TemperateDeciduousForest, Biome_TemperateMixedForest, Biome_Prairie,
		Biome_Veld, Biome_DesertShrubland, Biome_SagebrushSteppe,
	},
	Biome_Veld: {
		Biome_TemperateDeciduousForest, Biome_Pampas, Biome_Prairie, Biome_Savanna,
		Biome_Matorral, Biome_MediterraneanShrubland,
	},
	Biome_Prairie: {
		Biome_TemperateDeciduousForest, Biome_TemperateMixedForest, Biome_TemperateConiferousForest,
		Biome_Pampas, Biome_Veld, Biome_SagebrushSteppe, Biome_Steppe,
	},
	Biome_TemperateFen: {
		Biome_TemperateSwamp, Biome_CypressSwamp, Biome_ColdBog, Biome_ColdFen,
	},

	// Cold biomes
	Biome_BorealForest: {
		Biome_TemperateDeciduousForest, Biome_TemperateMixedForest, Biome_TemperateConiferousForest,
		Biome_Tundra, Biome_Alpine, Biome_ColdBog, Biome_Steppe,
	},
	Biome_Alpine: {
		Biome_BorealForest, Biome_TemperateConiferousForest, Biome_Tundra,
		Biome_ColdDesert, Biome_TropicalMontaneForest,
	},
	Biome_Tundra: {
		Biome_BorealForest, Biome_Alpine, Biome_ColdBog, Biome_ColdFen,
		Biome_ColdDesert, Biome_IceSheet, Biome_Steppe,
	},
	Biome_Steppe: {
		Biome_BorealForest, Biome_Tundra, Biome_ColdDesert, Biome_Prairie,
		Biome_SagebrushSteppe,
	},
	Biome_ColdBog: {
		Biome_BorealForest, Biome_Tundra, Biome_ColdFen, Biome_TemperateFen,
	},
	Biome_ColdFen: {
		Biome_Tundra, Biome_ColdBog, Biome_TemperateFen,
	},
	Biome_ColdDesert: {
		Biome_Tundra, Biome_Alpine, Biome_Steppe, Biome_IceSheet,
		Biome_SagebrushSteppe,
	},
	Biome_IceSheet: {
		Biome_Tundra, Biome_ColdDesert, Biome_SeaIce, Biome_Ocean,
	},
	Biome_SeaIce: {
		Biome_IceSheet, Biome_Ocean,
	},

	// Semi-arid and arid biomes
	Biome_SagebrushSteppe: {
		Biome_Steppe, Biome_ColdDesert, Biome_Prairie, Biome_Pampas,
		Biome_Matorral, Biome_DesertShrubland, Biome_HotDesert,
	},
	Biome_Matorral: {
		Biome_SagebrushSteppe, Biome_DesertShrubland, Biome_MediterraneanShrubland,
		Biome_Veld, Biome_Fynbos,
	},

	// Hot biomes
	Biome_MediterraneanShrubland: {
		Biome_TemperateDeciduousForest, Biome_TemperateMixedForest, Biome_Matorral,
		Biome_Fynbos, Biome_DesertShrubland, Biome_Veld,
	},
	Biome_Fynbos: {
		Biome_MediterraneanShrubland, Biome_Matorral, Biome_DesertShrubland,
		Biome_Ocean,
	},
	Biome_DesertShrubland: {
		Biome_SagebrushSteppe, Biome_Matorral, Biome_MediterraneanShrubland,
		Biome_Fynbos, Biome_HotDesert, Biome_Savanna,
	},
	Biome_HotDesert: {
		Biome_DesertShrubland, Biome_SagebrushSteppe, Biome_ExtremeDesert,
		Biome_Savanna,
	},
	Biome_ExtremeDesert: {
		Biome_HotDesert,
	},

	// Water biomes
	Biome_Ocean: {
		Biome_Lake, Biome_Mangrove, Biome_MangroveSwamp, Biome_TemperateRainforest,
		Biome_Fynbos, Biome_SeaIce, Biome_IceSheet,
	},
	Biome_Lake: {
		Biome_Ocean, Biome_River, Biome_Mangrove, Biome_MangroveSwamp,
	},
	Biome_River: {
		Biome_Lake, Biome_Ocean,
	},
}

// Function to assign biomes based on climate and terrain
func assignBiomes() {
	// First handle water bodies
	for y := range MapHeight {
		for x := range MapWidth {
			// Water bodies get special biome types
			if Map[y][x].altitude <= 0 {
				// Check if this is connected to the edge (ocean) or is a lake
				if x == 0 || y == 0 || x == MapWidth-1 || y == MapHeight-1 {
					Map[y][x].biome = Biome_Ocean
				} else {
					// Check if connected to edge
					isOcean := false
					// Simple flood fill to check ocean connection
					var processed [MapHeight][MapWidth]bool
					queue := []struct{ x, y int }{{x, y}}
					processed[y][x] = true

					for len(queue) > 0 {
						curr := queue[0]
						queue = queue[1:]

						// If we've reached an edge, this is ocean
						if curr.x == 0 || curr.y == 0 || curr.x == MapWidth-1 || curr.y == MapHeight-1 {
							isOcean = true
							break
						}

						// Check adjacent water tiles
						for dy := -1; dy <= 1; dy++ {
							for dx := -1; dx <= 1; dx++ {
								if dx == 0 && dy == 0 {
									continue
								}

								nx, ny := curr.x+dx, curr.y+dy
								if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight &&
									Map[ny][nx].altitude <= 0 && !processed[ny][nx] {
									queue = append(queue, struct{ x, y int }{nx, ny})
									processed[ny][nx] = true
								}
							}
						}
					}

					if isOcean {
						Map[y][x].biome = Biome_Ocean
					} else {
						Map[y][x].biome = Biome_Lake
					}
				}
				continue
			}

			// Rivers get their own biome
			if Map[y][x].landType == LandType_Water && Map[y][x].altitude > 0 {
				Map[y][x].biome = Biome_River
				continue
			}
		}
	}

	// Handle land tiles
	for y := range MapHeight {
		for x := range MapWidth {
			// Skip water tiles, already handled
			if Map[y][x].biome == Biome_Ocean || Map[y][x].biome == Biome_Lake || Map[y][x].biome == Biome_River {
				continue
			}

			// Get climate data for biome determination
			avgTemp := Map[y][x].climate.avgTemp
			avgRain := Map[y][x].climate.avgRain
			winterTemp := Map[y][x].climate.winterTemp
			summerTemp := Map[y][x].climate.summerTemp
			tempRange := math.Abs(summerTemp - winterTemp)

			// Get landType
			landType := Map[y][x].landType

			// Handle special cases first

			// 1. Alpine areas at high elevations
			if landType == LandType_Mountains && Map[y][x].altitude > 1.2 {
				Map[y][x].biome = Biome_Alpine
				continue
			}

			// 2. Sand dunes in deserts
			if landType == LandType_SandDunes {
				if avgTemp > 0.7 && avgRain < 0.2 {
					Map[y][x].biome = Biome_ExtremeDesert
				} else {
					Map[y][x].biome = Biome_HotDesert
				}
				continue
			}

			// 3. Wetlands and swamps
			if Map[y][x].hasMarsh {
				if avgTemp > 0.7 {
					if landType == LandType_Coastal {
						if Map[y][x].hasGrove {
							Map[y][x].biome = Biome_Mangrove
						} else {
							Map[y][x].biome = Biome_MangroveSwamp
						}
					} else if Map[y][x].hasGrove {
						Map[y][x].biome = Biome_TropicalSwampForest
					} else {
						Map[y][x].biome = Biome_TropicalSwamp
					}
				} else if avgTemp > 0.4 {
					if Map[y][x].hasGrove && summerTemp > 0.7 {
						Map[y][x].biome = Biome_CypressSwamp
					} else {
						Map[y][x].biome = Biome_TemperateSwamp
					}
				} else {
					// Cold wetlands
					if winterTemp < 0.25 {
						if avgRain > 0.7 {
							Map[y][x].biome = Biome_ColdBog
						} else {
							Map[y][x].biome = Biome_ColdFen
						}
					} else {
						Map[y][x].biome = Biome_TemperateFen
					}
				}
				continue
			}

			// Handle different climate zones

			// TROPICAL ZONE - hot year-round (avgTemp > 0.7, low seasonal variation)
			if avgTemp > 0.7 && tempRange < 0.3 {
				// Very wet tropical = rainforest
				if avgRain > 0.7 {
					if landType == LandType_Mountains || Map[y][x].altitude > 0.9 {
						Map[y][x].biome = Biome_TropicalMontaneForest
					} else if Map[y][x].altitude > 0.5 {
						Map[y][x].biome = Biome_TropicalEvergreenForest
					} else {
						Map[y][x].biome = Biome_TropicalRainforest
					}
				} else if avgRain > 0.5 { // Moderately wet tropical = seasonal forest or moist forest
					if (Map[y][x].climate.winterRain-Map[y][x].climate.summerRain) > 0.3 ||
						(Map[y][x].climate.summerRain-Map[y][x].climate.winterRain) > 0.3 {
						Map[y][x].biome = Biome_TropicalSeasonalForest
					} else {
						Map[y][x].biome = Biome_TropicalMoistForest
					}
				} else if avgRain > 0.3 { // Drier tropical = savanna
					Map[y][x].biome = Biome_Savanna
				} else if avgRain > 0.2 { // Very dry tropical = desert types
					Map[y][x].biome = Biome_DesertShrubland
				} else if avgRain > 0.1 {
					Map[y][x].biome = Biome_HotDesert
				} else {
					Map[y][x].biome = Biome_ExtremeDesert
				}
				continue
			}

			// SUBTROPICAL/WARM TEMPERATE ZONE (avgTemp 0.55-0.7)
			if avgTemp > 0.55 && avgTemp <= 0.7 {
				// Mediterranean climate (dry summers, wet winters)
				summerRain := Map[y][x].climate.summerRain
				winterRain := Map[y][x].climate.winterRain
				if summerRain < 0.3 && winterRain > 0.5 {
					if landType == LandType_Coastal {
						if avgRain > 0.45 {
							Map[y][x].biome = Biome_Fynbos
						} else {
							Map[y][x].biome = Biome_MediterraneanShrubland
						}
					} else {
						if avgRain > 0.4 {
							Map[y][x].biome = Biome_MediterraneanShrubland
						} else {
							Map[y][x].biome = Biome_Matorral
						}
					}
				} else if avgRain > 0.6 { // Humid subtropical
					Map[y][x].biome = Biome_TemperateDeciduousForest
				} else if avgRain > 0.45 { // Moderately wet subtropical
					if winterTemp < 0.4 {
						Map[y][x].biome = Biome_TemperateMixedForest
					} else {
						Map[y][x].biome = Biome_TemperateDeciduousForest
					}
				} else if avgRain > 0.3 { // Drier subtropical = temperate grasslands
					if landType == LandType_Plateaus {
						Map[y][x].biome = Biome_Veld
					} else {
						Map[y][x].biome = Biome_Prairie
					}
				} else if avgRain > 0.2 { // Very dry subtropical = semi-desert
					Map[y][x].biome = Biome_DesertShrubland
				} else {
					Map[y][x].biome = Biome_HotDesert
				}
				continue
			}

			// MID-TEMPERATE ZONE (avgTemp 0.4-0.55)
			if avgTemp > 0.4 && avgTemp <= 0.55 {
				// Wet temperate = rainforests & mixed forests
				if avgRain > 0.7 {
					if landType == LandType_Coastal {
						Map[y][x].biome = Biome_TemperateRainforest
					} else if winterTemp < 0.3 {
						Map[y][x].biome = Biome_TemperateMixedForest
					} else {
						Map[y][x].biome = Biome_TemperateDeciduousForest
					}
				} else if avgRain > 0.5 { // Moderately wet temperate
					if winterTemp < 0.3 {
						if landType == LandType_Mountains || landType == LandType_Hills && Map[y][x].altitude > 0.7 {
							Map[y][x].biome = Biome_TemperateConiferousForest
						} else {
							Map[y][x].biome = Biome_TemperateMixedForest
						}
					} else {
						Map[y][x].biome = Biome_TemperateDeciduousForest
					}
				} else if avgRain > 0.35 { // Moderate temperate
					if landType == LandType_Mountains || (landType == LandType_Hills && winterTemp < 0.25) {
						Map[y][x].biome = Biome_TemperateConiferousForest
					} else {
						Map[y][x].biome = Biome_Prairie
					}
				} else if avgRain > 0.25 { // Drier temperate
					Map[y][x].biome = Biome_SagebrushSteppe
				} else { // Very dry temperate
					Map[y][x].biome = Biome_ColdDesert
				}
				continue
			}

			// COLD TEMPERATE / BOREAL ZONE (avgTemp 0.25-0.4)
			if avgTemp > 0.25 && avgTemp <= 0.4 {
				// Wet boreal
				if avgRain > 0.6 {
					if summerTemp > 0.5 && landType != LandType_Mountains {
						Map[y][x].biome = Biome_TemperateMixedForest
					} else {
						Map[y][x].biome = Biome_BorealForest
					}
				} else if avgRain > 0.4 { // Moderate boreal
					if landType == LandType_Mountains && Map[y][x].altitude > 0.9 {
						Map[y][x].biome = Biome_Alpine
					} else {
						Map[y][x].biome = Biome_BorealForest
					}
				} else if avgRain > 0.25 { // Drier boreal
					if winterTemp < 0.15 {
						Map[y][x].biome = Biome_Steppe
					} else {
						Map[y][x].biome = Biome_SagebrushSteppe
					}
				} else { // Very dry boreal
					Map[y][x].biome = Biome_ColdDesert
				}
				continue
			}

			// ARCTIC/ALPINE ZONE (avgTemp <= 0.25)
			if avgTemp <= 0.25 {
				// Wet arctic/alpine
				if avgRain > 0.5 {
					if summerTemp > 0.4 {
						Map[y][x].biome = Biome_BorealForest
					} else if summerTemp > 0.3 {
						Map[y][x].biome = Biome_Tundra
					} else {
						if landType == LandType_Mountains {
							Map[y][x].biome = Biome_Alpine
						} else {
							Map[y][x].biome = Biome_Tundra
						}
					}
				} else if avgRain > 0.3 { // Moderate arctic/alpine
					if summerTemp > 0.4 && landType != LandType_Mountains {
						Map[y][x].biome = Biome_Steppe
					} else {
						Map[y][x].biome = Biome_Tundra
					}
				} else { // Drier arctic/alpine
					if winterTemp < 0.1 && avgRain < 0.2 {
						Map[y][x].biome = Biome_IceSheet
					} else {
						Map[y][x].biome = Biome_ColdDesert
					}
				}
				continue
			}
		}
	}

	// Refine biome assignments with local features
	for y := range MapHeight {
		for x := range MapWidth {
			// Adjust based on specific features

			// Groves tend to push toward more forested biomes
			if Map[y][x].hasGrove {
				// In very wet areas, upgrade to richer forest types
				if Map[y][x].biome == Biome_Savanna && Map[y][x].climate.avgRain > 0.4 {
					Map[y][x].biome = Biome_TropicalSeasonalForest
				}
				if Map[y][x].biome == Biome_Prairie && Map[y][x].climate.avgRain > 0.4 {
					if Map[y][x].climate.avgTemp > 0.6 {
						Map[y][x].biome = Biome_TemperateDeciduousForest
					} else if Map[y][x].climate.avgTemp > 0.4 {
						Map[y][x].biome = Biome_TemperateMixedForest
					} else {
						Map[y][x].biome = Biome_BorealForest
					}
				}
				if Map[y][x].biome == Biome_SagebrushSteppe && Map[y][x].climate.avgRain > 0.35 {
					if Map[y][x].climate.avgTemp > 0.45 {
						Map[y][x].biome = Biome_TemperateMixedForest
					} else if Map[y][x].climate.avgTemp > 0.3 {
						Map[y][x].biome = Biome_BorealForest
					}
				}
			}

			// Desert-specific features
			if Map[y][x].isDesert && !Map[y][x].hasGrove {
				if Map[y][x].climate.avgTemp > 0.7 && Map[y][x].climate.avgRain < 0.15 {
					Map[y][x].biome = Biome_ExtremeDesert
				} else if Map[y][x].climate.avgTemp > 0.6 && Map[y][x].climate.avgRain < 0.25 {
					Map[y][x].biome = Biome_HotDesert
				} else if Map[y][x].climate.avgTemp < 0.3 && Map[y][x].climate.avgRain < 0.25 {
					Map[y][x].biome = Biome_ColdDesert
				}
			}

			// Scrub pushes toward more arid biomes
			if Map[y][x].hasScrub && !Map[y][x].hasGrove {
				if Map[y][x].biome == Biome_Savanna && Map[y][x].climate.avgRain < 0.4 {
					Map[y][x].biome = Biome_DesertShrubland
				}
				if Map[y][x].biome == Biome_Prairie && Map[y][x].climate.avgRain < 0.35 {
					Map[y][x].biome = Biome_SagebrushSteppe
				}
				if Map[y][x].biome == Biome_TemperateDeciduousForest && Map[y][x].climate.avgRain < 0.55 {
					Map[y][x].biome = Biome_MediterraneanShrubland
				}
			}

			// Salt flats are desert features
			if Map[y][x].hasSaltFlat {
				if Map[y][x].climate.avgTemp > 0.6 {
					Map[y][x].biome = Biome_HotDesert
				} else if Map[y][x].climate.avgTemp > 0.3 {
					Map[y][x].biome = Biome_DesertShrubland
				} else {
					Map[y][x].biome = Biome_ColdDesert
				}
			}
		}
	}

	// Final smoothing pass to create more coherent biome regions
	// This is optional but creates more natural-looking transitions
	//smoothBiomes()
}

// smoothBiomes performs a biome smoothing pass to create more coherent regions
// and reduce unrealistic small patches or "noise" in the biome distribution
func smoothBiomes() {
	// Create a temporary copy of the biome map
	var tempBiomes [MapHeight][MapWidth]Biome

	// First, copy all existing biomes
	for y := range MapHeight {
		for x := range MapWidth {
			tempBiomes[y][x] = Map[y][x].biome
		}
	}

	// Define a small influence radius for smoothing
	smoothRadius := 2

	// Process each land tile
	for y := range MapHeight {
		for x := range MapWidth {
			// Skip water tiles (ocean, lakes, rivers)
			if Map[y][x].biome == Biome_Ocean || Map[y][x].biome == Biome_Lake || Map[y][x].biome == Biome_River {
				continue
			}

			// Skip tiles near map edges to avoid bounds checking
			if x < smoothRadius || y < smoothRadius || x >= MapWidth-smoothRadius || y >= MapHeight-smoothRadius {
				continue
			}

			// Get the current biome and landtype
			currentBiome := Map[y][x].biome
			currentLandType := Map[y][x].landType

			// Count occurrences of each biome in the neighborhood
			biomeCount := make(map[Biome]int)
			totalCount := 0

			// Check surrounding tiles
			for dy := -smoothRadius; dy <= smoothRadius; dy++ {
				for dx := -smoothRadius; dx <= smoothRadius; dx++ {
					// Skip the center tile
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := x+dx, y+dy
					neighborBiome := Map[ny][nx].biome

					// Skip water biomes for smoothing
					if neighborBiome == Biome_Ocean || neighborBiome == Biome_Lake || neighborBiome == Biome_River {
						continue
					}

					// Weight by distance - closer tiles have more influence
					weight := 3 - (abs(dx) + abs(dy))
					if weight < 1 {
						weight = 1
					}

					// Add to the count, weighted by distance
					biomeCount[neighborBiome] += weight
					totalCount += weight
				}
			}

			// If we have neighboring biomes to consider
			if totalCount > 0 {
				// Find the most common adjacent biome
				mostCommonBiome := currentBiome
				highestCount := 0

				for biome, count := range biomeCount {
					// Check if this biome can exist on the current land type
					canExistOnLandType := false
					for _, validLandType := range BiomeLandTypes[biome] {
						if validLandType == currentLandType {
							canExistOnLandType = true
							break
						}
					}

					// Only consider biomes that can exist on this land type
					// and are compatible (adjacent) with the current biome
					isCompatible := false
					for _, adjBiome := range AdjacentBiomes[currentBiome] {
						if biome == adjBiome {
							isCompatible = true
							break
						}
					}

					if canExistOnLandType && isCompatible && count > highestCount {
						mostCommonBiome = biome
						highestCount = count
					}
				}

				// Only change if there's a strong prevalence of another biome
				// and that biome is compatible with current landtype and climate
				if highestCount >= totalCount/3 && mostCommonBiome != currentBiome {
					// Check if the new biome is appropriate for this climate
					avgTemp := Map[y][x].climate.avgTemp
					avgRain := Map[y][x].climate.avgRain

					// Very simplified climate check - full check would be too complex here
					biomeIsClimateSuitable := true

					// Basic climate compatibility rules
					if (mostCommonBiome == Biome_HotDesert || mostCommonBiome == Biome_ExtremeDesert) &&
						(avgTemp < 0.6 || avgRain > 0.3) {
						biomeIsClimateSuitable = false
					}
					if (mostCommonBiome == Biome_TropicalRainforest || mostCommonBiome == Biome_TropicalMoistForest) &&
						(avgTemp < 0.65 || avgRain < 0.6) {
						biomeIsClimateSuitable = false
					}
					if (mostCommonBiome == Biome_BorealForest || mostCommonBiome == Biome_Tundra) &&
						avgTemp > 0.45 {
						biomeIsClimateSuitable = false
					}
					if mostCommonBiome == Biome_ColdDesert && avgTemp > 0.5 {
						biomeIsClimateSuitable = false
					}

					// Apply the change if climatically suitable
					if biomeIsClimateSuitable {
						tempBiomes[y][x] = mostCommonBiome
					}
				}
			}
		}
	}

	// Apply the changes
	for y := range MapHeight {
		for x := range MapWidth {
			// Only apply to land tiles
			if Map[y][x].biome != Biome_Ocean && Map[y][x].biome != Biome_Lake && Map[y][x].biome != Biome_River {
				Map[y][x].biome = tempBiomes[y][x]
			}
		}
	}
}
