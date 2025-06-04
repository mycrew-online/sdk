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
	TOGGLE_AIRCRAFT_EXIT_EVENT_ID  = 2002
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

// ToggleBattery1Handler handles toggling battery 1 in MSFS
func (mc *MonitorClient) ToggleBattery1Handler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current battery 1 switch state and toggle it
	mc.mutex.RLock()
	currentState := mc.currentData.Battery1Switch
	mc.mutex.RUnlock()

	// Toggle the state: 0 -> 1, 1 -> 0
	newState := int32(1)
	if currentState == 1 {
		newState = 0
	}

	// Set the SimVar using the registered definition ID
	if err := mc.sdk.SetSimVar(BATTERY1_SWITCH_DEFINE_ID, newState); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle battery 1: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Battery 1 toggled successfully",
	})
}

// ToggleBattery2Handler handles toggling battery 2 in MSFS
func (mc *MonitorClient) ToggleBattery2Handler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current battery 2 switch state and toggle it
	mc.mutex.RLock()
	currentState := mc.currentData.Battery2Switch
	mc.mutex.RUnlock()

	// Toggle the state: 0 -> 1, 1 -> 0
	newState := int32(1)
	if currentState == 1 {
		newState = 0
	}

	// Set the SimVar using the registered definition ID
	if err := mc.sdk.SetSimVar(BATTERY2_SWITCH_DEFINE_ID, newState); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle battery 2: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Battery 2 toggled successfully",
	})
}

// ToggleApuMasterSwitchHandler handles toggling APU master switch in MSFS
func (mc *MonitorClient) ToggleApuMasterSwitchHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current APU master switch state and toggle it
	mc.mutex.RLock()
	currentState := mc.currentData.ApuMasterSwitch
	mc.mutex.RUnlock()

	// Toggle the state: 0 -> 1, 1 -> 0
	newState := int32(1)
	if currentState == 1 {
		newState = 0
	}

	// Set the SimVar using the registered definition ID
	if err := mc.sdk.SetSimVar(APU_MASTER_SWITCH_DEFINE_ID, newState); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle APU master switch: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "APU master switch toggled successfully",
	})
}

// ToggleApuStartButtonHandler handles toggling APU start button in MSFS
func (mc *MonitorClient) ToggleApuStartButtonHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current APU start button state and toggle it
	mc.mutex.RLock()
	currentState := mc.currentData.ApuStartButton
	mc.mutex.RUnlock()

	// Toggle the state: 0 -> 1, 1 -> 0
	newState := int32(1)
	if currentState == 1 {
		newState = 0
	}

	// Set the SimVar using the registered definition ID
	if err := mc.sdk.SetSimVar(APU_START_BUTTON_DEFINE_ID, newState); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle APU start button: %v", err), http.StatusInternalServerError)
		return
	}
	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "APU start button toggled successfully",
	})
}

// ToggleAircraftExitHandler handles toggling aircraft exit (canopy) in MSFS
func (mc *MonitorClient) ToggleAircraftExitHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Transmit the TOGGLE_AIRCRAFT_EXIT event
	if err := mc.sdk.TransmitClientEvent(
		types.SIMCONNECT_OBJECT_ID_USER,
		TOGGLE_AIRCRAFT_EXIT_EVENT_ID,
		0, // No data value needed for toggle
		AIRCRAFT_NOTIFICATION_GROUP_ID,
		types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY,
	); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle aircraft exit: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Aircraft exit toggled successfully",
	})
}

// ToggleCabinNoSmokingAlertHandler handles toggling cabin no smoking alert switch in MSFS
func (mc *MonitorClient) ToggleCabinNoSmokingAlertHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current cabin no smoking switch state and cycle through states
	mc.mutex.RLock()
	currentState := mc.currentData.CabinNoSmokingSwitch
	mc.mutex.RUnlock()

	// Cycle through states: 2 (off) -> 1 (auto) -> 0 (on) -> 2 (off)
	var newState int32
	switch currentState {
	case 2: // off -> auto
		newState = 1
	case 1: // auto -> on
		newState = 0
	case 0: // on -> off
		newState = 2
	default: // unknown state, set to off
		newState = 2
	}

	// Set the SimVar using the registered definition ID
	if err := mc.sdk.SetSimVar(CABIN_NO_SMOKING_SWITCH_DEFINE_ID, newState); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle cabin no smoking alert: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Cabin no smoking alert toggled successfully",
	})
}

// ToggleCabinSeatbeltsAlertHandler handles toggling cabin seatbelts alert switch in MSFS
func (mc *MonitorClient) ToggleCabinSeatbeltsAlertHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current cabin seatbelts switch state and cycle through states
	mc.mutex.RLock()
	currentState := mc.currentData.CabinSeatbeltsSwitch
	mc.mutex.RUnlock()

	// Cycle through states: 2 (off) -> 1 (auto) -> 0 (on) -> 2 (off)
	var newState int32
	switch currentState {
	case 2: // off -> auto
		newState = 1
	case 1: // auto -> on
		newState = 0
	case 0: // on -> off
		newState = 2
	default: // unknown state, set to off
		newState = 2
	}

	// Set the SimVar using the registered definition ID
	if err := mc.sdk.SetSimVar(CABIN_SEATBELTS_SWITCH_DEFINE_ID, newState); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle cabin seatbelts alert: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Cabin seatbelts alert toggled successfully",
	})
}

// SetCabinNoSmokingAlertHandler handles setting cabin no smoking alert to a specific state
func (mc *MonitorClient) SetCabinNoSmokingAlertHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body to get the desired state
	var requestData struct {
		State int32 `json:"state"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate state value (0=ON, 1=AUTO, 2=OFF)
	if requestData.State < 0 || requestData.State > 2 {
		http.Error(w, "Invalid state value. Must be 0 (ON), 1 (AUTO), or 2 (OFF)", http.StatusBadRequest)
		return
	}

	// Set the SimVar using the registered definition ID
	if err := mc.sdk.SetSimVar(CABIN_NO_SMOKING_SWITCH_DEFINE_ID, requestData.State); err != nil {
		http.Error(w, fmt.Sprintf("Failed to set cabin no smoking alert: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Cabin no smoking alert set successfully",
		"state":   requestData.State,
	})
}

// SetCabinSeatbeltsAlertHandler handles setting cabin seatbelts alert to a specific state
func (mc *MonitorClient) SetCabinSeatbeltsAlertHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body to get the desired state
	var requestData struct {
		State int32 `json:"state"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate state value (0=ON, 1=AUTO, 2=OFF)
	if requestData.State < 0 || requestData.State > 2 {
		http.Error(w, "Invalid state value. Must be 0 (ON), 1 (AUTO), or 2 (OFF)", http.StatusBadRequest)
		return
	}

	// Set the SimVar using the registered definition ID
	if err := mc.sdk.SetSimVar(CABIN_SEATBELTS_SWITCH_DEFINE_ID, requestData.State); err != nil {
		http.Error(w, fmt.Sprintf("Failed to set cabin seatbelts alert: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Cabin seatbelts alert set successfully",
		"state":   requestData.State,
	})
}

// RegisterAircraftEvents registers aircraft control events with SimConnect
func (mc *MonitorClient) RegisterAircraftEvents() error { // Map external power event
	if err := mc.sdk.MapClientEventToSimEvent(TOGGLE_EXTERNAL_POWER_EVENT_ID, "TOGGLE_EXTERNAL_POWER"); err != nil {
		return fmt.Errorf("failed to map TOGGLE_EXTERNAL_POWER event: %v", err)
	}
	// Map aircraft exit (canopy) event
	if err := mc.sdk.MapClientEventToSimEvent(TOGGLE_AIRCRAFT_EXIT_EVENT_ID, "TOGGLE_AIRCRAFT_EXIT"); err != nil {
		return fmt.Errorf("failed to map TOGGLE_AIRCRAFT_EXIT event: %v", err)
	}

	// Add external power event to the notification group
	if err := mc.sdk.AddClientEventToNotificationGroup(AIRCRAFT_NOTIFICATION_GROUP_ID, TOGGLE_EXTERNAL_POWER_EVENT_ID, false); err != nil {
		return fmt.Errorf("failed to add event %d to notification group: %v", TOGGLE_EXTERNAL_POWER_EVENT_ID, err)
	}
	// Add aircraft exit event to the notification group
	if err := mc.sdk.AddClientEventToNotificationGroup(AIRCRAFT_NOTIFICATION_GROUP_ID, TOGGLE_AIRCRAFT_EXIT_EVENT_ID, false); err != nil {
		return fmt.Errorf("failed to add event %d to notification group: %v", TOGGLE_AIRCRAFT_EXIT_EVENT_ID, err)
	}

	// Set notification group priority after adding events
	if err := mc.sdk.SetNotificationGroupPriority(AIRCRAFT_NOTIFICATION_GROUP_ID, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST); err != nil {
		return fmt.Errorf("failed to set aircraft notification group priority: %v", err)
	}

	return nil
}
