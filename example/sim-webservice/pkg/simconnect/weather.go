package simconnect

import (
	"fmt"
	"log"
	"sync"
	"time"

	"sim-webservice/pkg/models"

	"github.com/mycrew-online/sdk/pkg/client"
	"github.com/mycrew-online/sdk/pkg/types"
)

// Constants for SimConnect variable definitions
const ( // Core Weather Variables (Row 1)
	TEMP_DEFINE_ID       = 1
	PRESSURE_DEFINE_ID   = 2
	WIND_SPEED_DEFINE_ID = 3
	WIND_DIR_DEFINE_ID   = 4

	// Time & Simulation Variables (Row 1.5 - New)
	ZULU_TIME_DEFINE_ID       = 31
	LOCAL_TIME_DEFINE_ID      = 32
	SIMULATION_TIME_DEFINE_ID = 33
	SIMULATION_RATE_DEFINE_ID = 34

	// Environmental Variables (Row 2)
	VISIBILITY_DEFINE_ID      = 5
	PRECIP_RATE_DEFINE_ID     = 6
	PRECIP_STATE_DEFINE_ID    = 7
	DENSITY_ALT_DEFINE_ID     = 8
	GROUND_ALT_DEFINE_ID      = 9
	MAGVAR_DEFINE_ID          = 10
	SEA_LEVEL_PRESS_DEFINE_ID = 11
	AMBIENT_DENSITY_DEFINE_ID = 12
	REALISM_DEFINE_ID         = 35

	// Position & Navigation Variables (Row 3)
	LATITUDE_DEFINE_ID       = 13
	LONGITUDE_DEFINE_ID      = 14
	ALTITUDE_DEFINE_ID       = 15
	GROUND_SPEED_DEFINE_ID   = 16
	HEADING_DEFINE_ID        = 17
	VERTICAL_SPEED_DEFINE_ID = 18

	// Airport/Navigation Info Variables (Row 4)
	NEAREST_AIRPORT_DEFINE_ID     = 19
	DISTANCE_TO_AIRPORT_DEFINE_ID = 20
	COM_FREQUENCY_DEFINE_ID       = 21
	NAV1_FREQUENCY_DEFINE_ID      = 22
	GPS_DISTANCE_DEFINE_ID        = 23
	GPS_ETE_DEFINE_ID             = 24
	// Flight Status Variables (Row 5)
	ON_GROUND_DEFINE_ID        = 25
	ON_RUNWAY_DEFINE_ID        = 26
	GPS_ACTIVE_DEFINE_ID       = 27
	AUTOPILOT_MASTER_DEFINE_ID = 28
	SURFACE_TYPE_DEFINE_ID     = 29
	INDICATED_SPEED_DEFINE_ID  = 30

	// Request IDs
	TEMP_REQUEST_ID                = 101
	PRESSURE_REQUEST_ID            = 102
	WIND_SPEED_REQUEST_ID          = 103
	WIND_DIR_REQUEST_ID            = 104
	ZULU_TIME_REQUEST_ID           = 131
	LOCAL_TIME_REQUEST_ID          = 132
	SIMULATION_TIME_REQUEST_ID     = 133
	SIMULATION_RATE_REQUEST_ID     = 134
	VISIBILITY_REQUEST_ID          = 105
	PRECIP_RATE_REQUEST_ID         = 106
	PRECIP_STATE_REQUEST_ID        = 107
	DENSITY_ALT_REQUEST_ID         = 108
	GROUND_ALT_REQUEST_ID          = 109
	MAGVAR_REQUEST_ID              = 110
	SEA_LEVEL_PRESS_REQUEST_ID     = 111
	AMBIENT_DENSITY_REQUEST_ID     = 112
	REALISM_REQUEST_ID             = 135
	LATITUDE_REQUEST_ID            = 113
	LONGITUDE_REQUEST_ID           = 114
	ALTITUDE_REQUEST_ID            = 115
	GROUND_SPEED_REQUEST_ID        = 116
	HEADING_REQUEST_ID             = 117
	VERTICAL_SPEED_REQUEST_ID      = 118
	NEAREST_AIRPORT_REQUEST_ID     = 119
	DISTANCE_TO_AIRPORT_REQUEST_ID = 120
	COM_FREQUENCY_REQUEST_ID       = 121
	NAV1_FREQUENCY_REQUEST_ID      = 122
	GPS_DISTANCE_REQUEST_ID        = 123
	GPS_ETE_REQUEST_ID             = 124
	ON_GROUND_REQUEST_ID           = 125
	ON_RUNWAY_REQUEST_ID           = 126
	GPS_ACTIVE_REQUEST_ID          = 127
	AUTOPILOT_MASTER_REQUEST_ID    = 128
	SURFACE_TYPE_REQUEST_ID        = 129
	INDICATED_SPEED_REQUEST_ID     = 130
)

// WeatherClient handles SimConnect communication for flight data
type WeatherClient struct {
	sdk            *client.Engine
	currentWeather models.FlightData
	mutex          sync.RWMutex
	dllPath        string // Store custom DLL path if provided
}

// NewWeatherClient creates a new weather client
func NewWeatherClient() *WeatherClient {
	return &WeatherClient{}
}

// NewWeatherClientWithDLL creates a new weather client with custom DLL path
func NewWeatherClientWithDLL(dllPath string) *WeatherClient {
	return &WeatherClient{
		dllPath: dllPath,
	}
}

// Connect establishes connection to SimConnect and registers variables
func (wc *WeatherClient) Connect() error {
	fmt.Println("🔗 Connecting to Microsoft Flight Simulator...")

	// Create new SimConnect client with custom DLL path if provided
	if wc.dllPath != "" {
		wc.sdk = client.NewWithCustomDLL("SimWebService", wc.dllPath).(*client.Engine)
	} else {
		wc.sdk = client.New("SimWebService").(*client.Engine)
	}

	// Connect to SimConnect
	if err := wc.sdk.Open(); err != nil {
		return fmt.Errorf("failed to connect to SimConnect: %v", err)
	}
	fmt.Println("✅ Connected to Microsoft Flight Simulator!")

	// Register weather variables
	fmt.Println("📝 Registering weather variables...")

	// Ambient Temperature
	if err := wc.sdk.RegisterSimVarDefinition(
		TEMP_DEFINE_ID,
		"AMBIENT TEMPERATURE",
		"Celsius",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT TEMPERATURE: %v", err)
	}

	// Ambient Pressure
	if err := wc.sdk.RegisterSimVarDefinition(
		PRESSURE_DEFINE_ID,
		"AMBIENT PRESSURE",
		"Inches of mercury",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT PRESSURE: %v", err)
	}

	// Wind Speed
	if err := wc.sdk.RegisterSimVarDefinition(
		WIND_SPEED_DEFINE_ID,
		"AMBIENT WIND VELOCITY",
		"Knots",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT WIND VELOCITY: %v", err)
	}
	// Wind Direction
	if err := wc.sdk.RegisterSimVarDefinition(
		WIND_DIR_DEFINE_ID,
		"AMBIENT WIND DIRECTION",
		"Degrees",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT WIND DIRECTION: %v", err)
	}

	// Environmental Variables (Row 2)

	// Ambient Visibility
	if err := wc.sdk.RegisterSimVarDefinition(
		VISIBILITY_DEFINE_ID,
		"AMBIENT VISIBILITY",
		"Meters",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT VISIBILITY: %v", err)
	}

	// Precipitation Rate
	if err := wc.sdk.RegisterSimVarDefinition(
		PRECIP_RATE_DEFINE_ID,
		"AMBIENT PRECIP RATE",
		"millimeters of water",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT PRECIP RATE: %v", err)
	}

	// Precipitation State
	if err := wc.sdk.RegisterSimVarDefinition(
		PRECIP_STATE_DEFINE_ID,
		"AMBIENT PRECIP STATE",
		"Mask",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT PRECIP STATE: %v", err)
	}

	// Density Altitude
	if err := wc.sdk.RegisterSimVarDefinition(
		DENSITY_ALT_DEFINE_ID,
		"DENSITY ALTITUDE",
		"ft",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register DENSITY ALTITUDE: %v", err)
	}

	// Ground Altitude
	if err := wc.sdk.RegisterSimVarDefinition(
		GROUND_ALT_DEFINE_ID,
		"GROUND ALTITUDE",
		"Meters",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register GROUND ALTITUDE: %v", err)
	}

	// Magnetic Variation
	if err := wc.sdk.RegisterSimVarDefinition(
		MAGVAR_DEFINE_ID,
		"MAGVAR",
		"Degrees",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register MAGVAR: %v", err)
	}

	// Sea Level Pressure
	if err := wc.sdk.RegisterSimVarDefinition(
		SEA_LEVEL_PRESS_DEFINE_ID,
		"SEA LEVEL PRESSURE",
		"Millibars",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register SEA LEVEL PRESSURE: %v", err)
	}
	// Ambient Density
	if err := wc.sdk.RegisterSimVarDefinition(
		AMBIENT_DENSITY_DEFINE_ID,
		"AMBIENT DENSITY",
		"Slugs per cubic feet",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT DENSITY: %v", err)
	}

	// Position & Navigation Variables (Row 3)

	// Aircraft Latitude
	if err := wc.sdk.RegisterSimVarDefinition(
		LATITUDE_DEFINE_ID,
		"PLANE LATITUDE",
		"degrees",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register PLANE LATITUDE: %v", err)
	}

	// Aircraft Longitude
	if err := wc.sdk.RegisterSimVarDefinition(
		LONGITUDE_DEFINE_ID,
		"PLANE LONGITUDE",
		"degrees",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register PLANE LONGITUDE: %v", err)
	}

	// Aircraft Altitude
	if err := wc.sdk.RegisterSimVarDefinition(
		ALTITUDE_DEFINE_ID,
		"PLANE ALTITUDE",
		"feet",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register PLANE ALTITUDE: %v", err)
	}

	// Ground Speed
	if err := wc.sdk.RegisterSimVarDefinition(
		GROUND_SPEED_DEFINE_ID,
		"GROUND VELOCITY",
		"knots",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register GROUND VELOCITY: %v", err)
	}

	// True Heading
	if err := wc.sdk.RegisterSimVarDefinition(
		HEADING_DEFINE_ID,
		"PLANE HEADING DEGREES TRUE",
		"degrees",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register PLANE HEADING DEGREES TRUE: %v", err)
	}

	// Vertical Speed
	if err := wc.sdk.RegisterSimVarDefinition(
		VERTICAL_SPEED_DEFINE_ID,
		"VERTICAL SPEED",
		"feet per second",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register VERTICAL SPEED: %v", err)
	}

	// Airport/Navigation Info Variables (Row 4)	// Nearest Airport
	if err := wc.sdk.RegisterSimVarDefinition(
		NEAREST_AIRPORT_DEFINE_ID,
		"FACILITY AIRPORT CLOSEST",
		"",
		types.SIMCONNECT_DATATYPE_STRINGV,
	); err != nil {
		return fmt.Errorf("failed to register FACILITY AIRPORT CLOSEST: %v", err)
	}

	// Distance to Airport
	if err := wc.sdk.RegisterSimVarDefinition(
		DISTANCE_TO_AIRPORT_DEFINE_ID,
		"ATC RUNWAY DISTANCE",
		"meters",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register ATC RUNWAY DISTANCE: %v", err)
	}

	// COM1 Frequency
	if err := wc.sdk.RegisterSimVarDefinition(
		COM_FREQUENCY_DEFINE_ID,
		"COM ACTIVE FREQUENCY:1",
		"MHz",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register COM ACTIVE FREQUENCY:1: %v", err)
	}

	// NAV1 Frequency
	if err := wc.sdk.RegisterSimVarDefinition(
		NAV1_FREQUENCY_DEFINE_ID,
		"NAV ACTIVE FREQUENCY:1",
		"MHz",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register NAV ACTIVE FREQUENCY:1: %v", err)
	}

	// GPS Distance to Waypoint
	if err := wc.sdk.RegisterSimVarDefinition(
		GPS_DISTANCE_DEFINE_ID,
		"GPS WP DISTANCE",
		"meters",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register GPS WP DISTANCE: %v", err)
	}

	// GPS ETE (Estimated Time Enroute)
	if err := wc.sdk.RegisterSimVarDefinition(
		GPS_ETE_DEFINE_ID,
		"GPS WP ETE",
		"seconds",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register GPS WP ETE: %v", err)
	}

	// Flight Status Variables (Row 5)

	// On Ground Status
	if err := wc.sdk.RegisterSimVarDefinition(
		ON_GROUND_DEFINE_ID,
		"SIM ON GROUND",
		"Bool",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register SIM ON GROUND: %v", err)
	}

	// On Runway Status
	if err := wc.sdk.RegisterSimVarDefinition(
		ON_RUNWAY_DEFINE_ID,
		"ON ANY RUNWAY",
		"Bool",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register ON ANY RUNWAY: %v", err)
	}

	// GPS Flight Plan Active
	if err := wc.sdk.RegisterSimVarDefinition(
		GPS_ACTIVE_DEFINE_ID,
		"GPS IS ACTIVE FLIGHT PLAN",
		"Bool",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register GPS IS ACTIVE FLIGHT PLAN: %v", err)
	}

	// Autopilot Master
	if err := wc.sdk.RegisterSimVarDefinition(
		AUTOPILOT_MASTER_DEFINE_ID,
		"AUTOPILOT MASTER",
		"Bool",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register AUTOPILOT MASTER: %v", err)
	}

	// Surface Type
	if err := wc.sdk.RegisterSimVarDefinition(
		SURFACE_TYPE_DEFINE_ID,
		"SURFACE TYPE",
		"Enum",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register SURFACE TYPE: %v", err)
	}
	// Indicated Airspeed
	if err := wc.sdk.RegisterSimVarDefinition(
		INDICATED_SPEED_DEFINE_ID,
		"AIRSPEED INDICATED",
		"knots",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AIRSPEED INDICATED: %v", err)
	}

	// Time & Simulation Variables (New Row 1.5)

	// Zulu Time
	if err := wc.sdk.RegisterSimVarDefinition(
		ZULU_TIME_DEFINE_ID,
		"ZULU TIME",
		"seconds",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register ZULU TIME: %v", err)
	}

	// Local Time
	if err := wc.sdk.RegisterSimVarDefinition(
		LOCAL_TIME_DEFINE_ID,
		"LOCAL TIME",
		"seconds",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register LOCAL TIME: %v", err)
	}

	// Simulation Time
	if err := wc.sdk.RegisterSimVarDefinition(
		SIMULATION_TIME_DEFINE_ID,
		"SIMULATION TIME",
		"seconds",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register SIMULATION TIME: %v", err)
	}

	// Simulation Rate
	if err := wc.sdk.RegisterSimVarDefinition(
		SIMULATION_RATE_DEFINE_ID,
		"SIMULATION RATE",
		"number",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register SIMULATION RATE: %v", err)
	}

	// Realism (added to Environmental Variables)
	if err := wc.sdk.RegisterSimVarDefinition(
		REALISM_DEFINE_ID,
		"REALISM",
		"percent",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		return fmt.Errorf("failed to register REALISM: %v", err)
	}

	fmt.Println("✅ Flight monitoring variables registered successfully!") // Start periodic data requests
	fmt.Println("⏰ Starting periodic flight monitoring (every second)...")

	// Core Weather Variables (Row 1)
	if err := wc.sdk.RequestSimVarDataPeriodic(TEMP_DEFINE_ID, TEMP_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start temperature monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(PRESSURE_DEFINE_ID, PRESSURE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start pressure monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(WIND_SPEED_DEFINE_ID, WIND_SPEED_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start wind speed monitoring: %v", err)
	}
	if err := wc.sdk.RequestSimVarDataPeriodic(WIND_DIR_DEFINE_ID, WIND_DIR_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start wind direction monitoring: %v", err)
	}

	// Time & Simulation Variables (Row 1.5)
	if err := wc.sdk.RequestSimVarDataPeriodic(ZULU_TIME_DEFINE_ID, ZULU_TIME_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start zulu time monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(LOCAL_TIME_DEFINE_ID, LOCAL_TIME_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start local time monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(SIMULATION_TIME_DEFINE_ID, SIMULATION_TIME_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start simulation time monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(SIMULATION_RATE_DEFINE_ID, SIMULATION_RATE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start simulation rate monitoring: %v", err)
	}

	// Environmental Variables (Row 2)
	if err := wc.sdk.RequestSimVarDataPeriodic(VISIBILITY_DEFINE_ID, VISIBILITY_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start visibility monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(PRECIP_RATE_DEFINE_ID, PRECIP_RATE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start precipitation rate monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(PRECIP_STATE_DEFINE_ID, PRECIP_STATE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start precipitation state monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(DENSITY_ALT_DEFINE_ID, DENSITY_ALT_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start density altitude monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(GROUND_ALT_DEFINE_ID, GROUND_ALT_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start ground altitude monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(MAGVAR_DEFINE_ID, MAGVAR_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start magnetic variation monitoring: %v", err)
	}
	if err := wc.sdk.RequestSimVarDataPeriodic(SEA_LEVEL_PRESS_DEFINE_ID, SEA_LEVEL_PRESS_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start sea level pressure monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(AMBIENT_DENSITY_DEFINE_ID, AMBIENT_DENSITY_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start ambient density monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(REALISM_DEFINE_ID, REALISM_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start realism monitoring: %v", err)
	}

	// Position & Navigation Variables (Row 3)
	if err := wc.sdk.RequestSimVarDataPeriodic(LATITUDE_DEFINE_ID, LATITUDE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start latitude monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(LONGITUDE_DEFINE_ID, LONGITUDE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start longitude monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(ALTITUDE_DEFINE_ID, ALTITUDE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start altitude monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(GROUND_SPEED_DEFINE_ID, GROUND_SPEED_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start ground speed monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(HEADING_DEFINE_ID, HEADING_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start heading monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(VERTICAL_SPEED_DEFINE_ID, VERTICAL_SPEED_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start vertical speed monitoring: %v", err)
	}

	// Airport/Navigation Info Variables (Row 4)
	if err := wc.sdk.RequestSimVarDataPeriodic(NEAREST_AIRPORT_DEFINE_ID, NEAREST_AIRPORT_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start nearest airport monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(DISTANCE_TO_AIRPORT_DEFINE_ID, DISTANCE_TO_AIRPORT_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start distance to airport monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(COM_FREQUENCY_DEFINE_ID, COM_FREQUENCY_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start COM frequency monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(NAV1_FREQUENCY_DEFINE_ID, NAV1_FREQUENCY_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start NAV1 frequency monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(GPS_DISTANCE_DEFINE_ID, GPS_DISTANCE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start GPS distance monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(GPS_ETE_DEFINE_ID, GPS_ETE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start GPS ETE monitoring: %v", err)
	}

	// Flight Status Variables (Row 5)
	if err := wc.sdk.RequestSimVarDataPeriodic(ON_GROUND_DEFINE_ID, ON_GROUND_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start on ground monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(ON_RUNWAY_DEFINE_ID, ON_RUNWAY_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start on runway monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(GPS_ACTIVE_DEFINE_ID, GPS_ACTIVE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start GPS active monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(AUTOPILOT_MASTER_DEFINE_ID, AUTOPILOT_MASTER_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start autopilot master monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(SURFACE_TYPE_DEFINE_ID, SURFACE_TYPE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start surface type monitoring: %v", err)
	}

	if err := wc.sdk.RequestSimVarDataPeriodic(INDICATED_SPEED_DEFINE_ID, INDICATED_SPEED_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start indicated speed monitoring: %v", err)
	}

	fmt.Println("✅ Periodic flight monitoring started!")

	// Start message processing in background
	go wc.processSimConnectMessages()

	return nil
}

// GetCurrentWeather returns the current weather data
func (wc *WeatherClient) GetCurrentWeather() models.WeatherData {
	wc.mutex.RLock()
	defer wc.mutex.RUnlock()
	return wc.currentWeather
}

// SetWeatherPreset applies a weather preset (placeholder for future implementation)
func (wc *WeatherClient) SetWeatherPreset(preset models.WeatherPreset) error {
	log.Printf("🌤️ Applying weather preset: %+v", preset)
	// TODO: Implement actual weather setting via SimConnect
	// This would require using different SimConnect APIs for weather control
	return nil
}

// Close closes the SimConnect connection
func (wc *WeatherClient) Close() error {
	if wc.sdk != nil {
		return wc.sdk.Close()
	}
	return nil
}

func (wc *WeatherClient) processSimConnectMessages() {
	messages := wc.sdk.Listen()
	if messages == nil {
		log.Fatal("❌ Failed to start listening for SimConnect messages")
	}

	for msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}

		// Only process SIMOBJECT_DATA messages
		msgType, exists := msgMap["type"]
		if !exists || msgType != "SIMOBJECT_DATA" {
			continue
		}

		// Check if we have parsed data
		parsedData, exists := msgMap["parsed_data"]
		if !exists {
			continue
		}

		// Cast to SimVarData
		simVarData, ok := parsedData.(*client.SimVarData)
		if !ok {
			continue
		}

		// Update weather data based on DefineID
		wc.updateWeatherData(simVarData)
	}
}

func (wc *WeatherClient) updateWeatherData(data *client.SimVarData) {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	// Handle different data types
	var floatValue float32
	var intValue uint32

	// Extract value based on type
	switch v := data.Value.(type) {
	case float32:
		floatValue = v
	case float64:
		floatValue = float32(v)
	case int32:
		intValue = uint32(v)
	case uint32:
		intValue = v
	default:
		return // Skip if we can't convert
	}

	// Update the appropriate field
	switch data.DefineID { // Core Weather Variables (Row 1)
	case TEMP_DEFINE_ID:
		wc.currentWeather.Temperature = floatValue
	case PRESSURE_DEFINE_ID:
		wc.currentWeather.Pressure = floatValue
	case WIND_SPEED_DEFINE_ID:
		wc.currentWeather.WindSpeed = floatValue
	case WIND_DIR_DEFINE_ID:
		wc.currentWeather.WindDirection = floatValue
	// Time & Simulation Variables (Row 1.5)
	case ZULU_TIME_DEFINE_ID:
		// Convert seconds to HH:MM:SS format
		hours := intValue / 3600
		minutes := (intValue % 3600) / 60
		seconds := intValue % 60
		wc.currentWeather.ZuluTime = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case LOCAL_TIME_DEFINE_ID:
		// Convert seconds to HH:MM:SS format
		hours := intValue / 3600
		minutes := (intValue % 3600) / 60
		seconds := intValue % 60
		wc.currentWeather.LocalTime = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case SIMULATION_TIME_DEFINE_ID:
		// Convert seconds to HH:MM:SS format
		hours := intValue / 3600
		minutes := (intValue % 3600) / 60
		seconds := intValue % 60
		wc.currentWeather.SimulationTime = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case SIMULATION_RATE_DEFINE_ID:
		wc.currentWeather.SimulationRate = float32(intValue)

	// Environmental Variables (Row 2)
	case VISIBILITY_DEFINE_ID:
		wc.currentWeather.Visibility = floatValue
	case PRECIP_RATE_DEFINE_ID:
		wc.currentWeather.PrecipRate = floatValue
	case PRECIP_STATE_DEFINE_ID:
		wc.currentWeather.PrecipState = intValue
	case DENSITY_ALT_DEFINE_ID:
		wc.currentWeather.DensityAltitude = floatValue
	case GROUND_ALT_DEFINE_ID:
		wc.currentWeather.GroundAltitude = floatValue
	case MAGVAR_DEFINE_ID:
		wc.currentWeather.MagVar = floatValue
	case SEA_LEVEL_PRESS_DEFINE_ID:
		wc.currentWeather.SeaLevelPress = floatValue
	case AMBIENT_DENSITY_DEFINE_ID:
		wc.currentWeather.AmbientDensity = floatValue
	case REALISM_DEFINE_ID:
		wc.currentWeather.Realism = float32(intValue)

	// Position & Navigation Variables (Row 3)
	case LATITUDE_DEFINE_ID:
		wc.currentWeather.Latitude = floatValue
	case LONGITUDE_DEFINE_ID:
		wc.currentWeather.Longitude = floatValue
	case ALTITUDE_DEFINE_ID:
		wc.currentWeather.Altitude = floatValue
	case GROUND_SPEED_DEFINE_ID:
		wc.currentWeather.GroundSpeed = floatValue
	case HEADING_DEFINE_ID:
		wc.currentWeather.Heading = floatValue
	case VERTICAL_SPEED_DEFINE_ID:
		wc.currentWeather.VerticalSpeed = floatValue
	// Airport/Navigation Info Variables (Row 4)
	case NEAREST_AIRPORT_DEFINE_ID:
		// For string values, we need special handling
		if strVal, ok := data.Value.(string); ok {
			wc.currentWeather.NearestAirport = strVal
		}
	case DISTANCE_TO_AIRPORT_DEFINE_ID:
		wc.currentWeather.DistanceToAirport = floatValue
	case COM_FREQUENCY_DEFINE_ID:
		wc.currentWeather.ComFrequency = floatValue
	case NAV1_FREQUENCY_DEFINE_ID:
		wc.currentWeather.Nav1Frequency = floatValue
	case GPS_DISTANCE_DEFINE_ID:
		wc.currentWeather.GpsDistance = floatValue
	case GPS_ETE_DEFINE_ID:
		wc.currentWeather.GpsEte = floatValue
	// Flight Status Variables (Row 5)
	case ON_GROUND_DEFINE_ID:
		if intValue != 0 {
			wc.currentWeather.OnGround = 1
		} else {
			wc.currentWeather.OnGround = 0
		}
	case ON_RUNWAY_DEFINE_ID:
		if intValue != 0 {
			wc.currentWeather.OnRunway = 1
		} else {
			wc.currentWeather.OnRunway = 0
		}
	case GPS_ACTIVE_DEFINE_ID:
		if intValue != 0 {
			wc.currentWeather.GpsActive = 1
		} else {
			wc.currentWeather.GpsActive = 0
		}
	case AUTOPILOT_MASTER_DEFINE_ID:
		if intValue != 0 {
			wc.currentWeather.AutopilotMaster = 1
		} else {
			wc.currentWeather.AutopilotMaster = 0
		}
	case SURFACE_TYPE_DEFINE_ID:
		wc.currentWeather.SurfaceType = intValue
	case INDICATED_SPEED_DEFINE_ID:
		wc.currentWeather.IndicatedSpeed = floatValue
	}

	// Update timestamp
	wc.currentWeather.LastUpdate = time.Now().Format("15:04:05")
}
