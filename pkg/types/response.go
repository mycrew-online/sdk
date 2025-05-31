package types

// SIMCONNECT_RECV is the base structure for all SimConnect messages
type SIMCONNECT_RECV struct {
	DwSize    uint32           // Size of the structure
	DwVersion uint32           // Version of SimConnect, matches SDK
	DwID      SimConnectRecvID // Message ID
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
