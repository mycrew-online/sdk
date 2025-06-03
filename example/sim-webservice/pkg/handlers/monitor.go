package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"sim-webservice/pkg/simconnect"
)

// MonitorHandler handles HTTP requests related to monitoring
type MonitorHandler struct {
	monitorClient *simconnect.MonitorClient
}

// NewMonitorHandler creates a new monitor handler
func NewMonitorHandler(monitorClient *simconnect.MonitorClient) *MonitorHandler {
	return &MonitorHandler{
		monitorClient: monitorClient,
	}
}

// HandleIndex serves the main web interface
func (mh *MonitorHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

// HandleMonitorAPI serves monitor data as JSON
func (mh *MonitorHandler) HandleMonitorAPI(w http.ResponseWriter, r *http.Request) {
	data := mh.monitorClient.GetCurrentData()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleCameraStateToggle handles setting the camera state
func (mh *MonitorHandler) HandleCameraStateToggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.SetCameraState(w, r)
}

// HandleExternalPowerToggle handles toggling external power
func (mh *MonitorHandler) HandleExternalPowerToggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.ToggleExternalPowerHandler(w, r)
}
