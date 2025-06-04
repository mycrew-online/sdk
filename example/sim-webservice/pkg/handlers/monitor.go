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

// HandleAircraftExitToggle handles toggling aircraft exit (canopy)
func (mh *MonitorHandler) HandleAircraftExitToggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.ToggleAircraftExitHandler(w, r)
}

// HandleCabinNoSmokingToggle handles toggling cabin no smoking alert
func (mh *MonitorHandler) HandleCabinNoSmokingToggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.ToggleCabinNoSmokingAlertHandler(w, r)
}

// HandleCabinSeatbeltsToggle handles toggling cabin seatbelts alert
func (mh *MonitorHandler) HandleCabinSeatbeltsToggle(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.ToggleCabinSeatbeltsAlertHandler(w, r)
}

// HandleCabinNoSmokingSet handles setting cabin no smoking alert to a specific state
func (mh *MonitorHandler) HandleCabinNoSmokingSet(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.SetCabinNoSmokingAlertHandler(w, r)
}

// HandleCabinSeatbeltsSet handles setting cabin seatbelts alert to a specific state
func (mh *MonitorHandler) HandleCabinSeatbeltsSet(w http.ResponseWriter, r *http.Request) {
	mh.monitorClient.SetCabinSeatbeltsAlertHandler(w, r)
}
