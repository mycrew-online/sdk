package types

type SimConnectPeriod uint32

// SIMCONNECT_CLIENT_DATA_PERIOD defines the frequency at which client data is sent or received
const (
	SIMCONNECT_PERIOD_NEVER        SimConnectPeriod = iota // Never send data
	SIMCONNECT_PERIOD_ONCE                                 // Send data once only
	SIMCONNECT_PERIOD_VISUAL_FRAME                         // Send data every visual frame
	SIMCONNECT_PERIOD_ON_SET                               // Send data when sim variables are changed
	SIMCONNECT_PERIOD_SECOND                               // Send data once per second
)

type SimConnectDataType uint32

// SIMCONNECT_DATATYPE defines the data types used in SimConnect communications
const (
	SIMCONNECT_DATATYPE_INVALID      SimConnectDataType = iota // Invalid data type
	SIMCONNECT_DATATYPE_INT32                                  // 32-bit integer
	SIMCONNECT_DATATYPE_INT64                                  // 64-bit integer
	SIMCONNECT_DATATYPE_FLOAT32                                // 32-bit float
	SIMCONNECT_DATATYPE_FLOAT64                                // 64-bit float
	SIMCONNECT_DATATYPE_STRING8                                // 8-byte string
	SIMCONNECT_DATATYPE_STRING32                               // 32-byte string
	SIMCONNECT_DATATYPE_STRING64                               // 64-byte string
	SIMCONNECT_DATATYPE_STRING128                              // 128-byte string
	SIMCONNECT_DATATYPE_STRING256                              // 256-byte string
	SIMCONNECT_DATATYPE_STRING260                              // 260-byte string
	SIMCONNECT_DATATYPE_STRINGV                                // Variable length string
	SIMCONNECT_DATATYPE_INITPOSITION                           // Initial position
	SIMCONNECT_DATATYPE_MARKERSTATE                            // Marker state
	SIMCONNECT_DATATYPE_WAYPOINT                               // Waypoint
	SIMCONNECT_DATATYPE_LATLONALT                              // Latitude, longitude, and altitude
	SIMCONNECT_DATATYPE_XYZ                                    // XYZ coordinates
)

// SIMCONNECT_OBJECT_ID defines object identifiers
const (
	SIMCONNECT_OBJECT_ID_USER uint32 = 0 // User aircraft
)

// SIMCONNECT_DATA_REQUEST_FLAG defines data request flags
const (
	SIMCONNECT_DATA_REQUEST_FLAG_DEFAULT uint32 = 0 // Default request flags
)

// SIMCONNECT_DATA_SET_FLAG defines data set flags
const (
	SIMCONNECT_DATA_SET_FLAG_DEFAULT uint32 = 0 // Default set flags
)

// SIMCONNECT_NOTIFICATION_GROUP_ID defines notification group priorities
const (
	SIMCONNECT_GROUP_PRIORITY_HIGHEST          uint32 = 1
	SIMCONNECT_GROUP_PRIORITY_HIGHEST_MASKABLE uint32 = 10000000
	SIMCONNECT_GROUP_PRIORITY_STANDARD         uint32 = 1900000000
	SIMCONNECT_GROUP_PRIORITY_DEFAULT          uint32 = 2000000000
	SIMCONNECT_GROUP_PRIORITY_LOWEST           uint32 = 4000000000
)

// SIMCONNECT_EVENT_FLAG defines event transmission flags
const (
	SIMCONNECT_EVENT_FLAG_DEFAULT             uint32 = 0
	SIMCONNECT_EVENT_FLAG_FAST_REPEAT_TIMER   uint32 = 1
	SIMCONNECT_EVENT_FLAG_SLOW_REPEAT_TIMER   uint32 = 2
	SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY uint32 = 4
)

// ClientEventID type for client-defined event identifiers
type ClientEventID uint32

// NotificationGroupID type for notification group identifiers
type NotificationGroupID uint32
