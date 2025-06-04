# Flight Monitor WebService Demo

A comprehensive real-time web interface for monitoring and controlling Microsoft Flight Simulator flight data using the SimConnect SDK.

![image (19)](https://github.com/user-attachments/assets/d131d57b-976b-40a7-a866-01b3a73b24dd)

## Features

- **Real-time Flight Monitoring**: View current flight conditions with 36+ variables including environmental data, position, navigation, aircraft status, and simulation time
- **Modern Web Interface**: Responsive UI built with Tailwind CSS featuring dark/light mode toggle
- **Camera Control**: Switch between different camera views (Cockpit, External, Drone, etc.) directly from the web interface
- **System Events Monitoring**: Track simulator state, pause status, aircraft loading, and flight status
- **Live Updates**: Flight data updates automatically every second with real-time timestamps
- **Custom DLL Support**: Specify custom SimConnect.dll path via command line for different MSFS installations
- **Comprehensive Flight Data**: 36 variables across 6 categories:
  - Environmental conditions (temperature, pressure, wind, precipitation)
  - Time & simulation data (Zulu time, local time, simulation rate)
  - Position & navigation (GPS, altitude, heading, speed)
  - Airport & frequency information (nearest airport, COM/NAV frequencies)
  - Flight status (ground state, autopilot, surface type)
  - Camera state and system events

## Getting Started

### Prerequisites

1. Microsoft Flight Simulator must be running
2. An aircraft must be loaded  
3. SimConnect SDK must be installed

### Running the Demo

#### Default Usage (Standard DLL Path)
```powershell
# Navigate to this directory
cd example\sim-webservice

# Build and run with default settings
go run cmd\server\main.go
```

#### Custom DLL Path
```powershell
# Use custom SimConnect.dll location
go run cmd\server\main.go -dll "C:\Custom\Path\SimConnect.dll"

# Example for MSFS 2020
go run cmd\server\main.go -dll "C:\MSFS SDK\SimConnect SDK\lib\SimConnect.dll"

# Example for custom installation
go run cmd\server\main.go -dll "D:\FlightSim\SDK\SimConnect.dll"
```

#### Build and Run
```powershell
# Build executable
go build -o flight-monitor.exe cmd\server\main.go

# Run with default DLL
.\flight-monitor.exe

# Run with custom DLL
.\flight-monitor.exe -dll "C:\Custom\Path\SimConnect.dll"
```

### Web Interface Features

Once the server is running, open `http://localhost:8080` in your browser to access:

- **Flight Dashboard**: Real-time monitoring grid showing all flight variables
- **Dark/Light Mode Toggle**: Switch between dark and light themes (top-right corner)
- **Camera Control**: Toggle camera views with dedicated button (next to theme toggle)
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Live Data Updates**: All data refreshes every second with timestamp display

### Command Line Options

- `-dll <path>`: Specify custom path to SimConnect.dll
  - If not provided, uses default path: `C:/MSFS 2024 SDK/SimConnect SDK/lib/SimConnect.dll`
  - Useful for custom installations or older MSFS versions

## Troubleshooting

### SimConnect DLL Issues

If you encounter DLL loading errors:

1. **Default Path Issues**: Use the `-dll` flag to specify the correct path
   ```powershell   # For MSFS 2020
   .\flight-monitor.exe -dll "C:\MSFS SDK\SimConnect SDK\lib\SimConnect.dll"
   
   # For custom installation
   .\flight-monitor.exe -dll "D:\YourPath\SimConnect.dll"
   ```

2. **Find Your SimConnect.dll**: Common locations:
   - MSFS 2024: `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`
   - MSFS 2020: `C:\MSFS SDK\SimConnect SDK\lib\SimConnect.dll`
   - Windows: `C:\Windows\System32\SimConnect.dll`

3. **Connection Failed**: Ensure:
   - Microsoft Flight Simulator is running
   - An aircraft is loaded (not in main menu)
   - SimConnect is enabled in simulator settings

### Common Error Messages

- `Failed to connect to SimConnect`: MSFS not running or aircraft not loaded
- `DLL load failed`: Wrong DLL path, use `-dll` flag with correct path
- `Port 8080 already in use`: Another application is using the port

## Current Implementation

**Complete Flight Monitoring Suite with 36+ Variables**:

**Row 1 - Core Environmental Data**:
- Ambient Temperature (°C)
- Barometric Pressure (inHg) 
- Wind Speed (knots)
- Wind Direction (degrees)

**Row 1.5 - Time & Simulation Data (NEW)**:
- Zulu Time (UTC)
- Local Time
- Simulation Time
- Simulation Rate (speed multiplier)

**Row 2 - Environmental Conditions**:
- Visibility (meters)
- Precipitation Rate (mm/h)
- Precipitation State (None/Rain/Snow)
- Density Altitude (ft)
- Ground Altitude (m)
- Magnetic Variation (degrees)
- Sea Level Pressure (mb)
- Air Density (slugs/ft³)
- Realism Setting (%)

**Row 3 - Position & Navigation**:
- Aircraft Latitude/Longitude (degrees)
- Altitude (feet)
- Ground Speed (knots)
- True Heading (degrees)
- Vertical Speed (ft/sec)

**Row 4 - Airport/Navigation Info**:
- Nearest Airport ID
- Distance to Airport (m)
- COM1 Frequency (MHz)
- NAV1 Frequency (MHz)
- GPS Distance to Waypoint (m)
- GPS ETE - Estimated Time Enroute (seconds)

**Row 5 - Flight Status**:
- On Ground Status (boolean)
- On Runway Status (boolean)
- GPS Flight Plan Active (boolean)
- Autopilot Master (boolean)
- Surface Type (enum)
- Indicated Airspeed (knots)

**Additional Features**:
- **Camera State Control**: Monitor and control camera views (Cockpit, External, Drone, etc.)
- **System Events**: Real-time monitoring of simulator state, pause status, aircraft/flight loading
- **Interactive Controls**: Camera view switching via web interface

## API Endpoints

The webservice provides several HTTP endpoints for interaction:

### Flight Data API
- **GET `/api/monitor`**: Returns current flight data as JSON
  ```json
  {
    "temperature": 15.2,
    "pressure": 29.92,
    "windSpeed": 8.5,
    "windDirection": 270.0,
    "zuluTime": "14:30:25",
    "localTime": "09:30:25",
    "simulationTime": "14:30:25",
    "simulationRate": 1.0,
    // ... all other flight variables
    "lastUpdate": "2025-06-04T14:30:25Z"
  }
  ```

### Camera Control API
- **POST `/api/camera`**: Set camera state
  ```json
  {
    "state": 3
  }
  ```
  Valid camera states:
  - `2`: Cockpit view
  - `3`: External/Chase view  
  - `4`: Drone camera
  - `5-10`: Other camera modes

### System Events API
- **GET `/api/system`**: Returns system events and simulator status
  ```json
  {
    "simRunning": true,
    "simPaused": false,
    "aircraftLoaded": true,
    "flightLoaded": true,
    "lastEventName": "Simulator Running",
    "lastEventTime": "2025-06-04T14:30:25Z"
  }
  ```

### Static Files
- **GET `/static/*`**: Serves CSS, JavaScript, and other static assets

## Technical Details

### Architecture
- **Port**: 8080 (HTTP server)
- **Update Frequency**: 1 second for all flight variables (SimConnect backend) and 1 second frontend polling
- **Data Variables**: 36+ flight simulation variables across 6 categories
- **Frontend**: Vanilla HTML/CSS with Tailwind CSS framework
- **Backend**: Go with SimConnect SDK integration
- **Real-time Updates**: JavaScript polling for live data streaming
- **Concurrency**: Single `Listen()` call with goroutine-based message processing

### Performance Characteristics
- **Memory Usage**: Low footprint with efficient data structures
- **CPU Usage**: Minimal overhead with optimized SimConnect integration
- **Network Traffic**: JSON payloads typically 2-4KB per update
- **Latency**: Sub-100ms response times for API calls
- **Concurrent Connections**: Supports multiple browser sessions

### File Structure
```
sim-webservice/
├── cmd/server/main.go          # Application entry point
├── pkg/
│   ├── handlers/monitor.go     # HTTP request handlers
│   ├── models/monitor.go       # Data structures
│   └── simconnect/
│       ├── monitor.go          # Core SimConnect integration
│       ├── camera.go           # Camera control functionality
│       ├── system_events.go    # System event handling
│       └── system.go           # System utilities
├── templates/index.html        # Web interface template
├── static/
│   ├── css/styles.css         # Custom CSS styles
│   └── js/
│       ├── main.js            # Main JavaScript functionality
│       └── system.js          # System event handling
└── README.md                   # This documentation
```

## Implementation Notes

### SimConnect Integration
The application uses the advanced SimConnect Go SDK with proper adherence to the **single `Listen()` call per client** pattern described in the [Advanced Usage Guide](../../docs/ADVANCED_USAGE.md). Key implementation details:

- **Single Connection**: One SimConnect client instance for all data
- **Efficient Processing**: Goroutine-based message handling with proper synchronization
- **Error Resilience**: Graceful handling of connection issues and simulator restarts
- **Resource Management**: Proper cleanup of periodic requests on shutdown

### Data Processing Pipeline
1. **Registration Phase**: All 36+ variables registered with SimConnect on startup
2. **Periodic Requests**: Each variable requested at 1-second intervals
3. **Message Processing**: Incoming data processed in dedicated goroutine
4. **Data Storage**: Thread-safe storage with read/write locks
5. **API Serving**: HTTP handlers serve current data with minimal latency

### Web Interface Features
- **Responsive Grid Layout**: 6-row layout optimized for different screen sizes
- **Dark/Light Theme**: Persistent user preference with system detection
- **Live Updates**: JavaScript polling for real-time data refresh
- **Camera Controls**: Interactive buttons for simulator camera manipulation
- **Status Indicators**: Visual feedback for connection status and data freshness

## Development & Customization

### Adding New Variables
To add additional SimConnect variables:

1. **Define Constants** in `pkg/simconnect/monitor.go`:
   ```go
   const (
       NEW_VAR_DEFINE_ID = 37
       NEW_VAR_REQUEST_ID = 137
   )
   ```

2. **Update Data Model** in `pkg/models/monitor.go`:
   ```go
   type FlightData struct {
       // ...existing fields...
       NewVariable float32 `json:"newVariable"`
   }
   ```

3. **Register Variable** in the `Connect()` method:
   ```go
   if err := mc.sdk.RegisterSimVarDefinition(
       NEW_VAR_DEFINE_ID,
       "SIMCONNECT_VARIABLE_NAME",
       "units",
       types.SIMCONNECT_DATATYPE_FLOAT32,
   ); err != nil {
       return fmt.Errorf("failed to register variable: %v", err)
   }
   ```

4. **Add Periodic Request**:
   ```go
   if err := mc.sdk.RequestSimVarDataPeriodic(
       NEW_VAR_DEFINE_ID, 
       NEW_VAR_REQUEST_ID, 
       types.SIMCONNECT_PERIOD_SECOND
   ); err != nil {
       return fmt.Errorf("failed to start monitoring: %v", err)
   }
   ```

5. **Handle Updates** in `processSimConnectMessages()`:
   ```go
   case NEW_VAR_REQUEST_ID:
       mc.currentData.NewVariable = simVar.Value.(float32)
   ```

### Custom API Endpoints
Add new endpoints by creating handlers in `pkg/handlers/` and registering them in `cmd/server/main.go`:

```go
http.HandleFunc("/api/custom", customHandler.HandleCustomAPI)
```

### Frontend Customization
- Modify `templates/index.html` for layout changes
- Update `static/css/styles.css` for custom styling
- Extend `static/js/main.js` for additional functionality

## Advanced Usage Patterns

### Production Deployment
For production environments, consider these enhancements:

```powershell
# Build optimized executable
go build -ldflags="-s -w" -o flight-monitor.exe cmd\server\main.go

# Run as Windows service (requires additional service wrapper)
# Or run with process manager like PM2 for Node.js equivalent
```

### Multiple Client Architecture
As described in the [Advanced Usage Guide](../../docs/ADVANCED_USAGE.md), you can extend this example to support multiple SimConnect clients for different purposes:

- **High-frequency monitoring**: Critical flight data
- **Environmental monitoring**: Weather and atmospheric data  
- **Control systems**: Aircraft systems control
- **Navigation data**: GPS and radio frequencies

### Integration with External Systems
The webservice can be extended to integrate with:

- **Flight tracking databases**: Log flight data to SQL databases
- **External APIs**: Weather services, airport information
- **Hardware interfaces**: Physical instrument panels
- **Stream overlays**: OBS Studio integration for streaming

### Custom Monitoring Presets
The application supports monitor presets (placeholder implementation):

```javascript
// POST to /api/monitor/preset
{
    "name": "Stormy Weather",
    "temperature": 5.0,
    "pressure": 28.8,
    "windSpeed": 25.0,
    "windDirection": 270.0
}
```

## Troubleshooting & FAQ

### Performance Issues
- **High CPU usage**: Check if multiple SimConnect applications are running
- **Memory leaks**: Restart application periodically in long-running scenarios
- **Slow updates**: Verify MSFS is not paused and aircraft is properly loaded

### Data Accuracy
- **Invalid readings**: Some variables may return default values when aircraft systems are off
- **Zero values**: Normal for certain variables when aircraft is cold and dark
- **String variables**: Airport names may be empty if no nearby airports

### Browser Compatibility
- **Modern browsers**: Chrome 80+, Firefox 75+, Safari 13+, Edge 80+
- **JavaScript required**: Application will not function with JavaScript disabled
- **Local network access**: Ensure firewall allows connections to port 8080

### Common Questions

**Q: Can I change the update frequency?**
A: Yes, modify the `types.SIMCONNECT_PERIOD_SECOND` parameter in the periodic requests to `SIMCONNECT_PERIOD_VISUAL_FRAME` for higher frequency or other period types.

**Q: How do I add custom weather data?**
A: Follow the "Adding New Variables" section and use weather-related SimConnect variables like `AMBIENT WIND GUSTS VELOCITY` or `AMBIENT TEMPERATURE GRADIENT`.

**Q: Can I control other aircraft systems?**
A: Yes, extend the camera control pattern to add controls for lights, engines, autopilot, etc. using SimConnect event mapping.

**Q: Does this work with X-Plane or other simulators?**
A: No, this is specifically designed for Microsoft Flight Simulator's SimConnect API. Other simulators would require different SDKs.

## Related Documentation

- [Advanced Usage Guide](../../docs/ADVANCED_USAGE.md) - Comprehensive patterns for production applications
- [API Documentation](../../docs/API.md) - Complete SDK API reference
- [Main SDK README](../../README.md) - Getting started with the SimConnect Go SDK
- [External Power Logger Example](../external-power-logger/README.md) - Simpler monitoring example

## Contributing

This example demonstrates best practices for:
- Single `Listen()` call architecture (required by SDK)
- Proper goroutine management and synchronization
- Comprehensive error handling and recovery
- Modern web interface development
- Production-ready code organization

When contributing improvements:
1. Follow the established patterns
2. Maintain backward compatibility
3. Add appropriate error handling
4. Update documentation
5. Test with different aircraft and scenarios

---

**Note**: This webservice example showcases the full capabilities of the SimConnect Go SDK while maintaining proper architecture patterns. It serves as both a functional flight monitoring tool and a reference implementation for building production applications.
