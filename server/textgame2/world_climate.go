package textgame2

import (
	"fmt"
	"math"
)

// TODO: Should the climate generation be more accurate by generating cloud and wind paths?

// Climate tracks seasonal variations in temperature and rainfall
type Climate struct {
	// Temperature for each season (0.0 = very cold, 1.0 = very hot)
	winterTemp float64
	springTemp float64
	summerTemp float64
	fallTemp   float64

	// Rainfall for each season (0.0 = extremely dry, 1.0 = extremely wet)
	winterRain float64
	springRain float64
	summerRain float64
	fallRain   float64

	// Annual averages (for convenience)
	avgTemp float64
	avgRain float64
}

// Season enum for accessing specific seasonal values
type Season int

const (
	Winter Season = iota
	Spring
	Summer
	Fall
)

// TemperatureResult holds the converted temperature values
type TemperatureResult struct {
	Celsius    float64
	Fahrenheit float64
}

// ConvertTemperature converts a normalized temperature value (0.0-1.0)
// to real-world Celsius and Fahrenheit temperatures
func ConvertTemperature(normalizedTemp float64) TemperatureResult {
	// Define the temperature scale
	// 0.0 maps to -30°C (extremely cold)
	// 1.0 maps to 45°C (extremely hot)

	// Linear mapping from normalized (0.0-1.0) to Celsius (-30 to 45)
	celsius := -30.0 + normalizedTemp*75.0

	// Convert Celsius to Fahrenheit
	fahrenheit := celsius*9.0/5.0 + 32.0

	return TemperatureResult{
		Celsius:    math.Round(celsius*10) / 10,    // Round to 1 decimal place
		Fahrenheit: math.Round(fahrenheit*10) / 10, // Round to 1 decimal place
	}
}

// GetSeasonalTemperatures returns the temperatures for all seasons in the given tile
func (t *Tile) GetSeasonalTemperatures() map[string]TemperatureResult {
	return map[string]TemperatureResult{
		"Winter": ConvertTemperature(t.climate.winterTemp),
		"Spring": ConvertTemperature(t.climate.springTemp),
		"Summer": ConvertTemperature(t.climate.summerTemp),
		"Fall":   ConvertTemperature(t.climate.fallTemp),
		"Annual": ConvertTemperature(t.climate.avgTemp),
	}
}

// GetTemperatureDescription returns a human-readable description of the temperature
func GetTemperatureDescription(normalizedTemp float64) string {
	temps := ConvertTemperature(normalizedTemp)

	// Create descriptive text based on temperature range
	if normalizedTemp < 0.15 {
		return fmt.Sprintf("Frigid (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	} else if normalizedTemp < 0.3 {
		return fmt.Sprintf("Very Cold (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	} else if normalizedTemp < 0.4 {
		return fmt.Sprintf("Cold (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	} else if normalizedTemp < 0.5 {
		return fmt.Sprintf("Cool (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	} else if normalizedTemp < 0.6 {
		return fmt.Sprintf("Mild (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	} else if normalizedTemp < 0.7 {
		return fmt.Sprintf("Warm (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	} else if normalizedTemp < 0.8 {
		return fmt.Sprintf("Hot (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	} else if normalizedTemp < 0.9 {
		return fmt.Sprintf("Very Hot (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	} else {
		return fmt.Sprintf("Scorching (%.1f°C/%.1f°F)", temps.Celsius, temps.Fahrenheit)
	}
}

// GetRainfallDescription returns a human-readable description of rainfall
func GetRainfallDescription(normalizedRain float64) string {
	// Convert normalized rainfall to approximate mm/year
	// 0.0 = ~0mm, 1.0 = ~3000mm (very wet rainforest)
	mmPerYear := normalizedRain * 3000

	if normalizedRain < 0.15 {
		return fmt.Sprintf("Arid (%.0f mm/year)", mmPerYear)
	} else if normalizedRain < 0.3 {
		return fmt.Sprintf("Very Dry (%.0f mm/year)", mmPerYear)
	} else if normalizedRain < 0.45 {
		return fmt.Sprintf("Dry (%.0f mm/year)", mmPerYear)
	} else if normalizedRain < 0.6 {
		// return fmt.Sprintf("Moderate (%.0f mm/year)", mmPerYear)
		return fmt.Sprintf("Moderately Precipitous (%.0f mm/year)", mmPerYear)
	} else if normalizedRain < 0.75 {
		return fmt.Sprintf("Wet (%.0f mm/year)", mmPerYear)
	} else if normalizedRain < 0.9 {
		return fmt.Sprintf("Very Wet (%.0f mm/year)", mmPerYear)
	} else {
		return fmt.Sprintf("Extremely Wet (%.0f mm/year)", mmPerYear)
	}
}
