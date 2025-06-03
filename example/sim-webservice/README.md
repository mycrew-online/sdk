# Weather WebService Demo

A real-time web interface for monitoring and controlling Microsoft Flight Simulator weather conditions using the SimConnect SDK.

## Features

- **Real-time Weather Monitoring**: View current weather conditions including temperature, pressure, wind speed/direction
- **Web Interface**: Modern, responsive UI built with Tailwind CSS
- **Live Updates**: Weather data updates automatically every second
- **Custom DLL Support**: Specify custom SimConnect.dll path via command line
- **Comprehensive Flight Data**: 30 variables including weather, navigation, and flight status

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
go build -o sim-webservice.exe cmd\server\main.go

# Run with default DLL
.\sim-webservice.exe

# Run with custom DLL
.\sim-webservice.exe -dll "C:\Custom\Path\SimConnect.dll"
```

Then open your browser to `http://localhost:8080`

### Command Line Options

- `-dll <path>`: Specify custom path to SimConnect.dll
  - If not provided, uses default path: `C:/MSFS 2024 SDK/SimConnect SDK/lib/SimConnect.dll`
  - Useful for custom installations or older MSFS versions

## Troubleshooting

### SimConnect DLL Issues

If you encounter DLL loading errors:

1. **Default Path Issues**: Use the `-dll` flag to specify the correct path
   ```powershell
   # For MSFS 2020
   .\sim-webservice.exe -dll "C:\MSFS SDK\SimConnect SDK\lib\SimConnect.dll"
   
   # For custom installation
   .\sim-webservice.exe -dll "D:\YourPath\SimConnect.dll"
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

**Complete Flight Monitoring Suite**:

**Weather Variables (Row 1)**:
- Ambient Temperature (°C)
- Barometric Pressure (inHg) 
- Wind Speed (knots)
- Wind Direction (degrees)

**Environmental Variables (Row 2)**:
- Visibility (meters)
- Precipitation Rate (mm/h)
- Precipitation State
- Density Altitude (ft)
- Ground Altitude (m)
- Magnetic Variation (degrees)
- Sea Level Pressure (mb)
- Air Density (slugs/ft³)

**Position & Navigation (Row 3)**:
- Aircraft Latitude/Longitude
- Altitude (feet)
- Ground Speed (knots)
- True Heading (degrees)
- Vertical Speed (ft/sec)

**Airport/Navigation Info (Row 4)**:
- Nearest Airport ID
- Distance to Airport (m)
- COM1 Frequency (MHz)
- NAV1 Frequency (MHz)
- GPS Distance to Waypoint (m)
- GPS ETE (seconds)

**Flight Status (Row 5)**:
- On Ground Status
- On Runway Status
- GPS Flight Plan Active
- Autopilot Master
- Surface Type
- Indicated Airspeed (knots)

## Technical Details

- **Port**: 8080
- **Update Frequency**: 1 second
- **Data Variables**: 30 flight simulation variables
- **Frontend**: Vanilla HTML/CSS with Tailwind CDN
- **Backend**: Go with SimConnect SDK integration
