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
const (
	// Core Weather Variables (Row 1)
	TEMP_DEFINE_ID       = 1
	PRESSURE_DEFINE_ID   = 2
	WIND_SPEED_DEFINE_ID = 3
	WIND_DIR_DEFINE_ID   = 4

	// Environmental Variables (Row 2)
	VISIBILITY_DEFINE_ID      = 5
	PRECIP_RATE_DEFINE_ID     = 6
	PRECIP_STATE_DEFINE_ID    = 7
	DENSITY_ALT_DEFINE_ID     = 8
	GROUND_ALT_DEFINE_ID      = 9
	MAGVAR_DEFINE_ID          = 10
	SEA_LEVEL_PRESS_DEFINE_ID = 11
	AMBIENT_DENSITY_DEFINE_ID = 12

	// Request IDs
	TEMP_REQUEST_ID            = 101
	PRESSURE_REQUEST_ID        = 102
	WIND_SPEED_REQUEST_ID      = 103
	WIND_DIR_REQUEST_ID        = 104
	VISIBILITY_REQUEST_ID      = 105
	PRECIP_RATE_REQUEST_ID     = 106
	PRECIP_STATE_REQUEST_ID    = 107
	DENSITY_ALT_REQUEST_ID     = 108
	GROUND_ALT_REQUEST_ID      = 109
	MAGVAR_REQUEST_ID          = 110
	SEA_LEVEL_PRESS_REQUEST_ID = 111
	AMBIENT_DENSITY_REQUEST_ID = 112
)

// WeatherClient handles SimConnect communication for weather data
type WeatherClient struct {
	sdk            *client.Engine
	currentWeather models.WeatherData
	mutex          sync.RWMutex
}

// NewWeatherClient creates a new weather client
func NewWeatherClient() *WeatherClient {
	return &WeatherClient{}
}

// Connect establishes connection to SimConnect and registers variables
func (wc *WeatherClient) Connect() error {
	fmt.Println("🔗 Connecting to Microsoft Flight Simulator...")

	// Create new SimConnect client
	wc.sdk = client.New("SimWebService").(*client.Engine)

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

	fmt.Println("✅ Environmental variables registered successfully!")
	// Start periodic data requests
	fmt.Println("⏰ Starting periodic environmental monitoring (every second)...")

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

	fmt.Println("✅ Periodic environmental monitoring started!")

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
	switch data.DefineID {
	// Core Weather Variables (Row 1)
	case TEMP_DEFINE_ID:
		wc.currentWeather.Temperature = floatValue
	case PRESSURE_DEFINE_ID:
		wc.currentWeather.Pressure = floatValue
	case WIND_SPEED_DEFINE_ID:
		wc.currentWeather.WindSpeed = floatValue
	case WIND_DIR_DEFINE_ID:
		wc.currentWeather.WindDirection = floatValue

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
	}

	// Update timestamp
	wc.currentWeather.LastUpdate = time.Now().Format("15:04:05")
}
