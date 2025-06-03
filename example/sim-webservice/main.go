package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mycrew-online/sdk/pkg/client"
	"github.com/mycrew-online/sdk/pkg/types"
)

// WeatherData holds the current weather information
type WeatherData struct {
	Temperature   float32 `json:"temperature"`   // Celsius
	Pressure      float32 `json:"pressure"`      // inHg
	WindSpeed     float32 `json:"windSpeed"`     // knots
	WindDirection float32 `json:"windDirection"` // degrees
	LastUpdate    string  `json:"lastUpdate"`
}

// WeatherPreset represents a weather configuration
type WeatherPreset struct {
	Name          string  `json:"name"`
	Temperature   float32 `json:"temperature"`
	Pressure      float32 `json:"pressure"`
	WindSpeed     float32 `json:"windSpeed"`
	WindDirection float32 `json:"windDirection"`
}

// Constants for SimConnect variable definitions
const (
	TEMP_DEFINE_ID       = 1
	PRESSURE_DEFINE_ID   = 2
	WIND_SPEED_DEFINE_ID = 3
	WIND_DIR_DEFINE_ID   = 4

	TEMP_REQUEST_ID       = 101
	PRESSURE_REQUEST_ID   = 102
	WIND_SPEED_REQUEST_ID = 103
	WIND_DIR_REQUEST_ID   = 104
)

var (
	currentWeather WeatherData
	weatherMutex   sync.RWMutex
	sdk            *client.Engine
)

func main() {
	fmt.Println("🌤️  Weather WebService Demo - Starting...")
	fmt.Println("   Real-time weather monitoring for Microsoft Flight Simulator")
	fmt.Println("   Open your browser to http://localhost:8080")
	fmt.Println()

	// Initialize SimConnect
	if err := initSimConnect(); err != nil {
		log.Fatalf("❌ Failed to initialize SimConnect: %v", err)
	}

	// Set up HTTP routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/weather", handleWeatherAPI)
	http.HandleFunc("/api/weather/preset", handleWeatherPreset)

	// Start the web server
	fmt.Println("🚀 Starting web server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Failed to start web server: %v", err)
	}
}

func initSimConnect() error {
	fmt.Println("🔗 Connecting to Microsoft Flight Simulator...")

	// Create new SimConnect client
	sdk = client.New("WeatherWebService").(*client.Engine)

	// Connect to SimConnect
	if err := sdk.Open(); err != nil {
		return fmt.Errorf("failed to connect to SimConnect: %v", err)
	}
	fmt.Println("✅ Connected to Microsoft Flight Simulator!")

	// Register weather variables
	fmt.Println("📝 Registering weather variables...")

	// Ambient Temperature
	if err := sdk.RegisterSimVarDefinition(
		TEMP_DEFINE_ID,
		"AMBIENT TEMPERATURE",
		"Celsius",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT TEMPERATURE: %v", err)
	}

	// Ambient Pressure
	if err := sdk.RegisterSimVarDefinition(
		PRESSURE_DEFINE_ID,
		"AMBIENT PRESSURE",
		"Inches of mercury",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT PRESSURE: %v", err)
	}

	// Wind Speed
	if err := sdk.RegisterSimVarDefinition(
		WIND_SPEED_DEFINE_ID,
		"AMBIENT WIND VELOCITY",
		"Knots",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT WIND VELOCITY: %v", err)
	}

	// Wind Direction
	if err := sdk.RegisterSimVarDefinition(
		WIND_DIR_DEFINE_ID,
		"AMBIENT WIND DIRECTION",
		"Degrees",
		types.SIMCONNECT_DATATYPE_FLOAT32,
	); err != nil {
		return fmt.Errorf("failed to register AMBIENT WIND DIRECTION: %v", err)
	}

	fmt.Println("✅ Weather variables registered successfully!")

	// Start periodic data requests
	fmt.Println("⏰ Starting periodic weather monitoring (every second)...")
	if err := sdk.RequestSimVarDataPeriodic(TEMP_DEFINE_ID, TEMP_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start temperature monitoring: %v", err)
	}

	if err := sdk.RequestSimVarDataPeriodic(PRESSURE_DEFINE_ID, PRESSURE_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start pressure monitoring: %v", err)
	}

	if err := sdk.RequestSimVarDataPeriodic(WIND_SPEED_DEFINE_ID, WIND_SPEED_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start wind speed monitoring: %v", err)
	}

	if err := sdk.RequestSimVarDataPeriodic(WIND_DIR_DEFINE_ID, WIND_DIR_REQUEST_ID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
		return fmt.Errorf("failed to start wind direction monitoring: %v", err)
	}

	fmt.Println("✅ Periodic weather monitoring started!")

	// Start message processing in background
	go processSimConnectMessages()

	return nil
}

func processSimConnectMessages() {
	messages := sdk.Listen()
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
		updateWeatherData(simVarData)
	}
}

func updateWeatherData(data *client.SimVarData) {
	weatherMutex.Lock()
	defer weatherMutex.Unlock()

	// Extract value as float32
	value, ok := data.Value.(float32)
	if !ok {
		// Try float64 conversion if direct float32 fails
		if val64, ok := data.Value.(float64); ok {
			value = float32(val64)
		} else {
			return // Skip if we can't convert to float
		}
	}

	// Update the appropriate field
	switch data.DefineID {
	case TEMP_DEFINE_ID:
		currentWeather.Temperature = value
	case PRESSURE_DEFINE_ID:
		currentWeather.Pressure = value
	case WIND_SPEED_DEFINE_ID:
		currentWeather.WindSpeed = value
	case WIND_DIR_DEFINE_ID:
		currentWeather.WindDirection = value
	}

	// Update timestamp
	currentWeather.LastUpdate = time.Now().Format("15:04:05")
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MSFS Weather Monitor</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            theme: {
                extend: {
                    colors: {
                        flight: {
                            50: '#f0f9ff',
                            500: '#0ea5e9',
                            700: '#0369a1',
                            900: '#0c4a6e'
                        }
                    }
                }
            }
        }
    </script>
</head>
<body class="bg-gradient-to-b from-sky-100 to-blue-200 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <!-- Header -->
        <div class="text-center mb-8">
            <h1 class="text-4xl font-bold text-flight-900 mb-2">🌤️ MSFS Weather Monitor</h1>
            <p class="text-flight-700">Real-time weather conditions from Microsoft Flight Simulator</p>
        </div>

        <!-- Weather Dashboard -->
        <div class="max-w-4xl mx-auto">
            <!-- Status Card -->
            <div class="bg-white rounded-lg shadow-lg p-6 mb-6">
                <div class="flex items-center justify-between">
                    <div class="flex items-center">
                        <div class="w-3 h-3 bg-green-500 rounded-full mr-3"></div>
                        <span class="text-sm font-medium text-gray-700">Connected to MSFS</span>
                    </div>
                    <div class="text-sm text-gray-500">
                        Last Update: <span id="lastUpdate">--:--:--</span>
                    </div>
                </div>
            </div>

            <!-- Weather Cards Grid -->
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <!-- Temperature Card -->
                <div class="bg-white rounded-lg shadow-lg p-6">
                    <div class="flex items-center justify-between mb-2">
                        <h3 class="text-lg font-semibold text-gray-700">Temperature</h3>
                        <span class="text-2xl">🌡️</span>
                    </div>
                    <div class="text-3xl font-bold text-flight-900">
                        <span id="temperature">--</span><span class="text-xl text-gray-500">°C</span>
                    </div>
                </div>

                <!-- Pressure Card -->
                <div class="bg-white rounded-lg shadow-lg p-6">
                    <div class="flex items-center justify-between mb-2">
                        <h3 class="text-lg font-semibold text-gray-700">Pressure</h3>
                        <span class="text-2xl">📊</span>
                    </div>
                    <div class="text-3xl font-bold text-flight-900">
                        <span id="pressure">--</span><span class="text-xl text-gray-500">inHg</span>
                    </div>
                </div>

                <!-- Wind Speed Card -->
                <div class="bg-white rounded-lg shadow-lg p-6">
                    <div class="flex items-center justify-between mb-2">
                        <h3 class="text-lg font-semibold text-gray-700">Wind Speed</h3>
                        <span class="text-2xl">💨</span>
                    </div>
                    <div class="text-3xl font-bold text-flight-900">
                        <span id="windSpeed">--</span><span class="text-xl text-gray-500">kts</span>
                    </div>
                </div>

                <!-- Wind Direction Card -->
                <div class="bg-white rounded-lg shadow-lg p-6">
                    <div class="flex items-center justify-between mb-2">
                        <h3 class="text-lg font-semibold text-gray-700">Wind Direction</h3>
                        <span class="text-2xl">🧭</span>
                    </div>
                    <div class="text-3xl font-bold text-flight-900">
                        <span id="windDirection">--</span><span class="text-xl text-gray-500">°</span>
                    </div>
                </div>
            </div>

            <!-- Info Panel -->
            <div class="bg-white rounded-lg shadow-lg p-6 mt-6">
                <h3 class="text-lg font-semibold text-gray-700 mb-3">📋 Status</h3>
                <div class="text-sm text-gray-600 space-y-1">
                    <div>• Monitoring 4 basic weather variables</div>
                    <div>• Updates every second from SimConnect</div>
                    <div>• Phase 1: Weather monitoring (read-only)</div>
                    <div class="text-gray-400">• Phase 2: Weather controls (coming soon)</div>
                </div>
            </div>

            <!-- Preset Weather Controls -->
            <div class="bg-white rounded-lg shadow-lg p-6 mt-6">
                <h3 class="text-lg font-semibold text-gray-700 mb-3">🎛️ Weather Controls</h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <!-- Preset Buttons -->
                    <div class="flex flex-col space-y-4">
                        <!-- Preset 1: Clear Sky -->
                        <button onclick="setWeatherPreset('Clear Sky')" class="preset-button bg-flight-500 text-white rounded-lg shadow-md px-4 py-2 transition-all duration-200 hover:bg-flight-600">
                            Clear Sky
                        </button>

                        <!-- Preset 2: Partly Cloudy -->
                        <button onclick="setWeatherPreset('Partly Cloudy')" class="preset-button bg-flight-500 text-white rounded-lg shadow-md px-4 py-2 transition-all duration-200 hover:bg-flight-600">
                            Partly Cloudy
                        </button>

                        <!-- Preset 3: Overcast -->
                        <button onclick="setWeatherPreset('Overcast')" class="preset-button bg-flight-500 text-white rounded-lg shadow-md px-4 py-2 transition-all duration-200 hover:bg-flight-600">
                            Overcast
                        </button>

                        <!-- Preset 4: Rain -->
                        <button onclick="setWeatherPreset('Rain')" class="preset-button bg-flight-500 text-white rounded-lg shadow-md px-4 py-2 transition-all duration-200 hover:bg-flight-600">
                            Rain
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Fetch weather data from API
        async function updateWeather() {
            try {
                const response = await fetch('/api/weather');
                const data = await response.json();
                
                // Update display values
                document.getElementById('temperature').textContent = data.temperature.toFixed(1);
                document.getElementById('pressure').textContent = data.pressure.toFixed(2);
                document.getElementById('windSpeed').textContent = data.windSpeed.toFixed(1);
                document.getElementById('windDirection').textContent = Math.round(data.windDirection);
                document.getElementById('lastUpdate').textContent = data.lastUpdate;
                
            } catch (error) {
                console.error('Failed to fetch weather data:', error);
            }
        }

        // Set weather preset
        async function setWeatherPreset(presetName) {
            const presets = {
                "Clear Sky": { temperature: 20, pressure: 29.92, windSpeed: 5, windDirection: 270 },
                "Partly Cloudy": { temperature: 15, pressure: 29.85, windSpeed: 10, windDirection: 180 },
                "Overcast": { temperature: 10, pressure: 29.80, windSpeed: 15, windDirection: 90 },
                "Rain": { temperature: 5, pressure: 29.70, windSpeed: 20, windDirection: 0 }
            };

            const preset = presets[presetName];
            if (!preset) return;

            try {
                const response = await fetch('/api/weather/preset', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(preset)
                });

                if (response.ok) {
                    console.log("Weather preset " + presetName + " applied.");
                    updateWeather(); // Refresh weather data
                } else {
                    console.error('Failed to apply weather preset:', response.statusText);
                }
            } catch (error) {
                console.error('Error applying weather preset:', error);
            }
        }

        // Update weather data every 2 seconds
        updateWeather(); // Initial load
        setInterval(updateWeather, 2000);
    </script>
</body>
</html>`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

func handleWeatherAPI(w http.ResponseWriter, r *http.Request) {
	weatherMutex.RLock()
	weather := currentWeather
	weatherMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weather)
}

func handleWeatherPreset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var preset WeatherPreset
	if err := json.NewDecoder(r.Body).Decode(&preset); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log the received preset
	log.Printf("Received weather preset: %+v\n", preset)

	// TODO: Apply the weather preset using SimConnect

	w.WriteHeader(http.StatusNoContent)
}
