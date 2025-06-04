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

// HandleBattery1Toggle handles toggling battery 1
func (mh *MonitorHandler) HandleBattery1Toggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.ToggleBattery1Handler(w, r)
}

// HandleBattery2Toggle handles toggling battery 2
func (mh *MonitorHandler) HandleBattery2Toggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.ToggleBattery2Handler(w, r)
}

// HandleApuMasterSwitchToggle handles toggling APU master switch
func (mh *MonitorHandler) HandleApuMasterSwitchToggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.ToggleApuMasterSwitchHandler(w, r)
}

// HandleApuStartButtonToggle handles toggling APU start button
func (mh *MonitorHandler) HandleApuStartButtonToggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.ToggleApuStartButtonHandler(w, r)
}
