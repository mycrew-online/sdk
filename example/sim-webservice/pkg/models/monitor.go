package models

// FlightData holds comprehensive flight monitoring information
type FlightData struct { // Core Environmental Data (Row 1)
	Temperature      float32 `json:"temperature"`      // Celsius
	SeaLevelPressure float32 `json:"seaLevelPressure"` // millibars (SEA LEVEL PRESSURE)
	WindSpeed        float32 `json:"windSpeed"`        // knots
	WindDirection    float32 `json:"windDirection"`    // degrees

	// Time & Simulation Variables (Row 1.5 - New)
	ZuluTime       string  `json:"zuluTime"`       // HH:MM:SS format
	LocalTime      string  `json:"localTime"`      // HH:MM:SS format
	SimulationTime string  `json:"simulationTime"` // HH:MM:SS format
	SimulationRate float32 `json:"simulationRate"` // multiplier (1.0x, 2.0x, etc.)

	// Environmental Conditions (Row 2)
	Visibility        float32 `json:"visibility"`        // meters
	PrecipRate        float32 `json:"precipRate"`        // millimeters of water
	PrecipState       uint32  `json:"precipState"`       // 2=None, 4=Rain, 8=Snow
	DensityAltitude   float32 `json:"densityAltitude"`   // feet
	GroundAltitude    float32 `json:"groundAltitude"`    // meters
	MagVar            float32 `json:"magVar"`            // degrees
	BarometerPressure float32 `json:"barometerPressure"` // inHg (BAROMETER PRESSURE)
	AmbientDensity    float32 `json:"ambientDensity"`    // slugs per cubic feet
	Realism           float32 `json:"realism"`           // percentage

	// Position & Navigation Data (Row 3)
	Latitude      float32 `json:"latitude"`      // degrees
	Longitude     float32 `json:"longitude"`     // degrees
	Altitude      float32 `json:"altitude"`      // feet
	GroundSpeed   float32 `json:"groundSpeed"`   // knots
	Heading       float32 `json:"heading"`       // degrees
	VerticalSpeed float32 `json:"verticalSpeed"` // feet per second

	// Airport/Navigation Info (Row 4)
	NearestAirport    string  `json:"nearestAirport"`    // airport name
	DistanceToAirport float32 `json:"distanceToAirport"` // meters
	ComFrequency      float32 `json:"comFrequency"`      // MHz
	Nav1Frequency     float32 `json:"nav1Frequency"`     // MHz
	GpsDistance       float32 `json:"gpsDistance"`       // meters
	GpsEte            float32 `json:"gpsEte"`            // seconds

	// Flight Status (Row 5)
	OnGround        uint32  `json:"onGround"`        // boolean as uint32
	OnRunway        uint32  `json:"onRunway"`        // boolean as uint32
	GpsActive       uint32  `json:"gpsActive"`       // boolean as uint32
	AutopilotMaster uint32  `json:"autopilotMaster"` // boolean as uint32
	SurfaceType     uint32  `json:"surfaceType"`     // enum
	IndicatedSpeed  float32 `json:"indicatedSpeed"`  // knots
	// Camera State
	CameraState            uint32 `json:"cameraState"`            // enum: 2=Cockpit, 3=External/Chase, 4=Drone, etc.	// Aircraft Systems
	ExternalPowerOn        uint32 `json:"externalPowerOn"`        // boolean as uint32
	ExternalPowerAvailable uint32 `json:"externalPowerAvailable"` // boolean as uint32 (EXTERNAL POWER AVAILABLE)
	// Battery Systems
	Battery1Switch  uint32  `json:"battery1Switch"`  // boolean as uint32 (0/1 off/on)
	Battery2Switch  uint32  `json:"battery2Switch"`  // boolean as uint32 (0/1 off/on)
	Battery1Voltage float32 `json:"battery1Voltage"` // volts
	Battery2Voltage float32 `json:"battery2Voltage"` // volts
	Battery1Charge  float32 `json:"battery1Charge"`  // percentage
	Battery2Charge  float32 `json:"battery2Charge"`  // percentage	// APU Systems
	ApuMasterSwitch uint32  `json:"apuMasterSwitch"` // boolean as uint32 (0/1 off/on)
	ApuStartButton  uint32  `json:"apuStartButton"`  // boolean as uint32 (0/1 off/on)
	// Aircraft Control Systems
	CanopyOpen           uint32 `json:"canopyOpen"`           // boolean as uint32 (0/1 closed/open)
	CabinNoSmokingSwitch uint32 `json:"cabinNoSmokingSwitch"` // boolean as uint32 (0/1 off/on)
	CabinSeatbeltsSwitch uint32 `json:"cabinSeatbeltsSwitch"` // boolean as uint32 (0/1 off/on)

	LastUpdate string `json:"lastUpdate"`
}

// MonitorPreset represents a monitor configuration
type MonitorPreset struct {
	Name          string  `json:"name"`
	Temperature   float32 `json:"temperature"`
	Pressure      float32 `json:"pressure"`
	WindSpeed     float32 `json:"windSpeed"`
	WindDirection float32 `json:"windDirection"`
}

// SimVarDefinition holds information about a SimConnect variable
type SimVarDefinition struct {
	DefineID  uint32
	Name      string
	Units     string
	DataType  interface{} // types.SimConnectDataType
	RequestID uint32
}
