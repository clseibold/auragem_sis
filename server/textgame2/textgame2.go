package textgame2

import (
	_ "embed"
	"fmt"
	"math"
	"path"
	"strconv"
	"strings"
	"time"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

//go:embed design.md
var designDocument string

const TickRealTimeDuration = time.Second
const InGameSecondsPerTick int = 4 // NOTE: I could get 7 in-game days per real-time day if I switched this to 7 igs per tick.
const WorkHours = 10
const TicksPerInGameDay float64 = 1 / float64(InGameSecondsPerTick) * 60 * 60 * 24
const TicksPerInGameWorkDay float64 = 1 / float64(InGameSecondsPerTick) * 60 * 60 * WorkHours
const AreaOfEachWorldTile = 10000000 // In square meters (10 square kilometers)

func TicksToInGameDuration(ticks int) time.Duration {
	return time.Duration(ticks*InGameSecondsPerTick) * time.Second
}

type Context struct {
	// previousTickTime time.Time
	inGameTime   time.Time
	ticker       *time.Ticker
	firstColony  *Colony
	secondColony *Colony
}

func NewContext() *Context {
	// TODO: Load in saved game states from save directory (including ticker and time information)
	generateWorldMap()

	context := new(Context)
	context.ticker = time.NewTicker(TickRealTimeDuration)
	context.firstColony = NewColony(context, "Test Colony", 6, true)
	context.secondColony = NewColony(context, "Second Test Colony", 6, false)
	context.inGameTime = time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC)
	return context
}

func (c *Context) Start() {
	go c.SimulationLoop()
}

func (c *Context) SimulationLoop() {
	for {
		<-c.ticker.C
		c.inGameTime = c.inGameTime.Add(TicksToInGameDuration(1))
		c.firstColony.Tick()
		c.secondColony.Tick()
	}
}

func (c *Context) Attach(s sis.ServeMux) {
	s.AddRoute("/", c.Homepage)
	s.AddRoute("/about/", c.About)
	s.AddRoute("/design/", c.DesignDocument)
	s.AddRoute("/world-map/", PrintWorldMap)
	s.AddRoute("/explore/", c.ExploreWorld)

	group := s.Group("/test/")
	group.AddRoute("/", c.firstColony.ColonyPage)
	group.AddRoute("/resource_zone/:id/", c.firstColony.ResourceZonePage)
	group.AddRoute("/resource_zone/:id/add_worker", c.firstColony.AddWorkerPage)
	group.AddRoute("/resource_zone/:id/remove_worker", c.firstColony.RemoveWorkerPage)

	group2 := s.Group("/second-test/")
	group2.AddRoute("/", c.secondColony.ColonyPage)
	group2.AddRoute("/resource_zone/:id/", c.secondColony.ResourceZonePage)
	group2.AddRoute("/resource_zone/:id/add_worker", c.secondColony.AddWorkerPage)
	group2.AddRoute("/resource_zone/:id/remove_worker", c.secondColony.RemoveWorkerPage)
}

func (c *Context) Homepage(request *sis.Request) {
	request.Heading(1, "Biomebound: Colony Stratum")
	request.Gemini("\n")
	request.Link("/about/", "About")
	request.Link("/test/", "Test Colony")
	request.Link("/second-test/", "Second Test Colony")
	request.Gemini("\n")
	request.Link("/world-map/", "World Map")
	request.Link("/explore/?25,25", "Explore World Map")
}

func (c *Context) About(request *sis.Request) {
	request.Heading(1, "About Game")
	request.Gemini(`
Biomebound is an MMO colony-management survival game written primarily for Gemini.

Four in-game days equals one real-time day.

Each colony has a set of resource zones. These are zones of resources that are harvested from the land. The resources available from resource zones are dependent on the location, biome, and weather of the colony.

=> / Homepage
=> /design/ Design Document
`) // TODO
}

func (c *Context) DesignDocument(request *sis.Request) {
	request.Gemini(designDocument)
}

func (c *Context) ExploreWorld(request *sis.Request) {
	xy, _ := request.Query()
	if xy == "" {
		xy = "25,25"
	}
	parts := strings.Split(xy, ",")
	x, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	y, _ := strconv.Atoi(strings.TrimSpace(parts[1]))

	tile := &Map[y][x]

	request.Heading(1, fmt.Sprintf("Explore Map: (%d, %d)", x, y))
	request.Gemini("\n")
	request.Gemini(GetTileDescription(x, y) + "\n")
	request.Gemini("\n")

	request.Gemini("```Statistics\n")
	request.Gemini("\nClimate Stats:\n")
	request.Gemini(fmt.Sprintf("Altitude:     %.2f\n", tile.altitude))
	request.Gemini(fmt.Sprintf("Avg Temp:     %.1f°C (%f)\n", ConvertTemperature(tile.climate.avgTemp).Celsius, tile.climate.avgTemp))
	request.Gemini(fmt.Sprintf("Avg Winter:   %.1f°C (%f)\n", ConvertTemperature(tile.climate.winterTemp).Celsius, tile.climate.winterTemp))
	request.Gemini(fmt.Sprintf("Avg Spring:   %.1f°C (%f)\n", ConvertTemperature(tile.climate.springTemp).Celsius, tile.climate.springTemp))
	request.Gemini(fmt.Sprintf("Avg Summer:   %.1f°C (%f)\n", ConvertTemperature(tile.climate.summerTemp).Celsius, tile.climate.summerTemp))
	request.Gemini(fmt.Sprintf("Avg Autumn:   %.1f°C (%f)\n", ConvertTemperature(tile.climate.fallTemp).Celsius, tile.climate.fallTemp))
	request.Gemini("```\n\n")

	landTypeNames := map[LandType]string{
		LandType_Water:     "Water",
		LandType_Plains:    "Plains",
		LandType_Hills:     "Hills",
		LandType_Valleys:   "Valleys",
		LandType_Plateaus:  "Plateaus",
		LandType_Mountains: "Mountains",
		LandType_Coastal:   "Coastal",
		LandType_SandDunes: "Sand Dunes",
	}

	if y-1 >= 0 {
		up := &Map[y-1][x]
		if up.landType != LandType_Water {
			request.Link(fmt.Sprintf("/explore/?%d,%d", x, y-1), fmt.Sprintf("Go Up (%s)", landTypeNames[up.landType]))
		}
	}
	if y+1 < MapHeight {
		down := &Map[y+1][x]
		if down.landType != LandType_Water {
			request.Link(fmt.Sprintf("/explore/?%d,%d", x, y+1), fmt.Sprintf("Go Down (%s)", landTypeNames[down.landType]))
		}
	}
	if x-1 >= 0 {
		left := &Map[y][x-1]
		if left.landType != LandType_Water {
			request.Link(fmt.Sprintf("/explore/?%d,%d", x-1, y), fmt.Sprintf("Go Left (%s)", landTypeNames[left.landType]))
		}
	}
	if x+1 < MapWidth {
		right := &Map[y][x+1]
		if right.landType != LandType_Water {
			request.Link(fmt.Sprintf("/explore/?%d,%d", x+1, y), fmt.Sprintf("Go Right (%s)", landTypeNames[right.landType]))
		}
	}
}

// TODO: When you create a new colony, create a background description of how the people got there.

func (colony *Colony) ColonyPage(request *sis.Request) {
	request.Heading(1, colony.name)
	request.Gemini("\n")

	// Tile Description
	request.Gemini(GetTileDescription(colony.tileLocation.X, colony.tileLocation.Y) + "\n")
	request.Gemini("\n")

	// Water and Food consumption per person

	unemployedAgents := 0
	for id, _ := range colony.agents {
		a := &colony.agents[id]
		if a.assignedZone == -1 {
			unemployedAgents += 1
		}
	}

	// Statistics
	request.Gemini("```Statistics\n")
	if colony.context.IsWorkTime() {
		request.Gemini(fmt.Sprintf("Date & Time:  %s (Work)\n", colony.context.inGameTime.Format(time.TimeOnly)))
	} else if colony.context.IsSleepTime() {
		request.Gemini(fmt.Sprintf("Date & Time:  %s (Sleep)\n", colony.context.inGameTime.Format(time.TimeOnly)))
	} else if colony.context.IsFreeTime() {
		request.Gemini(fmt.Sprintf("Date & Time:  %s (Free Time)\n", colony.context.inGameTime.Format(time.TimeOnly)))
	}
	request.Gemini(fmt.Sprintf("Map Location: (%d, %d)\n", colony.tileLocation.X, colony.tileLocation.Y))
	request.Gemini(fmt.Sprintf("Population:   %d (%d unemployed)\n", len(colony.agents), unemployedAgents))
	request.Gemini(fmt.Sprintf("Food:         %d (%+.2f/workday)\n", colony.resourceCounts[Resource_Berries], math.Ceil(colony.currentProduction[Resource_Berries]-colony.currentConsumption[Resource_Berries])*TicksPerInGameWorkDay)) // TODO: Count *all* food sources
	request.Gemini(fmt.Sprintf("Water:        %d (+0/workday)\n", colony.resourceCounts[Resource_Water]))
	request.Gemini(fmt.Sprintf("Oak Logs:     %d (%+.2f/workday | %+.2f)\n", colony.resourceCounts[Resource_Oak_Logs], math.Ceil(colony.currentProduction[Resource_Oak_Logs]-colony.currentConsumption[Resource_Oak_Logs])*TicksPerInGameWorkDay, LandResource_Forest_Oak.PerDayProductionPerAgent()/24*10))
	request.Gemini(fmt.Sprintf("Granite:      %d (%+.2f/workday)\n", colony.resourceCounts[Resource_Granite], math.Ceil(colony.currentProduction[Resource_Granite]-colony.currentConsumption[Resource_Granite])*TicksPerInGameWorkDay))
	request.Gemini(fmt.Sprintf("Coal:         %d (+0/workday)\n", colony.resourceCounts[Resource_Coal]))
	request.Gemini(fmt.Sprintf("Iron:         %d (+0/workday)\n", colony.resourceCounts[Resource_Iron]))
	// request.Gemini(fmt.Sprintf("Production Factor: %d\n", colony.productionFactor)) // The efficiency of all production in colony
	// request.Gemini(fmt.Sprintf("Next Update in")) // TODO: Get real-time duration till next building update.

	request.Gemini("\nClimate Stats:\n")
	request.Gemini(fmt.Sprintf("Altitude:     %.2f\n", Map[colony.tileLocation.Y][colony.tileLocation.X].altitude))
	request.Gemini(fmt.Sprintf("Avg Temp:     %.1f°C (%f)\n", ConvertTemperature(Map[colony.tileLocation.Y][colony.tileLocation.X].climate.avgTemp).Celsius, Map[colony.tileLocation.Y][colony.tileLocation.X].climate.avgTemp))
	request.Gemini(fmt.Sprintf("Avg Winter:   %.1f°C (%f)\n", ConvertTemperature(Map[colony.tileLocation.Y][colony.tileLocation.X].climate.winterTemp).Celsius, Map[colony.tileLocation.Y][colony.tileLocation.X].climate.winterTemp))
	request.Gemini(fmt.Sprintf("Avg Spring:   %.1f°C (%f)\n", ConvertTemperature(Map[colony.tileLocation.Y][colony.tileLocation.X].climate.springTemp).Celsius, Map[colony.tileLocation.Y][colony.tileLocation.X].climate.springTemp))
	request.Gemini(fmt.Sprintf("Avg Summer:   %.1f°C (%f)\n", ConvertTemperature(Map[colony.tileLocation.Y][colony.tileLocation.X].climate.summerTemp).Celsius, Map[colony.tileLocation.Y][colony.tileLocation.X].climate.summerTemp))
	request.Gemini(fmt.Sprintf("Avg Autumn:   %.1f°C (%f)\n", ConvertTemperature(Map[colony.tileLocation.Y][colony.tileLocation.X].climate.fallTemp).Celsius, Map[colony.tileLocation.Y][colony.tileLocation.X].climate.fallTemp))
	request.Gemini("```\n")
	request.Gemini("\n")

	// Pages
	request.Link("/build/", "Build")
	request.Link("/research/", "Research")
	// request.Link("/trade/", "Trade")
	// request.Link("/resources/", "Resources")
	// request.Link("/stats/", "Stats")
	// request.Link("/laws/", "Laws")

	// Resource Zones List
	request.Heading(2, "Natural Resource Zones")
	for i, zone := range colony.landResources {
		if zone.landResource == LandResource_Unknown {
			continue
		}

		if zone.amount == 0 {
			request.Link("/resource_zone/"+strconv.Itoa(i), zone.landResource.ToString()+" (depleted)")
		} else {
			request.Link("/resource_zone/"+strconv.Itoa(i), zone.landResource.ToString())
		}
	}

	// Action Links
}

func (colony *Colony) ResourceZonePage(request *sis.Request) {
	resourceId, _ := strconv.Atoi(request.GetParam("id"))
	zone := &colony.landResources[resourceId]

	request.Heading(1, "Resource Zone: "+zone.landResource.ToString())
	request.Gemini("\n")

	request.Gemini("```Statistics\n")
	request.Gemini(fmt.Sprintf("Workers: %d / 20", len(zone.workers)))
	request.Gemini(fmt.Sprintf("Amount Left to Harvest: %d", zone.amount))
	request.Gemini("```\n")

	request.Gemini("\n")
	request.Link(path.Join("/resource_zone/", request.GetParam("id"), "/add_worker"), "Add Worker")
	request.Link(path.Join("/resource_zone/", request.GetParam("id"), "/remove_worker"), "Remove Worker")
	request.Gemini("\n")
	request.Link("/", "Back to Colony Overview")
}

func (colony *Colony) AddWorkerPage(request *sis.Request) {
	resourceId, _ := strconv.Atoi(request.GetParam("id"))
	zone := &colony.landResources[resourceId]

	// Pick a (random) unemployed worker to add to the zone, if one is available
	for id := range colony.agents {
		a := &colony.agents[id]

		if a.assignedZone == -1 {
			zone.AddWorker(AgentId(id), a)
			break
		}
	}

	// Redirect back to the zone's page
	request.Redirect("/resource_zone/%s/", request.GetParam("id"))
}

func (colony *Colony) RemoveWorkerPage(request *sis.Request) {
	resourceId, _ := strconv.Atoi(request.GetParam("id"))
	zone := &colony.landResources[resourceId]
	zone.RemoveLastWorker(colony)

	request.Redirect("/resource_zone/%s/", request.GetParam("id"))
}

func (ctx *Context) IsWorkTime() bool {
	workTime_start := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 8, 0, 0, 0, ctx.inGameTime.Location())
	workTime_end := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 8+WorkHours, 0, 0, 0, ctx.inGameTime.Location())
	return ctx.inGameTime.Equal(workTime_start) || (ctx.inGameTime.After(workTime_start) && ctx.inGameTime.Before(workTime_end))
}
func (ctx *Context) IsFreeTime() bool {
	freeTime_start := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 8+WorkHours, 0, 0, 0, ctx.inGameTime.Location())
	midnight := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day()+1, 0, 0, 0, 0, ctx.inGameTime.Location()) // Tomorrow
	return ctx.inGameTime.Equal(freeTime_start) || (ctx.inGameTime.After(freeTime_start) && ctx.inGameTime.Before(midnight))
}
func (ctx *Context) IsSleepTime() bool {
	midnight := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 0, 0, 0, 0, ctx.inGameTime.Location())
	workTime_start := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 8, 0, 0, 0, ctx.inGameTime.Location())
	return ctx.inGameTime.Equal(midnight) || (ctx.inGameTime.After(midnight) && ctx.inGameTime.Before(workTime_start))
}
