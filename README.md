# mycrew-online/sdk

Go SDK for Microsoft Flight Simulator SimConnect interface, providing seamless integration with MSFS 2024/2020 for real-time aircraft data access, event handling, and simulator control.

| Package | `mycrew-online/sdk` |
| :-- | :-- |
| Go module | `github.com/mycrew-online/sdk` |
| Go version | ![Go Version](https://img.shields.io/badge/go-1.21+-blue) |
| Latest version | ![GitHub Release](https://img.shields.io/github/v/release/mycrew-online/sdk) |
| License | ![GitHub License](https://img.shields.io/github/license/mycrew-online/sdk) |

## Table of contents

- [Installation](#installation)
- [Usage](#usage)
- [Advanced Usage](#advanced-usage)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Contributing](#contributing)

## Installation

> This package requires Microsoft Flight Simulator 2024 or 2020 with SimConnect SDK installed.

```shell
$ go get github.com/mycrew-online/sdk
```

### Prerequisites

- **Microsoft Flight Simulator 2024/2020** - Must be running for connection
- **SimConnect SDK** - Typically installed at `C:/MSFS 2024 SDK/SimConnect SDK/lib/SimConnect.dll`
- **Go 1.21+** - Required for module support

## Usage

Import the package and create a connection to start accessing flight simulator data.

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

func main() {
    // Create a new SimConnect client
    sdk := client.New("MyFlightApp")
    defer sdk.Close()

    // Connect to Microsoft Flight Simulator
    if err := sdk.Open(); err != nil {
        fmt.Printf("Failed to connect: %v\n", err)
        return
    }
    fmt.Println("Connected to MSFS!")

    // Register a simulation variable
    err := sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
    if err != nil {
        fmt.Printf("Failed to register SimVar: %v\n", err)
        return
    }

    // Request data once
    err = sdk.RequestSimVarData(1, 100)
    if err != nil {
        fmt.Printf("Failed to request data: %v\n", err)
        return
    }

    // Listen for messages
    messages := sdk.Listen()
    timeout := time.After(5 * time.Second)

    for {
        select {
        case msg := <-messages:
            if msgMap, ok := msg.(map[string]any); ok {
                if msgMap["type"] == "SIMOBJECT_DATA" {
                    if data, exists := msgMap["parsed_data"]; exists {
                        if simVarData, ok := data.(*client.SimVarData); ok {
                            fmt.Printf("Altitude: %.2f feet\n", simVarData.Value)
                        }
                    }
                }
            }
        case <-timeout:
            fmt.Println("Timeout reached")
            return
        }
    }
}
```

## Advanced Usage

### Periodic Data Requests

Request continuous data updates at specified intervals instead of one-time requests.

```go
// Register airspeed variable
err := sdk.RegisterSimVarDefinition(2, "AIRSPEED INDICATED", "knots", types.SIMCONNECT_DATATYPE_FLOAT32)

// Request updates every visual frame (~30-60 FPS)
err = sdk.RequestSimVarDataPeriodic(2, 200, types.SIMCONNECT_PERIOD_VISUAL_FRAME)

// Stop periodic updates when done
err = sdk.StopPeriodicRequest(200)
```

#### Available Update Periods

- `SIMCONNECT_PERIOD_NEVER` - Never send data (used for stopping)
- `SIMCONNECT_PERIOD_ONCE` - Send data once only (default)
- `SIMCONNECT_PERIOD_VISUAL_FRAME` - Every visual frame (~30-60 FPS)
- `SIMCONNECT_PERIOD_ON_SET` - When variables change
- `SIMCONNECT_PERIOD_SECOND` - Once per second

### System Event Subscription

Subscribe to simulator events like pause, aircraft loading, etc.

```go
// Subscribe to pause events
err := sdk.SubscribeToSystemEvent(1010, "Pause")

// Subscribe to aircraft loading events  
err = sdk.SubscribeToSystemEvent(1020, "AircraftLoaded")

// Listen for events in the message channel
for msg := range messages {
    if msgMap, ok := msg.(map[string]any); ok {
        if msgMap["type"] == "EVENT" {
            if eventData, exists := msgMap["event"]; exists {
                if event, ok := eventData.(*types.EventData); ok {
                    fmt.Printf("Event: %s (Type: %s)\n", event.EventName, event.EventType)
                }
            }
        }
    }
}
```

### Client Events and Aircraft Control

Map and transmit client events to control aircraft systems.

```go
// Define event IDs
const (
    EVENT_ID_TOGGLE_EXTERNAL_POWER = 10011511
    EVENT_ID_TOGGLE_MASTER_BATTERY = 10025115
    EVENT_ID_TOGGLE_BEACON_LIGHTS  = 10041515
)

// Map client events to simulator events
err := sdk.MapClientEventToSimEvent(EVENT_ID_TOGGLE_EXTERNAL_POWER, "TOGGLE_EXTERNAL_POWER")
err = sdk.MapClientEventToSimEvent(EVENT_ID_TOGGLE_MASTER_BATTERY, "TOGGLE_MASTER_BATTERY")

// Create notification group and set priority
err = sdk.SetNotificationGroupPriority(2000, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST)

// Add events to notification group
err = sdk.AddClientEventToNotificationGroup(2000, EVENT_ID_TOGGLE_EXTERNAL_POWER, false)
err = sdk.AddClientEventToNotificationGroup(2000, EVENT_ID_TOGGLE_MASTER_BATTERY, false)

// Transmit events to control aircraft
err = sdk.TransmitClientEvent(
    types.SIMCONNECT_OBJECT_ID_USER, 
    EVENT_ID_TOGGLE_EXTERNAL_POWER, 
    0, 
    2000, 
    types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY,
)
```

### Setting Simulation Variables

Write data back to the simulator to control aircraft state.

```go
// Register a settable variable
err := sdk.RegisterSimVarDefinition(3, "CAMERA STATE", "Enum", types.SIMCONNECT_DATATYPE_INT32)

// Set camera state values
cameraStates := []int32{2, 3, 4, 6} // Different camera views
for _, state := range cameraStates {
    err := sdk.SetSimVar(3, state)
    if err != nil {
        fmt.Printf("Failed to set camera state: %v\n", err)
    }
    time.Sleep(2 * time.Second) // Wait between changes
}
```

### Multi-Variable Monitoring

Monitor multiple aircraft systems simultaneously.

```go
// Define electrical system variables
electricalVars := []struct {
    defineID uint32
    name     string
    units    string
    dataType types.SimConnectDataType
}{
    {7, "EXTERNAL POWER AVAILABLE", "Bool", types.SIMCONNECT_DATATYPE_INT32},
    {8, "EXTERNAL POWER ON", "Bool", types.SIMCONNECT_DATATYPE_INT32},
    {9, "ELECTRICAL MASTER BATTERY", "Bool", types.SIMCONNECT_DATATYPE_INT32},
    {10, "GENERAL ENG MASTER ALTERNATOR", "Bool", types.SIMCONNECT_DATATYPE_INT32},
    {11, "LIGHT BEACON", "Bool", types.SIMCONNECT_DATATYPE_INT32},
    {12, "LIGHT NAV", "Bool", types.SIMCONNECT_DATATYPE_INT32},
    {13, "ELECTRICAL MAIN BUS VOLTAGE", "Volts", types.SIMCONNECT_DATATYPE_FLOAT32},
    {14, "ELECTRICAL BATTERY LOAD", "Amperes", types.SIMCONNECT_DATATYPE_FLOAT32},
}

// Register and start periodic monitoring for all variables
for _, v := range electricalVars {
    err := sdk.RegisterSimVarDefinition(v.defineID, v.name, v.units, v.dataType)
    if err != nil {
        continue
    }
    
    // Monitor each variable every second
    requestID := 700 + v.defineID
    err = sdk.RequestSimVarDataPeriodic(v.defineID, requestID, types.SIMCONNECT_PERIOD_SECOND)
}
```

## API Reference

### Connection Interface

```go
type Connection interface {
    // Connection Management
    Open() error
    Close() error
    Listen() <-chan any
    
    // SimVar Operations  
    RegisterSimVarDefinition(defID uint32, varName string, units string, dataType types.SimConnectDataType) error
    RequestSimVarData(defID uint32, requestID uint32) error
    RequestSimVarDataPeriodic(defID uint32, requestID uint32, period types.SimConnectPeriod) error
    StopPeriodicRequest(requestID uint32) error
    SetSimVar(defID uint32, value interface{}) error
    
    // Event Management
    SubscribeToSystemEvent(eventID uint32, eventName string) error
    MapClientEventToSimEvent(eventID types.ClientEventID, eventName string) error
    AddClientEventToNotificationGroup(groupID types.NotificationGroupID, eventID types.ClientEventID, maskable bool) error
    SetNotificationGroupPriority(groupID types.NotificationGroupID, priority uint32) error
    TransmitClientEvent(objectID uint32, eventID types.ClientEventID, data uint32, groupID types.NotificationGroupID, flags uint32) error
}
```

### Data Types

```go
// Supported SimConnect data types
const (
    SIMCONNECT_DATATYPE_INT32    // 32-bit integer
    SIMCONNECT_DATATYPE_FLOAT32  // 32-bit float  
    SIMCONNECT_DATATYPE_STRINGV  // Variable-length string
    // ... additional types available
)

// Update periods for periodic requests
const (
    SIMCONNECT_PERIOD_NEVER        // Stop updates
    SIMCONNECT_PERIOD_ONCE         // Single request
    SIMCONNECT_PERIOD_VISUAL_FRAME // Every frame
    SIMCONNECT_PERIOD_ON_SET       // On value change
    SIMCONNECT_PERIOD_SECOND       // Every second
)
```

### Message Types

Messages received through the `Listen()` channel contain parsed data:

```go
// SimVar data message
type SimVarData struct {
    RequestID uint32        // Request identifier
    DefineID  uint32        // Variable definition ID
    Value     interface{}   // Parsed value (int32, float64, string)
}

// Event data message  
type EventData struct {
    GroupID   uint32  // Notification group ID
    EventID   uint32  // Event identifier
    EventData uint32  // Event data payload
    EventType string  // "system" or "client"
    EventName string  // Human-readable event name
}
```

## Examples

### Real-time Flight Dashboard

```go
package main

import (
    "fmt"
    "time"
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

func main() {
    sdk := client.New("FlightDashboard")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        panic(err)
    }

    // Register flight data variables
    variables := map[uint32]string{
        1: "PLANE ALTITUDE",
        2: "AIRSPEED INDICATED", 
        3: "HEADING INDICATOR",
        4: "VERTICAL SPEED",
    }

    for id, name := range variables {
        sdk.RegisterSimVarDefinition(id, name, "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
        sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
    }

    messages := sdk.Listen()
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    data := make(map[uint32]float64)

    for {
        select {
        case msg := <-messages:
            if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
                if parsed, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := parsed.(*client.SimVarData); ok {
                        data[simVar.DefineID] = simVar.Value.(float64)
                    }
                }
            }
        case <-ticker.C:
            fmt.Printf("\r🛩️  Alt: %.0f ft | Speed: %.0f kts | Hdg: %.0f° | VS: %.0f fpm", 
                data[1], data[2], data[3], data[4])
        }
    }
}
```

### Aircraft Systems Monitor

```go
func monitorElectricalSystems() {
    sdk := client.New("ElectricalMonitor")
    defer sdk.Close()
    
    if err := sdk.Open(); err != nil {
        panic(err)
    }

    // Register electrical system variables
    systems := map[uint32]string{
        10: "EXTERNAL POWER ON",
        11: "ELECTRICAL MASTER BATTERY", 
        12: "GENERAL ENG MASTER ALTERNATOR",
        13: "ELECTRICAL MAIN BUS VOLTAGE",
    }

    for id, name := range systems {
        dataType := types.SIMCONNECT_DATATYPE_INT32
        if id == 13 { // Voltage is float
            dataType = types.SIMCONNECT_DATATYPE_FLOAT32
        }
        
        sdk.RegisterSimVarDefinition(id, name, "Bool", dataType)
        sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
    }

    for msg := range sdk.Listen() {
        if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
            if parsed, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := parsed.(*client.SimVarData); ok {
                    systemName := systems[simVar.DefineID]
                    fmt.Printf("%s: %v\n", systemName, simVar.Value)
                }
            }
        }
    }
}
```

### Custom DLL Path

```go
// Use custom SimConnect DLL location
sdk := client.NewWithCustomDLL("MyApp", "D:/CustomPath/SimConnect.dll")
```

## Performance Considerations

### Update Frequency Guidelines

- **VISUAL_FRAME**: Use for critical flight instruments (airspeed, altitude, attitude)
- **SECOND**: Use for navigation data (position, heading, fuel levels)  
- **ON_SET**: Use for user-controlled settings (camera modes, autopilot states)

### Resource Management

- Always call `StopPeriodicRequest()` when data is no longer needed
- Use appropriate update frequencies to avoid overwhelming your application
- Consider grouping related variables with similar update requirements
- Use different RequestIDs to distinguish between data categories

### Error Handling

The SDK provides comprehensive error handling:

```go
// All operations return detailed errors
if err := sdk.RegisterSimVarDefinition(1, "INVALID_VAR", "units", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
    fmt.Printf("Registration failed: %v\n", err)
}

// Exception messages are propagated through the message channel
for msg := range messages {
    if msgMap, ok := msg.(map[string]any); ok {
        if msgMap["type"] == "EXCEPTION" {
            if exception, exists := msgMap["exception"]; exists {
                fmt.Printf("SimConnect Exception: %+v\n", exception)
            }
        }
    }
}
```

## Contributing

_Contributions are welcomed and must follow [Code of Conduct](https://github.com/mycrew-online/sdk?tab=coc-ov-file) and common [Contributions guidelines](https://github.com/mycrew-online/sdk/blob/main/docs/CONTRIBUTING.md)._

> If you'd like to report security issue please follow [security guidelines](https://github.com/mycrew-online/sdk?tab=security-ov-file).

---
<sup><sub>_All rights reserved &copy; mycrew-online and contributors_</sub></sup>
