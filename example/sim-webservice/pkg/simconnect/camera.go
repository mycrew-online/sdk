package simconnect

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mycrew-online/sdk/pkg/types"
)

// CameraStateHandler handles setting the camera state in MSFS
func (wc *WeatherClient) SetCameraState(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var requestBody struct {
		State int32 `json:"state"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Define valid camera states
	validStates := []int32{2, 3, 4, 5, 6, 7, 8, 9, 10}
	isValid := false

	for _, state := range validStates {
		if requestBody.State == state {
			isValid = true
			break
		}
	}

	if !isValid {
		http.Error(w, "Invalid camera state", http.StatusBadRequest)
		return
	}

	// Set the camera state in SimConnect
	if err := wc.sdk.SetSimVar(CAMERA_STATE_DEFINE_ID, requestBody.State); err != nil {
		http.Error(w, fmt.Sprintf("Failed to set camera state: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Camera state set to %d", requestBody.State),
	})
}

// RegisterCameraState registers the CAMERA_STATE variable with SimConnect
func (wc *WeatherClient) RegisterCameraState() error {
	// Register Camera State
	if err := wc.sdk.RegisterSimVarDefinition(
		CAMERA_STATE_DEFINE_ID,
		"CAMERA STATE",
		"Enum",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register CAMERA STATE: %v", err)
	}

	// Request periodic updates for camera state
	if err := wc.sdk.RequestSimVarDataPeriodic(
		CAMERA_STATE_DEFINE_ID,
		CAMERA_STATE_REQUEST_ID,
		types.SIMCONNECT_PERIOD_SECOND,
	); err != nil {
		return fmt.Errorf("failed to start camera state monitoring: %v", err)
	}

	return nil
}
