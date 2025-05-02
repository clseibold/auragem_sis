package textgame2

import (
	"fmt"
	"strings"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

const divider = ":"

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
	} else if query == "landtypes" {
		debugLandTypes(request)
		return
	} else if query == "withnumbers" {
		divider = ":"
		noNumbers = false
	}

	request.Heading(1, "World Map")
	request.Gemini("\n")
	if !showValues {
		if noNumbers {
			request.Link("/world-map?withborders", "Show With Map Numbers")
		} else {
			request.Link("/world-map", "Show Without Map Numbers")
		}
		request.Link("/world-map?values", "Show Values")
		request.Link("/world-map?landtypes", "Show Land Types")
	} else {
		if noNumbers {
			request.Link("/world-map?withborders", "Show With Map Numbers")
		} else {
			request.Link("/world-map", "Show Without Map Numbers")
		}
		request.Link("/world-map", "Show Terrain")
		request.Link("/world-map?landtypes", "Show Land Types")
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
				// Prefix
				if Map[y][x].hasPond {
					request.PlainText("o")
				} else if Map[y][x].hasStream {
					request.PlainText(".")
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

	request.PlainText("\nBase Perlin Noise:\n")
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
					request.PlainText(fmt.Sprintf("%5d"+divider, x))
				} else {
					request.PlainText(fmt.Sprintf("%2d"+divider, x))
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
				request.PlainText(fmt.Sprintf("%+.2f"+divider, MapPerlin[y][x].altitude))
			} else {
				altitude := MapPerlin[y][x].altitude
				request.PlainText(" ") // Prefix
				if altitude <= 0 {
					request.PlainText("~") // Water
				} else if altitude >= 1 {
					request.PlainText("▲") // Mountain
				} else if altitude >= 0.8 { // Foothills
					request.PlainText("n")
				} else if altitude >= 0.3 {
					request.PlainText("+")
				} else {
					request.PlainText(" ") // Plains
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
	request.PlainText("- ~: Water (lake/river)\n")
	request.PlainText("o: small pond\n")
	request.PlainText(".: small stream\n")
	request.PlainText("(space): Plains\n")
	request.PlainText("+: Hills\n")
	request.PlainText("n: Foothills\n")
	request.PlainText("⌄: Valleys\n")
	request.PlainText("≡: Plateaus\n")
	request.PlainText("▲: Mountains\n")
	request.PlainText("c: Coastal\n")
	request.PlainText("d: Sand Dunes\n")
	request.PlainText("```\n")
}

func debugLandTypes(request *sis.Request) {
	request.Heading(1, "Land Types Map")
	request.Gemini("\n")
	request.Link("/world-map/", "Back to World Map")
	request.Gemini("\n")

	// Count land types
	landTypeCounts := make(map[LandType]int)

	for y := range MapHeight {
		for x := range MapWidth {
			landTypeCounts[Map[y][x].landType]++
		}
	}

	// Print land type statistics
	request.Gemini("## Land Type Distribution\n\n")
	request.Gemini("```\n")
	request.Gemini("| Land Type | Count | Percentage |\n")
	request.Gemini("|-----------|-------|------------|\n")

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
		request.Gemini(fmt.Sprintf("| %-9s | %-5d | %-10.2f%% |\n", landTypeNames[lt], count, percentage))
	}
	request.Gemini("```\n")

	request.Gemini("\n## Land Types Map\n\n")
	request.Gemini("```\n")

	// Print map header
	request.PlainText("|  |")
	for x := range MapWidth {
		request.PlainText("%2d"+divider, x)
	}
	request.PlainText("\n")

	// Print the map
	for y := range MapHeight {
		request.PlainText(divider+"%2d"+divider, y)
		for x := range MapWidth {
			var symbol string

			switch Map[y][x].landType {
			case LandType_Water:
				symbol = "~"
			case LandType_Plains:
				symbol = " "
			case LandType_Hills:
				symbol = "+"
			case LandType_Valleys:
				symbol = "⌄"
			case LandType_Plateaus:
				symbol = "≡"
			case LandType_Mountains:
				symbol = "▲"
			case LandType_Coastal:
				symbol = "c"
			case LandType_SandDunes:
				symbol = "d"
			default:
				symbol = "?"
			}

			// Prefix
			if Map[y][x].hasPond {
				request.PlainText("o")
			} else if Map[y][x].hasStream {
				request.PlainText(".")
			} else {
				request.PlainText(" ")
			}
			request.PlainText("%s"+divider, symbol)
		}
		request.PlainText("\n")
	}

	request.PlainText("\nLegend:\n")
	request.PlainText("- ~: Water (lake/river)\n")
	request.PlainText("o: small pond\n")
	request.PlainText(".: small stream\n")
	request.PlainText("(space): Plains\n")
	request.PlainText("+: Hills\n")
	request.PlainText("n: Foothills\n")
	request.PlainText("⌄: Valleys\n")
	request.PlainText("≡: Plateaus\n")
	request.PlainText("▲: Mountains\n")
	request.PlainText("c: Coastal\n")
	request.PlainText("d: Sand Dunes\n")
	request.Gemini("```\n")
}
