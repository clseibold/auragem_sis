package textgame2

import (
	"fmt"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

func PrintWorldMap(request *sis.Request) {
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
	}

	request.Heading(1, "World Map")
	request.Gemini("\n")
	if !showValues {
		request.Link("/world-map?values", "Show Values")
		request.Link("/world-map?mountains", "Show Mountain Ranges")
		request.Link("/world-map?landtypes", "Show Land Types")
	} else {
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
		// Heading
		if y == 0 {
			if showValues {
				request.PlainText("|     |")
			} else {
				request.PlainText("|  |")
			}
			for x := range MapWidth {
				if showValues {
					request.PlainText("%5d|", x)
				} else {
					request.PlainText("%2d|", x)
				}
			}
			request.PlainText("\n")
			if showValues {
				request.PlainText("\n")
			}
		}

		if showValues {
			request.PlainText("|%5d|", y)
		} else {
			request.PlainText("|%2d|", y)
		}
		for x := range MapWidth {
			if showValues {
				request.PlainText(fmt.Sprintf("%+.2f|", Map[y][x].altitude))
			} else {
				// altitude := Map[y][x].altitude
				switch Map[y][x].landType {
				case LandType_Water:
					request.PlainText(" ~|")
				case LandType_Mountains:
					request.PlainText(" A|") // Mountain
				case LandType_Plateaus:
					request.PlainText(" =|") // Plateau
				case LandType_Hills:
					if Map[y][x].altitude >= 0.8 {
						request.PlainText(" n|") // High hills/foothills
					} else {
						request.PlainText(" +|") // Regular hills
					}
				case LandType_Valleys:
					request.PlainText(" v|") // Valley
				case LandType_Coastal:
					request.PlainText(" c|") // Coastal
				case LandType_Plains:
					request.PlainText("  |") // Plains
				default:
					request.PlainText("  |") // Default plains
				}
			}
		}
		request.PlainText("\n")
		if showValues {
			request.PlainText("\n")
		}
	}

	request.PlainText("\nBase Perlin Noise:\n")
	for y := range MapHeight {
		// Heading
		if y == 0 {
			if showValues {
				request.PlainText("|     |")
			} else {
				request.PlainText("|  |")
			}
			for x := range MapWidth {
				if showValues {
					request.PlainText(fmt.Sprintf("%5d|", x))
				} else {
					request.PlainText(fmt.Sprintf("%2d|", x))
				}
			}
			request.PlainText("\n")
			if showValues {
				request.PlainText("\n")
			}
		}

		// Values
		if showValues {
			request.PlainText("|%5d|", y)
		} else {
			request.PlainText("|%2d|", y)
		}
		for x := range MapWidth {
			if showValues {
				request.PlainText(fmt.Sprintf("%+.2f|", MapPerlin[y][x].altitude))
			} else {
				altitude := MapPerlin[y][x].altitude
				if altitude <= 0 {
					request.PlainText(" ~|") // Water
				} else if altitude >= 1 {
					request.PlainText(" A|") // Mountain
				} else if altitude >= 0.5 { // Hills?
					if Map[y][x].landType == LandType_Plateaus {
						request.PlainText(" =|")
					} else {
						request.PlainText(" n|")
					}
				} else if altitude >= 0.3 {
					request.PlainText(" +|")
				} else {
					request.PlainText("  |") // Plains
				}
			}
		}
		request.PlainText("\n")
		if showValues {
			request.PlainText("\n")
		}
	}
	request.Gemini("```\n")

	request.Gemini("Legend:\n")
	request.Gemini("- ~: Water\n")
	request.Gemini("- (space): Plains\n")
	request.Gemini("- +: Hills\n")
	request.Gemini("- v: Valleys\n")
	request.Gemini("- =: Plateaus\n")
	request.Gemini("- A: Mountains\n")
	request.Gemini("- c: Coastal\n")
	request.Gemini("- d: Sand Dunes\n")
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
		request.PlainText("%2d|", x)
	}
	request.PlainText("\n")

	// Print the map
	for y := range MapHeight {
		request.PlainText("|%2d|", y)
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
				symbol = "v"
			case LandType_Plateaus:
				symbol = "="
			case LandType_Mountains:
				symbol = "A"
			case LandType_Coastal:
				symbol = "c"
			case LandType_SandDunes:
				symbol = "d"
			default:
				symbol = "?"
			}

			request.PlainText(" %s|", symbol)
		}
		request.PlainText("\n")
	}

	request.Gemini("```\n")

	request.Gemini("Legend:\n")
	request.Gemini("- ~: Water\n")
	request.Gemini("- (space): Plains\n")
	request.Gemini("- +: Hills\n")
	request.Gemini("- v: Valleys\n")
	request.Gemini("- =: Plateaus\n")
	request.Gemini("- A: Mountains\n")
	request.Gemini("- c: Coastal\n")
	request.Gemini("- d: Sand Dunes\n")
}
