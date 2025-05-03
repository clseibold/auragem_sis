# Gemini Text Game 2

* Genre: Survival Strategy Colony-management Simulation.
* Game loop uses 1-second ticks, and all simulations are based on their number of ticks.
  * Simulations can be sped-up by decreasing the tick duration.
* When first starting, should be able to do a lot that will affect later game in different ways. But after the first few real-time days, the time it takes to do stuff should lengthen a little bit until it becomes consistent. This allows new players to have something to do when just starting out. The further along a player is, the more things they will be managing in general, so longer wait times won't matter nearly as much. Additionally, the cycle times of each production should be staggered.
* 1 real-time day = ~4 in-game days (igd)
* 1 tile on world map is 10 square kilometers

## Environment

Environment of colony determines the available resources and their amounts. A colony's environment is a tile on a large map that's relative to other colonies in the multiplayer online world. Colonies can trade with each other for resources.

TODO: Create a map of biomes/land types that people can choose from. It should probably be procedurally generated so it's infinite?

### Time

* Each location is part of a particular time zone based on the longitudinal location of the colony within the world map.
* The Sunrise and Sunset times of a colony are based on the latitude and longitude location of the colony within the world map, as well as the altitude.
* Colonists only work during the day, and sleep at night, unless you are in emergency overtime (where sleep deprivation can lead to increased illness, injury, stress, and lower productivity).

### General Biome Types and Temperatures
* Warm:
  * Forest - Temperate, Tropical
  * Swamp - Temperate, Tropical
* Hot:
  * Arid Shrubland
  * Desert
  * Extreme Desert
* Cold:
  * Boreal Forest
  * Cold Bog
  * Tundra
  * Ice Sheet
  * Sea Ice (Hopeless)

### Land Types
* Coastal
* Hills
* Fluvial
* Mountains
* Valley

### Seasons and Weather

### Resource Zones

Resource Zones are zones of a particular resource within your region that can be harvested either manually or with buildings/technologies.

NOTE: Each resource has a strength that will be used to calculate the strength of the products they are used for.

Wells can be created for water in regions that do not have lakes or ponds.

### Renewable Resource Growth Rates

Plants (including trees) and animals within the land naturally grow and reproduce on their own, and thus have growth rates.

Resource Zones can include forests, lakes, ponds, berry bushes, metal deposits, coal deposits, etc.

Each metal, stone, and wood material should have at least one product that only it exclusively can make. This allow for each resource to have some unique usefuleness and will entice trade between regions/colonies.

## Population, Building, and Resource caps

* Forces colonies to trade with each other
* Population levels out based on the resources and space
* Immigration influx could stress resources, but there's always ways to manage it.
* People may choose to leave a colony based on bad conditions. They end up going to other colonies of other players in the map.

## Resource Consumption

* Water and Food consumption are dependent on the health status of each person. Teens, ill, and stressed persons need more food and water.
* Some resources take other resources to make. These production buildings will use what's available. If what's available is less than the maximum needed, then the resource will be split among each building. Buildings that get less produce less.
* Production Buildings need to be placed in a **dependency graph** so that we can start at the leaves that only take land resources and traverse up the dependencies.

## Building and Construction Times

* Housing/Building Insulation (can use dried grass, aka. straw, for this)

## Colony Research

Each colony can specialize in a particular line of research, but they will not be able to do every type of research and must pick and choose. This entices trade between colonies.

## Colonist Stats

* Each colonist has a productivity that determins the efficiency of the production at their workplace. The colonist productivity is based on the physical and mental health, stress, illness, and sleep statistics.

## Energy

## Trade Routes and Constructing Them

Before you can trade with others, you must have the following:
* Trade routes that pass by your colony and theirs
* Designated merchant workers
* A merchant outpost building?

## Guided Tutorial and Wiki

* Guided tutorial to describe basics of game:
  * Resource collection and Storage
  * Builds
  * Research
  * Overall colony stats
  * Stats of each colonist
  * Job assignment and unemployment
  * First things to do when starting a colony: wood collection (forresters and tree cutters), fire, farming, research, food gathering, cooks, housing
  * Which food to start with for the particular environment the player's colony is in
* More Info (wiki)
  * Gardening
  * Housing
  * Illness and mental health
  * Mining
  * Population loss and Immigration
  * Trade
  * Energy production
* Missions/Objectives/Directives
  * Goals/Directives to get started with the game
* Achievements
  * Population achievements
  * Trade stats
