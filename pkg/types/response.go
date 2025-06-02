package types

// SIMCONNECT_RECV is the base structure for all SimConnect messages
type SIMCONNECT_RECV struct {
	DwSize    uint32           // Size of the structure
	DwVersion uint32           // Version of SimConnect, matches SDK
	DwID      SimConnectRecvID // Message ID
}

// SIMCONNECT_RECV_SIMOBJECT_DATA represents SimObject data received from SimConnect
type SIMCONNECT_RECV_SIMOBJECT_DATA struct {
	SIMCONNECT_RECV        // Inherits from base structure
	DwRequestID     uint32 // ID of the client defined request
	DwObjectID      uint32 // ID of the client defined object
	DwDefineID      uint32 // ID of the client defined data definition
	DwFlags         uint32 // Flags that were set for this data request
	DwEntryNumber   uint32 // Index number of this object (1-based)
	DwOutOf         uint32 // Total number of objects being returned
	DwDefineCount   uint32 // Number of 8-byte elements in the data array
	DwData          uint32 // Start of data array (actual data follows)
}

// SIMCONNECT_RECV_EXCEPTION represents exception information from SimConnect
type SIMCONNECT_RECV_EXCEPTION struct {
	SIMCONNECT_RECV        // Inherits from base structure
	DwException     uint32 // Exception code
	DwSendID        uint32 // ID of the packet that caused the exception
	DwIndex         uint32 // Index number for some exceptions
}

// ExceptionData represents a parsed SimConnect exception for channel messages
type ExceptionData struct {
	ExceptionCode SimConnectException `json:"exception_code"` // Numeric exception code
	ExceptionName string              `json:"exception_name"` // Human-readable exception name
	Description   string              `json:"description"`    // Detailed description of the exception
	SendID        uint32              `json:"send_id"`        // ID of the packet that caused the exception
	Index         uint32              `json:"index"`          // Index number for some exceptions
	Severity      string              `json:"severity"`       // "warning", "error", "critical"
}

// SIMCONNECT_RECV_EVENT represents event information received from SimConnect
// Field names match official SimConnect documentation
type SIMCONNECT_RECV_EVENT struct {
	SIMCONNECT_RECV        // Inherits from base structure
	UGroupID        uint32 // ID of the client defined group (uGroupID in official docs)
	UEventID        uint32 // ID of the client defined event (uEventID in official docs)
	DwData          uint32 // Event data - usually zero, but some events require additional qualification
}

// EventData represents a parsed SimConnect event for channel messages
type EventData struct {
	GroupID   uint32 `json:"group_id"`   // ID of the group the event belongs to
	EventID   uint32 `json:"event_id"`   // ID of the event
	EventData uint32 `json:"event_data"` // Event-specific data value
	EventName string `json:"event_name"` // Human-readable event name (if available)
	EventType string `json:"event_type"` // Type of event: "system", "client", or "unknown"
}

type SimConnectRecvID uint32

// SIMCONNECT_RECV_ID defines all possible message types that can be received from SimConnect
const (
	SIMCONNECT_RECV_ID_NULL                             SimConnectRecvID = iota // Null message
	SIMCONNECT_RECV_ID_EXCEPTION                                                // Exception information
	SIMCONNECT_RECV_ID_OPEN                                                     // Connection established
	SIMCONNECT_RECV_ID_QUIT                                                     // Connection closed
	SIMCONNECT_RECV_ID_EVENT                                                    // Event information
	SIMCONNECT_RECV_ID_EVENT_OBJECT_ADDREMOVE                                   // Object added or removed
	SIMCONNECT_RECV_ID_EVENT_FILENAME                                           // Filename event
	SIMCONNECT_RECV_ID_EVENT_FRAME                                              // Frame event
	SIMCONNECT_RECV_ID_SIMOBJECT_DATA                                           // SimObject data
	SIMCONNECT_RECV_ID_SIMOBJECT_DATA_BYTYPE                                    // SimObject data by type
	SIMCONNECT_RECV_ID_WEATHER_OBSERVATION                                      // Weather observation
	SIMCONNECT_RECV_ID_CLOUD_STATE                                              // Cloud state
	SIMCONNECT_RECV_ID_ASSIGNED_OBJECT_ID                                       // Assigned object ID
	SIMCONNECT_RECV_ID_RESERVED_KEY                                             // Reserved key
	SIMCONNECT_RECV_ID_CUSTOM_ACTION                                            // Custom action
	SIMCONNECT_RECV_ID_SYSTEM_STATE                                             // System state
	SIMCONNECT_RECV_ID_CLIENT_DATA                                              // Client data
	SIMCONNECT_RECV_ID_EVENT_WEATHER_MODE                                       // Weather mode event
	SIMCONNECT_RECV_ID_AIRPORT_LIST                                             // Airport list
	SIMCONNECT_RECV_ID_VOR_LIST                                                 // VOR list
	SIMCONNECT_RECV_ID_NDB_LIST                                                 // NDB list
	SIMCONNECT_RECV_ID_WAYPOINT_LIST                                            // Waypoint list
	SIMCONNECT_RECV_ID_EVENT_MULTIPLAYER_SERVER_STARTED                         // Multiplayer server started
	SIMCONNECT_RECV_ID_EVENT_MULTIPLAYER_CLIENT_STARTED                         // Multiplayer client started
	SIMCONNECT_RECV_ID_EVENT_MULTIPLAYER_SESSION_ENDED                          // Multiplayer session ended
	SIMCONNECT_RECV_ID_EVENT_RACE_END                                           // Race end event
	SIMCONNECT_RECV_ID_EVENT_RACE_LAP                                           // Race lap event
	SIMCONNECT_RECV_ID_PICK                                                     // Pick event
	SIMCONNECT_RECV_ID_EVENT_EX1                                                // Extended event 1
	SIMCONNECT_RECV_ID_FACILITY_DATA                                            // Facility data
	SIMCONNECT_RECV_ID_FACILITY_DATA_END                                        // Facility data end
	SIMCONNECT_RECV_ID_FACILITY_MINIMAL_LIST                                    // Facility minimal list
	SIMCONNECT_RECV_ID_JETWAY_DATA                                              // Jetway data
	SIMCONNECT_RECV_ID_CONTROLLERS_LIST                                         // Controllers list
	SIMCONNECT_RECV_ID_ACTION_CALLBACK                                          // Action callback
	SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENTS                                   // Enumerate input events
	SIMCONNECT_RECV_ID_GET_INPUT_EVENT                                          // Get input event
	SIMCONNECT_RECV_ID_SUBSCRIBE_INPUT_EVENT                                    // Subscribe to input event
	SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENT_PARAMS                             // Enumerate input event parameters
)
