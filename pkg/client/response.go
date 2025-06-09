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
func (e *Engine) parseSimObjectData(ppData uintptr, pcbData uint32) *SimVarData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_SIMOBJECT_DATA structure
	simObjData := (*types.SIMCONNECT_RECV_SIMOBJECT_DATA)(unsafe.Pointer(ppData))
	if simObjData.DwID != types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA {
		return nil
	}

	// Look up the expected data type for this DefineID (thread-safe)
	e.mu.RLock()
	dataType, exists := e.dataTypeRegistry[simObjData.DwDefineID]
	e.mu.RUnlock()
	if !exists {
		// Fallback to FLOAT32 if not found
		dataType = types.SIMCONNECT_DATATYPE_FLOAT32
	}
	var value interface{}
	headerSize := unsafe.Sizeof(*simObjData)

	// Parse based on the registered data type - now supports all 17 SIMCONNECT_DATATYPE values
	switch dataType {
	// === NUMERIC TYPES ===
	case types.SIMCONNECT_DATATYPE_FLOAT32:
		// For FLOAT32: 4-byte value stored in DwData field
		float32Value := *(*float32)(unsafe.Pointer(&simObjData.DwData))
		value = float64(float32Value)

	case types.SIMCONNECT_DATATYPE_FLOAT64:
		// For FLOAT64: 8-byte double precision value after header
		if pcbData >= uint32(headerSize)+8 {
			dataPtr := ppData + uintptr(headerSize)
			value = *(*float64)(unsafe.Pointer(dataPtr))
		} else {
			value = float64(0.0)
		}

	case types.SIMCONNECT_DATATYPE_INT32:
		// For INT32: 4-byte integer stored in DwData field
		int32Value := *(*int32)(unsafe.Pointer(&simObjData.DwData))
		value = int32Value

	case types.SIMCONNECT_DATATYPE_INT64:
		// For INT64: 8-byte integer after header
		if pcbData >= uint32(headerSize)+8 {
			dataPtr := ppData + uintptr(headerSize)
			value = *(*int64)(unsafe.Pointer(dataPtr))
		} else {
			value = int64(0)
		}

	// === STRING TYPES ===
	case types.SIMCONNECT_DATATYPE_STRINGV:
		// For STRINGV: Variable-length string data comes after the header
		value = e.parseVariableString(ppData, pcbData, headerSize)

	case types.SIMCONNECT_DATATYPE_STRING8:
		value = e.parseFixedString(ppData, pcbData, headerSize, 8)
	case types.SIMCONNECT_DATATYPE_STRING32:
		value = e.parseFixedString(ppData, pcbData, headerSize, 32)
	case types.SIMCONNECT_DATATYPE_STRING64:
		value = e.parseFixedString(ppData, pcbData, headerSize, 64)
	case types.SIMCONNECT_DATATYPE_STRING128:
		value = e.parseFixedString(ppData, pcbData, headerSize, 128)
	case types.SIMCONNECT_DATATYPE_STRING256:
		value = e.parseFixedString(ppData, pcbData, headerSize, 256)
	case types.SIMCONNECT_DATATYPE_STRING260:
		value = e.parseFixedString(ppData, pcbData, headerSize, 260)

	// === STRUCTURE TYPES ===
	case types.SIMCONNECT_DATATYPE_INITPOSITION:
		value = e.parseInitPosition(ppData, pcbData, headerSize)
	case types.SIMCONNECT_DATATYPE_MARKERSTATE:
		value = e.parseMarkerState(ppData, pcbData, headerSize)
	case types.SIMCONNECT_DATATYPE_WAYPOINT:
		value = e.parseWaypoint(ppData, pcbData, headerSize)
	case types.SIMCONNECT_DATATYPE_LATLONALT:
		value = e.parseLatLonAlt(ppData, pcbData, headerSize)
	case types.SIMCONNECT_DATATYPE_XYZ:
		value = e.parseXYZ(ppData, pcbData, headerSize)

	case types.SIMCONNECT_DATATYPE_INVALID:
		// Invalid data type - return nil value
		value = nil

	default:
		// Enhanced fallback with type information for debugging
		value = e.parseUnknownType(ppData, pcbData, headerSize, dataType)
	}

	return &SimVarData{
		RequestID: simObjData.DwRequestID,
		DefineID:  simObjData.DwDefineID,
		Value:     value,
	}
}

// parseSimConnectData processes incoming SimConnect messages for debugging
/*func parseSimConnectData(ppData uintptr, pcbData uint32, engine *Engine) {
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
}*/

// parseSimConnectToChannelMessage converts SimConnect data to a channel message
func (e *Engine) parseSimConnectToChannelMessage(ppData uintptr, pcbData uint32) any {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast the pointer to the base SIMCONNECT_RECV structure
	recv := (*types.SIMCONNECT_RECV)(unsafe.Pointer(ppData))

	// Debug: also call parseSimConnectData for console output
	//parseSimConnectData(ppData, pcbData, engine)

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
		if simVarData := e.parseSimObjectData(ppData, pcbData); simVarData != nil {
			msg["parsed_data"] = simVarData
		}
	}

	// For SIMOBJECT_DATA_BYTYPE, add the parsed values directly
	if recv.DwID == types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA_BYTYPE {
		if simVarData := e.parseSimObjectData(ppData, pcbData); simVarData != nil {
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
		if eventData := e.parseEventData(ppData, pcbData); eventData != nil {
			msg["event"] = eventData
		}
	}

	// For EVENT_EX1, add the parsed extended event data
	if recv.DwID == types.SIMCONNECT_RECV_ID_EVENT_EX1 {
		if eventExData := e.parseEventExData(ppData, pcbData); eventExData != nil {
			msg["event_ex"] = eventExData
		}
	}

	// For ASSIGNED_OBJECT_ID, add the parsed object data
	if recv.DwID == types.SIMCONNECT_RECV_ID_ASSIGNED_OBJECT_ID {
		if objectData := e.parseAssignedObjectData(ppData, pcbData); objectData != nil {
			msg["assigned_object"] = objectData
		}
	}

	// For SYSTEM_STATE, add the parsed system state data
	if recv.DwID == types.SIMCONNECT_RECV_ID_SYSTEM_STATE {
		if stateData := e.parseSystemStateData(ppData, pcbData); stateData != nil {
			msg["system_state"] = stateData
		}
	}

	// For CLIENT_DATA, add the parsed client data
	if recv.DwID == types.SIMCONNECT_RECV_ID_CLIENT_DATA {
		if clientData := e.parseClientData(ppData, pcbData); clientData != nil {
			msg["client_data"] = clientData
		}
	}
	// For CUSTOM_ACTION, add the parsed custom action data
	if recv.DwID == types.SIMCONNECT_RECV_ID_CUSTOM_ACTION {
		if actionData := e.parseCustomActionData(ppData, pcbData); actionData != nil {
			msg["custom_action"] = actionData
		}
	}

	// === NEW CRITICAL EVENT PARSERS ===

	// For EVENT_OBJECT_ADDREMOVE, add the parsed object event data
	if recv.DwID == types.SIMCONNECT_RECV_ID_EVENT_OBJECT_ADDREMOVE {
		if objData := e.parseObjectAddRemoveData(ppData, pcbData); objData != nil {
			msg["object_event"] = objData
		}
	}

	// For EVENT_FILENAME, add the parsed filename event data
	if recv.DwID == types.SIMCONNECT_RECV_ID_EVENT_FILENAME {
		if filenameData := e.parseFilenameEventData(ppData, pcbData); filenameData != nil {
			msg["filename_event"] = filenameData
		}
	}

	// For EVENT_FRAME, add the parsed frame event data
	if recv.DwID == types.SIMCONNECT_RECV_ID_EVENT_FRAME {
		if frameData := e.parseFrameEventData(ppData, pcbData); frameData != nil {
			msg["frame_event"] = frameData
		}
	}

	// For FACILITY_DATA, add the parsed facility data
	if recv.DwID == types.SIMCONNECT_RECV_ID_FACILITY_DATA {
		if facilityData := e.parseFacilityData(ppData, pcbData); facilityData != nil {
			msg["facility_data"] = facilityData
		}
	}

	// For PICK events, add the parsed pick event data
	if recv.DwID == types.SIMCONNECT_RECV_ID_PICK {
		if pickData := e.parsePickEventData(ppData, pcbData); pickData != nil {
			msg["pick_event"] = pickData
		}
	}

	// === GENERIC FALLBACK HANDLING ===
	// Handle any unhandled message types with basic raw data extraction
	if !e.isHandledMessageType(recv.DwID) {
		msg["unhandled"] = true
		msg["raw_data"] = e.extractRawMessageData(ppData, pcbData)

		// Optional: Log unhandled message types for monitoring
		// This helps identify which message types are actually being received
		// but not yet implemented
		if e.shouldLogUnhandledMessage(recv.DwID) {
			// Note: In production, you might want to use a proper logger
			// and rate-limit these messages to avoid spam
			_ = recv.DwID // Placeholder - replace with actual logging if needed
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
	case types.SIMCONNECT_RECV_ID_EVENT_EX1:
		return "EVENT_EX1"
	case types.SIMCONNECT_RECV_ID_EVENT_OBJECT_ADDREMOVE:
		return "EVENT_OBJECT_ADDREMOVE"
	case types.SIMCONNECT_RECV_ID_EVENT_FILENAME:
		return "EVENT_FILENAME"
	case types.SIMCONNECT_RECV_ID_EVENT_FRAME:
		return "EVENT_FRAME"
	case types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA:
		return "SIMOBJECT_DATA"
	case types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA_BYTYPE:
		return "SIMOBJECT_DATA_BYTYPE"
	case types.SIMCONNECT_RECV_ID_ASSIGNED_OBJECT_ID:
		return "ASSIGNED_OBJECT_ID"
	case types.SIMCONNECT_RECV_ID_RESERVED_KEY:
		return "RESERVED_KEY"
	case types.SIMCONNECT_RECV_ID_CUSTOM_ACTION:
		return "CUSTOM_ACTION"
	case types.SIMCONNECT_RECV_ID_SYSTEM_STATE:
		return "SYSTEM_STATE"
	case types.SIMCONNECT_RECV_ID_CLIENT_DATA:
		return "CLIENT_DATA"
	case types.SIMCONNECT_RECV_ID_WEATHER_OBSERVATION:
		return "WEATHER_OBSERVATION"
	case types.SIMCONNECT_RECV_ID_CLOUD_STATE:
		return "CLOUD_STATE"
	case types.SIMCONNECT_RECV_ID_EVENT_WEATHER_MODE:
		return "EVENT_WEATHER_MODE"
	case types.SIMCONNECT_RECV_ID_AIRPORT_LIST:
		return "AIRPORT_LIST"
	case types.SIMCONNECT_RECV_ID_VOR_LIST:
		return "VOR_LIST"
	case types.SIMCONNECT_RECV_ID_NDB_LIST:
		return "NDB_LIST"
	case types.SIMCONNECT_RECV_ID_WAYPOINT_LIST:
		return "WAYPOINT_LIST"
	case types.SIMCONNECT_RECV_ID_EVENT_MULTIPLAYER_SERVER_STARTED:
		return "EVENT_MULTIPLAYER_SERVER_STARTED"
	case types.SIMCONNECT_RECV_ID_EVENT_MULTIPLAYER_CLIENT_STARTED:
		return "EVENT_MULTIPLAYER_CLIENT_STARTED"
	case types.SIMCONNECT_RECV_ID_EVENT_MULTIPLAYER_SESSION_ENDED:
		return "EVENT_MULTIPLAYER_SESSION_ENDED"
	case types.SIMCONNECT_RECV_ID_EVENT_RACE_END:
		return "EVENT_RACE_END"
	case types.SIMCONNECT_RECV_ID_EVENT_RACE_LAP:
		return "EVENT_RACE_LAP"
	case types.SIMCONNECT_RECV_ID_PICK:
		return "PICK"
	case types.SIMCONNECT_RECV_ID_FACILITY_DATA:
		return "FACILITY_DATA"
	case types.SIMCONNECT_RECV_ID_FACILITY_DATA_END:
		return "FACILITY_DATA_END"
	case types.SIMCONNECT_RECV_ID_FACILITY_MINIMAL_LIST:
		return "FACILITY_MINIMAL_LIST"
	case types.SIMCONNECT_RECV_ID_JETWAY_DATA:
		return "JETWAY_DATA"
	case types.SIMCONNECT_RECV_ID_CONTROLLERS_LIST:
		return "CONTROLLERS_LIST"
	case types.SIMCONNECT_RECV_ID_ACTION_CALLBACK:
		return "ACTION_CALLBACK"
	case types.SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENTS:
		return "ENUMERATE_INPUT_EVENTS"
	case types.SIMCONNECT_RECV_ID_GET_INPUT_EVENT:
		return "GET_INPUT_EVENT"
	case types.SIMCONNECT_RECV_ID_SUBSCRIBE_INPUT_EVENT:
		return "SUBSCRIBE_INPUT_EVENT"
	case types.SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENT_PARAMS:
		return "ENUMERATE_INPUT_EVENT_PARAMS"
	default:
		return "UNKNOWN"
	}
}

// parseEventData extracts event data from SIMCONNECT_RECV_EVENT message
func (e *Engine) parseEventData(ppData uintptr, pcbData uint32) *types.EventData {
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
		GroupID:   eventData.UGroupID,
		EventID:   eventData.UEventID,
		EventData: eventData.DwData,
		EventType: eventData.DwID, // Default type
	}

	// Classify event type and resolve event name based on EventID
	//result.EventType, result.EventName = classifyEvent(eventData.UEventID, eventData.UGroupID)

	return result
}

// parseEventExData extracts extended event data from SIMCONNECT_RECV_EVENT_EX1 message
func (e *Engine) parseEventExData(ppData uintptr, pcbData uint32) *types.EventExData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_EVENT_EX1 structure
	eventData := (*types.SIMCONNECT_RECV_EVENT_EX1)(unsafe.Pointer(ppData))
	if eventData.DwID != types.SIMCONNECT_RECV_ID_EVENT_EX1 {
		return nil
	}

	// Create extended event data structure for channel message
	result := &types.EventExData{
		GroupID: eventData.UGroupID,
		EventID: eventData.UEventID,
		Data: []uint32{
			eventData.DwData0,
			eventData.DwData1,
			eventData.DwData2,
			eventData.DwData3,
			eventData.DwData4,
		},
	}

	return result
}

// parseAssignedObjectData extracts assigned object ID data from SIMCONNECT_RECV_ASSIGNED_OBJECT_ID message
func (e *Engine) parseAssignedObjectData(ppData uintptr, pcbData uint32) *types.AssignedObjectData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_ASSIGNED_OBJECT_ID structure
	objData := (*types.SIMCONNECT_RECV_ASSIGNED_OBJECT_ID)(unsafe.Pointer(ppData))
	if objData.DwID != types.SIMCONNECT_RECV_ID_ASSIGNED_OBJECT_ID {
		return nil
	}

	// Create assigned object data structure for channel message
	result := &types.AssignedObjectData{
		ObjectID:  objData.DwObjectID,
		RequestID: objData.DwRequestID,
	}

	return result
}

// parseSystemStateData extracts system state data from SIMCONNECT_RECV_SYSTEM_STATE message
func (e *Engine) parseSystemStateData(ppData uintptr, pcbData uint32) *types.SystemStateData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_SYSTEM_STATE structure
	stateData := (*types.SIMCONNECT_RECV_SYSTEM_STATE)(unsafe.Pointer(ppData))
	if stateData.DwID != types.SIMCONNECT_RECV_ID_SYSTEM_STATE {
		return nil
	}

	// Convert float from uint32 representation
	floatValue := *(*float32)(unsafe.Pointer(&stateData.DwFloat))

	// Convert string from byte array (find null terminator)
	stringValue := ""
	for i, b := range stateData.SzString {
		if b == 0 {
			stringValue = string(stateData.SzString[:i])
			break
		}
	}

	// Create system state data structure for channel message
	result := &types.SystemStateData{
		RequestID:    stateData.DwRequestID,
		IntegerValue: stateData.DwInteger,
		FloatValue:   floatValue,
		StringValue:  stringValue,
	}

	return result
}

// parseClientData extracts client data from SIMCONNECT_RECV_CLIENT_DATA message
func (e *Engine) parseClientData(ppData uintptr, pcbData uint32) *types.ClientData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_CLIENT_DATA structure
	clientData := (*types.SIMCONNECT_RECV_CLIENT_DATA)(unsafe.Pointer(ppData))
	if clientData.DwID != types.SIMCONNECT_RECV_ID_CLIENT_DATA {
		return nil
	}

	// For client data, we need to parse the actual data based on the definition
	// For now, we'll store the raw data pointer and size
	var data interface{}
	headerSize := unsafe.Sizeof(*clientData)
	if pcbData > uint32(headerSize) {
		// Calculate data location and available bytes
		dataPtr := ppData + uintptr(headerSize)
		dataLen := pcbData - uint32(headerSize)

		// For basic implementation, store as byte slice
		dataBytes := make([]byte, dataLen)
		for i := uint32(0); i < dataLen; i++ {
			dataBytes[i] = *(*byte)(unsafe.Pointer(dataPtr + uintptr(i)))
		}
		data = dataBytes
	}

	// Create client data structure for channel message
	result := &types.ClientData{
		RequestID:    clientData.DwRequestID,
		DefineID:     clientData.DwDefineID,
		EntryNumber:  clientData.DwEntryNumber,
		TotalEntries: clientData.DwOutOf,
		Data:         data,
	}

	return result
}

// parseCustomActionData extracts custom action data from SIMCONNECT_RECV_CUSTOM_ACTION message
func (e *Engine) parseCustomActionData(ppData uintptr, pcbData uint32) *types.CustomActionData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_CUSTOM_ACTION structure
	actionData := (*types.SIMCONNECT_RECV_CUSTOM_ACTION)(unsafe.Pointer(ppData))
	if actionData.DwID != types.SIMCONNECT_RECV_ID_CUSTOM_ACTION {
		return nil
	}

	// Create custom action data structure for channel message
	result := &types.CustomActionData{
		GuidRequestID: actionData.DwGuidRequestID,
		UserRequestID: actionData.DwURequestID,
		Result:        actionData.DwResult,
	}

	return result
}

// === NEW CRITICAL EVENT PARSERS ===

// parseObjectAddRemoveData extracts object add/remove event data from SIMCONNECT_RECV_EVENT_OBJECT_ADDREMOVE message
func (e *Engine) parseObjectAddRemoveData(ppData uintptr, pcbData uint32) *types.ObjectAddRemoveData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_EVENT_OBJECT_ADDREMOVE structure
	objEvent := (*types.SIMCONNECT_RECV_EVENT_OBJECT_ADDREMOVE)(unsafe.Pointer(ppData))
	if objEvent.DwID != types.SIMCONNECT_RECV_ID_EVENT_OBJECT_ADDREMOVE {
		return nil
	}

	// Determine action based on event ID (this is application-specific)
	action := "unknown"
	// Note: In practice, you would map specific event IDs to "added" or "removed"
	// This requires knowledge of your registered event IDs

	// Create object add/remove data structure for channel message
	result := &types.ObjectAddRemoveData{
		EventID:  objEvent.UEventID,
		ObjectID: objEvent.DwData,
		Action:   action,
	}

	return result
}

// parseFilenameEventData extracts filename event data from SIMCONNECT_RECV_EVENT_FILENAME message
func (e *Engine) parseFilenameEventData(ppData uintptr, pcbData uint32) *types.FilenameEventData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_EVENT_FILENAME structure
	filenameEvent := (*types.SIMCONNECT_RECV_EVENT_FILENAME)(unsafe.Pointer(ppData))
	if filenameEvent.DwID != types.SIMCONNECT_RECV_ID_EVENT_FILENAME {
		return nil
	}

	// Convert filename from byte array (find null terminator)
	filename := ""
	for i, b := range filenameEvent.SzFileName {
		if b == 0 {
			filename = string(filenameEvent.SzFileName[:i])
			break
		}
	}

	// Create filename event data structure for channel message
	result := &types.FilenameEventData{
		EventID:  filenameEvent.UEventID,
		Flags:    filenameEvent.DwFlags,
		GroupID:  filenameEvent.DwGroupID,
		Filename: filename,
	}

	return result
}

// parseFrameEventData extracts frame event data from SIMCONNECT_RECV_EVENT_FRAME message
func (e *Engine) parseFrameEventData(ppData uintptr, pcbData uint32) *types.FrameEventData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_EVENT_FRAME structure
	frameEvent := (*types.SIMCONNECT_RECV_EVENT_FRAME)(unsafe.Pointer(ppData))
	if frameEvent.DwID != types.SIMCONNECT_RECV_ID_EVENT_FRAME {
		return nil
	}

	// Create frame event data structure for channel message
	result := &types.FrameEventData{
		FrameRate: frameEvent.DwFrameRate,
		SimSpeed:  frameEvent.DwSimSpeed,
	}

	return result
}

// parseFacilityData extracts facility data from SIMCONNECT_RECV_FACILITY_DATA message
func (e *Engine) parseFacilityData(ppData uintptr, pcbData uint32) *types.FacilityData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_FACILITY_DATA structure
	facilityData := (*types.SIMCONNECT_RECV_FACILITY_DATA)(unsafe.Pointer(ppData))
	if facilityData.DwID != types.SIMCONNECT_RECV_ID_FACILITY_DATA {
		return nil
	}

	// The actual facility data follows the header
	headerSize := unsafe.Sizeof(*facilityData)
	var data interface{}

	if pcbData > uint32(headerSize) {
		// Extract raw data bytes for now
		// In practice, this would be parsed based on the specific facility type
		dataLen := pcbData - uint32(headerSize)
		dataPtr := ppData + uintptr(headerSize)
		dataBytes := make([]byte, dataLen)
		for i := uint32(0); i < dataLen; i++ {
			dataBytes[i] = *(*byte)(unsafe.Pointer(dataPtr + uintptr(i)))
		}
		data = dataBytes
	}

	// Create facility data structure for channel message
	result := &types.FacilityData{
		RequestID:    facilityData.DwRequestID,
		ArraySize:    facilityData.DwArraySize,
		EntryNumber:  facilityData.DwEntryNumber,
		TotalEntries: facilityData.DwOutOf,
		Data:         data,
	}

	return result
}

// parsePickEventData extracts pick event data from SIMCONNECT_RECV_PICK message
func (e *Engine) parsePickEventData(ppData uintptr, pcbData uint32) *types.PickEventData {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast to the proper SIMCONNECT_RECV_PICK structure
	pickEvent := (*types.SIMCONNECT_RECV_PICK)(unsafe.Pointer(ppData))
	if pickEvent.DwID != types.SIMCONNECT_RECV_ID_PICK {
		return nil
	}

	// Create pick event data structure for channel message
	result := &types.PickEventData{
		ObjectID:   pickEvent.DwObjectID,
		PickType:   pickEvent.DwPickType,
		PickSource: pickEvent.DwPickSource,
	}

	return result
}

// === GENERIC FALLBACK HELPER FUNCTIONS ===

// isHandledMessageType checks if a message type has a specific parser implemented
func (e *Engine) isHandledMessageType(messageType types.SimConnectRecvID) bool {
	switch messageType {
	case types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA,
		types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA_BYTYPE,
		types.SIMCONNECT_RECV_ID_EXCEPTION,
		types.SIMCONNECT_RECV_ID_EVENT,
		types.SIMCONNECT_RECV_ID_EVENT_EX1,
		types.SIMCONNECT_RECV_ID_ASSIGNED_OBJECT_ID,
		types.SIMCONNECT_RECV_ID_SYSTEM_STATE,
		types.SIMCONNECT_RECV_ID_CLIENT_DATA,
		types.SIMCONNECT_RECV_ID_CUSTOM_ACTION,
		// New critical parsers
		types.SIMCONNECT_RECV_ID_EVENT_OBJECT_ADDREMOVE,
		types.SIMCONNECT_RECV_ID_EVENT_FILENAME,
		types.SIMCONNECT_RECV_ID_EVENT_FRAME,
		types.SIMCONNECT_RECV_ID_FACILITY_DATA,
		types.SIMCONNECT_RECV_ID_PICK:
		return true
	default:
		return false
	}
}

// extractRawMessageData extracts basic information from unhandled message types
func (e *Engine) extractRawMessageData(ppData uintptr, pcbData uint32) map[string]interface{} {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Extract just the basic header information safely
	recv := (*types.SIMCONNECT_RECV)(unsafe.Pointer(ppData))

	rawData := map[string]interface{}{
		"header_size":    recv.DwSize,
		"header_version": recv.DwVersion,
		"message_id":     recv.DwID,
		"total_bytes":    pcbData,
	}

	// Extract first few bytes of payload data if available
	headerSize := unsafe.Sizeof(*recv)
	if pcbData > uint32(headerSize) {
		payloadSize := pcbData - uint32(headerSize)
		if payloadSize > 0 {
			// Limit to first 16 bytes to avoid large data dumps
			maxBytes := uint32(16)
			if payloadSize < maxBytes {
				maxBytes = payloadSize
			}

			payloadPtr := ppData + uintptr(headerSize)
			payload := make([]byte, maxBytes)
			for i := uint32(0); i < maxBytes; i++ {
				payload[i] = *(*byte)(unsafe.Pointer(payloadPtr + uintptr(i)))
			}
			rawData["payload_preview"] = payload
			rawData["payload_size"] = payloadSize
		}
	}

	return rawData
}

// shouldLogUnhandledMessage determines if an unhandled message type should be logged
// This helps with rate limiting and focusing on important unhandled messages
func (e *Engine) shouldLogUnhandledMessage(messageType types.SimConnectRecvID) bool {
	// Skip logging for common/expected unhandled message types that we don't need
	switch messageType {
	case types.SIMCONNECT_RECV_ID_NULL,
		types.SIMCONNECT_RECV_ID_OPEN,
		types.SIMCONNECT_RECV_ID_QUIT:
		return false // These are common and not critical to log
	default:
		// Log other unhandled types so we can see what's being missed
		return true
	}
}

// Helper functions for parsing different SimConnect data types

// parseVariableString parses SIMCONNECT_DATATYPE_STRINGV - variable length string
func (e *Engine) parseVariableString(ppData uintptr, pcbData uint32, headerSize uintptr) string {
	if pcbData <= uint32(headerSize) {
		return ""
	}

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
	return string(stringBytes)
}

// parseFixedString parses fixed-length string types (STRING8, STRING32, etc.)
func (e *Engine) parseFixedString(ppData uintptr, pcbData uint32, headerSize uintptr, maxLen int) string {
	expectedSize := uint32(headerSize) + uint32(maxLen)
	if pcbData < expectedSize {
		// Not enough data, return empty string
		return ""
	}

	// Read fixed-length string data
	stringDataPtr := ppData + uintptr(headerSize)
	stringBytes := make([]byte, maxLen)

	for i := 0; i < maxLen; i++ {
		b := *(*byte)(unsafe.Pointer(stringDataPtr + uintptr(i)))
		if b == 0 {
			// Found null terminator
			stringBytes = stringBytes[:i]
			break
		}
		stringBytes[i] = b
	}

	return string(stringBytes)
}

// parseInitPosition parses SIMCONNECT_DATATYPE_INITPOSITION structure
func (e *Engine) parseInitPosition(ppData uintptr, pcbData uint32, headerSize uintptr) *types.InitPosition {
	expectedSize := uint32(headerSize) + uint32(unsafe.Sizeof(types.InitPosition{}))
	if pcbData < expectedSize {
		return nil
	}

	dataPtr := ppData + uintptr(headerSize)
	initPos := (*types.InitPosition)(unsafe.Pointer(dataPtr))

	// Return a copy to avoid pointer issues
	return &types.InitPosition{
		Latitude:  initPos.Latitude,
		Longitude: initPos.Longitude,
		Altitude:  initPos.Altitude,
		Pitch:     initPos.Pitch,
		Bank:      initPos.Bank,
		Heading:   initPos.Heading,
		OnGround:  initPos.OnGround,
		Airspeed:  initPos.Airspeed,
	}
}

// parseMarkerState parses SIMCONNECT_DATATYPE_MARKERSTATE structure
func (e *Engine) parseMarkerState(ppData uintptr, pcbData uint32, headerSize uintptr) *types.MarkerState {
	expectedSize := uint32(headerSize) + uint32(unsafe.Sizeof(types.MarkerState{}))
	if pcbData < expectedSize {
		return nil
	}

	dataPtr := ppData + uintptr(headerSize)
	marker := (*types.MarkerState)(unsafe.Pointer(dataPtr))

	// Return a copy to avoid pointer issues
	result := &types.MarkerState{
		Latitude:  marker.Latitude,
		Longitude: marker.Longitude,
		Altitude:  marker.Altitude,
		Flags:     marker.Flags,
		Heading:   marker.Heading,
		Speed:     marker.Speed,
		Bank:      marker.Bank,
		Pitch:     marker.Pitch,
	}

	// Copy the name byte array
	copy(result.Name[:], marker.Name[:])

	return result
}

// parseWaypoint parses SIMCONNECT_DATATYPE_WAYPOINT structure
func (e *Engine) parseWaypoint(ppData uintptr, pcbData uint32, headerSize uintptr) *types.Waypoint {
	expectedSize := uint32(headerSize) + uint32(unsafe.Sizeof(types.Waypoint{}))
	if pcbData < expectedSize {
		return nil
	}

	dataPtr := ppData + uintptr(headerSize)
	waypoint := (*types.Waypoint)(unsafe.Pointer(dataPtr))

	// Return a copy to avoid pointer issues
	return &types.Waypoint{
		Latitude:  waypoint.Latitude,
		Longitude: waypoint.Longitude,
		Altitude:  waypoint.Altitude,
		Flags:     waypoint.Flags,
		Speed:     waypoint.Speed,
		Throttle:  waypoint.Throttle,
	}
}

// parseLatLonAlt parses SIMCONNECT_DATATYPE_LATLONALT structure
func (e *Engine) parseLatLonAlt(ppData uintptr, pcbData uint32, headerSize uintptr) *types.LatLonAlt {
	expectedSize := uint32(headerSize) + uint32(unsafe.Sizeof(types.LatLonAlt{}))
	if pcbData < expectedSize {
		return nil
	}

	dataPtr := ppData + uintptr(headerSize)
	latLonAlt := (*types.LatLonAlt)(unsafe.Pointer(dataPtr))

	// Return a copy to avoid pointer issues
	return &types.LatLonAlt{
		Latitude:  latLonAlt.Latitude,
		Longitude: latLonAlt.Longitude,
		Altitude:  latLonAlt.Altitude,
	}
}

// parseXYZ parses SIMCONNECT_DATATYPE_XYZ structure
func (e *Engine) parseXYZ(ppData uintptr, pcbData uint32, headerSize uintptr) *types.XYZ {
	expectedSize := uint32(headerSize) + uint32(unsafe.Sizeof(types.XYZ{}))
	if pcbData < expectedSize {
		return nil
	}

	dataPtr := ppData + uintptr(headerSize)
	xyz := (*types.XYZ)(unsafe.Pointer(dataPtr))

	// Return a copy to avoid pointer issues
	return &types.XYZ{
		X: xyz.X,
		Y: xyz.Y,
		Z: xyz.Z,
	}
}

// parseUnknownType handles unknown or unsupported data types with enhanced debugging
func (e *Engine) parseUnknownType(ppData uintptr, pcbData uint32, headerSize uintptr, dataType types.SimConnectDataType) interface{} {
	// For debugging: Log the unknown type (commented out to avoid stdout interference)
	// fmt.Printf("⚠️ Unknown/unsupported data type: %d, falling back to FLOAT32\n", dataType)

	// Fallback to FLOAT32 parsing for unknown types
	simObjData := (*types.SIMCONNECT_RECV_SIMOBJECT_DATA)(unsafe.Pointer(ppData))
	float32Value := *(*float32)(unsafe.Pointer(&simObjData.DwData))
	return float64(float32Value)
}
