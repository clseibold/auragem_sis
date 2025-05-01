package textgame2

// These are the resources that are part of the land, excluding animals, and are usually
// held within resource zones. They are harvested into regular resources.
type LandResource uint8

const (
	LandResource_Unknown LandResource = iota
	LandResource_Dirt

	// Water Types
	LandResource_Pond
	LandResource_Lake_Vertical
	LandResource_Lake_Horizontal

	// Tree Types
	LandResource_Forest_Oak

	// Fuel Ore
	LandResource_Coal

	// Sone
	LandResource_Clay

	LandResource_Granite
	LandResource_Limestone
	LandResource_Sandstone
	LandResource_Marble
	LandResource_Slate

	// Naturally-ocurring Metal Ores
	LandResource_Iron
	LandResource_Aluminum
	LandResource_Zinc
	LandResource_Copper
	LandResource_Nickel
	LandResource_Tin
	LandResource_Silver
	LandResource_Gold

	// Plants
	LandResource_Haygrass
	LandResource_RawRice
	LandResource_Berries
	LandResource_Potatoes
	LandResource_Corn
	LandResource_Agave
	LandResource_Mushrooms
	LandResource_Strawberries

	LandResource_Max
)

func (resource LandResource) ToString() string {
	switch resource {
	case LandResource_Unknown:
		return "Unknown"
	case LandResource_Dirt:
		return "Dirt"
	case LandResource_Pond:
		return "Pond"
	case LandResource_Lake_Vertical:
		return "Lake (Vertical)"
	case LandResource_Lake_Horizontal:
		return "Lake (Horizontal)"
	case LandResource_Forest_Oak:
		return "Oak Forest"
	case LandResource_Coal:
		return "Coal"
	case LandResource_Clay:
		return "Clay"
	case LandResource_Granite:
		return "Granite"
	case LandResource_Limestone:
		return "Limestone"
	case LandResource_Sandstone:
		return "Sandstone"
	case LandResource_Marble:
		return "Marble"
	case LandResource_Slate:
		return "Slate"
	case LandResource_Iron:
		return "Iron"
	case LandResource_Aluminum:
		return "Aluminum"
	case LandResource_Zinc:
		return "Zinc"
	case LandResource_Copper:
		return "Copper"
	case LandResource_Nickel:
		return "Nickel"
	case LandResource_Tin:
		return "Tin"
	case LandResource_Silver:
		return "Silver"
	case LandResource_Gold:
		return "Gold"
	case LandResource_Haygrass:
		return "Haygrass"
	case LandResource_RawRice:
		return "Raw Rice"
	case LandResource_Berries:
		return "Berries"
	case LandResource_Potatoes:
		return "Potatoes"
	case LandResource_Corn:
		return "Corn"
	case LandResource_Agave:
		return "Agave"
	case LandResource_Mushrooms:
		return "Mushrooms"
	case LandResource_Strawberries:
		return "Strawberries"
	default:
		return "Unknown Resource"
	}
}

// These are the resources that have been harvested or crafted.
type Resource uint8 // uint8 max is 255, uint16 max is 65535

const (
	// Basics
	Resource_Water Resource = iota
	Resource_ResearchPoints

	// Wood and Fuel
	Resource_Wood_Oak
	Resource_Coal

	// Stone
	Resource_Clay // Used for pottery and bricks

	Resource_Granite   // Most common stone type on the earth. Igneous rock composed of quartz, feldspar, and mica. Durable and strong. Often found in mountain ranges and the continental crust.
	Resource_Limestone // Common in sedimentary environments. Composed of calcium carbonate.
	Resource_Sandstone // Sedimentary rock composed mainly of sand-sized mineral particles or rock fragments. Commonly used in construction.
	Resource_Marble    // Metamorphic rock formed from limestone. Used in sculptures and high-end construction.
	Resource_Slate     // Fine-grained metamorphic rock that originates from Shale (heat and pressure over time turns Shale into Slate). Strong and durable, resistant to weathering and suitable for outdoor applications. Water-resistant.
	// Resource_Conglomerate // Sedimentary rock found in riverbeds, composed of rounded fragments of various sizes cemented together.
	// Resource_Quartzite // Hard metamorphic rock that originates from sandstone. Known for durability, used in construction, and as decoration.
	// Resource_Shale // Sedimentary rock formed from clay and silt. Used for bricks and tiles.
	// Resource_Basalt // Strong volcanic rock, used in construction and as aggregate in concrete.
	// Resrouce_Pumice // Lightweight volcanic rock
	// Resource_Flint
	// Resource_Soapstone
	// Resource_Quartz // Crystal-like stone

	// Naturally-ocurring Metals
	Resource_Iron     // Used in construction, manufacturing, and transportation, especially to make steel.
	Resource_Aluminum // Used in packaging, transportation, construction, and consumer goods.
	Resource_Zinc     // Used to galvanize steel to prevent corrosion, as well as in alloys and batteries.
	Resource_Copper   // Exclusively used in electrical wiring, plumbing, and electronics due to excellent conductivity.
	Resource_Nickel   // Used in stainless steel production, batteries (e.g., nickel-cadmium), and various alloys.
	Resource_Tin      // Combined with Copper to make bronze, coated on steel to prevent corrosion, soldering metal parts for electronics or plumbing, and glassmaking.
	Resource_Silver   // Used in electronics, jewelry, and photograpy.
	Resource_Gold     // Used in jewelry, electronics, and as an investment.

	// Resource_Titanium // Lightweight and strong. Used in video games for armor, high-end weapons, and equipment.
	// Resource_Obsidian // Volcanic glass

	// Fabrics and Wools
	Resource_Cloth
	Resource_SheepWool

	// Raw Plant Foods
	Resource_Hay
	Resource_RawRice
	Resource_Berries
	Resource_Potatoes
	Resource_Corn
	Resource_Agave
	Resource_Mushrooms
	Resource_Strawberries

	// Raw Meat Foods
	Resource_Milk
	Resource_Meat
	Resource_InsectMeat

	// --- Crafted Resources ---
	Resource_Steel  // Crafted from Iron and a small percentage of Carbon. Smelt pig iron combined with carbon (in the form of coke, a carbon derived fro coal) and other elements. Alloy with manganese, nickel, chromium, or vanadium, for different types of steel. Finally, refine and cast into various shapes (sheets, bars, and beams).
	Resource_Bronze // Alloy of Copper and Tin

	Resource_Max
)

// Note: Some foods when eaten raw can give food poisoning

type ResourceZoneId int

// A resource zone is a zone on the land of a particular resource that can be harvested
// either manually or with buildings. Each colony region can have up to 10 different resource zones.
type ResourceZone struct {
	id       ResourceZoneId
	resource LandResource
	amount   uint
	workers  []AgentId
}

func (zone *ResourceZone) AddWorker(id AgentId, a *Agent) {
	// If already assigned a zone, return
	if a.assignedZone != -1 {
		return
	}

	a.assignedZone = zone.id
	zone.workers = append(zone.workers, id)
}

func (zone *ResourceZone) RemoveWorker(id AgentId, a *Agent) {
	if a.assignedZone == -1 {
		return
	}

	// Reset the agent's assignedZone
	a.assignedZone = -1

	// Remove the worker by replacing it with the last element and slicing off the last element
	for i, workerId := range zone.workers {
		if workerId == id {
			zone.workers[i] = zone.workers[len(zone.workers)-1]
			zone.workers = zone.workers[:len(zone.workers)-1]
			break
		}
	}
}

func (zone *ResourceZone) RemoveLastWorker(colony *Colony) {
	lastWorkerId := zone.workers[len(zone.workers)-1]
	colony.agents[lastWorkerId].assignedZone = -1
	zone.workers = zone.workers[:len(zone.workers)-1]
}

func NewResourceZone(id ResourceZoneId, resource LandResource, amount uint) ResourceZone {
	return ResourceZone{resource: resource, amount: amount, workers: make([]AgentId, 0, 20)}
}
