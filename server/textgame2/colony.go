package textgame2

import "fmt"

// TODO: Buildings and Agents should have a number of ticks that they've been working/turned on for when we switch to production
// going over multiple ticks (a cycle).

var beginnerResourceCounts = [Resource_Max]uint{
	Resource_Wood_Oak: 0,
}

type Colony struct {
	// TODO: agents are array-of-structs atm. Potentially turn into struct of arrays later. Each agent is a state machine?
	context        *Context
	agents         []Agent
	resourceCounts [Resource_Max]uint // Current resources in storage
	landResources  [10]ResourceZone   // Available resource zones from land
	//landResources  [LandResource_Max]uint // Available resources from land
	buildings []Building

	// Production and consumption for current tick, committed to storage at the start of the next tick.
	currentProduction  [Resource_Max]uint
	currentConsumption [Resource_Max]uint

	// resourceConsumers [Resource_Max]*Node
	// landResourceProducers [LandResource_Max]*Node
	// resourceProducers [Resource_Max]*Node
}

func NewColony(context *Context, initialPopulationSize uint) *Colony {
	colony := new(Colony)
	colony.context = context
	colony.agents = make([]Agent, initialPopulationSize)
	colony.resourceCounts = beginnerResourceCounts
	colony.buildings = make([]Building, 0)
	for i := range colony.agents {
		a := &colony.agents[i]

		// TODO: randomize name, age, gender, and sexual orientation
		a.name = fmt.Sprintf("Unknown%2d Unknown", i)
		a.age = 20
		if i < len(colony.agents)/2 {
			a.gender = AgentGender_Male
			a.sexualAttraction = [AgentGender_Max]bool{
				AgentGender_Female: true,
			}
		} else {
			a.gender = AgentGender_Female
			a.sexualAttraction = [AgentGender_Max]bool{
				AgentGender_Male: true,
			}
		}

		a.food = 100
		a.health = 100
		a.state = AgentState_Idle
		a.stress = 20 // 20% stress from starting a new colony off with nothing
		a.familyID = i
		a.assignedZone = -1
	}

	colony.landResources[0] = NewResourceZone(0, LandResource_Forest_Oak, 20000)

	return colony
}

func (colony *Colony) Tick() {
	// Commit previous tick's resource production and consumption numbers to storage.
	// This should always be the very first thing done in a tick.
	colony.CommitProductionAndConsumption()

	// The next thing is to update the work/idle/sleep state of each agent based on the current time, day, etc. and their assigned workplace.
	if colony.context.IsWorkTime() {
		for id, _ := range colony.agents {
			a := &colony.agents[id]
			if a.assignedZone != -1 {
				a.state = AgentState_Work
			} else {
				a.state = AgentState_Idle
			}
		}
	} else if colony.context.IsSleepTime() {
		for id, _ := range colony.agents {
			a := &colony.agents[id]
			a.state = AgentState_Sleep
		}
	} else if colony.context.IsFreeTime() {
		for id, _ := range colony.agents {
			a := &colony.agents[id]
			a.state = AgentState_Idle
		}
	}

	// Design TODO: The current tick's consumption always takes from storage and never from the current tick's production?

	// Go through all resource zones (and their buildings/technologies) to set next tick's initial resource production and consumption numbers
	for i, _ := range colony.landResources {
		zone := &colony.landResources[i]
		if zone.resource == LandResource_Unknown {
			continue
		}

		var productionFromZone uint = colony.productionFromZone(zone)

		switch zone.resource {
		case LandResource_Dirt:
		case LandResource_Pond:
		case LandResource_Lake_Vertical:
		case LandResource_Lake_Horizontal:
		case LandResource_Forest_Oak:
			colony.currentProduction[Resource_Wood_Oak] = productionFromZone
			colony.currentConsumption[Resource_Wood_Oak] = 0
		case LandResource_Coal:
		case LandResource_Clay:
		case LandResource_Granite:
		case LandResource_Limestone:
		case LandResource_Sandstone:
		case LandResource_Marble:
		case LandResource_Slate:
		case LandResource_Iron:
		case LandResource_Aluminum:
		case LandResource_Zinc:
		case LandResource_Copper:
		case LandResource_Nickel:
		case LandResource_Tin:
		case LandResource_Silver:
		case LandResource_Gold:
		case LandResource_Haygrass:
		case LandResource_RawRice:
		case LandResource_Berries:
		case LandResource_Potatoes:
		case LandResource_Corn:
		case LandResource_Agave:
		case LandResource_Mushrooms:
		case LandResource_Strawberries:
		default:

		}
	}

	// Design Note TODO:
	// Go through rest of buildings to add to the resource production and consumption numbers.
	// Consumption of buildings should be bounded by the storage plus the production of other buildings.
	// So, we should iterate over all the buildings to get the ones that can use just the storage. Then, we iterate again
	// for the buildings that can consume the production of the storage-using-only buildings along with the rest of the storage.
	// OR RATHER we should probably have some dependency graph here so that we can traverse over it starting from the leaves.

	// TODO: Agents also consume things (like food, water, and clothes, at the bare minimum). These need to be factored in to the consumption numbers.
}

// Commits the previous tick's production and consumption to storage at the start of each tick.
func (colony *Colony) CommitProductionAndConsumption() {
	// Go through every resource zone and adjust their amount numbers based on what is being produced from each zone (which is not necessarily the same as the resource's production value).
	// Readjust the production numbers based on the amount (if necessary, but this should already be done in the Tick() function!)
	for i := range colony.landResources {
		zone := &colony.landResources[i]
		if zone.resource == LandResource_Unknown {
			continue
		}

		var productionFromZone uint = colony.productionFromZone(zone)

		switch zone.resource {
		case LandResource_Dirt:
		case LandResource_Pond:
		case LandResource_Lake_Vertical:
		case LandResource_Lake_Horizontal:
		case LandResource_Forest_Oak:
			if productionFromZone >= zone.amount {
				zone.amount = 0

				// Readjust based on overflow amount from production of zone
				diff := productionFromZone - zone.amount
				colony.currentProduction[Resource_Wood_Oak] -= diff
			} else {
				zone.amount -= productionFromZone
			}
		case LandResource_Coal:
		case LandResource_Clay:
		case LandResource_Granite:
		case LandResource_Limestone:
		case LandResource_Sandstone:
		case LandResource_Marble:
		case LandResource_Slate:
		case LandResource_Iron:
		case LandResource_Aluminum:
		case LandResource_Zinc:
		case LandResource_Copper:
		case LandResource_Nickel:
		case LandResource_Tin:
		case LandResource_Silver:
		case LandResource_Gold:
		case LandResource_Haygrass:
		case LandResource_RawRice:
		case LandResource_Berries:
		case LandResource_Potatoes:
		case LandResource_Corn:
		case LandResource_Agave:
		case LandResource_Mushrooms:
		case LandResource_Strawberries:
		default:

		}
	}

	// Now commit the production and consumption to storage
	for resource := range Resource_Max {
		if colony.currentConsumption[resource] >= colony.resourceCounts[resource]+colony.currentProduction[resource] {
			colony.resourceCounts[resource] = 0
		} else {
			colony.resourceCounts[resource] = uint(int(colony.resourceCounts[resource]) + int(colony.currentProduction[resource]) - int(colony.currentConsumption[resource]))
		}
	}
}

// Calculates the current tick's production from a zone, taking into account the zone's amount, the number of workers, *and* the state and stats of each worker.
func (colony *Colony) productionFromZone(zone *ResourceZone) uint {
	cap := zone.amount

	var numberOfActiveWorkers uint = 0
	for _, workerId := range zone.workers {
		worker := &colony.agents[workerId]
		if worker.state == AgentState_Work {
			numberOfActiveWorkers += 1
		}
	}

	return min(1*numberOfActiveWorkers, cap)
}
