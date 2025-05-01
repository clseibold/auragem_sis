package textgame2

import (
	"fmt"
)

// TODO: Buildings and Agents should have a number of ticks that they've been working/turned on for when we switch to production
// going over multiple ticks (a cycle).

var beginnerResourceCounts = [Resource_Max]uint{
	Resource_Wood_Oak: 0,
}

type Colony struct {
	// TODO: agents are array-of-structs atm. Potentially turn into struct of arrays later. Each agent is a state machine?
	context        *Context
	name           string
	agents         []Agent
	resourceCounts [Resource_Max]uint // Current resources in storage
	landResources  [10]ResourceZone   // Available resource zones from land
	//landResources  [LandResource_Max]uint // Available resources from land
	buildings []Building

	// Production and consumption for current tick, the whole integer committed to storage at the start of the next tick.
	currentProduction  [Resource_Max]float64
	currentConsumption [Resource_Max]float64

	// resourceConsumers [Resource_Max]*Node
	// landResourceProducers [LandResource_Max]*Node
	// resourceProducers [Resource_Max]*Node
}

func NewColony(context *Context, name string, initialPopulationSize uint) *Colony {
	colony := new(Colony)
	colony.context = context
	colony.name = name
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

	colony.landResources[0] = NewResourceZone(0, LandResource_Forest_Oak, 30000)
	colony.landResources[1] = NewResourceZone(1, LandResource_Granite, 20000)
	colony.landResources[2] = NewResourceZone(2, LandResource_Berries, 40000)

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
		if zone.landResource == LandResource_Unknown {
			continue
		}

		// Remove the whole integer but keep the fractional parts in the consumption and production.
		colony.currentProduction[zone.landResource.ToResource()] = colony.currentProduction[zone.landResource.ToResource()] - float64(uint(colony.currentProduction[zone.landResource.ToResource()]))
		colony.currentConsumption[zone.landResource.ToResource()] = colony.currentConsumption[zone.landResource.ToResource()] - float64(uint(colony.currentConsumption[zone.landResource.ToResource()]))

		// Add to the fractional parts this tick's production and consumption values.
		var productionFromZone float64 = colony.productionFromZone(zone)

		colony.currentProduction[zone.landResource.ToResource()] += productionFromZone
		colony.currentConsumption[zone.landResource.ToResource()] = 0
	}

	// Design Note TODO:
	// Go through rest of buildings to add to the resource production and consumption numbers.
	// Consumption of buildings should be bounded by the storage plus the production of other buildings.
	// So, we should iterate over all the buildings to get the ones that can use just the storage. Then, we iterate again
	// for the buildings that can consume the production of the storage-using-only buildings along with the rest of the storage.
	// OR RATHER we should probably have some dependency graph here so that we can traverse over it starting from the leaves.

	// TODO: Agents also consume things (like food, water, and clothes, at the bare minimum). These need to be factored in to the consumption numbers.
	foodConsumption := float64(len(colony.agents)) * float64(.000555)
	colony.currentConsumption[Resource_Berries] = min(foodConsumption, float64(colony.resourceCounts[Resource_Berries]))
}

// Commits the previous tick's production and consumption to storage at the start of each tick.
func (colony *Colony) CommitProductionAndConsumption() {
	// Go through every resource zone and adjust their amount numbers based on what is being produced from each zone (which is not necessarily the same as the resource's production value).
	// Readjust the production numbers based on the amount (if necessary, but this should already be done in the Tick() function!)
	for i := range colony.landResources {
		zone := &colony.landResources[i]
		if zone.landResource == LandResource_Unknown {
			continue
		}

		var productionFromZone float64 = colony.productionFromZone(zone)
		productionWhole := uint(productionFromZone)

		if productionWhole >= zone.amount {
			zone.amount = 0

			// Readjust based on overflow amount from production of zone
			diff := productionWhole - zone.amount
			colony.currentProduction[zone.landResource.ToResource()] -= float64(diff)
		} else {
			zone.amount -= productionWhole
		}
	}

	// Now commit the production and consumption to storage
	for resource := range Resource_Max {
		productionWhole := uint(colony.currentProduction[resource])
		consumptionWhole := uint(colony.currentConsumption[resource])
		if consumptionWhole >= colony.resourceCounts[resource]+productionWhole {
			colony.resourceCounts[resource] = 0
		} else {
			colony.resourceCounts[resource] = uint(int(colony.resourceCounts[resource]) + int(productionWhole) - int(consumptionWhole))
		}
	}
}

// Calculates the current tick's production from a zone, taking into account the zone's amount, the number of workers, *and* the state and stats of each worker.
func (colony *Colony) productionFromZone(zone *ResourceZone) float64 {
	cap := float64(zone.amount)

	var numberOfActiveWorkers float64 = 0
	for _, workerId := range zone.workers {
		worker := &colony.agents[workerId]
		if worker.state == AgentState_Work {
			numberOfActiveWorkers += 1
		}
	}

	return min(productionPerDayToPerTicks(zone.landResource.PerDayProductionPerAgent())*numberOfActiveWorkers, cap)
}

func productionPerDayToPerTicks(perDay float64) float64 {
	return perDay / 24 / 60 / 60 * float64(InGameSecondsPerTick)
}
