package client

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/mycrew-online/sdk/pkg/types"
)

// RegisterSimVarDefinition registers a single simulation variable to a data definition with specified data type
// This enhanced version tracks the data type for proper parsing later
func (e *Engine) RegisterSimVarDefinition(defID uint32, varName string, units string, dataType types.SimConnectDataType) error {
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

// RequestSimVarDataPeriodic requests data for a previously registered sim variable with a specified frequency
// This allows for continuous data updates at the specified period
func (e *Engine) RequestSimVarDataPeriodic(defID uint32, requestID uint32, period types.SimConnectPeriod) error {
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

	// Call SimConnect_RequestDataOnSimObject with the specified period
	hresult, _, _ := SimConnect_RequestDataOnSimObject.Call(
		uintptr(handle),                          // hSimConnect
		uintptr(requestID),                       // RequestID
		uintptr(defID),                           // DefineID
		uintptr(types.SIMCONNECT_OBJECT_ID_USER), // ObjectID (user aircraft)
		uintptr(period),                          // Period (periodic request)
		uintptr(types.SIMCONNECT_DATA_REQUEST_FLAG_DEFAULT), // Flags
		0, // origin
		0, // interval
		0, // limit
	)

	if !IsHRESULTSuccess(uint32(hresult)) {
		return fmt.Errorf("SimConnect_RequestDataOnSimObject periodic failed: 0x%08X", uint32(hresult))
	}
	return nil
}

// StopPeriodicRequest stops a periodic data request by requesting it with SIMCONNECT_PERIOD_NEVER
func (e *Engine) StopPeriodicRequest(requestID uint32) error {
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

	// Call SimConnect_RequestDataOnSimObject with NEVER period to stop updates
	hresult, _, _ := SimConnect_RequestDataOnSimObject.Call(
		uintptr(handle),                          // hSimConnect
		uintptr(requestID),                       // RequestID
		0,                                        // DefineID (can be 0 when stopping)
		uintptr(types.SIMCONNECT_OBJECT_ID_USER), // ObjectID (user aircraft)
		uintptr(types.SIMCONNECT_PERIOD_NEVER),   // Period (NEVER to stop)
		uintptr(types.SIMCONNECT_DATA_REQUEST_FLAG_DEFAULT), // Flags
		0, // origin
		0, // interval
		0, // limit
	)

	if !IsHRESULTSuccess(uint32(hresult)) {
		return fmt.Errorf("SimConnect_RequestDataOnSimObject stop failed: 0x%08X", uint32(hresult))
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
		return fmt.Errorf("defID %d not found in data type registry - call RegisterSimVarDefinition first", defID)
	}
	// Convert the value to the proper binary format based on data type
	var dataPtr unsafe.Pointer
	var dataSize uint32

	switch dataType {
	case types.SIMCONNECT_DATATYPE_INVALID:
		return fmt.Errorf("cannot set data with INVALID data type for defID %d", defID)

	case types.SIMCONNECT_DATATYPE_INT32:
		var int32Value int32
		switch v := value.(type) {
		case int32:
			int32Value = v
		case int:
			int32Value = int32(v)
		case int64:
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

	case types.SIMCONNECT_DATATYPE_INT64:
		var int64Value int64
		switch v := value.(type) {
		case int64:
			int64Value = v
		case int:
			int64Value = int64(v)
		case int32:
			int64Value = int64(v)
		case float64:
			int64Value = int64(v)
		case float32:
			int64Value = int64(v)
		default:
			return fmt.Errorf("cannot convert %T to int64 for defID %d", value, defID)
		}
		dataPtr = unsafe.Pointer(&int64Value)
		dataSize = 8

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
		case int64:
			float32Value = float32(v)
		default:
			return fmt.Errorf("cannot convert %T to float32 for defID %d", value, defID)
		}
		dataPtr = unsafe.Pointer(&float32Value)
		dataSize = 4

	case types.SIMCONNECT_DATATYPE_FLOAT64:
		var float64Value float64
		switch v := value.(type) {
		case float64:
			float64Value = v
		case float32:
			float64Value = float64(v)
		case int32:
			float64Value = float64(v)
		case int:
			float64Value = float64(v)
		case int64:
			float64Value = float64(v)
		default:
			return fmt.Errorf("cannot convert %T to float64 for defID %d", value, defID)
		}
		dataPtr = unsafe.Pointer(&float64Value)
		dataSize = 8

	case types.SIMCONNECT_DATATYPE_STRINGV:
		var stringValue string
		switch v := value.(type) {
		case string:
			stringValue = v
		default:
			return fmt.Errorf("cannot convert %T to string for defID %d", value, defID)
		}
		// For variable strings, include null terminator
		stringBytes := []byte(stringValue + "\x00")
		dataPtr = unsafe.Pointer(&stringBytes[0])
		dataSize = uint32(len(stringBytes))

	case types.SIMCONNECT_DATATYPE_STRING8:
		stringBytes, err := e.prepareFixedString(value, 8, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(&stringBytes[0])
		dataSize = 8

	case types.SIMCONNECT_DATATYPE_STRING32:
		stringBytes, err := e.prepareFixedString(value, 32, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(&stringBytes[0])
		dataSize = 32

	case types.SIMCONNECT_DATATYPE_STRING64:
		stringBytes, err := e.prepareFixedString(value, 64, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(&stringBytes[0])
		dataSize = 64

	case types.SIMCONNECT_DATATYPE_STRING128:
		stringBytes, err := e.prepareFixedString(value, 128, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(&stringBytes[0])
		dataSize = 128

	case types.SIMCONNECT_DATATYPE_STRING256:
		stringBytes, err := e.prepareFixedString(value, 256, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(&stringBytes[0])
		dataSize = 256

	case types.SIMCONNECT_DATATYPE_STRING260:
		stringBytes, err := e.prepareFixedString(value, 260, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(&stringBytes[0])
		dataSize = 260

	case types.SIMCONNECT_DATATYPE_INITPOSITION:
		initPos, err := e.prepareInitPosition(value, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(initPos)
		dataSize = uint32(unsafe.Sizeof(types.InitPosition{}))

	case types.SIMCONNECT_DATATYPE_MARKERSTATE:
		markerState, err := e.prepareMarkerState(value, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(markerState)
		dataSize = uint32(unsafe.Sizeof(types.MarkerState{}))

	case types.SIMCONNECT_DATATYPE_WAYPOINT:
		waypoint, err := e.prepareWaypoint(value, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(waypoint)
		dataSize = uint32(unsafe.Sizeof(types.Waypoint{}))

	case types.SIMCONNECT_DATATYPE_LATLONALT:
		latLonAlt, err := e.prepareLatLonAlt(value, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(latLonAlt)
		dataSize = uint32(unsafe.Sizeof(types.LatLonAlt{}))

	case types.SIMCONNECT_DATATYPE_XYZ:
		xyz, err := e.prepareXYZ(value, defID)
		if err != nil {
			return err
		}
		dataPtr = unsafe.Pointer(xyz)
		dataSize = uint32(unsafe.Sizeof(types.XYZ{}))

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

// Helper functions for preparing complex data types for SetSimVar

// prepareFixedString prepares a fixed-length string for setting
func (e *Engine) prepareFixedString(value interface{}, maxLength int, defID uint32) ([]byte, error) {
	var stringValue string
	switch v := value.(type) {
	case string:
		stringValue = v
	default:
		return nil, fmt.Errorf("cannot convert %T to string for defID %d", value, defID)
	}

	// Truncate if too long
	if len(stringValue) >= maxLength {
		stringValue = stringValue[:maxLength-1]
	}

	// Create fixed-length byte array with null terminator
	stringBytes := make([]byte, maxLength)
	copy(stringBytes, []byte(stringValue))
	// Ensure null termination (array is zero-initialized)
	return stringBytes, nil
}

// prepareInitPosition prepares an InitPosition structure for setting
func (e *Engine) prepareInitPosition(value interface{}, defID uint32) (*types.InitPosition, error) {
	switch v := value.(type) {
	case types.InitPosition:
		return &v, nil
	case *types.InitPosition:
		return v, nil
	case map[string]interface{}:
		// Allow setting from a map/JSON-like structure
		initPos := &types.InitPosition{}
		if lat, ok := v["latitude"]; ok {
			if latFloat, ok := lat.(float64); ok {
				initPos.Latitude = latFloat
			}
		}
		if lon, ok := v["longitude"]; ok {
			if lonFloat, ok := lon.(float64); ok {
				initPos.Longitude = lonFloat
			}
		}
		if alt, ok := v["altitude"]; ok {
			if altFloat, ok := alt.(float64); ok {
				initPos.Altitude = altFloat
			}
		}
		if pitch, ok := v["pitch"]; ok {
			if pitchFloat, ok := pitch.(float64); ok {
				initPos.Pitch = pitchFloat
			}
		}
		if bank, ok := v["bank"]; ok {
			if bankFloat, ok := bank.(float64); ok {
				initPos.Bank = bankFloat
			}
		}
		if heading, ok := v["heading"]; ok {
			if headingFloat, ok := heading.(float64); ok {
				initPos.Heading = headingFloat
			}
		}
		if onGround, ok := v["onGround"]; ok {
			if onGroundBool, ok := onGround.(bool); ok {
				if onGroundBool {
					initPos.OnGround = 1
				} else {
					initPos.OnGround = 0
				}
			}
		}
		if airspeed, ok := v["airspeed"]; ok {
			if airspeedFloat, ok := airspeed.(float64); ok {
				initPos.Airspeed = uint32(airspeedFloat)
			}
		}
		return initPos, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to InitPosition for defID %d", value, defID)
	}
}

// prepareMarkerState prepares a MarkerState structure for setting
func (e *Engine) prepareMarkerState(value interface{}, defID uint32) (*types.MarkerState, error) {
	switch v := value.(type) {
	case types.MarkerState:
		return &v, nil
	case *types.MarkerState:
		return v, nil
	case map[string]interface{}:
		// Allow setting from a map/JSON-like structure
		markerState := &types.MarkerState{}
		if name, ok := v["name"]; ok {
			if nameStr, ok := name.(string); ok {
				if len(nameStr) >= 64 {
					nameStr = nameStr[:63]
				}
				copy(markerState.Name[:], []byte(nameStr))
			}
		}
		if lat, ok := v["latitude"]; ok {
			if latFloat, ok := lat.(float64); ok {
				markerState.Latitude = latFloat
			}
		}
		if lon, ok := v["longitude"]; ok {
			if lonFloat, ok := lon.(float64); ok {
				markerState.Longitude = lonFloat
			}
		}
		if alt, ok := v["altitude"]; ok {
			if altFloat, ok := alt.(float64); ok {
				markerState.Altitude = altFloat
			}
		}
		if flags, ok := v["flags"]; ok {
			if flagsInt, ok := flags.(int); ok {
				markerState.Flags = uint32(flagsInt)
			} else if flagsFloat, ok := flags.(float64); ok {
				markerState.Flags = uint32(flagsFloat)
			}
		}
		if heading, ok := v["heading"]; ok {
			if headingFloat, ok := heading.(float64); ok {
				markerState.Heading = headingFloat
			}
		}
		if speed, ok := v["speed"]; ok {
			if speedFloat, ok := speed.(float64); ok {
				markerState.Speed = speedFloat
			}
		}
		if bank, ok := v["bank"]; ok {
			if bankFloat, ok := bank.(float64); ok {
				markerState.Bank = bankFloat
			}
		}
		if pitch, ok := v["pitch"]; ok {
			if pitchFloat, ok := pitch.(float64); ok {
				markerState.Pitch = pitchFloat
			}
		}
		return markerState, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to MarkerState for defID %d", value, defID)
	}
}

// prepareWaypoint prepares a Waypoint structure for setting
func (e *Engine) prepareWaypoint(value interface{}, defID uint32) (*types.Waypoint, error) {
	switch v := value.(type) {
	case types.Waypoint:
		return &v, nil
	case *types.Waypoint:
		return v, nil
	case map[string]interface{}:
		// Allow setting from a map/JSON-like structure
		waypoint := &types.Waypoint{}
		if lat, ok := v["latitude"]; ok {
			if latFloat, ok := lat.(float64); ok {
				waypoint.Latitude = latFloat
			}
		}
		if lon, ok := v["longitude"]; ok {
			if lonFloat, ok := lon.(float64); ok {
				waypoint.Longitude = lonFloat
			}
		}
		if alt, ok := v["altitude"]; ok {
			if altFloat, ok := alt.(float64); ok {
				waypoint.Altitude = altFloat
			}
		}
		if flags, ok := v["flags"]; ok {
			if flagsInt, ok := flags.(int); ok {
				waypoint.Flags = uint32(flagsInt)
			} else if flagsFloat, ok := flags.(float64); ok {
				waypoint.Flags = uint32(flagsFloat)
			}
		}
		if speed, ok := v["speed"]; ok {
			if speedFloat, ok := speed.(float64); ok {
				waypoint.Speed = speedFloat
			}
		}
		if throttle, ok := v["throttle"]; ok {
			if throttleFloat, ok := throttle.(float64); ok {
				waypoint.Throttle = throttleFloat
			}
		}
		return waypoint, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to Waypoint for defID %d", value, defID)
	}
}

// prepareLatLonAlt prepares a LatLonAlt structure for setting
func (e *Engine) prepareLatLonAlt(value interface{}, defID uint32) (*types.LatLonAlt, error) {
	switch v := value.(type) {
	case types.LatLonAlt:
		return &v, nil
	case *types.LatLonAlt:
		return v, nil
	case map[string]interface{}:
		// Allow setting from a map/JSON-like structure
		latLonAlt := &types.LatLonAlt{}
		if lat, ok := v["latitude"]; ok {
			if latFloat, ok := lat.(float64); ok {
				latLonAlt.Latitude = latFloat
			}
		}
		if lon, ok := v["longitude"]; ok {
			if lonFloat, ok := lon.(float64); ok {
				latLonAlt.Longitude = lonFloat
			}
		}
		if alt, ok := v["altitude"]; ok {
			if altFloat, ok := alt.(float64); ok {
				latLonAlt.Altitude = altFloat
			}
		}
		return latLonAlt, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to LatLonAlt for defID %d", value, defID)
	}
}

// prepareXYZ prepares an XYZ structure for setting
func (e *Engine) prepareXYZ(value interface{}, defID uint32) (*types.XYZ, error) {
	switch v := value.(type) {
	case types.XYZ:
		return &v, nil
	case *types.XYZ:
		return v, nil
	case map[string]interface{}:
		// Allow setting from a map/JSON-like structure
		xyz := &types.XYZ{}
		if x, ok := v["x"]; ok {
			if xFloat, ok := x.(float64); ok {
				xyz.X = xFloat
			}
		}
		if y, ok := v["y"]; ok {
			if yFloat, ok := y.(float64); ok {
				xyz.Y = yFloat
			}
		}
		if z, ok := v["z"]; ok {
			if zFloat, ok := z.(float64); ok {
				xyz.Z = zFloat
			}
		}
		return xyz, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to XYZ for defID %d", value, defID)
	}
}
