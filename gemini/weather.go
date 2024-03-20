package gemini

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"gitlab.com/clseibold/auragem_sis/config"
	sis "gitlab.com/clseibold/smallnetinformationservices"
)

var apiKey = config.WeatherApiKey

func handleWeather(g sis.ServerHandle) {
	publishDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-19T13:51:00", time.Local)
	g.AddRoute("/weather", func(request sis.Request) {
		request.Redirect("/weather/")
	})
	g.AddRoute("/weather/", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: publishDate, UpdateDate: time.Now(), Language: "en", Abstract: "# Weather\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}
		iqAirResponse := getNearestLocation(request)
		request.Gemini(fmt.Sprintf(`# Weather for %s, %s, %s

%s

Temperature: %.1f°C / %.1f°F
Pressure: %d hPa
Humidity: %d%%
Wind Speed: %.2f m/s
US AQI: %d (%s)

Powered by IQAir
`, iqAirResponse.Data.City, iqAirResponse.Data.State, iqAirResponse.Data.Country, getIconCodeDescription(iqAirResponse.Data.Current.Weather.IconCode), iqAirResponse.Data.Current.Weather.Temperature, celsiusToFahrenheit(iqAirResponse.Data.Current.Weather.Temperature), iqAirResponse.Data.Current.Weather.Pressure, iqAirResponse.Data.Current.Weather.Humidity, iqAirResponse.Data.Current.Weather.WindSpeed, iqAirResponse.Data.Current.Pollution.AQI_US, aqiDescription(iqAirResponse.Data.Current.Pollution.AQI_US)))
	})
}

// ----- IQAir API ----------

// Gets nearest city location using IP Address geolocation
// http://api.airvisual.com/v2/nearest_city?key={{YOUR_API_KEY}}
func getNearestLocation(request sis.Request) IQAirResponse {
	url := "http://api.airvisual.com/v2/nearest_city?key=" + apiKey + "&x-forwarded-for=" + request.IP
	if !IsPublicIP(net.ParseIP(request.IP)) {
		url = "http://api.airvisual.com/v2/nearest_city?key=" + apiKey
	}
	fmt.Printf("IP: %s\n", request.IP)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
		return IQAirResponse{}
	}

	req.Header.Add("User-Agent", "AuraGem")
	if IsPublicIP(net.ParseIP(request.IP)) {
		req.Header.Add("x-forwarded-for", request.IP)
	}
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		fmt.Println(err)
		return IQAirResponse{}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println(err)
		return IQAirResponse{}
	}

	var iqAirResponse IQAirResponse
	jsonErr := json.Unmarshal(body, &iqAirResponse)
	fmt.Printf("Status: %s", iqAirResponse.Status) // TODO: Check for statuses like call_limit_reached, ip_location_failed, no_nearest_station, or too_many_requests
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return IQAirResponse{}
	}

	return iqAirResponse
}

type IQAirResponse struct {
	Status string            `json:"status"`
	Data   IQAirResponseData `json:"data"`
}

type IQAirResponseData struct {
	City     string           `json:"city"`
	State    string           `json:"state"`
	Country  string           `json:"country"`
	Location IQAirLocation    `json:"location"`
	Forecast ForecastData     `json:"forecasts"`
	Current  IQAirCurrentInfo `json:"current"`
	History  HistoryData      `json:"history"`
}

type IQAirLocation struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type IQAirCurrentInfo struct {
	Weather   IQAirWeather   `json:"weather"`
	Pollution IQAirPollution `json:"pollution"`
}

type IQAirWeather struct {
	Timestamp       time.Time `json:"ts"`
	Temperature     float64   `json:"tp"`
	Temperature_Min float64   `json:"tp_min"`
	Pressure        int       `json:"pr"`
	Humidity        int       `json:"hu"`
	WindSpeed       float64   `json:"ws"`
	WindDirection   int       `json:"wd"`
	IconCode        string    `json:"ic"`
}

type IQAirPollution struct {
	Timestamp    time.Time `json:"ts"`
	AQI_US       int       `json:"aqius"`
	Main_US      string    `json:"mainus"`
	AQI_Chinese  int       `json:"aqicn"`
	Main_Chinese string    `json:"maincn"`
	P2           IQAirP2   `json:"p2"`
}

type IQAirP2 struct {
	Conc        float64 `json:"conc"`
	AQI_US      int     `json:"aqius"`
	AQI_Chinese int     `json:"aqicn"`
}

type ForecastData struct {
	Timestamp       time.Time `json:"ts"`
	AQI_US          int       `json:"aqius"`
	AQI_Chinese     int       `json:"aqicn"`
	Temperature     float64   `json:"tp"`
	Temperature_Min float64   `json:"tp_min"`
	Pressure        int       `json:"pr"`
	Humidity        int       `json:"hu"`
	WindSpeed       float64   `json:"ws"`
	WindDirection   int       `json:"wd"`
	IconCode        string    `json:"ic"`
}
type HistoryData struct {
	Weather   []IQAirWeather   `json:"weather"`
	Pollution []IQAirPollution `json:"pollution"`
}

func getIconCodeDescription(iconCode string) string {
	switch iconCode {
	case "01d":
		return "☀️ Clear sky (day)"
	case "01n":
		return "🌙 Clear sky (night)"
	case "02d":
		return "⛅ Few clouds (day)"
	case "02n":
		return "☁️ Few clouds (night)"
	case "03d", "03n":
		return "☁️ Scattered clouds"
	case "04d", "04n":
		return "☁️ Broken clouds"
	case "09d", "09n":
		return "🌧️ Shower rain"
	case "10d":
		return "🌦️ Rain (day)"
	case "10n":
		return "🌧️ Rain (night)"
	case "11d", "11n":
		return "🌩️ Thunderstorm"
	case "13d", "13n":
		return "🌨️ Snow"
	case "50d", "50n":
		return "🌫️ Mist"
	default:
		return ""
	}
}

func aqiDescription(aqi int) string {
	if aqi < 50 {
		return "Good"
	} else if aqi <= 100 {
		return "Moderate"
	} else if aqi <= 150 {
		return "Unhealthy for Sensitive Groups"
	} else if aqi <= 200 {
		return "Unhealthy"
	} else if aqi <= 300 {
		return "Very Unhealthy"
	} else if aqi <= 500 {
		return "Hazardous"
	}
	return ""
}

func celsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 9.0 / 5.0) + 32.0
}

// ----- Get Public IP when request comes from local network ------

func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

type IP struct {
	Query string
}

func getip2() string {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return err.Error()
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err.Error()
	}

	var ip IP
	json.Unmarshal(body, &ip)

	return ip.Query
}
