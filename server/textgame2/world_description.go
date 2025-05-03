package textgame2

import (
	"fmt"
	"math"
	"strings"
)

// GetTileDescription generates a descriptive paragraph about a tile
func GetTileDescription(x, y int) string {
	if x < 0 || x >= MapWidth || y < 0 || y >= MapHeight {
		return "An unknown area beyond the mapped world."
	}

	tile := &Map[y][x]
	description := strings.Builder{}

	// Start with land type and basic geographic description
	landDescription := getLandTypeDescription(tile)
	description.WriteString(landDescription)

	// Add biome description
	biomeDescription := getBiomeDescription(tile)
	description.WriteString(" " + biomeDescription)

	// Add natural features (water features, vegetation, etc.)
	featureDescription := getFeatureDescription(tile)
	if featureDescription != "" {
		description.WriteString(" " + featureDescription)
	}

	// Add climate information
	climateDescription := getClimateDescription(tile)
	description.WriteString(" " + climateDescription)

	// Add seasonal details or additional sensory information
	atmosphericDetails := getAtmosphericDetails(tile)
	description.WriteString(" " + atmosphericDetails)

	return description.String()
}

// getLandTypeDescription describes the basic terrain
func getLandTypeDescription(tile *Tile) string {
	switch tile.landType {
	case LandType_Water:
		if tile.altitude <= -0.5 {
			return "The deep waters here reflect the sky in their dark blue expanse."
		} else if tile.biome == Biome_Ocean {
			return "You're at the edge of a vast ocean, with waves gently lapping against the shore."
		} else if tile.biome == Biome_Lake {
			return "A serene lake spreads before you, its surface occasionally broken by ripples."
		} else {
			return "This area is covered by shallow water, clear enough to see the bottom in places."
		}
	case LandType_Plains:
		if tile.altitude < 0.15 {
			return "You stand on low-lying flatlands that stretch out in gentle undulations."
		} else if tile.isDesert {
			return "A flat expanse of parched earth extends in all directions."
		} else {
			return "The open plains here offer wide vistas with few obstacles."
		}
	case LandType_Hills:
		if tile.altitude >= 0.8 {
			return "These rolling foothills rise steadily toward the mountains visible in the distance."
		} else {
			return "The landscape is dominated by gently rolling hills that create a patchwork of slopes and valleys."
		}
	case LandType_Valleys:
		if tile.hasFloodArea {
			return "You're in a broad river valley with rich alluvial soil deposited by seasonal floods."
		} else {
			return "This sheltered valley is nestled between higher terrain, protected from harsh winds."
		}
	case LandType_Plateaus:
		if tile.altitude > 0.7 {
			return "You stand atop a high plateau with impressive views of the surrounding lands."
		} else {
			return "This elevated plateau rises abruptly from the surrounding terrain, creating a distinct tabletop landscape."
		}
	case LandType_Mountains:
		if tile.altitude > 1.3 {
			return "Massive mountain peaks tower overhead, their summits often shrouded in clouds."
		} else {
			return "The rugged mountain terrain here is characterized by steep slopes and rocky outcroppings."
		}
	case LandType_Coastal:
		return "This coastal area marks the transition between land and sea, with the shoreline visible nearby."
	case LandType_SandDunes:
		return "Rippling sand dunes dominate the landscape, their shapes constantly shifting with the wind."
	default:
		return "The terrain here is of an unusual character, difficult to categorize."
	}
}

// getBiomeDescription provides details about the biome
func getBiomeDescription(tile *Tile) string {
	switch tile.biome {
	// Warm biomes
	case Biome_TropicalRainforest:
		return "The tropical rainforest here is extraordinarily diverse, with a dense canopy of vegetation year-round and rich, quickly recycled soil beneath. Countless species make their home among the multiple layers of plant life, from the forest floor to the emergent trees that pierce the canopy."
	case Biome_TropicalSeasonalForest:
		return "This tropical seasonal forest transitions dramatically between wet and dry periods. During the rainy season, the vegetation is lush and green, while in the dry months, many trees shed their leaves to conserve water. The forest floor is covered with a layer of rapidly decomposing leaf litter."
	case Biome_TropicalMontaneForest:
		return "The cooler temperatures of this high-elevation tropical forest create a unique ecosystem. Trees are shorter than in lowland forests, often gnarled and heavily laden with epiphytes, mosses, and lichens that thrive in the persistent cloud moisture."
	case Biome_TropicalMoistForest:
		return "The tropical moist forest surrounds you with verdant vegetation. While not as perpetually wet as a rainforest, it maintains high humidity and supports diverse plant and animal life with its nutrient-poor but rapidly recycling soil."
	case Biome_TropicalEvergreenForest:
		return "Massive evergreen trees form the foundation of this tropical forest. Unlike deciduous forests, the foliage here remains green year-round, creating a constant shade that limits undergrowth to shade-tolerant species."
	case Biome_Savanna:
		return "The savanna stretches out with its characteristic mix of scattered trees and sweeping grasslands. During wet seasons, the landscape turns green with fresh growth, while the dry season brings golden hues as grasses cure in the sun. Large grazing animals often inhabit these spaces."
	case Biome_TropicalSwampForest:
		return "Towering trees rise from the waterlogged soil of this swamp forest. Their massive buttress roots provide stability in the soft ground, while hanging lianas and epiphytes create a complex vertical ecosystem above the dark, nutrient-rich waters."
	case Biome_TropicalSwamp:
		return "The tropical swamp is a world of water and scattered vegetation. Specialized plants that can tolerate the waterlogged conditions thrive here, creating habitat for a diversity of creatures adapted to this amphibious environment."
	case Biome_Mangrove:
		return "Distinctive mangrove trees dominate this coastal wetland, their tangled roots creating a barrier between land and sea. The brackish water is rich in organic matter, supporting juvenile fish and countless invertebrates that thrive among the protective roots."

	// Temperate biomes
	case Biome_TemperateDeciduousForest:
		return "Deciduous trees create a canopy that changes dramatically with the seasons—from spring buds to full summer foliage, brilliant autumn colors, and the stark silhouettes of winter. The forest floor cycles through understory blooms and layers of fallen leaves."
	case Biome_TemperateMixedForest:
		return "This mixed forest blends the evergreen resilience of conifers with the seasonal changes of deciduous trees. This diversity creates varied habitats and a forest that never completely loses its foliage, even in winter."
	case Biome_TemperateRainforest:
		return "Mist frequently envelops this lush temperate rainforest. Massive ancient trees are draped with mosses and epiphytes, while the understory features dense ferns and shade-tolerant shrubs thriving in the cool, perpetually moist environment."
	case Biome_TemperateConiferousForest:
		return "Tall coniferous trees form a dense canopy over a forest floor typically carpeted with needles. The evergreen nature of these trees allows them to photosynthesize year-round, while their resinous wood resists decay and many pests."
	case Biome_TemperateSwamp:
		return "Standing water defines this temperate swamp, with water-tolerant trees and shrubs creating islands of vegetation. The water level fluctuates seasonally, influencing the cycles of plant growth and animal activity."
	case Biome_CypressSwamp:
		return "Ancient cypress trees rise from the murky waters, their distinctive knees protruding above the surface. Spanish moss drapes from branches, creating an otherworldly atmosphere in this specialized wetland ecosystem."
	case Biome_MangroveSwamp:
		return "The brackish water supports a forest of mangroves whose complex root systems protect the coastline and create a nursery for countless marine species. The boundary between land and sea is blurred in this transitional ecosystem."
	case Biome_Pampas:
		return "Vast temperate grasslands stretch to the horizon, dominated by tall grasses that wave in the wind like a golden sea. Few trees interrupt this landscape, which supports grazing animals and birds that nest among the protective grass cover."
	case Biome_Veld:
		return "This temperate grassland features both grasses and scattered drought-resistant shrubs. The soil supports a diverse community of grasses that have adapted to withstand periodic drought and occasional fires that maintain the open character of the landscape."
	case Biome_Prairie:
		return "The prairie extends with mixed grasses of varying heights, creating a complex structure despite the apparent simplicity. The dense root systems form deep, rich soil teeming with life, while seasonal blooms of wildflowers add splashes of color."
	case Biome_TemperateFen:
		return "Groundwater seepage creates this wetland rich in minerals and supporting specialized plant communities. Unlike bogs, the water flow brings nutrients that support a diverse ecosystem of sedges, grasses, and wildflowers."

	// Cold biomes
	case Biome_BorealForest:
		return "Coniferous trees adapted to harsh northern conditions dominate this forest. The growing season is short but intense, with the evergreen trees ready to photosynthesize whenever conditions allow. The forest floor is carpeted with mosses, lichens, and acid-loving shrubs."
	case Biome_Alpine:
		return "Above the treeline, hardy alpine plants hug the ground to survive the harsh conditions. Most growth happens in a brief summer window, with plants often featuring dense hairs, waxy coatings, or rosette forms to survive the extremes of temperature and wind."
	case Biome_Tundra:
		return "The tundra appears stark, but closer inspection reveals a community of lichens, mosses, and low-growing plants with specially adapted root systems that can function in the shallow active layer above the permafrost."
	case Biome_Steppe:
		return "This cold, semi-arid grassland features short grasses and herbs that can withstand both the temperature extremes and limited precipitation. The open landscape experiences significant winds that further challenge plant growth."
	case Biome_ColdBog:
		return "Waterlogged and acidic, this northern bog accumulates partially decayed organic matter as peat. Specialized plants like sphagnum moss, sedges, and insectivorous species thrive in the nutrient-poor conditions."
	case Biome_ColdFen:
		return "Unlike the acidic bog, this cold fen receives mineral-rich groundwater that supports sedges, specialized grasses, and a variety of wildflowers adapted to the saturated soil conditions."
	case Biome_ColdDesert:
		return "This high-elevation or high-latitude desert receives little precipitation, and what moisture does fall often comes as snow. The sparse vegetation consists of widely spaced shrubs and grasses adapted to both cold and drought."
	case Biome_IceSheet:
		return "A vast expanse of permanent ice covers the landscape, its thickness obscuring the terrain beneath. Only at the margins can any life be found, where specialized microorganisms might survive in protected niches."
	case Biome_SeaIce:
		return "The frozen surface of the sea forms a dynamic platform that changes with the seasons. Though it appears lifeless, the ice edge and areas beneath support a specialized community of algae, invertebrates, and the larger species that feed on them."

	// Semi-arid biomes
	case Biome_SagebrushSteppe:
		return "Aromatic sagebrush dominates this semi-arid landscape, interspersed with bunchgrasses and occasional flowering plants that create a sparse but well-adapted community. The plants are spaced to maximize access to limited water resources."
	case Biome_Matorral:
		return "Drought-resistant shrubs with hard, evergreen leaves characterize this Mediterranean scrubland. Plants have adapted to survive long dry summers and mild, wetter winters, often featuring aromatic compounds that discourage herbivory."

	// Hot biomes
	case Biome_MediterraneanShrubland:
		return "Aromatic, drought-adapted shrubs with leathery leaves dominate this ecosystem, which experiences hot, dry summers and mild, wet winters. Many plants are fire-adapted, with some requiring the heat of periodic wildfires to release their seeds."
	case Biome_Fynbos:
		return "This highly diverse shrubland features fine-leaved, evergreen plants adapted to nutrient-poor soils. Many species have specialized relationships with insects for pollination and ants for seed dispersal, contributing to the exceptional biodiversity."
	case Biome_DesertShrubland:
		return "Widely spaced shrubs adapted to extreme aridity characterize this landscape. Plants have evolved diverse strategies to conserve water, from waxy coatings and reduced leaf surfaces to specialized metabolic pathways and deep root systems."
	case Biome_HotDesert:
		return "The harsh desert environment supports specialized plants and animals that have evolved remarkable adaptations for water conservation and heat regulation. Vegetation is sparse, with considerable bare ground between plants."
	case Biome_ExtremeDesert:
		return "Only the most hardy organisms survive in this hyperarid environment. Most plant life is confined to temporary water sources or exists as seeds waiting for the rare precipitation events that trigger brief bursts of growth and reproduction."

	// Water biomes
	case Biome_Ocean:
		return "The open ocean extends beyond sight, its surface hiding the complex ecosystems beneath. The water column supports a range of life forms, from microscopic plankton to large marine species."
	case Biome_Lake:
		return "This freshwater lake creates a distinct ecosystem compared to the surrounding land. The water supports aquatic plants, fish, and a variety of organisms that depend on the relatively stable environment it provides."
	case Biome_River:
		return "The flowing water of the river shapes the surrounding landscape while supporting species adapted to life in current. The ecosystem changes from headwaters to mouth, creating a diverse ribbon of habitat."

	default:
		return "The biological community here represents a unique combination of plants and animals adapted to the local conditions."
	}
}

// getFeatureDescription describes special geographic features
func getFeatureDescription(tile *Tile) string {
	features := make([]string, 0)

	// Water features
	if tile.hasStream {
		features = append(features, "A small stream cuts through the area, its clear water flowing over smooth stones")
	}
	if tile.hasPond {
		features = append(features, "A tranquil pond reflects the sky, its edges ringed with characteristic vegetation")
	}
	if tile.hasSpring {
		features = append(features, "A natural spring bubbles up from underground, creating a small oasis of fresh water")
	}
	if tile.hasMarsh && !tile.hasPond {
		features = append(features, "Soggy marsh ground squelches underfoot, supporting specialized water-tolerant plants")
	} else if tile.hasMarsh && tile.hasPond {
		features = append(features, "The margins of the pond transition into a marshy area with reeds and sedges")
	}

	// Plains features
	if tile.hasGrove {
		if tile.isDesert {
			features = append(features, "A small grove of drought-resistant trees provides rare shade in this arid landscape")
		} else if tile.biome == Biome_Savanna {
			features = append(features, "A clustered grove of acacia trees stands out against the open grassland")
		} else {
			features = append(features, "A pleasant grove of trees creates a shaded microhabitat")
		}
	}
	if tile.hasMeadow {
		if tile.isDesert {
			features = append(features, "A small patch of wildflowers has bloomed following recent rains")
		} else {
			features = append(features, "A colorful meadow of wildflowers attracts numerous pollinators")
		}
	}
	if tile.hasScrub {
		if tile.isDesert {
			features = append(features, "Scattered drought-resistant shrubs dot the landscape")
		} else {
			features = append(features, "Hardy scrub vegetation covers portions of the ground")
		}
	}
	if tile.hasRocks {
		features = append(features, "Weathered rock outcroppings break through the surface, providing habitat for specialized species")
	}
	if tile.hasGameTrail {
		features = append(features, "Well-worn animal trails crisscross the area, revealing the movements of local wildlife")
	}
	if tile.hasFloodArea {
		features = append(features, "Evidence of seasonal flooding is visible in the terrain and vegetation patterns")
	}
	if tile.hasSaltFlat {
		features = append(features, "A crusty salt flat indicates high evaporation and mineral accumulation")
	}

	if len(features) == 0 {
		return ""
	} else if len(features) == 1 {
		return features[0] + "."
	} else {
		// Join multiple features with appropriate punctuation
		result := features[0]
		for i := 1; i < len(features)-1; i++ {
			result += ", " + features[i]
		}
		result += ", and " + features[len(features)-1] + "."
		return result
	}
}

// getClimateDescription describes the climate
func getClimateDescription(tile *Tile) string {
	// Get temperature and rainfall descriptions
	tempDesc := GetTemperatureDescription(tile.climate.avgTemp)
	rainDesc := GetRainfallDescription(tile.climate.avgRain)

	// Extract just the descriptive part without the technical values
	tempPart := strings.Split(tempDesc, " (")[0]
	rainPart := strings.Split(rainDesc, " (")[0]

	// Create climate description
	return fmt.Sprintf("The climate is generally %s and %s.", strings.ToLower(tempPart), strings.ToLower(rainPart))
}

// getAtmosphericDetails adds sensory information and seasonal changes
func getAtmosphericDetails(tile *Tile) string {
	// Detect significant seasonal changes
	tempRange := math.Abs(tile.climate.summerTemp - tile.climate.winterTemp)
	rainRange := math.Max(
		math.Max(tile.climate.winterRain, tile.climate.springRain),
		math.Max(tile.climate.summerRain, tile.climate.fallRain),
	) - math.Min(
		math.Min(tile.climate.winterRain, tile.climate.springRain),
		math.Min(tile.climate.summerRain, tile.climate.fallRain),
	)

	// Check for extreme seasons
	hasExtremeSummer := tile.climate.summerTemp > 0.8
	hasExtremeWinter := tile.climate.winterTemp < 0.2

	// Choose atmospheric details based on biome, climate, and features
	if tile.landType == LandType_Water {
		if tile.biome == Biome_Ocean {
			return "The rhythmic sound of waves provides a constant backdrop, while the briny scent of seawater fills the air."
		} else {
			return "The water creates a microclimate that moderates temperature extremes compared to the surrounding landscape."
		}
	} else if tile.isDesert {
		if hasExtremeSummer {
			return "Daytime temperatures soar to brutal extremes, with significant cooling at night due to rapid heat loss."
		} else {
			return "The arid conditions create striking clarity in the air, allowing distant features to be seen with unusual sharpness."
		}
	} else if tempRange > 0.5 && rainRange > 0.4 {
		// Areas with dramatic seasonal changes
		return "The landscape transforms dramatically through the seasons, from lush growth during wet periods to dormancy during dry or cold times."
	} else if hasExtremeWinter {
		return "Winters bring harsh conditions that severely test the resilience of the local ecosystem."
	} else if hasExtremeSummer {
		return "Summer heat dominates the yearly cycle, with plants and animals having adapted various strategies to cope with high temperatures."
	} else if tile.climate.avgRain > 0.7 {
		return "The persistent moisture creates a humid atmosphere where the rich scents of vegetation and soil hang in the air."
	} else if tile.landType == LandType_Mountains {
		return "Weather conditions can change rapidly here due to the mountainous terrain, with local microclimates creating varied ecological niches."
	} else {
		return "The changing light throughout the day reveals different aspects of this environment, from morning dew to afternoon shadows."
	}
}
