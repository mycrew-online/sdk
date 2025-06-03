package simconnect

import (
	"time"

	"github.com/mycrew-online/sdk/pkg/types"
)

// updateSystemEvents processes system events from the simulator
func (wc *WeatherClient) updateSystemEvents(event *types.EventData) {
	// Get write lock for system events
	wc.systemEvents.mutex.Lock()
	defer wc.systemEvents.mutex.Unlock()

	// Update based on event ID
	switch event.EventID {
	case SIM_STATE_EVENT_ID: // Sim state event (running/stopped)
		if event.EventData == 1 {
			wc.systemEvents.SimRunning = true
			wc.systemEvents.LastEventName = "Simulator Running"
		} else {
			wc.systemEvents.SimRunning = false
			wc.systemEvents.LastEventName = "Simulator Stopped"
		}

	case PAUSE_EVENT_ID: // Pause event
		if event.EventData == 1 {
			wc.systemEvents.SimPaused = true
			wc.systemEvents.LastEventName = "Simulator Paused"
		} else {
			wc.systemEvents.SimPaused = false
			wc.systemEvents.LastEventName = "Simulator Resumed"
		}

	case AIRCRAFT_LOADED_EVENT_ID: // Aircraft loaded
		wc.systemEvents.AircraftLoaded = true
		wc.systemEvents.LastEventName = "Aircraft Loaded"

	case FLIGHT_LOADED_EVENT_ID: // Flight loaded
		wc.systemEvents.FlightLoaded = true
		wc.systemEvents.LastEventName = "Flight Loaded"
	}

	// Update timestamp
	wc.systemEvents.LastEventTime = time.Now()
}
