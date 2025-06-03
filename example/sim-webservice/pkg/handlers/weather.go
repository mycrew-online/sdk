package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"sim-webservice/pkg/simconnect"
)

// WeatherHandler handles HTTP requests related to weather
type WeatherHandler struct {
	weatherClient *simconnect.WeatherClient
}

// NewWeatherHandler creates a new weather handler
func NewWeatherHandler(weatherClient *simconnect.WeatherClient) *WeatherHandler {
	return &WeatherHandler{
		weatherClient: weatherClient,
	}
}

// HandleIndex serves the main web interface
func (wh *WeatherHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

// HandleWeatherAPI serves weather data as JSON
func (wh *WeatherHandler) HandleWeatherAPI(w http.ResponseWriter, r *http.Request) {
	weather := wh.weatherClient.GetCurrentWeather()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weather)
}

// HandleCameraStateToggle handles setting the camera state
func (wh *WeatherHandler) HandleCameraStateToggle(w http.ResponseWriter, r *http.Request) {
	wh.weatherClient.SetCameraState(w, r)
}
