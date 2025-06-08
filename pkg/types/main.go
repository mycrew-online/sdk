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

// Complex SimConnect data structure definitions
// These correspond to the SIMCONNECT_DATATYPE_* structure types

// InitPosition represents SIMCONNECT_DATA_INITPOSITION structure
type InitPosition struct {
	Latitude  float64 `json:"latitude"`  // Latitude in degrees
	Longitude float64 `json:"longitude"` // Longitude in degrees
	Altitude  float64 `json:"altitude"`  // Altitude in feet
	Pitch     float64 `json:"pitch"`     // Pitch in degrees
	Bank      float64 `json:"bank"`      // Bank in degrees
	Heading   float64 `json:"heading"`   // Heading in degrees
	OnGround  uint32  `json:"on_ground"` // 1 if on ground, 0 if airborne
	Airspeed  uint32  `json:"airspeed"`  // Indicated airspeed in knots
}

// MarkerState represents SIMCONNECT_DATA_MARKERSTATE structure
type MarkerState struct {
	Name      [64]byte `json:"name"`      // Marker name (null-terminated string)
	Latitude  float64  `json:"latitude"`  // Latitude in degrees
	Longitude float64  `json:"longitude"` // Longitude in degrees
	Altitude  float64  `json:"altitude"`  // Altitude in feet
	Flags     uint32   `json:"flags"`     // Marker flags
	Heading   float64  `json:"heading"`   // Heading in degrees
	Speed     float64  `json:"speed"`     // Speed in knots
	Bank      float64  `json:"bank"`      // Bank angle in degrees
	Pitch     float64  `json:"pitch"`     // Pitch angle in degrees
}

// Waypoint represents SIMCONNECT_DATA_WAYPOINT structure
type Waypoint struct {
	Latitude  float64 `json:"latitude"`  // Latitude in degrees
	Longitude float64 `json:"longitude"` // Longitude in degrees
	Altitude  float64 `json:"altitude"`  // Altitude in feet
	Flags     uint32  `json:"flags"`     // Waypoint flags
	Speed     float64 `json:"speed"`     // Speed in knots
	Throttle  float64 `json:"throttle"`  // Throttle percentage (0.0-1.0)
}

// LatLonAlt represents SIMCONNECT_DATA_LATLONALT structure
type LatLonAlt struct {
	Latitude  float64 `json:"latitude"`  // Latitude in degrees
	Longitude float64 `json:"longitude"` // Longitude in degrees
	Altitude  float64 `json:"altitude"`  // Altitude in feet
}

// XYZ represents SIMCONNECT_DATA_XYZ structure
type XYZ struct {
	X float64 `json:"x"` // X coordinate
	Y float64 `json:"y"` // Y coordinate
	Z float64 `json:"z"` // Z coordinate
}
