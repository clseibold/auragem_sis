package textgame2

import (
	"fmt"
	"strings"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

const globalDivider = "|"

func PrintWorldMap(request *sis.Request) {
	divider := " "
	noNumbers := true

	showValues := false
	query, _ := request.Query()
	if query == "values" {
		showValues = true
		/*} else if query == "mountains" { // OUTDATED and broken
		debugMountainDimensions(request)
		return*/
	} else if query == "withnumbers" {
		divider = globalDivider
		noNumbers = false
	}

	// Count land types for land type distribution chart
	landTypeCounts := make(map[LandType]int)

	for y := range MapHeight {
		for x := range MapWidth {
			landTypeCounts[Map[y][x].landType]++
		}
	}

	request.Heading(1, "World Map")
	request.Gemini("\n")
	if !showValues {
		if noNumbers {
			request.Link("/world-map?withnumbers", "Show With Map Numbers")
		} else {
			request.Link("/world-map", "Show Without Map Numbers")
		}
		request.Link("/world-map?values", "Show Values")
	} else {
		if noNumbers {
			request.Link("/world-map?withnumbers", "Show With Map Numbers")
		} else {
			request.Link("/world-map", "Show Without Map Numbers")
		}
		request.Link("/world-map", "Show Terrain")
	}
	request.Gemini("\n")

	request.Gemini("```\n")
	// Print the peaks
	request.PlainText("Peaks: ")
	for _, peak := range MapPeaks {
		request.PlainText("(%d, %d) ", peak.peakX, peak.peakY)
	}
	request.PlainText("\n")

	// Print the lowest and highest tiles
	request.PlainText("\nLowest and Highest Altitudes on Map with Mountain Peaks:\n")
	lowest, highest := getMapLowestAndHighestPoints()
	request.PlainText("Lowest Tile Altitude: %+.2f\n", lowest.altitude)
	request.PlainText("Highest Tile Altitude: %+.2f\n", highest.altitude)

	request.PlainText("\nWith Full Terrain Generation:\n")
	for y := range MapHeight {
		// Heading/Top border
		if y == 0 && !noNumbers {
			if showValues {
				request.PlainText(divider + "     " + divider)
			} else {
				request.PlainText(divider + "  " + divider)
			}
			for x := range MapWidth {
				if showValues {
					request.PlainText("%5d"+divider, x)
				} else {
					request.PlainText("%2d"+divider, x)
				}
			}
			request.PlainText("\n")
			if showValues {
				request.PlainText("\n")
			}
		} else if y == 0 && noNumbers {
			request.PlainText(strings.Repeat("-", (MapWidth+2)*3))
			request.PlainText("\n")
		}

		if !noNumbers {
			if showValues {
				request.PlainText(divider+"%5d"+divider, y)
			} else {
				request.PlainText(divider+"%2d"+divider, y)
			}
		} else { // Left Border
			request.PlainText("|")
		}
		for x := range MapWidth {
			if showValues {
				request.PlainText(fmt.Sprintf("%+.2f"+divider, Map[y][x].altitude))
			} else {
				tile := &Map[y][x]
				// Prefix
				if tile.hasSpring && tile.hasPond {
					request.PlainText("⊙")
				} else if tile.hasSpring && tile.hasStream {
					request.PlainText("⊗")
				} else if tile.hasSpring {
					request.PlainText("⊕")
				} else if tile.hasMarsh && tile.hasPond {
					request.PlainText("⊛")
				} else if tile.hasPond {
					request.PlainText("o")
				} else if tile.hasMarsh {
					request.PlainText("≈")
				} else if tile.hasStream {
					request.PlainText(".")
				} else if tile.hasGrove {
					request.PlainText("Υ")
				} else if tile.hasMeadow {
					request.PlainText("⁎")
				} else if tile.hasScrub {
					request.PlainText("⌿")
				} else if tile.hasRocks {
					request.PlainText("◊")
				} else if tile.hasSaltFlat {
					request.PlainText("□")
				} else if tile.hasFloodArea {
					request.PlainText("∴")
				} else if tile.hasGameTrail {
					request.PlainText("-")
				} else {
					request.PlainText(" ")
				}

				switch Map[y][x].landType {
				case LandType_Water:
					request.PlainText("~")
				case LandType_Mountains:
					request.PlainText("▲") // Mountain
				case LandType_Plateaus:
					request.PlainText("≡") // Plateau
				case LandType_Hills:
					if Map[y][x].altitude >= 0.8 {
						request.PlainText("n") // High hills/foothills
					} else {
						request.PlainText("+") // Regular hills
					}
				case LandType_Valleys:
					request.PlainText("⌄") // Valley
				case LandType_Coastal:
					request.PlainText("c") // Coastal
				case LandType_Plains:
					request.PlainText(" ") // Plains
				case LandType_SandDunes:
					request.PlainText("s") // Sand Dunes
				default:
					request.PlainText(" ") // Default plains
				}
				request.PlainText(divider)
			}
		}

		if noNumbers { // Right Border
			request.PlainText("|")
		}

		request.PlainText("\n")
		if showValues {
			request.PlainText("\n")
		}

		// Bottom border
		if noNumbers && y == MapWidth-1 {
			request.PlainText(strings.Repeat("-", (MapWidth+2)*3))
			request.PlainText("\n")
		}
	}

	request.PlainText("\nLegend:\n")
	request.PlainText("o : Small pond (contained within a tile)\n")
	request.PlainText(". : Small stream (width contained within a tile)\n")
	request.PlainText("⊕ : Spring\n")
	request.PlainText("⊗ : Spring with stream\n")
	request.PlainText("⊙ : Spring with pond\n")
	request.PlainText("≈ : Marsh (soggy ground)\n")
	request.PlainText("⊛ : Marsh with pond\n")

	request.Gemini("Υ : Grove of trees\n")
	request.Gemini("⁎ : Flower meadow\n")
	request.Gemini("⌿ : Scrubland\n")
	request.Gemini("◊ : Rock outcroppings\n")
	request.Gemini("□ : Salt flat\n")
	request.Gemini("∴ : Seasonal flood area\n")
	request.Gemini("- : Game trail\n")

	request.PlainText(" ~: Water (lake/river)\n")
	request.PlainText(" (space): Plains\n")
	request.PlainText(" +: Hills\n")
	request.PlainText(" n: Foothills\n")
	request.PlainText(" ⌄: Valleys\n")
	request.PlainText(" ≡: Plateaus\n")
	request.PlainText(" ▲: Mountains\n")
	request.PlainText(" c: Coastal\n")
	request.PlainText(" d: Sand Dunes\n")

	request.PlainText("\nLand Type Distribution:\n")
	request.Gemini("| Land Type  | Count | Percentage |\n")
	request.Gemini("|------------|-------|------------|\n")

	totalTiles := MapWidth * MapHeight

	// Print in a specific order for readability
	landTypes := []LandType{
		LandType_Water,
		LandType_Plains,
		LandType_Hills,
		LandType_Valleys,
		LandType_Plateaus,
		LandType_Mountains,
		LandType_Coastal,
		LandType_SandDunes,
	}

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

	for _, lt := range landTypes {
		count := landTypeCounts[lt]
		percentage := float64(count) / float64(totalTiles) * 100.0
		request.Gemini(fmt.Sprintf("| %-10s | %-5d | %-9.2f%% |\n", landTypeNames[lt], count, percentage))
	}
	request.PlainText("```\n")
}
