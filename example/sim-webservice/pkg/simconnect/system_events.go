package simconnect

import (
	"time"

	"github.com/mycrew-online/sdk/pkg/types"
)

// updateSystemEvents processes system events from the simulator
func (mc *MonitorClient) updateSystemEvents(event *types.EventData) {
	// Get write lock for system events
	mc.systemEvents.mutex.Lock()
	defer mc.systemEvents.mutex.Unlock()

	// Update based on event ID
	switch event.EventID {
	case SIM_STATE_EVENT_ID: // Sim state event (running/stopped)
		if event.EventData == 1 {
			mc.systemEvents.SimRunning = true
			mc.systemEvents.LastEventName = "Simulator Running"
		} else {
			mc.systemEvents.SimRunning = false
			mc.systemEvents.LastEventName = "Simulator Stopped"
		}

	case PAUSE_EVENT_ID: // Pause event
		if event.EventData == 1 {
			mc.systemEvents.SimPaused = true
			mc.systemEvents.LastEventName = "Simulator Paused"
		} else {
			mc.systemEvents.SimPaused = false
			mc.systemEvents.LastEventName = "Simulator Resumed"
		}

	case AIRCRAFT_LOADED_EVENT_ID: // Aircraft loaded
		mc.systemEvents.AircraftLoaded = true
		mc.systemEvents.LastEventName = "Aircraft Loaded"

	case FLIGHT_LOADED_EVENT_ID: // Flight loaded
		mc.systemEvents.FlightLoaded = true
		mc.systemEvents.LastEventName = "Flight Loaded"
	}

	// Update timestamp
	mc.systemEvents.LastEventTime = time.Now()
}
