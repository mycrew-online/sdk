package client

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/mycrew-online/sdk/pkg/types"
)

// AddSimVar adds a single simulation variable to a data definition with specified data type
// This enhanced version tracks the data type for proper parsing later
func (e *Engine) AddSimVar(defID uint32, varName string, units string, dataType types.SimConnectDataType) error {
	// Thread-safe check for connection
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to simulator")
	}

	// Convert strings to C-style for SimConnect
	varNamePtr, err := syscall.BytePtrFromString(varName)
	if err != nil {
		return fmt.Errorf("invalid variable name: %v", err)
	}

	unitsPtr, err := syscall.BytePtrFromString(units)
	if err != nil {
		return fmt.Errorf("invalid units: %v", err)
	}

	// Thread-safe access to handle
	e.mu.RLock()
	handle := e.handle
	e.mu.RUnlock()

	// Call SimConnect_AddToDataDefinition with the specified data type
	hresult, _, _ := SimConnect_AddToDataDefinition.Call(
		uintptr(handle),                     // hSimConnect
		uintptr(defID),                      // DefineID
		uintptr(unsafe.Pointer(varNamePtr)), // DatumName
		uintptr(unsafe.Pointer(unitsPtr)),   // UnitsName
		uintptr(dataType),                   // DatumType (now configurable)
		0,                                   // fEpsilon
		0,                                   // DatumID
	)

	if !IsHRESULTSuccess(uint32(hresult)) {
		return fmt.Errorf("SimConnect_AddToDataDefinition failed: 0x%08X", uint32(hresult))
	}

	// Store the data type mapping for later parsing (thread-safe)
	e.mu.Lock()
	e.dataTypeRegistry[defID] = dataType
	e.mu.Unlock()

	return nil
}

// RequestSimVarData requests data for a previously registered sim variable
// This is the next baby step - actually get the data
func (e *Engine) RequestSimVarData(defID uint32, requestID uint32) error {
	// Thread-safe check for connection
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to simulator")
	}

	// Thread-safe access to handle
	e.mu.RLock()
	handle := e.handle
	e.mu.RUnlock()
	// Call SimConnect_RequestDataOnSimObject
	hresult, _, _ := SimConnect_RequestDataOnSimObject.Call(
		uintptr(handle),                                     // hSimConnect
		uintptr(requestID),                                  // RequestID
		uintptr(defID),                                      // DefineID
		uintptr(types.SIMCONNECT_OBJECT_ID_USER),            // ObjectID (user aircraft)
		uintptr(types.SIMCONNECT_PERIOD_ONCE),               // Period (one-time request)
		uintptr(types.SIMCONNECT_DATA_REQUEST_FLAG_DEFAULT), // Flags
		0, // origin
		0, // interval
		0, // limit
	)

	if !IsHRESULTSuccess(uint32(hresult)) {
		return fmt.Errorf("SimConnect_RequestDataOnSimObject failed: 0x%08X", uint32(hresult))
	}
	return nil
}
