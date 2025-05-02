package textgame2

type Biome uint8

const (
	// Warm
	Biome_TemperateForest Biome = iota
	Biome_TemperateSwamp
	Biome_TropicalForest
	Biome_TropicalSwamp

	// Biome_Savanna - aka. tropical grassland
	// Biome_Marsh - dominated by herbaceous (non-woody) plants like grasses and reeds.

	// Cold
	Biome_BorealForest // aka. Taiga
	Biome_ColdBog
	Biome_Tundra
	Biome_IceSheet
	Biome_SeaIce // Hopeless

	// Hot
	Biome_AridShrubland
	Biome_Desert
	Biome_ExtremeDesert

	// Uncategorized
	Biome_Grassland // Prarie, Steppe, Savanna, and Pampas
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

var LandTypesOfBiomes = [Biome_Max][]LandType{
	Biome_TemperateForest: []LandType{
		LandType_Hills,
		LandType_Plains,
		LandType_Valleys,
		LandType_Plateaus,
		LandType_Coastal,
	},
	Biome_TemperateSwamp: []LandType{
		LandType_Plains, // floodplains
		LandType_Valleys,
		LandType_Coastal, // Wetlands
		// Marshy areas
	},
	Biome_TropicalForest: []LandType{
		LandType_Hills,
		LandType_Mountains,
		LandType_Plains, // lowland
		LandType_Valleys,
	},
	Biome_TropicalSwamp: []LandType{
		LandType_Plains,  // floodplains
		LandType_Coastal, // wetlands
		// Marshes
		// low-lying areas
	},

	// Cold
	Biome_BorealForest: []LandType{
		LandType_Hills,
		LandType_Mountains,
		LandType_Valleys,
		LandType_Plateaus,
	},
	Biome_ColdBog: []LandType{
		LandType_Plains, // flat
		LandType_Coastal,
		// low-lying areas
		// wetlands
	},
	Biome_Tundra: []LandType{
		LandType_Mountains, // aka. alpine tundra
		LandType_Plains,    // flat
		LandType_Valleys,
		LandType_Coastal,
	},
	Biome_IceSheet: []LandType{ // TODO
		LandType_Hills,
		LandType_Mountains,
		LandType_Plains,
		LandType_Valleys,
		LandType_Plateaus,
		LandType_Coastal,
	},
	Biome_SeaIce: []LandType{ // TODO
		LandType_Hills,
		LandType_Mountains,
		LandType_Plains,
		LandType_Valleys,
		LandType_Plateaus,
		LandType_Coastal,
	},

	// Hot
	Biome_AridShrubland: []LandType{
		LandType_Hills,
		LandType_Valleys,
		LandType_Plateaus,
		// rocky outcrops
	},
	Biome_Desert: []LandType{
		LandType_SandDunes,
		LandType_Valleys,
		LandType_Plateaus, // rocky
		LandType_Coastal,
	},
	Biome_ExtremeDesert: []LandType{
		LandType_SandDunes,
		LandType_Mountains,
		LandType_Valleys,
		LandType_Plateaus, // rocky
		LandType_Coastal,
	},

	// Uncategorized
	Biome_Grassland: []LandType{
		LandType_Hills,  // rolling
		LandType_Plains, // flat
		LandType_Valleys,
		LandType_Plateaus,
	},
}

var AdjacentBiomes = [Biome_Max][]Biome{
	Biome_TemperateForest: []Biome{
		Biome_TemperateForest,
		Biome_TemperateSwamp,
		Biome_TropicalForest,

		// Cold
		Biome_BorealForest, // aka. Taiga
		// Biome_IceSheet,
		// Biome_SeaIce, // Hopeless

		// Hot
		Biome_AridShrubland,

		// Uncategorized
		Biome_Grassland,
	},
	Biome_TemperateSwamp: []Biome{
		Biome_TemperateForest,
		Biome_TemperateSwamp,

		// Cold
		Biome_BorealForest, // aka. Taiga
		// Biome_IceSheet,
		// Biome_SeaIce, // Hopeless

		// Uncategorized
		Biome_Grassland,
	},
	Biome_TropicalForest: []Biome{
		Biome_TemperateForest,
		Biome_TropicalForest,
		Biome_TropicalSwamp,

		// Biome_Savanna

		// Cold
		// Biome_IceSheet,
		// Biome_SeaIce, // Hopeless

		// Hot
		Biome_AridShrubland,
	},
	Biome_TropicalSwamp: []Biome{
		Biome_TropicalForest,
		Biome_TropicalSwamp,

		// Biome_Savanna

		// Cold
		// Biome_IceSheet,
		// Biome_SeaIce, // Hopeless
	},

	// Cold
	Biome_BorealForest: []Biome{
		Biome_TemperateForest,

		// Cold
		Biome_BorealForest, // aka. Taiga
		Biome_ColdBog,
		Biome_Tundra,
		Biome_IceSheet,
		Biome_SeaIce, // Hopeless

		// Uncategorized
		Biome_Grassland,
	},
	Biome_ColdBog: []Biome{
		Biome_TemperateSwamp,

		// Cold
		Biome_BorealForest, // aka. Taiga
		Biome_ColdBog,
		Biome_Tundra,
		Biome_IceSheet,
		Biome_SeaIce, // Hopeless
	},
	Biome_Tundra: []Biome{
		Biome_TemperateForest,

		// Cold
		Biome_BorealForest, // aka. Taiga
		Biome_ColdBog,
		Biome_IceSheet,
		Biome_SeaIce, // Hopeless
	},
	Biome_IceSheet: []Biome{ // TODO
		Biome_TemperateForest,
		Biome_TemperateSwamp,
		Biome_TropicalForest,
		Biome_TropicalSwamp,

		// Cold
		Biome_BorealForest, // aka. Taiga
		Biome_ColdBog,
		Biome_Tundra,
		Biome_IceSheet,
		Biome_SeaIce, // Hopeless

		// Hot
		Biome_AridShrubland,
		Biome_Desert,
		Biome_ExtremeDesert,

		// Uncategorized
		Biome_Grassland,
	},
	Biome_SeaIce: []Biome{ // TODO
		Biome_TemperateForest,
		Biome_TemperateSwamp,
		Biome_TropicalForest,
		Biome_TropicalSwamp,

		// Cold
		Biome_BorealForest, // aka. Taiga
		Biome_ColdBog,
		Biome_Tundra,
		Biome_IceSheet,
		Biome_SeaIce, // Hopeless

		// Hot
		Biome_AridShrubland,
		Biome_Desert,
		Biome_ExtremeDesert,

		// Uncategorized
		Biome_Grassland,
	},

	// Hot
	Biome_AridShrubland: []Biome{
		Biome_TemperateForest,
		Biome_TropicalForest,

		// Hot
		Biome_AridShrubland,
		Biome_Desert,
		Biome_ExtremeDesert,

		// Uncategorized
		Biome_Grassland,
	},
	Biome_Desert: []Biome{
		// Hot
		Biome_AridShrubland,
		Biome_Desert,
		Biome_ExtremeDesert,

		// Uncategorized
		Biome_Grassland,
	},
	Biome_ExtremeDesert: []Biome{
		// Hot
		Biome_AridShrubland,
		Biome_Desert,
		Biome_ExtremeDesert,

		// Uncategorized
		Biome_Grassland,
	},

	// Uncategorized
	Biome_Grassland: []Biome{
		Biome_TemperateForest,
		Biome_TropicalForest,

		// Cold
		Biome_BorealForest, // aka. Taiga
		Biome_IceSheet,
		Biome_SeaIce, // Hopeless

		// Hot
		Biome_AridShrubland,
		Biome_Desert,
		Biome_ExtremeDesert,

		// Uncategorized
		Biome_Grassland,
	},
}
