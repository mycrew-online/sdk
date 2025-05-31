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
	SIMCONNECT_EXCEPTION_INTERNAL                                                     // Internal error
)
