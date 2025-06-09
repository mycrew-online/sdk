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
- [Quick Start](#quick-start)
- [Basic Usage](#basic-usage)
- [Features](#features)
- [Documentation](#documentation)
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

## Quick Start

Get up and running with basic flight data monitoring in just a few lines of code.

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

func main() {
    // Create and connect to SimConnect
    sdk := client.New("QuickStart")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        panic(fmt.Sprintf("Failed to connect: %v", err))
    }
    fmt.Println("✅ Connected to MSFS!")

    // Register altitude monitoring
    err := sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
    if err != nil {
        panic(fmt.Sprintf("Failed to register variable: %v", err))
    }

    // Request continuous updates
    err = sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_SECOND)
    if err != nil {
        panic(fmt.Sprintf("Failed to request data: %v", err))
    }

    // Listen for altitude data
    messages := sdk.Listen()
    timeout := time.After(10 * time.Second)

    fmt.Println("📊 Monitoring altitude for 10 seconds...")
    for {
        select {
        case msg := <-messages:
            if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        fmt.Printf("✈️ Altitude: %.0f feet\n", simVar.Value)
                    }
                }
            }
        case <-timeout:
            fmt.Println("👋 Monitoring complete!")
            return
        }
    }
}
```

## Basic Usage

### Connection Management

```go
// Standard connection
sdk := client.New("MyApp")
defer sdk.Close()

// Custom DLL path (if needed)
sdk := client.NewWithCustomDLL("MyApp", "D:/Custom/SimConnect.dll")

// Open connection
if err := sdk.Open(); err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
```

### Reading Flight Data

```go
// Register flight instruments
sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(2, "AIRSPEED INDICATED", "knots", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(3, "HEADING INDICATOR", "degrees", types.SIMCONNECT_DATATYPE_FLOAT32)

// Request periodic updates
sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_SECOND)
sdk.RequestSimVarDataPeriodic(2, 200, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
sdk.RequestSimVarDataPeriodic(3, 300, types.SIMCONNECT_PERIOD_SECOND)

// Process messages
messages := sdk.Listen()
for msg := range messages {
    if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
        if data, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := data.(*client.SimVarData); ok {
                switch simVar.DefineID {
                case 1:
                    fmt.Printf("Altitude: %.0f feet\n", simVar.Value)
                case 2:
                    fmt.Printf("Airspeed: %.0f knots\n", simVar.Value)
                case 3:
                    fmt.Printf("Heading: %.0f degrees\n", simVar.Value)
                }
            }
        }
    }
}
```

### Controlling Aircraft Systems

```go
// Map and trigger aircraft events
sdk.MapClientEventToSimEvent(1001, "TOGGLE_MASTER_BATTERY")
sdk.AddClientEventToNotificationGroup(3000, 1001, false)
sdk.SetNotificationGroupPriority(3000, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST)

// Toggle master battery
sdk.TransmitClientEvent(types.SIMCONNECT_OBJECT_ID_USER, 1001, 0, 3000, types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY)
```

### Event Monitoring

```go
// Subscribe to system events
sdk.SubscribeToSystemEvent(2001, "Pause")
sdk.SubscribeToSystemEvent(2002, "Crashed")

// Handle events in message loop
for msg := range messages {
    if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "EVENT" {
        if eventData, exists := msgMap["event"]; exists {
            if event, ok := eventData.(*types.EventData); ok {
                fmt.Printf("System event: %d, data: %d\n", event.EventID, event.EventData)
            }
        }
    }
}
```

## Features

### 🎯 **Core Capabilities**
- **Real-time Data Access** - Monitor 100+ aircraft variables with configurable update rates
- **Aircraft Control** - Send commands to control lights, engines, autopilot, and more
- **System Events** - Subscribe to simulator state changes (pause, crash, aircraft loaded)
- **Custom DLL Support** - Use custom SimConnect DLL installations

### 📊 **Data Types Supported**
- **Numeric**: INT32, FLOAT32, FLOAT64, INT64
- **Strings**: Variable length and fixed-length strings (8-260 characters)
- **Structures**: InitPosition, MarkerState, Waypoint, LatLonAlt, XYZ
- **All 17 SimConnect data types** with automatic parsing

### ⚡ **Performance Features**
- **Efficient Memory Usage** - Zero-copy parsing where possible
- **Configurable Update Rates** - From real-time (60fps) to periodic (seconds)
- **Concurrent Processing** - Thread-safe message handling
- **Automatic Cleanup** - Resource management with proper disposal

### 🛡️ **Reliability Features**
- **Exception Handling** - Detailed SimConnect exception reporting with recovery suggestions
- **Connection Recovery** - Robust connection management with retry logic
- **Message Validation** - Type-safe message parsing with comprehensive error handling
- **Resource Tracking** - Automatic cleanup of periodic requests and event subscriptions

## Documentation

| Document | Description |
|----------|-------------|
| **[Getting Started](docs/GETTING_STARTED.md)** | Detailed setup guide with troubleshooting |
| **[Examples](docs/EXAMPLES.md)** | Complete working examples for common scenarios |
| **[API Reference](docs/API.md)** | Complete API documentation with parameters and return values |
| **[Advanced Usage](docs/ADVANCED_USAGE.md)** | Advanced patterns, concurrent processing, and production architectures |
| **[Performance Guide](docs/PERFORMANCE.md)** | Optimization techniques and best practices |
| **[SimVars Reference](docs/SIMVARS.md)** | Common SimConnect variables with units and descriptions |

### Architecture Guides
- **[Multiple Clients](docs/MULTIPLE_CLIENTS.md)** - When and how to use multiple SDK instances
- **[Error Handling](docs/ERROR_HANDLING.md)** - Comprehensive error handling strategies
- **[Production Deployment](docs/PRODUCTION.md)** - Production-ready patterns and monitoring

## Examples

The `example/` directory contains complete working applications:

| Example | Description | Complexity |
|---------|-------------|------------|
| **[basic-monitoring](example/basic-monitoring/)** | Simple flight data monitoring | Beginner |
| **[aircraft-controller](example/aircraft-controller/)** | Control aircraft systems | Intermediate |
| **[sim-webservice](example/sim-webservice/)** | Web-based dashboard with real-time updates | Advanced |
| **[multi-client-system](example/multi-client-system/)** | Complex multi-client architecture | Expert |

### Running Examples

```bash
# Basic monitoring
cd example/basic-monitoring
go run main.go

# Web dashboard
cd example/sim-webservice
go run cmd/server/main.go
# Open http://localhost:8080

# See full example list
ls example/
```

## Quick API Reference

### Essential Methods

```go
// Connection
sdk := client.New("AppName")
err := sdk.Open()
defer sdk.Close()
messages := sdk.Listen() // Call only once per client

// Data requests
sdk.RegisterSimVarDefinition(id, varName, units, dataType)
sdk.RequestSimVarDataPeriodic(defineID, requestID, period)
sdk.StopPeriodicRequest(requestID)

// Events
sdk.SubscribeToSystemEvent(eventID, eventName)
sdk.MapClientEventToSimEvent(eventID, simEventName)
sdk.TransmitClientEvent(objectID, eventID, data, groupID, flags)
```

### Update Periods

- `SIMCONNECT_PERIOD_VISUAL_FRAME` - Every frame (~30-60 FPS)
- `SIMCONNECT_PERIOD_SECOND` - Every second
- `SIMCONNECT_PERIOD_ON_SET` - When value changes
- `SIMCONNECT_PERIOD_ONCE` - Single request

### Common Data Types

- `SIMCONNECT_DATATYPE_FLOAT32` - Most aircraft variables
- `SIMCONNECT_DATATYPE_INT32` - Boolean states, integer values
- `SIMCONNECT_DATATYPE_STRINGV` - Variable-length strings

**Need more details?** See the [complete API reference](docs/API.md).

## Performance Tips

### ✅ **Best Practices**
```go
// Use appropriate update frequencies
sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_VISUAL_FRAME) // Critical instruments
sdk.RequestSimVarDataPeriodic(2, 200, types.SIMCONNECT_PERIOD_SECOND)       // Navigation data

// Always cleanup periodic requests
defer func() {
    sdk.StopPeriodicRequest(100)
    sdk.StopPeriodicRequest(200)
}()

// Use single Listen() call with fan-out pattern for multiple processors
messages := sdk.Listen()
// Distribute messages to specialized processors - see Advanced Usage guide
```

### ❌ **Common Pitfalls**
```go
// DON'T: Multiple Listen() calls return the same channel
messages1 := sdk.Listen()
messages2 := sdk.Listen() // Same channel as messages1!

// DON'T: Multiple goroutines on same channel causes message loss
go func() { for msg := range messages { /* Only gets some messages */ } }()
go func() { for msg := range messages { /* Only gets some messages */ } }()
```

**For detailed performance optimization, see [Performance Guide](docs/PERFORMANCE.md).**

## Contributing

Contributions are welcomed and must follow [Code of Conduct](https://github.com/mycrew-online/sdk?tab=coc-ov-file) and common [Contributions guidelines](https://github.com/mycrew-online/sdk/blob/main/docs/CONTRIBUTING.md).

> If you'd like to report security issue please follow [security guidelines](https://github.com/mycrew-online/sdk?tab=security-ov-file).

---
<sup><sub>_All rights reserved &copy; mycrew-online and contributors_</sub></sup>
