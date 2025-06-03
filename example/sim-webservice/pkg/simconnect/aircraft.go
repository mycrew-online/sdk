package simconnect

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mycrew-online/sdk/pkg/types"
)

// Constants for aircraft event IDs
const (
	TOGGLE_EXTERNAL_POWER_EVENT_ID = 2001
)

// Constants for aircraft notification group
const (
	AIRCRAFT_NOTIFICATION_GROUP_ID = 2000
)

// ToggleExternalPowerHandler handles toggling external power in MSFS
func (mc *MonitorClient) ToggleExternalPowerHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Transmit the TOGGLE_EXTERNAL_POWER event
	if err := mc.sdk.TransmitClientEvent(
		types.SIMCONNECT_OBJECT_ID_USER,
		TOGGLE_EXTERNAL_POWER_EVENT_ID,
		0, // No data value needed for toggle
		AIRCRAFT_NOTIFICATION_GROUP_ID,
		types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY,
	); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle external power: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "External power toggled successfully",
	})
}

// RegisterAircraftEvents registers aircraft control events with SimConnect
func (mc *MonitorClient) RegisterAircraftEvents() error {
	// Map the client event to the sim event
	if err := mc.sdk.MapClientEventToSimEvent(TOGGLE_EXTERNAL_POWER_EVENT_ID, "TOGGLE_EXTERNAL_POWER"); err != nil {
		return fmt.Errorf("failed to map TOGGLE_EXTERNAL_POWER event: %v", err)
	}

	// Set notification group priority
	if err := mc.sdk.SetNotificationGroupPriority(AIRCRAFT_NOTIFICATION_GROUP_ID, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST); err != nil {
		return fmt.Errorf("failed to set aircraft notification group priority: %v", err)
	}

	// Add the event to the notification group
	if err := mc.sdk.AddClientEventToNotificationGroup(AIRCRAFT_NOTIFICATION_GROUP_ID, TOGGLE_EXTERNAL_POWER_EVENT_ID, false); err != nil {
		return fmt.Errorf("failed to add TOGGLE_EXTERNAL_POWER to notification group: %v", err)
	}

	return nil
}
