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
- [Complete Examples](#complete-examples)
- [Advanced Usage](#advanced-usage)
- [API Reference](#api-reference)
- [Documentation](#documentation)
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

## Complete Examples

### Example 1: Real-Time Flight Dashboard

A comprehensive flight monitoring application that tracks multiple aircraft systems simultaneously.

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

type FlightDashboard struct {
    sdk        client.Connection
    flightData map[string]interface{}
    lastUpdate time.Time
}

func NewFlightDashboard() *FlightDashboard {
    return &FlightDashboard{
        sdk:        client.New("FlightDashboard"),
        flightData: make(map[string]interface{}),
    }
}

func (fd *FlightDashboard) Start() error {
    // Connect to simulator
    if err := fd.sdk.Open(); err != nil {
        return fmt.Errorf("connection failed: %v", err)
    }

    // Define flight instruments to monitor
    instruments := []struct {
        id     uint32
        name   string
        units  string
        label  string
        period types.SimConnectPeriod
    }{
        {1, "PLANE ALTITUDE", "feet", "Altitude", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
        {2, "AIRSPEED INDICATED", "knots", "Airspeed", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
        {3, "HEADING INDICATOR", "degrees", "Heading", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
        {4, "VERTICAL SPEED", "feet per minute", "Vertical Speed", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
        {5, "PLANE PITCH DEGREES", "degrees", "Pitch", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
        {6, "PLANE BANK DEGREES", "degrees", "Bank", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
        {7, "NAV OBS", "degrees", "OBS", types.SIMCONNECT_PERIOD_SECOND},
        {8, "GPS GROUND SPEED", "knots", "Ground Speed", types.SIMCONNECT_PERIOD_SECOND},
        {9, "FUEL TOTAL QUANTITY", "gallons", "Fuel Total", types.SIMCONNECT_PERIOD_SECOND},
        {10, "ENG RPM", "rpm", "Engine RPM", types.SIMCONNECT_PERIOD_SECOND},
    }

    // Register all instruments
    fmt.Println("📊 Registering flight instruments...")
    for _, instrument := range instruments {
        if err := fd.sdk.RegisterSimVarDefinition(
            instrument.id, 
            instrument.name, 
            instrument.units, 
            types.SIMCONNECT_DATATYPE_FLOAT32,
        ); err != nil {
            return fmt.Errorf("failed to register %s: %v", instrument.label, err)
        }

        if err := fd.sdk.RequestSimVarDataPeriodic(
            instrument.id, 
            instrument.id*100, 
            instrument.period,
        ); err != nil {
            return fmt.Errorf("failed to request %s data: %v", instrument.label, err)
        }
    }

    // Subscribe to critical system events
    systemEvents := []struct {
        id   uint32
        name string
    }{
        {2001, "Pause"},
        {2002, "Crashed"},
        {2003, "AircraftLoaded"},
        {2004, "FlightLoaded"},
    }

    fmt.Println("📡 Subscribing to system events...")
    for _, event := range systemEvents {
        if err := fd.sdk.SubscribeToSystemEvent(event.id, event.name); err != nil {
            log.Printf("⚠️ Failed to subscribe to %s: %v", event.name, err)
        }
    }

    fmt.Println("✅ Dashboard initialized successfully!")
    return nil
}

func (fd *FlightDashboard) Run() {
    // Set up graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Start message processing
    messages := fd.sdk.Listen()
    if messages == nil {
        log.Fatal("❌ Failed to start message listening")
    }

    // Display ticker for regular dashboard updates
    displayTicker := time.NewTicker(1 * time.Second)
    defer displayTicker.Stop()

    fmt.Println("🎯 Flight Dashboard Active - Press Ctrl+C to stop")
    fmt.Println(strings.Repeat("=", 80))

    for {
        select {
        case <-sigChan:
            fmt.Println("\n🛑 Shutdown signal received...")
            fd.cleanup()
            return

        case <-displayTicker.C:
            fd.displayDashboard()

        case msg, ok := <-messages:
            if !ok {
                fmt.Println("❌ Message channel closed")
                return
            }
            fd.processMessage(msg)
        }
    }
}

func (fd *FlightDashboard) processMessage(msg any) {
    msgMap, ok := msg.(map[string]any)
    if !ok {
        return
    }

    switch msgMap["type"] {
    case "SIMOBJECT_DATA":
        if parsedData, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := parsedData.(*client.SimVarData); ok {
                fd.updateFlightData(simVar)
            }
        }

    case "EVENT":
        if eventData, exists := msgMap["event"]; exists {
            if event, ok := eventData.(*types.EventData); ok {
                fd.handleSystemEvent(event)
            }
        }

    case "EXCEPTION":
        if exception, exists := msgMap["exception"]; exists {
            log.Printf("❌ SimConnect Exception: %+v", exception)
        }
    }
}

func (fd *FlightDashboard) updateFlightData(simVar *client.SimVarData) {
    // Map DefineID to readable names
    dataLabels := map[uint32]string{
        1:  "altitude",
        2:  "airspeed",
        3:  "heading",
        4:  "verticalSpeed",
        5:  "pitch",
        6:  "bank",
        7:  "obs",
        8:  "groundSpeed",
        9:  "fuelTotal",
        10: "engineRPM",
    }

    if label, exists := dataLabels[simVar.DefineID]; exists {
        fd.flightData[label] = simVar.Value
        fd.lastUpdate = time.Now()
    }
}

func (fd *FlightDashboard) handleSystemEvent(event *types.EventData) {
    switch event.EventName {
    case "Pause":
        if event.EventData == 1 {
            fmt.Println("\n⏸️ SIMULATOR PAUSED")
        } else {
            fmt.Println("\n▶️ SIMULATOR RESUMED")
        }
    case "Crashed":
        fmt.Printf("\n💥 AIRCRAFT CRASHED (Type: %d)\n", event.EventData)
    case "AircraftLoaded":
        fmt.Printf("\n✈️ AIRCRAFT LOADED (ID: %d)\n", event.EventData)
    case "FlightLoaded":
        fmt.Printf("\n🛫 FLIGHT PLAN LOADED (ID: %d)\n", event.EventData)
    }
}

func (fd *FlightDashboard) displayDashboard() {
    if len(fd.flightData) == 0 {
        fmt.Printf("\r⏳ Waiting for flight data... ")
        return
    }

    // Clear line and display current data
    fmt.Printf("\r\033[K") // Clear current line
    
    // Primary flight display
    altitude := fd.getFloat("altitude")
    airspeed := fd.getFloat("airspeed")
    heading := fd.getFloat("heading")
    verticalSpeed := fd.getFloat("verticalSpeed")
    
    fmt.Printf("✈️ ALT: %6.0fft | SPD: %3.0fkts | HDG: %03.0f° | VS: %+5.0ffpm",
        altitude, airspeed, heading, verticalSpeed)
    
    // Attitude and engine data on next update cycle
    if time.Since(fd.lastUpdate) < 2*time.Second {
        pitch := fd.getFloat("pitch")
        bank := fd.getFloat("bank")
        engineRPM := fd.getFloat("engineRPM")
        fuel := fd.getFloat("fuelTotal")
        
        fmt.Printf(" | PITCH: %+4.1f° | BANK: %+4.1f° | RPM: %.0f | FUEL: %.1fgal",
            pitch, bank, engineRPM, fuel)
    }
}

func (fd *FlightDashboard) getFloat(key string) float64 {
    if val, exists := fd.flightData[key]; exists {
        if f, ok := val.(float64); ok {
            return f
        }
        if f, ok := val.(float32); ok {
            return float64(f)
        }
    }
    return 0.0
}

func (fd *FlightDashboard) cleanup() {
    fmt.Println("🧹 Cleaning up dashboard...")
    
    // Stop all periodic requests
    for i := uint32(1); i <= 10; i++ {
        if err := fd.sdk.StopPeriodicRequest(i * 100); err != nil {
            log.Printf("⚠️ Failed to stop request %d: %v", i*100, err)
        }
    }
    
    fmt.Println("✅ Dashboard stopped")
}

func main() {
    dashboard := NewFlightDashboard()
    defer dashboard.sdk.Close()

    if err := dashboard.Start(); err != nil {
        log.Fatalf("❌ Failed to start dashboard: %v", err)
    }

    dashboard.Run()
}
```

### Example 2: Aircraft Systems Controller

Complete aircraft electrical and lighting system control with real-time feedback.

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

type AircraftController struct {
    sdk          client.Connection
    systemStates map[string]bool
    lastActivity time.Time
}

// Event IDs for aircraft controls
const (
    EVENT_TOGGLE_EXTERNAL_POWER = 10001
    EVENT_TOGGLE_MASTER_BATTERY = 10002
    EVENT_TOGGLE_BEACON_LIGHTS  = 10003
    EVENT_TOGGLE_NAV_LIGHTS     = 10004
    EVENT_TOGGLE_STROBE_LIGHTS  = 10005
    EVENT_TOGGLE_TAXI_LIGHTS    = 10006
)

func NewAircraftController() *AircraftController {
    return &AircraftController{
        sdk:          client.New("AircraftController"),
        systemStates: make(map[string]bool),
    }
}

func (ac *AircraftController) Initialize() error {
    if err := ac.sdk.Open(); err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }

    // Register system state variables for monitoring
    systemVars := []struct {
        id    uint32
        name  string
        label string
    }{
        {11, "EXTERNAL POWER ON", "ExternalPower"},
        {12, "ELECTRICAL MASTER BATTERY", "MasterBattery"},
        {13, "LIGHT BEACON", "BeaconLights"},
        {14, "LIGHT NAV", "NavLights"},
        {15, "LIGHT STROBE", "StrobeLights"},
        {16, "LIGHT TAXI", "TaxiLights"},
        {17, "ELECTRICAL MAIN BUS VOLTAGE", "BusVoltage"},
        {18, "ELECTRICAL BATTERY LOAD", "BatteryLoad"},
    }

    fmt.Println("📊 Registering system monitoring...")
    for _, sys := range systemVars {
        dataType := types.SIMCONNECT_DATATYPE_INT32
        units := "Bool"
        
        if sys.id >= 17 { // Voltage and current are floats
            dataType = types.SIMCONNECT_DATATYPE_FLOAT32
            if sys.id == 17 {
                units = "Volts"
            } else {
                units = "Amperes"
            }
        }

        if err := ac.sdk.RegisterSimVarDefinition(sys.id, sys.name, units, dataType); err != nil {
            return fmt.Errorf("failed to register %s: %v", sys.label, err)
        }

        if err := ac.sdk.RequestSimVarDataPeriodic(sys.id, sys.id*100, types.SIMCONNECT_PERIOD_SECOND); err != nil {
            return fmt.Errorf("failed to request %s data: %v", sys.label, err)
        }
    }

    // Set up control events
    controlEvents := map[types.ClientEventID]string{
        EVENT_TOGGLE_EXTERNAL_POWER: "TOGGLE_EXTERNAL_POWER",
        EVENT_TOGGLE_MASTER_BATTERY: "TOGGLE_MASTER_BATTERY",
        EVENT_TOGGLE_BEACON_LIGHTS:  "TOGGLE_BEACON_LIGHTS",
        EVENT_TOGGLE_NAV_LIGHTS:     "TOGGLE_NAV_LIGHTS",
        EVENT_TOGGLE_STROBE_LIGHTS:  "TOGGLE_STROBE_LIGHTS",
        EVENT_TOGGLE_TAXI_LIGHTS:    "TOGGLE_TAXI_LIGHTS",
    }

    fmt.Println("🎮 Setting up control events...")
    for eventID, eventName := range controlEvents {
        if err := ac.sdk.MapClientEventToSimEvent(eventID, eventName); err != nil {
            return fmt.Errorf("failed to map event %s: %v", eventName, err)
        }

        if err := ac.sdk.AddClientEventToNotificationGroup(3000, eventID, false); err != nil {
            return fmt.Errorf("failed to add event %s to group: %v", eventName, err)
        }
    }

    if err := ac.sdk.SetNotificationGroupPriority(3000, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST); err != nil {
        return fmt.Errorf("failed to set group priority: %v", err)
    }

    fmt.Println("✅ Aircraft controller initialized")
    return nil
}

func (ac *AircraftController) RunControlDemo() {
    messages := ac.sdk.Listen()
    if messages == nil {
        log.Fatal("❌ Failed to start listening")
    }

    // Monitor systems and run demo sequence
    go ac.processMessages(messages)

    // Wait for initial system state
    fmt.Println("⏳ Reading initial system state...")
    time.Sleep(2 * time.Second)

    // Run electrical system startup sequence
    ac.runStartupSequence()

    // Run lighting demo
    ac.runLightingDemo()

    // Run shutdown sequence
    ac.runShutdownSequence()
}

func (ac *AircraftController) processMessages(messages <-chan any) {
    for msg := range messages {
        msgMap, ok := msg.(map[string]any)
        if !ok {
            continue
        }

        switch msgMap["type"] {
        case "SIMOBJECT_DATA":
            if parsedData, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := parsedData.(*client.SimVarData); ok {
                    ac.updateSystemState(simVar)
                }
            }

        case "EVENT":
            if eventData, exists := msgMap["event"]; exists {
                if event, ok := eventData.(*types.EventData); ok {
                    ac.handleControlEvent(event)
                }
            }

        case "EXCEPTION":
            if exception, exists := msgMap["exception"]; exists {
                log.Printf("❌ Exception: %+v", exception)
            }
        }
    }
}

func (ac *AircraftController) updateSystemState(simVar *client.SimVarData) {
    var systemName string
    switch simVar.DefineID {
    case 11:
        systemName = "ExternalPower"
    case 12:
        systemName = "MasterBattery"
    case 13:
        systemName = "BeaconLights"
    case 14:
        systemName = "NavLights"
    case 15:
        systemName = "StrobeLights"
    case 16:
        systemName = "TaxiLights"
    case 17:
        fmt.Printf("🔋 Bus Voltage: %.1fV\n", simVar.Value)
        return
    case 18:
        fmt.Printf("⚡ Battery Load: %.1fA\n", simVar.Value)
        return
    default:
        return
    }

    // Update boolean system states
    if intVal, ok := simVar.Value.(int32); ok {
        ac.systemStates[systemName] = intVal != 0
        status := "❌ OFF"
        if intVal != 0 {
            status = "✅ ON"
        }
        fmt.Printf("📊 %s: %s\n", systemName, status)
    }
}

func (ac *AircraftController) handleControlEvent(event *types.EventData) {
    fmt.Printf("🎮 Control Event Executed: %s (Data: %d)\n", event.EventName, event.EventData)
    ac.lastActivity = time.Now()
}

func (ac *AircraftController) runStartupSequence() {
    fmt.Println("\n🚀 Starting Electrical System Startup Sequence...")
    
    // 1. Connect external power
    fmt.Println("🔌 Step 1: Connecting external power...")
    ac.executeControl(EVENT_TOGGLE_EXTERNAL_POWER, "Toggling External Power")
    time.Sleep(2 * time.Second)

    // 2. Enable master battery
    fmt.Println("🔋 Step 2: Enabling master battery...")
    ac.executeControl(EVENT_TOGGLE_MASTER_BATTERY, "Toggling Master Battery")
    time.Sleep(2 * time.Second)

    fmt.Println("✅ Electrical system startup complete!")
}

func (ac *AircraftController) runLightingDemo() {
    fmt.Println("\n💡 Running Lighting System Demo...")
    
    lights := []struct {
        event types.ClientEventID
        name  string
    }{
        {EVENT_TOGGLE_BEACON_LIGHTS, "Beacon Lights"},
        {EVENT_TOGGLE_NAV_LIGHTS, "Navigation Lights"},
        {EVENT_TOGGLE_STROBE_LIGHTS, "Strobe Lights"},
        {EVENT_TOGGLE_TAXI_LIGHTS, "Taxi Lights"},
    }

    // Turn on all lights sequentially
    for _, light := range lights {
        fmt.Printf("💡 Activating %s...\n", light.name)
        ac.executeControl(light.event, fmt.Sprintf("Toggling %s", light.name))
        time.Sleep(1 * time.Second)
    }

    time.Sleep(3 * time.Second)

    // Turn off all lights
    for _, light := range lights {
        fmt.Printf("💡 Deactivating %s...\n", light.name)
        ac.executeControl(light.event, fmt.Sprintf("Toggling %s", light.name))
        time.Sleep(1 * time.Second)
    }

    fmt.Println("✅ Lighting demo complete!")
}

func (ac *AircraftController) runShutdownSequence() {
    fmt.Println("\n🛑 Starting System Shutdown Sequence...")
    
    // 1. Disable master battery
    fmt.Println("🔋 Step 1: Disabling master battery...")
    ac.executeControl(EVENT_TOGGLE_MASTER_BATTERY, "Toggling Master Battery")
    time.Sleep(2 * time.Second)

    // 2. Disconnect external power
    fmt.Println("🔌 Step 2: Disconnecting external power...")
    ac.executeControl(EVENT_TOGGLE_EXTERNAL_POWER, "Toggling External Power")
    time.Sleep(2 * time.Second)

    fmt.Println("✅ System shutdown complete!")
}

func (ac *AircraftController) executeControl(eventID types.ClientEventID, description string) {
    fmt.Printf("🎮 Executing: %s\n", description)
    
    err := ac.sdk.TransmitClientEvent(
        types.SIMCONNECT_OBJECT_ID_USER,
        eventID,
        0, // No data value needed for toggle events
        3000,
        types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY,
    )
    
    if err != nil {
        fmt.Printf("❌ Failed to execute %s: %v\n", description, err)
    } else {
        fmt.Printf("✅ %s executed successfully\n", description)
    }
}

func (ac *AircraftController) Cleanup() {
    fmt.Println("🧹 Cleaning up controller...")
    
    // Stop all system monitoring
    for i := uint32(11); i <= 18; i++ {
        if err := ac.sdk.StopPeriodicRequest(i * 100); err != nil {
            log.Printf("⚠️ Failed to stop request %d: %v", i*100, err)
        }
    }
    
    fmt.Println("✅ Controller cleanup complete")
}

func main() {
    controller := NewAircraftController()
    defer controller.sdk.Close()
    defer controller.Cleanup()

    if err := controller.Initialize(); err != nil {
        log.Fatalf("❌ Failed to initialize controller: %v", err)
    }

    controller.RunControlDemo()
}
```

### Example 3: Environmental Data Logger

Comprehensive environmental and weather monitoring with data persistence.

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

type EnvironmentalLogger struct {
    sdk        client.Connection
    dataPoints []EnvironmentalData
    logFile    *os.File
}

type EnvironmentalData struct {
    Timestamp    time.Time `json:"timestamp"`
    Temperature  float64   `json:"temperature_celsius"`
    Pressure     float64   `json:"pressure_inHg"`
    WindSpeed    float64   `json:"wind_speed_knots"`
    WindDirection float64  `json:"wind_direction_degrees"`
    Visibility   float64   `json:"visibility_meters"`
    Precipitation float64  `json:"precipitation_rate"`
    DensityAlt   float64   `json:"density_altitude_feet"`
    GroundElev   float64   `json:"ground_elevation_feet"`
    MagVariation float64   `json:"magnetic_variation_degrees"`
    SeaLevelPres float64   `json:"sea_level_pressure_mb"`
    AirDensity   float64   `json:"air_density_slugs"`
    TrueAirspeed float64   `json:"true_airspeed_knots"`
}

func NewEnvironmentalLogger() *EnvironmentalLogger {
    return &EnvironmentalLogger{
        sdk: client.New("EnvironmentalLogger"),
    }
}

func (el *EnvironmentalLogger) Initialize() error {
    if err := el.sdk.Open(); err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }

    // Create log file with timestamp
    fileName := fmt.Sprintf("environmental_data_%s.json", 
        time.Now().Format("2006-01-02_15-04-05"))
    
    var err error
    el.logFile, err = os.Create(fileName)
    if err != nil {
        return fmt.Errorf("failed to create log file: %v", err)
    }

    // Register environmental variables
    envVars := []struct {
        id    uint32
        name  string
        units string
    }{
        {20, "AMBIENT TEMPERATURE", "celsius"},
        {21, "AMBIENT PRESSURE", "inHg"},
        {22, "AMBIENT WIND VELOCITY", "knots"},
        {23, "AMBIENT WIND DIRECTION", "degrees"},
        {24, "AMBIENT VISIBILITY", "meters"},
        {25, "AMBIENT PRECIP RATE", "mm per hour"},
        {26, "AMBIENT DENSITY ALTITUDE", "feet"},
        {27, "GROUND ALTITUDE", "feet"},
        {28, "MAGVAR", "degrees"},
        {29, "SEA LEVEL PRESSURE", "millibars"},
        {30, "AMBIENT AIR DENSITY", "slugs per cubic feet"},
        {31, "AIRSPEED TRUE", "knots"},
    }

    fmt.Println("🌍 Registering environmental variables...")
    for _, envVar := range envVars {
        if err := el.sdk.RegisterSimVarDefinition(
            envVar.id, 
            envVar.name, 
            envVar.units, 
            types.SIMCONNECT_DATATYPE_FLOAT32,
        ); err != nil {
            return fmt.Errorf("failed to register %s: %v", envVar.name, err)
        }

        // Request updates every 5 seconds for environmental data
        if err := el.sdk.RequestSimVarDataPeriodic(
            envVar.id, 
            envVar.id*100, 
            types.SIMCONNECT_PERIOD_SECOND,
        ); err != nil {
            return fmt.Errorf("failed to request %s data: %v", envVar.name, err)
        }
    }

    fmt.Printf("✅ Environmental logger initialized - logging to %s\n", fileName)
    return nil
}

func (el *EnvironmentalLogger) RunLogging(duration time.Duration) {
    messages := el.sdk.Listen()
    if messages == nil {
        log.Fatal("❌ Failed to start listening")
    }

    // Initialize current data point
    currentData := EnvironmentalData{
        Timestamp: time.Now(),
    }

    // Logging ticker - save data every 30 seconds
    logTicker := time.NewTicker(30 * time.Second)
    defer logTicker.Stop()

    // Display ticker - show current conditions every 5 seconds
    displayTicker := time.NewTicker(5 * time.Second)
    defer displayTicker.Stop()

    // Run for specified duration
    endTime := time.Now().Add(duration)

    fmt.Printf("📊 Environmental logging started - will run for %v\n", duration)
    fmt.Println("🌡️  Monitoring: Temperature, Pressure, Wind, Visibility, and more...")
    fmt.Println(strings.Repeat("=", 80))

    for time.Now().Before(endTime) {
        select {
        case <-logTicker.C:
            el.logDataPoint(currentData)

        case <-displayTicker.C:
            el.displayCurrentConditions(currentData)

        case msg, ok := <-messages:
            if !ok {
                fmt.Println("❌ Message channel closed")
                return
            }
            el.updateEnvironmentalData(msg, &currentData)

        case <-time.After(1 * time.Second):
            // Continue loop to check end time
        }
    }

    fmt.Println("\n⏰ Logging duration completed")
    el.logDataPoint(currentData) // Final log entry
}

func (el *EnvironmentalLogger) updateEnvironmentalData(msg any, data *EnvironmentalData) {
    msgMap, ok := msg.(map[string]any)
    if !ok {
        return
    }

    if msgMap["type"] != "SIMOBJECT_DATA" {
        return
    }

    parsedData, exists := msgMap["parsed_data"]
    if !exists {
        return
    }

    simVar, ok := parsedData.(*client.SimVarData)
    if !ok {
        return
    }

    // Update timestamp on any new data
    data.Timestamp = time.Now()

    // Map DefineID to data fields
    value := el.convertToFloat64(simVar.Value)
    
    switch simVar.DefineID {
    case 20:
        data.Temperature = value
    case 21:
        data.Pressure = value
    case 22:
        data.WindSpeed = value
    case 23:
        data.WindDirection = value
    case 24:
        data.Visibility = value
    case 25:
        data.Precipitation = value
    case 26:
        data.DensityAlt = value
    case 27:
        data.GroundElev = value
    case 28:
        data.MagVariation = value
    case 29:
        data.SeaLevelPres = value
    case 30:
        data.AirDensity = value
    case 31:
        data.TrueAirspeed = value
    }
}

func (el *EnvironmentalLogger) convertToFloat64(value interface{}) float64 {
    switch v := value.(type) {
    case float32:
        return float64(v)
    case float64:
        return v
    case int32:
        return float64(v)
    case int:
        return float64(v)
    default:
        return 0.0
    }
}

func (el *EnvironmentalLogger) displayCurrentConditions(data EnvironmentalData) {
    fmt.Printf("\r\033[K") // Clear line
    
    // Display key environmental conditions
    fmt.Printf("🌡️ %.1f°C | 🌪️ %.0f@%.0f° | 👁️ %.0fm | ⬆️ %.0fft | 🧭 %.1f° | TAS: %.0fkts",
        data.Temperature,
        data.WindSpeed, data.WindDirection,
        data.Visibility,
        data.DensityAlt,
        data.MagVariation,
        data.TrueAirspeed)
}

func (el *EnvironmentalLogger) logDataPoint(data EnvironmentalData) {
    el.dataPoints = append(el.dataPoints, data)
    
    // Write to JSON log file
    jsonData, err := json.Marshal(data)
    if err != nil {
        log.Printf("❌ Failed to marshal data: %v", err)
        return
    }
    
    // Write JSON line
    if _, err := el.logFile.WriteString(string(jsonData) + "\n"); err != nil {
        log.Printf("❌ Failed to write to log file: %v", err)
        return
    }
    
    el.logFile.Sync() // Ensure data is written to disk
    
    fmt.Printf("\n📝 Logged data point %d at %s\n", 
        len(el.dataPoints), 
        data.Timestamp.Format("15:04:05"))
}

func (el *EnvironmentalLogger) GenerateSummary() {
    if len(el.dataPoints) == 0 {
        fmt.Println("📊 No data points collected")
        return
    }

    fmt.Println("\n📊 Environmental Data Summary")
    fmt.Println(strings.Repeat("=", 50))

    // Calculate averages and ranges
    var (
        tempSum, tempMin, tempMax           = 0.0, 999.0, -999.0
        windSpeedSum, windSpeedMin, windSpeedMax = 0.0, 999.0, -999.0
        pressureSum, pressureMin, pressureMax = 0.0, 999.0, -999.0
    )

    for _, dp := range el.dataPoints {
        // Temperature
        tempSum += dp.Temperature
        if dp.Temperature < tempMin { tempMin = dp.Temperature }
        if dp.Temperature > tempMax { tempMax = dp.Temperature }
        
        // Wind speed
        windSpeedSum += dp.WindSpeed
        if dp.WindSpeed < windSpeedMin { windSpeedMin = dp.WindSpeed }
        if dp.WindSpeed > windSpeedMax { windSpeedMax = dp.WindSpeed }
        
        // Pressure
        pressureSum += dp.Pressure
        if dp.Pressure < pressureMin { pressureMin = dp.Pressure }
        if dp.Pressure > pressureMax { pressureMax = dp.Pressure }
    }

    count := float64(len(el.dataPoints))
    
    fmt.Printf("📊 Data Points Collected: %d\n", len(el.dataPoints))
    fmt.Printf("🌡️ Temperature: Avg %.1f°C | Range %.1f°C to %.1f°C\n", 
        tempSum/count, tempMin, tempMax)
    fmt.Printf("🌪️ Wind Speed: Avg %.1f kts | Range %.1f to %.1f kts\n", 
        windSpeedSum/count, windSpeedMin, windSpeedMax)
    fmt.Printf("📊 Pressure: Avg %.2f inHg | Range %.2f to %.2f inHg\n", 
        pressureSum/count, pressureMin, pressureMax)
    fmt.Printf("⏱️ Logging Duration: %v\n", 
        el.dataPoints[len(el.dataPoints)-1].Timestamp.Sub(el.dataPoints[0].Timestamp))
}

func (el *EnvironmentalLogger) Cleanup() {
    fmt.Println("\n🧹 Cleaning up environmental logger...")
    
    // Stop all periodic requests
    for i := uint32(20); i <= 31; i++ {
        if err := el.sdk.StopPeriodicRequest(i * 100); err != nil {
            log.Printf("⚠️ Failed to stop request %d: %v", i*100, err)
        }
    }
    
    // Close log file
    if el.logFile != nil {
        el.logFile.Close()
    }
    
    el.GenerateSummary()
    fmt.Println("✅ Environmental logger cleanup complete")
}

func main() {
    logger := NewEnvironmentalLogger()
    defer logger.sdk.Close()
    defer logger.Cleanup()

    if err := logger.Initialize(); err != nil {
        log.Fatalf("❌ Failed to initialize logger: %v", err)
    }

    // Log environmental data for 5 minutes
    logger.RunLogging(5 * time.Minute)
}
```

## Advanced Usage

For complex scenarios including multiple listeners, concurrent monitoring, and production patterns, see the [Advanced Usage Guide](docs/ADVANCED_USAGE.md).

Key advanced topics covered:
- **Multiple Listeners and Goroutines**: How to properly use `Listen()` with concurrent processing
- **Fan-Out Message Distribution**: Distributing messages to specialized processors  
- **Multiple Client Architecture**: When and how to use multiple SDK clients
- **Production Patterns**: Resilient monitoring with automatic recovery
- **Performance Optimization**: High-throughput processing and memory efficiency

### Quick Advanced Example: Proper Concurrent System Monitoring

**Important**: When using multiple goroutines with SimConnect, never have multiple goroutines read directly from the same message channel. Each message goes to only ONE goroutine, causing message loss. Use a fan-out pattern instead.

```go
package main

import (
    "fmt"
    "sync"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

func main() {
    sdk := client.New("ConcurrentMonitor")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        panic(err)
    }

    // Register multiple systems
    systems := []struct {
        id   uint32
        name string
    }{
        {1, "PLANE ALTITUDE"},
        {2, "AIRSPEED INDICATED"},
        {3, "ELECTRICAL MASTER BATTERY"},
        {4, "LIGHT BEACON"},
    }

    for _, sys := range systems {
        sdk.RegisterSimVarDefinition(sys.id, sys.name, "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
        sdk.RequestSimVarDataPeriodic(sys.id, sys.id*100, types.SIMCONNECT_PERIOD_SECOND)
    }    // Single Listen() call - shared by all processors
    messages := sdk.Listen()
    
    // Use a fan-out pattern to ensure all processors get all messages
    flightDataCh := make(chan interface{}, 100)
    electricalCh := make(chan interface{}, 100)
    
    var wg sync.WaitGroup
    
    // Message distributor - ensures all relevant messages reach all processors
    wg.Add(1)
    go func() {
        defer wg.Done()
        defer close(flightDataCh)
        defer close(electricalCh)
        
        for msg := range messages {
            if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        // Send flight data to flight processor
                        if simVar.DefineID <= 2 {
                            select {
                            case flightDataCh <- msg:
                            default: // Don't block
                            }
                        }
                        // Send electrical data to electrical processor
                        if simVar.DefineID >= 3 {
                            select {
                            case electricalCh <- msg:
                            default: // Don't block
                            }
                        }
                    }
                }
            }
        }
    }()
    
    // Flight data processor
    wg.Add(1)
    go func() {
        defer wg.Done()
        for msg := range flightDataCh {
            if msgMap, ok := msg.(map[string]any); ok {
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        fmt.Printf("✈️ Flight Data - ID:%d Value:%.1f\n", simVar.DefineID, simVar.Value)
                    }
                }
            }
        }
    }()
    
    // Electrical system processor
    wg.Add(1)
    go func() {
        defer wg.Done()
        for msg := range electricalCh {
            if msgMap, ok := msg.(map[string]any); ok {
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        fmt.Printf("⚡ Electrical - ID:%d Value:%.0f\n", simVar.DefineID, simVar.Value)
                    }
                }
            }
        }
    }()

    time.Sleep(10 * time.Second)
    
    // Cleanup
    for _, sys := range systems {
        sdk.StopPeriodicRequest(sys.id * 100)
    }
    
    wg.Wait()
}
```

### Custom DLL Path

When using non-standard SimConnect installations:

```go
// Use custom SimConnect DLL location
sdk := client.NewWithCustomDLL("MyApp", "D:/CustomPath/SimConnect.dll")
```

## API Reference

Complete API documentation is available in [docs/API.md](docs/API.md).

### Quick Reference

#### Connection Management
```go
sdk := client.New("AppName")                    // Create client
sdk := client.NewWithCustomDLL("App", "path")   // Custom DLL path
err := sdk.Open()                               // Connect
defer sdk.Close()                               // Cleanup
messages := sdk.Listen()                        // Get message channel (call once)
```

#### SimVar Operations
```go
// Register and request data
sdk.RegisterSimVarDefinition(id, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RequestSimVarData(defineID, requestID)                        // One-time request
sdk.RequestSimVarDataPeriodic(defineID, requestID, period)        // Continuous updates
sdk.StopPeriodicRequest(requestID)                                // Stop updates
sdk.SetSimVar(defineID, value)                                    // Set variable value
```

#### Event Management
```go
// System events
sdk.SubscribeToSystemEvent(eventID, "Pause")

// Client events (aircraft control)
sdk.MapClientEventToSimEvent(eventID, "TOGGLE_MASTER_BATTERY")
sdk.AddClientEventToNotificationGroup(groupID, eventID, false)
sdk.SetNotificationGroupPriority(groupID, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST)
sdk.TransmitClientEvent(objectID, eventID, data, groupID, flags)
```

#### Message Processing
```go
for msg := range messages {
    msgMap := msg.(map[string]any)
    switch msgMap["type"] {
    case "SIMOBJECT_DATA":
        simVar := msgMap["parsed_data"].(*client.SimVarData)
        fmt.Printf("ID: %d, Value: %v\n", simVar.DefineID, simVar.Value)
    case "EVENT":
        event := msgMap["event"].(*types.EventData)
        fmt.Printf("Event: %s, Data: %d\n", event.EventName, event.EventData)
    case "EXCEPTION":
        exception := msgMap["exception"]
        fmt.Printf("Exception: %+v\n", exception)
    }
}
```

#### Update Periods
- `SIMCONNECT_PERIOD_NEVER` - Stop updates
- `SIMCONNECT_PERIOD_ONCE` - Single request  
- `SIMCONNECT_PERIOD_VISUAL_FRAME` - Every frame (~30-60 FPS)
- `SIMCONNECT_PERIOD_SECOND` - Every second
- `SIMCONNECT_PERIOD_ON_SET` - When value changes

#### Data Types
- `SIMCONNECT_DATATYPE_INT32` - 32-bit integer
- `SIMCONNECT_DATATYPE_FLOAT32` - 32-bit float
- `SIMCONNECT_DATATYPE_FLOAT64` - 64-bit float  
- `SIMCONNECT_DATATYPE_STRINGV` - Variable-length string

## Documentation

| Document | Description |
|----------|-------------|
| [API.md](docs/API.md) | Complete API reference with all methods, parameters, and examples |
| [ADVANCED_USAGE.md](docs/ADVANCED_USAGE.md) | Advanced patterns including multiple listeners, concurrent monitoring, and production architectures |

### Example Projects

The `example/` directory contains complete working applications:

| Example | Description | Key Features |
|---------|-------------|--------------|
| **external-power-logger** | Monitors external power state with detailed logging | Periodic data requests, statistics tracking, graceful shutdown |
| **sim-webservice** | Web-based real-time monitoring dashboard | HTTP API, multiple system monitoring, concurrent processing |

To run the examples:

```bash
# External power logger
cd example/external-power-logger
go run main.go

# Web service dashboard  
cd example/sim-webservice
go run cmd/server/main.go
# Open http://localhost:8080 in your browser
```

## Performance Considerations

### Update Frequency Guidelines

Choose appropriate update frequencies based on data criticality and system performance:

```go
// Critical flight instruments - high frequency
sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_VISUAL_FRAME) // Altitude
sdk.RequestSimVarDataPeriodic(2, 200, types.SIMCONNECT_PERIOD_VISUAL_FRAME) // Airspeed

// Navigation data - medium frequency  
sdk.RequestSimVarDataPeriodic(3, 300, types.SIMCONNECT_PERIOD_SECOND) // GPS position
sdk.RequestSimVarDataPeriodic(4, 400, types.SIMCONNECT_PERIOD_SECOND) // Fuel quantity

// User settings - only when changed
sdk.RequestSimVarDataPeriodic(5, 500, types.SIMCONNECT_PERIOD_ON_SET) // Autopilot state
```

### Resource Management Best Practices

```go
func efficientMonitoring() {
    sdk := client.New("EfficientApp")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        panic(err)
    }

    // Track request IDs for cleanup
    activeRequests := []uint32{}

    // Register multiple variables efficiently
    variables := []struct {
        id     uint32
        name   string
        period types.SimConnectPeriod
    }{
        {1, "PLANE ALTITUDE", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
        {2, "AIRSPEED INDICATED", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
        {3, "FUEL TOTAL QUANTITY", types.SIMCONNECT_PERIOD_SECOND},
    }

    for _, v := range variables {
        sdk.RegisterSimVarDefinition(v.id, v.name, "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
        requestID := v.id * 100
        sdk.RequestSimVarDataPeriodic(v.id, requestID, v.period)
        activeRequests = append(activeRequests, requestID)
    }

    // Process messages efficiently
    messages := sdk.Listen()
    for msg := range messages {
        processMessageFast(msg) // Keep processing minimal
    }

    // Always cleanup periodic requests
    defer func() {
        for _, requestID := range activeRequests {
            sdk.StopPeriodicRequest(requestID)
        }
    }()
}

func processMessageFast(msg any) {
    // Fast path for common message types
    if msgMap, ok := msg.(map[string]any); ok {
        if msgMap["type"] == "SIMOBJECT_DATA" {
            if data, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := data.(*client.SimVarData); ok {
                    // Process immediately without additional allocations
                    switch simVar.DefineID {
                    case 1: // Altitude
                        updateAltitude(simVar.Value.(float64))
                    case 2: // Airspeed  
                        updateAirspeed(simVar.Value.(float64))
                    }
                }
            }
        }
    }
}
```

### Error Handling and Recovery

```go
func robustConnection() error {
    sdk := client.New("RobustApp")
    defer sdk.Close()

    // Retry connection with exponential backoff
    maxRetries := 5
    for attempt := 1; attempt <= maxRetries; attempt++ {
        if err := sdk.Open(); err != nil {
            if attempt == maxRetries {
                return fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, err)
            }
            
            backoff := time.Duration(attempt) * time.Second
            fmt.Printf("Connection attempt %d failed, retrying in %v...\n", attempt, backoff)
            time.Sleep(backoff)
            continue
        }
        break
    }

    // Monitor for exceptions and handle them appropriately
    messages := sdk.Listen()
    for msg := range messages {
        if msgMap, ok := msg.(map[string]any); ok {
            switch msgMap["type"] {
            case "EXCEPTION":
                if exception, exists := msgMap["exception"]; exists {
                    handleException(exception)
                }
            case "SIMOBJECT_DATA":
                // Normal processing
                processData(msgMap)
            }
        }
    }

    return nil
}

func handleException(exception any) {
    // Log exception details and take appropriate action
    fmt.Printf("⚠️ SimConnect Exception: %+v\n", exception)
    
    // Implement recovery logic based on exception type
    // Some exceptions are recoverable, others require reconnection
}
```

### Memory Efficiency

```go
// Use buffered channels for high-throughput scenarios
func highThroughputProcessing() {
    const bufferSize = 1000
    dataChannel := make(chan *client.SimVarData, bufferSize)
    
    // Process messages in batches
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    var batch []*client.SimVarData
    
    for {
        select {
        case data := <-dataChannel:
            batch = append(batch, data)
            
        case <-ticker.C:
            if len(batch) > 0 {
                processBatch(batch)
                batch = batch[:0] // Reset slice but keep underlying array
            }
        }
    }
}

func processBatch(batch []*client.SimVarData) {
    // Process multiple data points efficiently
    for _, data := range batch {
        // Fast processing without allocations
        _ = data
    }
}
```

## Contributing

_Contributions are welcomed and must follow [Code of Conduct](https://github.com/mycrew-online/sdk?tab=coc-ov-file) and common [Contributions guidelines](https://github.com/mycrew-online/sdk/blob/main/docs/CONTRIBUTING.md)._

> If you'd like to report security issue please follow [security guidelines](https://github.com/mycrew-online/sdk?tab=security-ov-file).

---
<sup><sub>_All rights reserved &copy; mycrew-online and contributors_</sub></sup>
