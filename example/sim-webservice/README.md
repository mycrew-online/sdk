# Weather WebService Demo

A real-time web interface for monitoring and controlling Microsoft Flight Simulator weather conditions using the SimConnect SDK.

## Features

- **Real-time Weather Monitoring**: View current weather conditions including temperature, pressure, wind speed/direction
- **Web Interface**: Modern, responsive UI built with Tailwind CSS
- **Live Updates**: Weather data updates automatically every second
- **Small Demo**: Starting with basic weather variables for end-to-end testing

## Getting Started

### Prerequisites

1. Microsoft Flight Simulator must be running
2. An aircraft must be loaded  
3. SimConnect SDK must be installed

### Running the Demo

```powershell
# Navigate to this directory
cd example\weather-webservice

# Build and run
go run main.go
```

Then open your browser to `http://localhost:8080`

## Current Implementation

**Phase 1**: Basic weather monitoring
- Ambient Temperature (°C)
- Barometric Pressure (inHg) 
- Wind Speed (knots)
- Wind Direction (degrees)

**Next Phases** (planned):
- Weather controls/settings
- Additional weather variables
- Extended UI features

## Technical Details

- **Port**: 8080
- **Update Frequency**: 1 second
- **WebSocket**: Real-time data streaming
- **Frontend**: Vanilla HTML/CSS with Tailwind CDN
