package types

type SimConnectException uint32

// SIMCONNECT_EXCEPTION defines all possible exception codes returned by SimConnect
const (
	SIMCONNECT_EXCEPTION_NONE                              SimConnectException = iota // No error
	SIMCONNECT_EXCEPTION_ERROR                                                        // General error
	SIMCONNECT_EXCEPTION_SIZE_MISMATCH                                                // Size mismatch error
	SIMCONNECT_EXCEPTION_UNRECOGNIZED_ID                                              // Unrecognized ID
	SIMCONNECT_EXCEPTION_UNOPENED                                                     // SimConnect client not opened
	SIMCONNECT_EXCEPTION_VERSION_MISMATCH                                             // Version mismatch
	SIMCONNECT_EXCEPTION_TOO_MANY_GROUPS                                              // Too many groups
	SIMCONNECT_EXCEPTION_NAME_UNRECOGNIZED                                            // Name not recognized
	SIMCONNECT_EXCEPTION_TOO_MANY_EVENT_NAMES                                         // Too many event names
	SIMCONNECT_EXCEPTION_EVENT_ID_DUPLICATE                                           // Event ID already exists
	SIMCONNECT_EXCEPTION_TOO_MANY_MAPS                                                // Too many maps
	SIMCONNECT_EXCEPTION_TOO_MANY_OBJECTS                                             // Too many objects
	SIMCONNECT_EXCEPTION_TOO_MANY_REQUESTS                                            // Too many requests
	SIMCONNECT_EXCEPTION_WEATHER_INVALID_PORT                                         // Invalid weather port
	SIMCONNECT_EXCEPTION_WEATHER_INVALID_METAR                                        // Invalid METAR string
	SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_GET_OBSERVATION                            // Unable to get weather observation
	SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_CREATE_STATION                             // Unable to create weather station
	SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_REMOVE_STATION                             // Unable to remove weather station
	SIMCONNECT_EXCEPTION_INVALID_DATA_TYPE                                            // Invalid data type
	SIMCONNECT_EXCEPTION_INVALID_DATA_SIZE                                            // Invalid data size
	SIMCONNECT_EXCEPTION_DATA_ERROR                                                   // Data error
	SIMCONNECT_EXCEPTION_INVALID_ARRAY                                                // Invalid array
	SIMCONNECT_EXCEPTION_CREATE_OBJECT_FAILED                                         // Failed to create object
	SIMCONNECT_EXCEPTION_LOAD_FLIGHTPLAN_FAILED                                       // Failed to load flight plan
	SIMCONNECT_EXCEPTION_OPERATION_INVALID_FOR_OBJECT_TYPE                            // Operation invalid for object type
	SIMCONNECT_EXCEPTION_ILLEGAL_OPERATION                                            // Illegal operation
	SIMCONNECT_EXCEPTION_ALREADY_SUBSCRIBED                                           // Already subscribed
	SIMCONNECT_EXCEPTION_INVALID_ENUM                                                 // Invalid enum
	SIMCONNECT_EXCEPTION_DEFINITION_ERROR                                             // Definition error
	SIMCONNECT_EXCEPTION_DUPLICATE_ID                                                 // Duplicate ID
	SIMCONNECT_EXCEPTION_DATUM_ID                                                     // Datum ID error
	SIMCONNECT_EXCEPTION_OUT_OF_BOUNDS                                                // Out of bounds
	SIMCONNECT_EXCEPTION_ALREADY_CREATED                                              // Already created
	SIMCONNECT_EXCEPTION_OBJECT_OUTSIDE_REALITY_BUBBLE                                // Object outside reality bubble
	SIMCONNECT_EXCEPTION_OBJECT_CONTAINER                                             // Object container error
	SIMCONNECT_EXCEPTION_OBJECT_AI                                                    // AI object error
	SIMCONNECT_EXCEPTION_OBJECT_ATC                                                   // ATC object error
	SIMCONNECT_EXCEPTION_OBJECT_SCHEDULE                                              // Object schedule error
	SIMCONNECT_EXCEPTION_JETWAY_DATA                                                  // Jetway data error
	SIMCONNECT_EXCEPTION_ACTION_NOT_FOUND                                             // Action not found
	SIMCONNECT_EXCEPTION_NOT_AN_ACTION                                                // Not an action
	SIMCONNECT_EXCEPTION_INCORRECT_ACTION_PARAMS                                      // Incorrect action parameters
	SIMCONNECT_EXCEPTION_GET_INPUT_EVENT_FAILED                                       // Failed to get input event
	SIMCONNECT_EXCEPTION_SET_INPUT_EVENT_FAILED                                       // Failed to set input event
)

// GetExceptionName returns the human-readable name for an exception code
func GetExceptionName(code SimConnectException) string {
	switch code {
	case SIMCONNECT_EXCEPTION_NONE:
		return "NONE"
	case SIMCONNECT_EXCEPTION_ERROR:
		return "ERROR"
	case SIMCONNECT_EXCEPTION_SIZE_MISMATCH:
		return "SIZE_MISMATCH"
	case SIMCONNECT_EXCEPTION_UNRECOGNIZED_ID:
		return "UNRECOGNIZED_ID"
	case SIMCONNECT_EXCEPTION_UNOPENED:
		return "UNOPENED"
	case SIMCONNECT_EXCEPTION_VERSION_MISMATCH:
		return "VERSION_MISMATCH"
	case SIMCONNECT_EXCEPTION_TOO_MANY_GROUPS:
		return "TOO_MANY_GROUPS"
	case SIMCONNECT_EXCEPTION_NAME_UNRECOGNIZED:
		return "NAME_UNRECOGNIZED"
	case SIMCONNECT_EXCEPTION_TOO_MANY_EVENT_NAMES:
		return "TOO_MANY_EVENT_NAMES"
	case SIMCONNECT_EXCEPTION_EVENT_ID_DUPLICATE:
		return "EVENT_ID_DUPLICATE"
	case SIMCONNECT_EXCEPTION_TOO_MANY_MAPS:
		return "TOO_MANY_MAPS"
	case SIMCONNECT_EXCEPTION_TOO_MANY_OBJECTS:
		return "TOO_MANY_OBJECTS"
	case SIMCONNECT_EXCEPTION_TOO_MANY_REQUESTS:
		return "TOO_MANY_REQUESTS"
	case SIMCONNECT_EXCEPTION_WEATHER_INVALID_PORT:
		return "WEATHER_INVALID_PORT"
	case SIMCONNECT_EXCEPTION_WEATHER_INVALID_METAR:
		return "WEATHER_INVALID_METAR"
	case SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_GET_OBSERVATION:
		return "WEATHER_UNABLE_TO_GET_OBSERVATION"
	case SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_CREATE_STATION:
		return "WEATHER_UNABLE_TO_CREATE_STATION"
	case SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_REMOVE_STATION:
		return "WEATHER_UNABLE_TO_REMOVE_STATION"
	case SIMCONNECT_EXCEPTION_INVALID_DATA_TYPE:
		return "INVALID_DATA_TYPE"
	case SIMCONNECT_EXCEPTION_INVALID_DATA_SIZE:
		return "INVALID_DATA_SIZE"
	case SIMCONNECT_EXCEPTION_DATA_ERROR:
		return "DATA_ERROR"
	case SIMCONNECT_EXCEPTION_INVALID_ARRAY:
		return "INVALID_ARRAY"
	case SIMCONNECT_EXCEPTION_CREATE_OBJECT_FAILED:
		return "CREATE_OBJECT_FAILED"
	case SIMCONNECT_EXCEPTION_LOAD_FLIGHTPLAN_FAILED:
		return "LOAD_FLIGHTPLAN_FAILED"
	case SIMCONNECT_EXCEPTION_OPERATION_INVALID_FOR_OBJECT_TYPE:
		return "OPERATION_INVALID_FOR_OBJECT_TYPE"
	case SIMCONNECT_EXCEPTION_ILLEGAL_OPERATION:
		return "ILLEGAL_OPERATION"
	case SIMCONNECT_EXCEPTION_ALREADY_SUBSCRIBED:
		return "ALREADY_SUBSCRIBED"
	case SIMCONNECT_EXCEPTION_INVALID_ENUM:
		return "INVALID_ENUM"
	case SIMCONNECT_EXCEPTION_DEFINITION_ERROR:
		return "DEFINITION_ERROR"
	case SIMCONNECT_EXCEPTION_DUPLICATE_ID:
		return "DUPLICATE_ID"
	case SIMCONNECT_EXCEPTION_DATUM_ID:
		return "DATUM_ID"
	case SIMCONNECT_EXCEPTION_OUT_OF_BOUNDS:
		return "OUT_OF_BOUNDS"
	case SIMCONNECT_EXCEPTION_ALREADY_CREATED:
		return "ALREADY_CREATED"
	case SIMCONNECT_EXCEPTION_OBJECT_OUTSIDE_REALITY_BUBBLE:
		return "OBJECT_OUTSIDE_REALITY_BUBBLE"
	case SIMCONNECT_EXCEPTION_OBJECT_CONTAINER:
		return "OBJECT_CONTAINER"
	case SIMCONNECT_EXCEPTION_OBJECT_AI:
		return "OBJECT_AI"
	case SIMCONNECT_EXCEPTION_OBJECT_ATC:
		return "OBJECT_ATC"
	case SIMCONNECT_EXCEPTION_OBJECT_SCHEDULE:
		return "OBJECT_SCHEDULE"
	case SIMCONNECT_EXCEPTION_JETWAY_DATA:
		return "JETWAY_DATA"
	case SIMCONNECT_EXCEPTION_ACTION_NOT_FOUND:
		return "ACTION_NOT_FOUND"
	case SIMCONNECT_EXCEPTION_NOT_AN_ACTION:
		return "NOT_AN_ACTION"
	case SIMCONNECT_EXCEPTION_INCORRECT_ACTION_PARAMS:
		return "INCORRECT_ACTION_PARAMS"
	case SIMCONNECT_EXCEPTION_GET_INPUT_EVENT_FAILED:
		return "GET_INPUT_EVENT_FAILED"
	case SIMCONNECT_EXCEPTION_SET_INPUT_EVENT_FAILED:
		return "SET_INPUT_EVENT_FAILED"
	default:
		return "UNKNOWN"
	}
}

// GetExceptionDescription returns a detailed description for an exception code
func GetExceptionDescription(code SimConnectException) string {
	switch code {
	case SIMCONNECT_EXCEPTION_NONE:
		return "No error occurred"
	case SIMCONNECT_EXCEPTION_ERROR:
		return "An unspecific error has occurred"
	case SIMCONNECT_EXCEPTION_SIZE_MISMATCH:
		return "The size of the data provided does not match the size required"
	case SIMCONNECT_EXCEPTION_UNRECOGNIZED_ID:
		return "The client event, request ID, data definition ID, or object ID was not recognized"
	case SIMCONNECT_EXCEPTION_UNOPENED:
		return "Communication with the SimConnect server has not been opened"
	case SIMCONNECT_EXCEPTION_VERSION_MISMATCH:
		return "A versioning error has occurred"
	case SIMCONNECT_EXCEPTION_TOO_MANY_GROUPS:
		return "The maximum number of groups allowed has been reached (max: 20)"
	case SIMCONNECT_EXCEPTION_NAME_UNRECOGNIZED:
		return "The simulation event name is not recognized"
	case SIMCONNECT_EXCEPTION_TOO_MANY_EVENT_NAMES:
		return "The maximum number of event names allowed has been reached (max: 1000)"
	case SIMCONNECT_EXCEPTION_EVENT_ID_DUPLICATE:
		return "The event ID has been used already"
	case SIMCONNECT_EXCEPTION_TOO_MANY_MAPS:
		return "The maximum number of mappings allowed has been reached (max: 20)"
	case SIMCONNECT_EXCEPTION_TOO_MANY_OBJECTS:
		return "The maximum number of objects allowed has been reached (max: 1000)"
	case SIMCONNECT_EXCEPTION_TOO_MANY_REQUESTS:
		return "The maximum number of requests allowed has been reached (max: 1000)"
	case SIMCONNECT_EXCEPTION_WEATHER_INVALID_PORT:
		return "Invalid weather port (deprecated)"
	case SIMCONNECT_EXCEPTION_WEATHER_INVALID_METAR:
		return "Invalid METAR string (deprecated)"
	case SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_GET_OBSERVATION:
		return "Unable to get weather observation (deprecated)"
	case SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_CREATE_STATION:
		return "Unable to create weather station (deprecated)"
	case SIMCONNECT_EXCEPTION_WEATHER_UNABLE_TO_REMOVE_STATION:
		return "Unable to remove weather station (deprecated)"
	case SIMCONNECT_EXCEPTION_INVALID_DATA_TYPE:
		return "The data type requested does not apply to the type of data requested"
	case SIMCONNECT_EXCEPTION_INVALID_DATA_SIZE:
		return "The size of the data provided is not what is expected"
	case SIMCONNECT_EXCEPTION_DATA_ERROR:
		return "A generic data error occurred"
	case SIMCONNECT_EXCEPTION_INVALID_ARRAY:
		return "An invalid array has been sent"
	case SIMCONNECT_EXCEPTION_CREATE_OBJECT_FAILED:
		return "The attempt to create an AI object failed"
	case SIMCONNECT_EXCEPTION_LOAD_FLIGHTPLAN_FAILED:
		return "The specified flight plan could not be found or loaded"
	case SIMCONNECT_EXCEPTION_OPERATION_INVALID_FOR_OBJECT_TYPE:
		return "The operation requested does not apply to the object type"
	case SIMCONNECT_EXCEPTION_ILLEGAL_OPERATION:
		return "The operation requested cannot be completed"
	case SIMCONNECT_EXCEPTION_ALREADY_SUBSCRIBED:
		return "The client has already subscribed to that event"
	case SIMCONNECT_EXCEPTION_INVALID_ENUM:
		return "The member of the enumeration provided was not valid"
	case SIMCONNECT_EXCEPTION_DEFINITION_ERROR:
		return "There is a problem with a data definition"
	case SIMCONNECT_EXCEPTION_DUPLICATE_ID:
		return "The ID has already been used"
	case SIMCONNECT_EXCEPTION_DATUM_ID:
		return "The datum ID is not recognized"
	case SIMCONNECT_EXCEPTION_OUT_OF_BOUNDS:
		return "The radius given was outside the acceptable range"
	case SIMCONNECT_EXCEPTION_ALREADY_CREATED:
		return "A client data area with the requested name has already been created"
	case SIMCONNECT_EXCEPTION_OBJECT_OUTSIDE_REALITY_BUBBLE:
		return "The object location is outside the reality bubble"
	case SIMCONNECT_EXCEPTION_OBJECT_CONTAINER:
		return "Error with the container system for the object"
	case SIMCONNECT_EXCEPTION_OBJECT_AI:
		return "Error with the AI system for the object"
	case SIMCONNECT_EXCEPTION_OBJECT_ATC:
		return "Error with the ATC system for the object"
	case SIMCONNECT_EXCEPTION_OBJECT_SCHEDULE:
		return "Error with object scheduling"
	case SIMCONNECT_EXCEPTION_JETWAY_DATA:
		return "Error retrieving jetway data"
	case SIMCONNECT_EXCEPTION_ACTION_NOT_FOUND:
		return "The given action cannot be found"
	case SIMCONNECT_EXCEPTION_NOT_AN_ACTION:
		return "The given action does not exist"
	case SIMCONNECT_EXCEPTION_INCORRECT_ACTION_PARAMS:
		return "Wrong parameters have been given to the action"
	case SIMCONNECT_EXCEPTION_GET_INPUT_EVENT_FAILED:
		return "Wrong name/hash passed to GetInputEvent"
	case SIMCONNECT_EXCEPTION_SET_INPUT_EVENT_FAILED:
		return "Wrong name/hash passed to SetInputEvent"
	default:
		return "Unknown exception type"
	}
}

// GetExceptionSeverity returns the severity level for an exception code
func GetExceptionSeverity(code SimConnectException) string {
	switch code {
	case SIMCONNECT_EXCEPTION_NONE:
		return "info"
	case SIMCONNECT_EXCEPTION_UNOPENED, SIMCONNECT_EXCEPTION_VERSION_MISMATCH:
		return "critical"
	case SIMCONNECT_EXCEPTION_TOO_MANY_GROUPS, SIMCONNECT_EXCEPTION_TOO_MANY_EVENT_NAMES,
		SIMCONNECT_EXCEPTION_TOO_MANY_MAPS, SIMCONNECT_EXCEPTION_TOO_MANY_OBJECTS,
		SIMCONNECT_EXCEPTION_TOO_MANY_REQUESTS, SIMCONNECT_EXCEPTION_ALREADY_SUBSCRIBED,
		SIMCONNECT_EXCEPTION_ALREADY_CREATED, SIMCONNECT_EXCEPTION_DUPLICATE_ID:
		return "warning"
	default:
		return "error"
	}
}

// IsException checks if a message contains an exception and returns it
func IsException(msg any) (*ExceptionData, bool) {
	if msgMap, ok := msg.(map[string]any); ok {
		if msgMap["type"] == "EXCEPTION" {
			if exception, exists := msgMap["exception"]; exists {
				if exceptionData, ok := exception.(*ExceptionData); ok {
					return exceptionData, true
				}
			}
		}
	}
	return nil, false
}

// IsCriticalException checks if an exception is critical severity
func IsCriticalException(exception *ExceptionData) bool {
	return exception.Severity == "critical"
}

// IsErrorException checks if an exception is error severity
func IsErrorException(exception *ExceptionData) bool {
	return exception.Severity == "error"
}

// IsWarningException checks if an exception is warning severity
func IsWarningException(exception *ExceptionData) bool {
	return exception.Severity == "warning"
}
