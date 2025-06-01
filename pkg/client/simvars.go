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

// SetSimVar sets data on a simulation object for a previously registered sim variable
// Baby Step 3A: Generic method that uses the data type registry for proper type conversion
func (e *Engine) SetSimVar(defID uint32, value interface{}) error {
	// Thread-safe check for connection
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to simulator")
	}

	// Look up the expected data type for this DefineID (thread-safe)
	e.mu.RLock()
	dataType, exists := e.dataTypeRegistry[defID]
	handle := e.handle
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("defID %d not found in data type registry - call AddSimVar first", defID)
	}

	// Convert the value to the proper binary format based on data type
	var dataPtr unsafe.Pointer
	var dataSize uint32

	switch dataType {
	case types.SIMCONNECT_DATATYPE_INT32:
		var int32Value int32
		switch v := value.(type) {
		case int32:
			int32Value = v
		case int:
			int32Value = int32(v)
		case float64:
			int32Value = int32(v)
		case float32:
			int32Value = int32(v)
		default:
			return fmt.Errorf("cannot convert %T to int32 for defID %d", value, defID)
		}
		dataPtr = unsafe.Pointer(&int32Value)
		dataSize = 4

	case types.SIMCONNECT_DATATYPE_FLOAT32:
		var float32Value float32
		switch v := value.(type) {
		case float32:
			float32Value = v
		case float64:
			float32Value = float32(v)
		case int32:
			float32Value = float32(v)
		case int:
			float32Value = float32(v)
		default:
			return fmt.Errorf("cannot convert %T to float32 for defID %d", value, defID)
		}
		dataPtr = unsafe.Pointer(&float32Value)
		dataSize = 4

	case types.SIMCONNECT_DATATYPE_STRINGV:
		var stringValue string
		switch v := value.(type) {
		case string:
			stringValue = v
		default:
			return fmt.Errorf("cannot convert %T to string for defID %d", value, defID)
		}
		// For strings, we need to include null terminator
		stringBytes := []byte(stringValue + "\x00")
		dataPtr = unsafe.Pointer(&stringBytes[0])
		dataSize = uint32(len(stringBytes))

	default:
		return fmt.Errorf("unsupported data type %d for defID %d", dataType, defID)
	}

	// Call SimConnect_SetDataOnSimObject
	hresult, _, _ := SimConnect_SetDataOnSimObject.Call(
		uintptr(handle),                                 // hSimConnect
		uintptr(defID),                                  // DefineID
		uintptr(types.SIMCONNECT_OBJECT_ID_USER),        // ObjectID (user aircraft)
		uintptr(types.SIMCONNECT_DATA_SET_FLAG_DEFAULT), // Flags
		0,                 // ArrayCount (0 for single values)
		uintptr(dataSize), // cbUnitSize
		uintptr(dataPtr),  // pDataSet
	)

	if !IsHRESULTSuccess(uint32(hresult)) {
		return fmt.Errorf("SimConnect_SetDataOnSimObject failed: 0x%08X", uint32(hresult))
	}

	return nil
}
