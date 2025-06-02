package client

import (
	"unsafe"

	"github.com/mycrew-online/sdk/pkg/types"
)

// SimVarData represents data received for a sim variable
type SimVarData struct {
	RequestID uint32
	DefineID  uint32
	Value     interface{} // Support multiple data types: float64, int32, string, etc.
}

// parseSimObjectData extracts sim variable data from SIMOBJECT_DATA message
// Now type-aware - looks up the expected data type for proper parsing
func parseSimObjectData(ppData uintptr, pcbData uint32, engine *Engine) *SimVarData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_SIMOBJECT_DATA structure
	simObjData := (*types.SIMCONNECT_RECV_SIMOBJECT_DATA)(unsafe.Pointer(ppData))
	if simObjData.DwID != types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA {
		return nil
	}

	// Look up the expected data type for this DefineID (thread-safe)
	engine.mu.RLock()
	dataType, exists := engine.dataTypeRegistry[simObjData.DwDefineID]
	engine.mu.RUnlock()
	if !exists {
		// Fallback to FLOAT32 if not found
		dataType = types.SIMCONNECT_DATATYPE_FLOAT32
	}

	var value interface{}
	// Parse based on the registered data type
	switch dataType {
	case types.SIMCONNECT_DATATYPE_FLOAT32:
		// For FLOAT32: 4-byte value stored in DwData field
		float32Value := *(*float32)(unsafe.Pointer(&simObjData.DwData))
		value = float64(float32Value)
	case types.SIMCONNECT_DATATYPE_INT32:
		// For INT32: 4-byte integer stored in DwData field
		int32Value := *(*int32)(unsafe.Pointer(&simObjData.DwData))
		value = int32Value // Store as actual int32, not converted to float64
	case types.SIMCONNECT_DATATYPE_STRINGV:
		// For STRINGV: Variable-length string data comes after the header
		// The string starts immediately after the SIMCONNECT_RECV_SIMOBJECT_DATA structure
		headerSize := unsafe.Sizeof(*simObjData)
		if pcbData > uint32(headerSize) {
			// Calculate string data location and available bytes
			stringDataPtr := ppData + uintptr(headerSize)
			stringDataLen := pcbData - uint32(headerSize)

			// Read the null-terminated string
			stringBytes := make([]byte, stringDataLen)
			for i := uint32(0); i < stringDataLen; i++ {
				b := *(*byte)(unsafe.Pointer(stringDataPtr + uintptr(i)))
				if b == 0 {
					// Found null terminator
					stringBytes = stringBytes[:i]
					break
				}
				stringBytes[i] = b
			}
			value = string(stringBytes)
		} else {
			value = "" // Empty string if no data
		}
	default:
		// Fallback to FLOAT32 for unknown types
		float32Value := *(*float32)(unsafe.Pointer(&simObjData.DwData))
		value = float64(float32Value)
	}

	return &SimVarData{
		RequestID: simObjData.DwRequestID,
		DefineID:  simObjData.DwDefineID,
		Value:     value,
	}
}

// parseSimConnectData processes incoming SimConnect messages for debugging
func parseSimConnectData(ppData uintptr, pcbData uint32, engine *Engine) {
	if ppData == 0 || pcbData == 0 {
		// fmt.Println("No data received")
		return
	}

	// Cast the pointer to the base SIMCONNECT_RECV structure
	recv := (*types.SIMCONNECT_RECV)(unsafe.Pointer(ppData))
	// fmt.Printf("Received message - Size: %d, Version: %d, ID: %d\n",
	//	recv.DwSize, recv.DwVersion, recv.DwID)	// Check what type of message we received based on the ID
	switch recv.DwID {
	case types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA:
		// fmt.Println("📊 SIMOBJECT_DATA received")
		if data := parseSimObjectData(ppData, pcbData, engine); data != nil {
			// Look up data type for proper formatting
			engine.mu.RLock()
			dataType, exists := engine.dataTypeRegistry[data.DefineID]
			engine.mu.RUnlock()

			if !exists {
				dataType = types.SIMCONNECT_DATATYPE_FLOAT32
			}
			// Format value based on data type and actual value type
			// All logging has been commented out to prevent stdout interference
			_ = dataType // Suppress unused variable warning
			_ = data     // Suppress unused variable warning
		}
	case types.SIMCONNECT_RECV_ID_OPEN:
		// fmt.Println("🔓 OPEN confirmation received")

	case types.SIMCONNECT_RECV_ID_EXCEPTION:
		// fmt.Println("❌ EXCEPTION received")
		// Parse the exception details with enhanced error reporting
		if ppData != 0 && pcbData >= uint32(unsafe.Sizeof(types.SIMCONNECT_RECV_EXCEPTION{})) {
			exceptionData := (*types.SIMCONNECT_RECV_EXCEPTION)(unsafe.Pointer(ppData))
			// fmt.Printf("   🔍 Exception Code: %d, SendID: %d, Index: %d\n",
			//	exceptionData.DwException, exceptionData.DwSendID, exceptionData.DwIndex)

			// Provide detailed exception descriptions based on the fetched documentation
			switch types.SimConnectException(exceptionData.DwException) {
			case types.SIMCONNECT_EXCEPTION_NONE:
				// fmt.Println("   📋 NONE: No error occurred")
			case types.SIMCONNECT_EXCEPTION_ERROR:
				// fmt.Println("   📋 ERROR: An unspecific error has occurred")
			case types.SIMCONNECT_EXCEPTION_SIZE_MISMATCH:
				// fmt.Println("   📋 SIZE_MISMATCH: The size of the data provided does not match the size required")
			case types.SIMCONNECT_EXCEPTION_UNRECOGNIZED_ID:
				// fmt.Println("   📋 UNRECOGNIZED_ID: The client event, request ID, data definition ID, or object ID was not recognized")
			case types.SIMCONNECT_EXCEPTION_UNOPENED:
				// fmt.Println("   📋 UNOPENED: Communication with the SimConnect server has not been opened")
			case types.SIMCONNECT_EXCEPTION_VERSION_MISMATCH:
				// fmt.Println("   📋 VERSION_MISMATCH: A versioning error has occurred")
			case types.SIMCONNECT_EXCEPTION_TOO_MANY_GROUPS:
				// fmt.Println("   📋 TOO_MANY_GROUPS: The maximum number of groups allowed has been reached (max: 20)")
			case types.SIMCONNECT_EXCEPTION_NAME_UNRECOGNIZED:
				// fmt.Println("   📋 NAME_UNRECOGNIZED: The simulation event name is not recognized")
			case types.SIMCONNECT_EXCEPTION_TOO_MANY_EVENT_NAMES:
				// fmt.Println("   📋 TOO_MANY_EVENT_NAMES: The maximum number of event names allowed has been reached (max: 1000)")
			case types.SIMCONNECT_EXCEPTION_EVENT_ID_DUPLICATE:
				// fmt.Println("   📋 EVENT_ID_DUPLICATE: The event ID has been used already")
			case types.SIMCONNECT_EXCEPTION_TOO_MANY_MAPS:
				// fmt.Println("   📋 TOO_MANY_MAPS: The maximum number of mappings allowed has been reached (max: 20)")
			case types.SIMCONNECT_EXCEPTION_TOO_MANY_OBJECTS:
				// fmt.Println("   📋 TOO_MANY_OBJECTS: The maximum number of objects allowed has been reached (max: 1000)")
			case types.SIMCONNECT_EXCEPTION_TOO_MANY_REQUESTS:
				// fmt.Println("   📋 TOO_MANY_REQUESTS: The maximum number of requests allowed has been reached (max: 1000)")
			case types.SIMCONNECT_EXCEPTION_INVALID_DATA_TYPE:
				// fmt.Println("   📋 INVALID_DATA_TYPE: The data type requested does not apply to the type of data requested")
			case types.SIMCONNECT_EXCEPTION_INVALID_DATA_SIZE:
				// fmt.Println("   📋 INVALID_DATA_SIZE: The size of the data provided is not what is expected")
			case types.SIMCONNECT_EXCEPTION_DATA_ERROR:
				// fmt.Println("   📋 DATA_ERROR: A generic data error occurred")
			case types.SIMCONNECT_EXCEPTION_INVALID_ARRAY:
				// fmt.Println("   📋 INVALID_ARRAY: An invalid array has been sent")
			case types.SIMCONNECT_EXCEPTION_CREATE_OBJECT_FAILED:
				// fmt.Println("   📋 CREATE_OBJECT_FAILED: The attempt to create an AI object failed")
			case types.SIMCONNECT_EXCEPTION_LOAD_FLIGHTPLAN_FAILED:
				// fmt.Println("   📋 LOAD_FLIGHTPLAN_FAILED: The specified flight plan could not be found or loaded")
			case types.SIMCONNECT_EXCEPTION_OPERATION_INVALID_FOR_OBJECT_TYPE:
				// fmt.Println("   📋 OPERATION_INVALID_FOR_OBJECT_TYPE: The operation requested does not apply to the object type")
			case types.SIMCONNECT_EXCEPTION_ILLEGAL_OPERATION:
				// fmt.Println("   📋 ILLEGAL_OPERATION: The operation requested cannot be completed")
			case types.SIMCONNECT_EXCEPTION_ALREADY_SUBSCRIBED:
				// fmt.Println("   📋 ALREADY_SUBSCRIBED: The client has already subscribed to that event")
			case types.SIMCONNECT_EXCEPTION_INVALID_ENUM:
				// fmt.Println("   📋 INVALID_ENUM: The member of the enumeration provided was not valid")
			case types.SIMCONNECT_EXCEPTION_DEFINITION_ERROR:
				// fmt.Println("   📋 DEFINITION_ERROR: There is a problem with a data definition")
			case types.SIMCONNECT_EXCEPTION_DUPLICATE_ID:
				// fmt.Println("   📋 DUPLICATE_ID: The ID has already been used")
			case types.SIMCONNECT_EXCEPTION_DATUM_ID:
				// fmt.Println("   📋 DATUM_ID: The datum ID is not recognized")
			case types.SIMCONNECT_EXCEPTION_OUT_OF_BOUNDS:
				// fmt.Println("   📋 OUT_OF_BOUNDS: The radius given was outside the acceptable range")
			case types.SIMCONNECT_EXCEPTION_ALREADY_CREATED:
				// fmt.Println("   📋 ALREADY_CREATED: A client data area with the requested name has already been created")
			case types.SIMCONNECT_EXCEPTION_OBJECT_OUTSIDE_REALITY_BUBBLE:
				// fmt.Println("   📋 OBJECT_OUTSIDE_REALITY_BUBBLE: The object location is outside the reality bubble")
			case types.SIMCONNECT_EXCEPTION_OBJECT_CONTAINER:
				// fmt.Println("   📋 OBJECT_CONTAINER: Error with the container system for the object")
			case types.SIMCONNECT_EXCEPTION_OBJECT_AI:
				// fmt.Println("   📋 OBJECT_AI: Error with the AI system for the object")
			case types.SIMCONNECT_EXCEPTION_OBJECT_ATC:
				// fmt.Println("   📋 OBJECT_ATC: Error with the ATC system for the object")
			case types.SIMCONNECT_EXCEPTION_OBJECT_SCHEDULE:
				// fmt.Println("   📋 OBJECT_SCHEDULE: Error with object scheduling")
			case types.SIMCONNECT_EXCEPTION_JETWAY_DATA:
				// fmt.Println("   📋 JETWAY_DATA: Error retrieving jetway data")
			case types.SIMCONNECT_EXCEPTION_ACTION_NOT_FOUND:
				// fmt.Println("   📋 ACTION_NOT_FOUND: The given action cannot be found")
			case types.SIMCONNECT_EXCEPTION_NOT_AN_ACTION:
				// fmt.Println("   📋 NOT_AN_ACTION: The given action does not exist")
			case types.SIMCONNECT_EXCEPTION_INCORRECT_ACTION_PARAMS:
				// fmt.Println("   📋 INCORRECT_ACTION_PARAMS: Wrong parameters have been given to the action")
			case types.SIMCONNECT_EXCEPTION_GET_INPUT_EVENT_FAILED:
				// fmt.Println("   📋 GET_INPUT_EVENT_FAILED: Wrong name/hash passed to GetInputEvent")
			case types.SIMCONNECT_EXCEPTION_SET_INPUT_EVENT_FAILED:
				// fmt.Println("   📋 SET_INPUT_EVENT_FAILED: Wrong name/hash passed to SetInputEvent")
			default:
				// fmt.Printf("   📋 Unknown exception type: %d\n", exceptionData.DwException)
			}
		}
	case types.SIMCONNECT_RECV_ID_SYSTEM_STATE:
		// fmt.Println("🔧 SYSTEM_STATE received")
	case types.SIMCONNECT_RECV_ID_EVENT:
		// fmt.Println("📡 EVENT received")
		if eventData := parseEventData(ppData, pcbData, engine); eventData != nil {
			// Log event reception for debugging
			// fmt.Printf("   🎯 Event ID: %d, Group: %d, Data: %d, Type: %s\n",
			//	eventData.EventID, eventData.GroupID, eventData.EventData, eventData.EventType)
			_ = eventData // Suppress unused variable warning
		}

	case types.SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENTS:
		// fmt.Println("🎮 ENUMERATE_INPUT_EVENTS received")

	case types.SIMCONNECT_RECV_ID_SUBSCRIBE_INPUT_EVENT:
		// fmt.Println("🔗 SUBSCRIBE_INPUT_EVENT received")

	case types.SIMCONNECT_RECV_ID_QUIT:
		// fmt.Println("👋 QUIT received")

	default:
		// fmt.Printf("❓ Unknown message type: %d\n", recv.DwID)
	}
}

// parseSimConnectToChannelMessage converts SimConnect data to a channel message
func parseSimConnectToChannelMessage(ppData uintptr, pcbData uint32, engine *Engine) any {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast the pointer to the base SIMCONNECT_RECV structure
	recv := (*types.SIMCONNECT_RECV)(unsafe.Pointer(ppData))

	// Debug: also call parseSimConnectData for console output
	parseSimConnectData(ppData, pcbData, engine)

	// Create a simple message structure for the channel
	msg := map[string]any{
		"size":       recv.DwSize,
		"version":    recv.DwVersion,
		"type":       getMessageTypeName(recv.DwID),
		"id":         recv.DwID,
		"data":       ppData,
		"size_bytes": pcbData,
	}

	// For SIMOBJECT_DATA, add the parsed values directly
	if recv.DwID == types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA {
		if simVarData := parseSimObjectData(ppData, pcbData, engine); simVarData != nil {
			msg["parsed_data"] = simVarData
		}
	}
	// For EXCEPTION, add the parsed exception data
	if recv.DwID == types.SIMCONNECT_RECV_ID_EXCEPTION {
		if pcbData >= uint32(unsafe.Sizeof(types.SIMCONNECT_RECV_EXCEPTION{})) {
			exceptionData := (*types.SIMCONNECT_RECV_EXCEPTION)(unsafe.Pointer(ppData))

			// Convert to SimConnectException type
			exceptionCode := types.SimConnectException(exceptionData.DwException)
			// Create structured exception data using our helper functions
			exceptionInfo := &types.ExceptionData{
				ExceptionCode: exceptionCode,
				ExceptionName: types.GetExceptionName(exceptionCode),
				Description:   types.GetExceptionDescription(exceptionCode),
				SendID:        exceptionData.DwSendID,
				Index:         exceptionData.DwIndex,
				Severity:      types.GetExceptionSeverity(exceptionCode),
			}

			msg["exception"] = exceptionInfo
		}
	}

	// For EVENT, add the parsed event data
	if recv.DwID == types.SIMCONNECT_RECV_ID_EVENT {
		if eventData := parseEventData(ppData, pcbData, engine); eventData != nil {
			msg["event"] = eventData
		}
	}

	return msg
}

// getMessageTypeName converts message ID to readable string
func getMessageTypeName(id types.SimConnectRecvID) string {
	switch id {
	case types.SIMCONNECT_RECV_ID_NULL:
		return "NULL"
	case types.SIMCONNECT_RECV_ID_EXCEPTION:
		return "EXCEPTION"
	case types.SIMCONNECT_RECV_ID_OPEN:
		return "OPEN"
	case types.SIMCONNECT_RECV_ID_QUIT:
		return "QUIT"
	case types.SIMCONNECT_RECV_ID_EVENT:
		return "EVENT"
	case types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA:
		return "SIMOBJECT_DATA"
	case types.SIMCONNECT_RECV_ID_SYSTEM_STATE:
		return "SYSTEM_STATE"
	case types.SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENTS:
		return "ENUMERATE_INPUT_EVENTS"
	case types.SIMCONNECT_RECV_ID_SUBSCRIBE_INPUT_EVENT:
		return "SUBSCRIBE_INPUT_EVENT"
	default:
		return "UNKNOWN"
	}
}

// parseEventData extracts event data from SIMCONNECT_RECV_EVENT message
func parseEventData(ppData uintptr, pcbData uint32, engine *Engine) *types.EventData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_EVENT structure
	eventData := (*types.SIMCONNECT_RECV_EVENT)(unsafe.Pointer(ppData))
	if eventData.DwID != types.SIMCONNECT_RECV_ID_EVENT {
		return nil
	}

	// Create event data structure for channel message
	result := &types.EventData{
		GroupID:   eventData.DwGroupID,
		EventID:   eventData.DwEventID,
		EventData: eventData.DwData,
		EventType: "unknown", // Default type
	} // TODO: Add event name lookup when event registry is implemented
	// For now, we'll use the event ID and assume they're system events
	// since we're starting with system event subscriptions
	result.EventType = "system"

	return result
}

// isSystemEvent determines if an event name represents a system event
// System events are predefined events from the simulator
func isSystemEvent(eventName string) bool {
	systemEvents := map[string]bool{
		"Paused":                true,
		"Unpaused":              true,
		"Sim":                   true,
		"SimStart":              true,
		"SimStop":               true,
		"AircraftLoaded":        true,
		"FlightLoaded":          true,
		"FlightSaved":           true,
		"FlightPlanActivated":   true,
		"FlightPlanDeactivated": true,
		"PositionChanged":       true,
		"CrashReset":            true,
		"Crashed":               true,
		"CustomMission":         true,
		"MissionCompleted":      true,
		"WeatherModeChanged":    true,
		"View":                  true,
		"Sound":                 true,
		"ObjectAdded":           true,
		"ObjectRemoved":         true,
		// Add more system events as needed
	}
	return systemEvents[eventName]
}
