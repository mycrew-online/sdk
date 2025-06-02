package client

import (
	"fmt"
	"syscall"
	"unsafe"
)

// SubscribeToSystemEvent subscribes to a system event from SimConnect
// This is Baby Step 1: Basic system event subscription (read-only)
// eventID: Unique ID to identify this event subscription
// eventName: Name of the system event (e.g., "Paused", "Sim", "AircraftLoaded")
func (e *Engine) SubscribeToSystemEvent(eventID uint32, eventName string) error {
	// Thread-safe check for connection
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to simulator")
	}

	// Convert event name to C-style string
	eventNamePtr, err := syscall.BytePtrFromString(eventName)
	if err != nil {
		return fmt.Errorf("invalid event name: %v", err)
	}

	// Thread-safe access to handle
	e.mu.RLock()
	handle := e.handle
	e.mu.RUnlock()

	// Call SimConnect_SubscribeToSystemEvent
	// HRESULT SimConnect_SubscribeToSystemEvent(HANDLE hSimConnect, SIMCONNECT_CLIENT_EVENT_ID EventID, LPCSTR SystemEventName)
	hresult, _, _ := SimConnect_SubscribeToSystemEvent.Call(
		uintptr(handle),                       // hSimConnect
		uintptr(eventID),                      // EventID
		uintptr(unsafe.Pointer(eventNamePtr)), // SystemEventName
	)

	if !IsHRESULTSuccess(uint32(hresult)) {
		return fmt.Errorf("SimConnect_SubscribeToSystemEvent failed: 0x%08X", uint32(hresult))
	}

	return nil
}
