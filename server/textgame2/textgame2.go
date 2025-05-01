package textgame2

import (
	"fmt"
	"path"
	"strconv"
	"time"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

const TickRealTimeDuration = time.Second
const InGameSecondsPerTick = 4

func TicksToInGameDuration(ticks int) time.Duration {
	return time.Duration(ticks*InGameSecondsPerTick) * time.Second
}

type Context struct {
	// previousTickTime time.Time
	inGameTime  time.Time
	ticker      *time.Ticker
	firstColony *Colony
}

func NewContext() *Context {
	// TODO: Load in saved game states from save directory (including ticker and time information)

	context := new(Context)
	context.ticker = time.NewTicker(TickRealTimeDuration)
	context.firstColony = NewColony(context, "Test Colony", 6)
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
	}
}

func (c *Context) Attach(s sis.ServeMux) {
	s.AddRoute("/", c.Homepage)
	group := s.Group("/test/")
	group.AddRoute("/", c.firstColony.ColonyPage)
	group.AddRoute("/resource_zone/:id/", c.firstColony.ResourceZonePage)
	group.AddRoute("/resource_zone/:id/add_worker", c.firstColony.AddWorkerPage)
	group.AddRoute("/resource_zone/:id/remove_worker", c.firstColony.RemoveWorkerPage)
}

func (c *Context) Homepage(request *sis.Request) {
	request.Heading(1, "Colony-Management Simulation Game")
	request.Gemini("\n")
	request.Link("/test/", "Test Colony")
}

func (colony *Colony) ColonyPage(request *sis.Request) {
	request.Heading(1, colony.name)
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
		request.Gemini(fmt.Sprintf("Date & Time: %s (Work)\n", colony.context.inGameTime.Format(time.TimeOnly)))
	} else if colony.context.IsSleepTime() {
		request.Gemini(fmt.Sprintf("Date & Time: %s (Sleep)\n", colony.context.inGameTime.Format(time.TimeOnly)))
	} else if colony.context.IsFreeTime() {
		request.Gemini(fmt.Sprintf("Date & Time: %s (Free Time)\n", colony.context.inGameTime.Format(time.TimeOnly)))
	}
	request.Gemini(fmt.Sprintf("Population:  %d (%d unemployed)\n", len(colony.agents), unemployedAgents))
	//request.Gemini(fmt.Sprintf("Food: %d\n", colony.resourceCounts[Resource_Food]))
	request.Gemini(fmt.Sprintf("Water:       %d (+0/cycle)\n", colony.resourceCounts[Resource_Water]))
	request.Gemini(fmt.Sprintf("Oak Wood:    %d (+0/cycle)\n", colony.resourceCounts[Resource_Wood_Oak]))
	request.Gemini(fmt.Sprintf("Coal:        %d (+0/cycle)\n", colony.resourceCounts[Resource_Coal]))
	request.Gemini(fmt.Sprintf("Iron:        %d (+0/cycle)\n", colony.resourceCounts[Resource_Iron]))
	// request.Gemini(fmt.Sprintf("Production Factor: %d\n", colony.productionFactor)) // The efficiency of all production in colony
	// request.Gemini(fmt.Sprintf("Next Update in")) // TODO: Get real-time duration till next building update.
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
	request.Heading(2, "Resource Zones")
	for i, zone := range colony.landResources {
		if zone.resource == LandResource_Unknown {
			continue
		}

		request.Link("/resource_zone/"+strconv.Itoa(i), zone.resource.ToString())
	}

	// Action Links
}

func (colony *Colony) ResourceZonePage(request *sis.Request) {
	resourceId, _ := strconv.Atoi(request.GetParam("id"))
	zone := &colony.landResources[resourceId]

	request.Heading(1, "Resource Zone: "+zone.resource.ToString())
	request.Gemini("\n")

	request.Gemini("```Statistics\n")
	request.Gemini(fmt.Sprintf("Workers: %d / 20", len(zone.workers)))
	request.Gemini(fmt.Sprintf("Amount Left to Harvest: %d", zone.amount))
	request.Gemini("```\n")

	request.Gemini("\n")
	request.Link(path.Join("/resource_zone/", request.GetParam("id"), "/add_worker"), "Add Worker")
	request.Link(path.Join("/resource_zone/", request.GetParam("id"), "/remove_worker"), "Remove Worker")
}

func (colony *Colony) AddWorkerPage(request *sis.Request) {
	resourceId, _ := strconv.Atoi(request.GetParam("id"))
	zone := &colony.landResources[resourceId]

	// Pick a (random) worker to add to the zone
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
	workTime_end := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 12+6, 0, 0, 0, ctx.inGameTime.Location())
	return ctx.inGameTime.Equal(workTime_start) || (ctx.inGameTime.After(workTime_start) && ctx.inGameTime.Before(workTime_end))
}
func (ctx *Context) IsFreeTime() bool {
	freeTime_start := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 12+6, 0, 0, 0, ctx.inGameTime.Location())
	midnight := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day()+1, 0, 0, 0, 0, ctx.inGameTime.Location()) // Tomorrow
	return ctx.inGameTime.Equal(freeTime_start) || (ctx.inGameTime.After(freeTime_start) && ctx.inGameTime.Before(midnight))
}
func (ctx *Context) IsSleepTime() bool {
	midnight := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 0, 0, 0, 0, ctx.inGameTime.Location())
	workTime_start := time.Date(ctx.inGameTime.Year(), ctx.inGameTime.Month(), ctx.inGameTime.Day(), 8, 0, 0, 0, ctx.inGameTime.Location())
	return ctx.inGameTime.Equal(midnight) || (ctx.inGameTime.After(midnight) && ctx.inGameTime.Before(workTime_start))
}
