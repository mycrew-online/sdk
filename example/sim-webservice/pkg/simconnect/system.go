package simconnect

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Constants for system event IDs
const (
	// System Event Define IDs
	SIM_STATE_EVENT_ID       = 1010
	PAUSE_EVENT_ID           = 1020
	AIRCRAFT_LOADED_EVENT_ID = 1030
	FLIGHT_LOADED_EVENT_ID   = 1040
	SIM_START_STOP_EVENT_ID  = 1050
)

// SystemEvents stores the current state of simulator system events
type SystemEvents struct {
	SimRunning     bool      `json:"simRunning"`
	SimPaused      bool      `json:"simPaused"`
	AircraftLoaded bool      `json:"aircraftLoaded"`
	FlightLoaded   bool      `json:"flightLoaded"`
	LastEventTime  time.Time `json:"lastEventTime"`
	LastEventName  string    `json:"lastEventName"`
	mutex          sync.RWMutex
}

// RegisterSystemEvents subscribes to system events from the simulator
func (wc *WeatherClient) RegisterSystemEvents() error {
	// Subscribe to Sim state events (running/stopped)
	if err := wc.sdk.SubscribeToSystemEvent(SIM_STATE_EVENT_ID, "Sim"); err != nil {
		return fmt.Errorf("failed to subscribe to Sim state events: %v", err)
	}

	// Subscribe to Pause events
	if err := wc.sdk.SubscribeToSystemEvent(PAUSE_EVENT_ID, "Pause"); err != nil {
		return fmt.Errorf("failed to subscribe to Pause events: %v", err)
	}

	// Subscribe to AircraftLoaded events
	if err := wc.sdk.SubscribeToSystemEvent(AIRCRAFT_LOADED_EVENT_ID, "AircraftLoaded"); err != nil {
		return fmt.Errorf("failed to subscribe to AircraftLoaded events: %v", err)
	}

	// Subscribe to FlightLoaded events
	if err := wc.sdk.SubscribeToSystemEvent(FLIGHT_LOADED_EVENT_ID, "FlightLoaded"); err != nil {
		return fmt.Errorf("failed to subscribe to FlightLoaded events: %v", err)
	}

	return nil
}

// GetSystemEventsHandler returns the current system events state
func (wc *WeatherClient) GetSystemEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get a read lock on the system events
	wc.systemEvents.mutex.RLock()
	defer wc.systemEvents.mutex.RUnlock()

	// Return the current system events
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wc.systemEvents)
}
