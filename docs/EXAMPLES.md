# Examples Guide

Complete working code samples for common SimConnect scenarios, from basic monitoring to advanced applications.

> **Note**: This guide contains code snippets and samples. For a complete working project example, see the **sim-webservice** in the `example/` directory.

## Table of Contents

- [Basic Examples](#basic-examples)
- [Flight Monitoring](#flight-monitoring)
- [Aircraft Control](#aircraft-control)
- [System Events](#system-events)
- [Advanced Patterns](#advanced-patterns)
- [Real-World Applications](#real-world-applications)

## Basic Examples

### Simple Altitude Monitor

Monitor aircraft altitude with basic error handling:

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

func main() {
    sdk := client.New("AltitudeMonitor")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    fmt.Println("✅ Connected to MSFS")

    // Register altitude variable
    err := sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
    if err != nil {
        log.Fatalf("Failed to register altitude: %v", err)
    }

    // Request updates every second
    err = sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_SECOND)
    if err != nil {
        log.Fatalf("Failed to request data: %v", err)
    }

    fmt.Println("📊 Monitoring altitude (Ctrl+C to stop)...")

    messages := sdk.Listen()
    for msg := range messages {
        if msgMap, ok := msg.(map[string]any); ok {
            switch msgMap["type"] {
            case "SIMOBJECT_DATA":
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        fmt.Printf("✈️ Altitude: %.0f feet\n", simVar.Value)
                    }
                }
            case "EXCEPTION":
                if exception, exists := msgMap["exception"]; exists {
                    fmt.Printf("⚠️ Exception: %+v\n", exception)
                }
            }
        }
    }
}
```

### Multi-Variable Basic Monitor

Monitor multiple flight parameters simultaneously:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

type FlightData struct {
    Altitude  float64
    Airspeed  float64
    Heading   float64
    VSpeed    float64
}

func main() {
    sdk := client.New("BasicMonitor")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        log.Fatalf("Connection failed: %v", err)
    }

    // Register flight instruments
    instruments := map[uint32]string{
        1: "PLANE ALTITUDE",
        2: "AIRSPEED INDICATED", 
        3: "HEADING INDICATOR",
        4: "VERTICAL SPEED",
    }

    for id, varName := range instruments {
        units := "feet"
        if id == 2 || id == 4 { // Airspeed and VS
            units = "knots"
        } else if id == 3 { // Heading
            units = "degrees"
        }

        err := sdk.RegisterSimVarDefinition(id, varName, units, types.SIMCONNECT_DATATYPE_FLOAT32)
        if err != nil {
            log.Fatalf("Failed to register %s: %v", varName, err)
        }

        err = sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
        if err != nil {
            log.Fatalf("Failed to request %s: %v", varName, err)
        }
    }

    // Graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    flightData := &FlightData{}
    messages := sdk.Listen()

    fmt.Println("📊 Basic Flight Monitor Active")
    fmt.Println("Press Ctrl+C to stop")

    for {
        select {
        case <-sigChan:
            fmt.Println("\n🛑 Stopping monitor...")
            for id := range instruments {
                sdk.StopPeriodicRequest(id * 100)
            }
            return

        case msg := <-messages:
            if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        updateFlightData(flightData, simVar)
                        displayFlightData(flightData)
                    }
                }
            }
        }
    }
}

func updateFlightData(fd *FlightData, simVar *client.SimVarData) {
    switch simVar.DefineID {
    case 1:
        fd.Altitude = simVar.Value.(float64)
    case 2:
        fd.Airspeed = simVar.Value.(float64)
    case 3:
        fd.Heading = simVar.Value.(float64)
    case 4:
        fd.VSpeed = simVar.Value.(float64)
    }
}

func displayFlightData(fd *FlightData) {
    fmt.Printf("\r\033[K✈️ ALT: %6.0fft | SPD: %3.0fkts | HDG: %03.0f° | VS: %+5.0ffpm",
        fd.Altitude, fd.Airspeed, fd.Heading, fd.VSpeed)
}
```

## Flight Monitoring

### Real-Time Flight Dashboard

A comprehensive flight monitoring application:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
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
    if err := fd.sdk.Open(); err != nil {
        return fmt.Errorf("connection failed: %v", err)
    }

    // Define comprehensive flight instruments
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
        {7, "GPS GROUND SPEED", "knots", "Ground Speed", types.SIMCONNECT_PERIOD_SECOND},
        {8, "FUEL TOTAL QUANTITY", "gallons", "Fuel Total", types.SIMCONNECT_PERIOD_SECOND},
        {9, "ENG RPM", "rpm", "Engine RPM", types.SIMCONNECT_PERIOD_SECOND},
        {10, "GEAR HANDLE POSITION", "Bool", "Gear", types.SIMCONNECT_PERIOD_ON_SET},
    }

    fmt.Println("📊 Registering flight instruments...")
    for _, instrument := range instruments {
        dataType := types.SIMCONNECT_DATATYPE_FLOAT32
        if instrument.units == "Bool" {
            dataType = types.SIMCONNECT_DATATYPE_INT32
        }

        if err := fd.sdk.RegisterSimVarDefinition(
            instrument.id, 
            instrument.name, 
            instrument.units, 
            dataType,
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
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    messages := fd.sdk.Listen()
    if messages == nil {
        log.Fatal("❌ Failed to start message listening")
    }

    displayTicker := time.NewTicker(500 * time.Millisecond)
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
    dataLabels := map[uint32]string{
        1:  "altitude",     2:  "airspeed",    3:  "heading",
        4:  "verticalSpeed", 5:  "pitch",      6:  "bank",
        7:  "groundSpeed",   8:  "fuelTotal",  9:  "engineRPM",
        10: "gear",
    }

    if label, exists := dataLabels[simVar.DefineID]; exists {
        fd.flightData[label] = simVar.Value
        fd.lastUpdate = time.Now()
    }
}

func (fd *FlightDashboard) handleSystemEvent(event *types.EventData) {
    // Handle system events (pause, crash, etc.)
    fmt.Printf("\n📡 System Event: ID %d, Data: %d\n", event.EventID, event.EventData)
}

func (fd *FlightDashboard) displayDashboard() {
    if len(fd.flightData) == 0 {
        fmt.Printf("\r⏳ Waiting for flight data...")
        return
    }

    fmt.Printf("\r\033[K") // Clear line
    
    // Primary flight display
    altitude := fd.getFloat("altitude")
    airspeed := fd.getFloat("airspeed")
    heading := fd.getFloat("heading")
    verticalSpeed := fd.getFloat("verticalSpeed")
    
    fmt.Printf("✈️ ALT: %6.0fft | SPD: %3.0fkts | HDG: %03.0f° | VS: %+5.0ffpm",
        altitude, airspeed, heading, verticalSpeed)
    
    // Additional data
    if time.Since(fd.lastUpdate) < 2*time.Second {
        pitch := fd.getFloat("pitch")
        bank := fd.getFloat("bank")
        engineRPM := fd.getFloat("engineRPM")
        fuel := fd.getFloat("fuelTotal")
        
        fmt.Printf(" | ATT: %+4.1f°/%+4.1f° | RPM: %.0f | FUEL: %.1fgal",
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

## Aircraft Control

### Basic Aircraft Systems Controller

Control aircraft electrical and lighting systems:

```go
package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

type AircraftController struct {
    sdk          client.Connection
    systemStates map[string]bool
}

const (
    EVENT_TOGGLE_MASTER_BATTERY = 10001
    EVENT_TOGGLE_BEACON_LIGHTS  = 10002
    EVENT_TOGGLE_NAV_LIGHTS     = 10003
    EVENT_TOGGLE_STROBE_LIGHTS  = 10004
    EVENT_TOGGLE_TAXI_LIGHTS    = 10005
    EVENT_TOGGLE_LANDING_LIGHTS = 10006
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
        {11, "ELECTRICAL MASTER BATTERY", "MasterBattery"},
        {12, "LIGHT BEACON", "BeaconLights"},
        {13, "LIGHT NAV", "NavLights"},
        {14, "LIGHT STROBE", "StrobeLights"},
        {15, "LIGHT TAXI", "TaxiLights"},
        {16, "LIGHT LANDING", "LandingLights"},
        {17, "ELECTRICAL MAIN BUS VOLTAGE", "BusVoltage"},
    }

    fmt.Println("📊 Registering system monitoring...")
    for _, sys := range systemVars {
        dataType := types.SIMCONNECT_DATATYPE_INT32
        units := "Bool"
        
        if sys.id == 17 { // Voltage is float
            dataType = types.SIMCONNECT_DATATYPE_FLOAT32
            units = "Volts"
        }

        if err := ac.sdk.RegisterSimVarDefinition(sys.id, sys.name, units, dataType); err != nil {
            return fmt.Errorf("failed to register %s: %v", sys.label, err)
        }

        if err := ac.sdk.RequestSimVarDataPeriodic(sys.id, sys.id*100, types.SIMCONNECT_PERIOD_ON_SET); err != nil {
            return fmt.Errorf("failed to request %s data: %v", sys.label, err)
        }
    }

    // Set up control events
    controlEvents := map[types.ClientEventID]string{
        EVENT_TOGGLE_MASTER_BATTERY: "TOGGLE_MASTER_BATTERY",
        EVENT_TOGGLE_BEACON_LIGHTS:  "TOGGLE_BEACON_LIGHTS",
        EVENT_TOGGLE_NAV_LIGHTS:     "TOGGLE_NAV_LIGHTS",
        EVENT_TOGGLE_STROBE_LIGHTS:  "TOGGLE_STROBE_LIGHTS",
        EVENT_TOGGLE_TAXI_LIGHTS:    "TOGGLE_TAXI_LIGHTS",
        EVENT_TOGGLE_LANDING_LIGHTS: "TOGGLE_LANDING_LIGHTS",
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

func (ac *AircraftController) RunInteractive() {
    messages := ac.sdk.Listen()
    
    // Process messages in background
    go func() {
        for msg := range messages {
            ac.processMessage(msg)
        }
    }()

    scanner := bufio.NewScanner(os.Stdin)
    
    fmt.Println("\n🎮 Aircraft Systems Controller")
    fmt.Println("Commands: 1=Battery, 2=Beacon, 3=Nav, 4=Strobe, 5=Taxi, 6=Landing, s=Status, q=Quit")
    
    for {
        fmt.Print("\n> ")
        if !scanner.Scan() {
            break
        }
        
        command := strings.TrimSpace(scanner.Text())
        if command == "q" || command == "quit" {
            break
        }
        
        if command == "s" || command == "status" {
            ac.displayStatus()
            continue
        }
        
        if num, err := strconv.Atoi(command); err == nil {
            ac.executeCommand(num)
        } else {
            fmt.Println("❌ Invalid command")
        }
    }
    
    fmt.Println("👋 Controller stopped")
}

func (ac *AircraftController) executeCommand(command int) {
    var eventID types.ClientEventID
    var description string
    
    switch command {
    case 1:
        eventID = EVENT_TOGGLE_MASTER_BATTERY
        description = "Master Battery"
    case 2:
        eventID = EVENT_TOGGLE_BEACON_LIGHTS
        description = "Beacon Lights"
    case 3:
        eventID = EVENT_TOGGLE_NAV_LIGHTS
        description = "Nav Lights"
    case 4:
        eventID = EVENT_TOGGLE_STROBE_LIGHTS
        description = "Strobe Lights"
    case 5:
        eventID = EVENT_TOGGLE_TAXI_LIGHTS
        description = "Taxi Lights"
    case 6:
        eventID = EVENT_TOGGLE_LANDING_LIGHTS
        description = "Landing Lights"
    default:
        fmt.Println("❌ Invalid command number")
        return
    }

    err := ac.sdk.TransmitClientEvent(
        types.SIMCONNECT_OBJECT_ID_USER,
        eventID,
        0,
        3000,
        types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY,
    )
    
    if err != nil {
        fmt.Printf("❌ Failed to toggle %s: %v\n", description, err)
    } else {
        fmt.Printf("🎯 Toggled %s\n", description)
    }
}

func (ac *AircraftController) processMessage(msg any) {
    msgMap, ok := msg.(map[string]any)
    if !ok {
        return
    }

    if msgMap["type"] == "SIMOBJECT_DATA" {
        if parsedData, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := parsedData.(*client.SimVarData); ok {
                ac.updateSystemState(simVar)
            }
        }
    }
}

func (ac *AircraftController) updateSystemState(simVar *client.SimVarData) {
    stateLabels := map[uint32]string{
        11: "battery", 12: "beacon", 13: "nav",
        14: "strobe", 15: "taxi", 16: "landing",
    }

    if label, exists := stateLabels[simVar.DefineID]; exists {
        if intVal, ok := simVar.Value.(int32); ok {
            ac.systemStates[label] = intVal != 0
        }
    }
}

func (ac *AircraftController) displayStatus() {
    fmt.Println("\n📊 Current System Status:")
    systems := []struct {
        key   string
        label string
    }{
        {"battery", "Master Battery"},
        {"beacon", "Beacon Lights"},
        {"nav", "Nav Lights"},
        {"strobe", "Strobe Lights"},
        {"taxi", "Taxi Lights"},
        {"landing", "Landing Lights"},
    }

    for _, sys := range systems {
        status := "❌ OFF"
        if ac.systemStates[sys.key] {
            status = "✅ ON"
        }
        fmt.Printf("  %s: %s\n", sys.label, status)
    }
}

func main() {
    controller := NewAircraftController()
    defer controller.sdk.Close()

    if err := controller.Initialize(); err != nil {
        log.Fatalf("❌ Failed to initialize controller: %v", err)
    }

    controller.RunInteractive()
}
```

## System Events

### Event Monitor and Handler

Monitor and respond to simulator system events:

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

type EventMonitor struct {
    sdk           client.Connection
    eventCounts   map[string]int
    lastEventTime map[string]time.Time
}

func NewEventMonitor() *EventMonitor {
    return &EventMonitor{
        sdk:           client.New("EventMonitor"),
        eventCounts:   make(map[string]int),
        lastEventTime: make(map[string]time.Time),
    }
}

func (em *EventMonitor) Initialize() error {
    if err := em.sdk.Open(); err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }

    // Subscribe to comprehensive set of system events
    systemEvents := []struct {
        id   uint32
        name string
        desc string
    }{
        {1001, "Pause", "Simulator pause state"},
        {1002, "Crashed", "Aircraft crash events"},
        {1003, "AircraftLoaded", "New aircraft loaded"},
        {1004, "FlightLoaded", "Flight plan loaded"},
        {1005, "FlightSaved", "Flight state saved"},
        {1006, "FlightPlanActivated", "Flight plan activated"},
        {1007, "FlightPlanDeactivated", "Flight plan deactivated"},
        {1008, "Sim", "Simulator state changes"},
        {1009, "Sound", "Sound system events"},
        {1010, "View", "View/camera changes"},
        {1011, "WeatherModeChanged", "Weather mode changed"},
        {1012, "AirportListReceived", "Airport data received"},
        {1013, "VorListReceived", "VOR data received"},
        {1014, "NdbListReceived", "NDB data received"},
        {1015, "WaypointListReceived", "Waypoint data received"},
    }

    fmt.Println("📡 Subscribing to system events...")
    for _, event := range systemEvents {
        if err := em.sdk.SubscribeToSystemEvent(event.id, event.name); err != nil {
            log.Printf("⚠️ Failed to subscribe to %s: %v", event.name, err)
        } else {
            fmt.Printf("  ✅ Subscribed to %s (%s)\n", event.name, event.desc)
        }
    }

    fmt.Println("✅ Event monitor initialized")
    return nil
}

func (em *EventMonitor) Run() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    messages := em.sdk.Listen()
    if messages == nil {
        log.Fatal("❌ Failed to start message listening")
    }

    // Statistics ticker
    statsTicker := time.NewTicker(30 * time.Second)
    defer statsTicker.Stop()

    fmt.Println("\n🎯 Event Monitor Active - Press Ctrl+C to stop")
    fmt.Println("Waiting for system events...")

    for {
        select {
        case <-sigChan:
            fmt.Println("\n🛑 Stopping event monitor...")
            em.displayFinalStats()
            return

        case <-statsTicker.C:
            em.displayStats()

        case msg, ok := <-messages:
            if !ok {
                fmt.Println("❌ Message channel closed")
                return
            }
            em.processMessage(msg)
        }
    }
}

func (em *EventMonitor) processMessage(msg any) {
    msgMap, ok := msg.(map[string]any)
    if !ok {
        return
    }

    switch msgMap["type"] {
    case "EVENT":
        if eventData, exists := msgMap["event"]; exists {
            if event, ok := eventData.(*types.EventData); ok {
                em.handleEvent(event)
            }
        }
    case "EXCEPTION":
        if exception, exists := msgMap["exception"]; exists {
            fmt.Printf("❌ SimConnect Exception: %+v\n", exception)
        }
    }
}

func (em *EventMonitor) handleEvent(event *types.EventData) {
    eventName := em.getEventName(event.EventID)
    if eventName == "" {
        eventName = fmt.Sprintf("Unknown_%d", event.EventID)
    }

    // Update statistics
    em.eventCounts[eventName]++
    em.lastEventTime[eventName] = time.Now()

    // Handle specific events
    switch event.EventID {
    case 1001: // Pause
        status := "RESUMED"
        if event.EventData == 1 {
            status = "PAUSED"
        }
        fmt.Printf("⏸️ Simulator %s (Data: %d)\n", status, event.EventData)

    case 1002: // Crashed
        fmt.Printf("💥 AIRCRAFT CRASHED! (Type: %d)\n", event.EventData)

    case 1003: // AircraftLoaded
        fmt.Printf("✈️ Aircraft loaded (ID: %d)\n", event.EventData)

    case 1004: // FlightLoaded
        fmt.Printf("🛫 Flight plan loaded (ID: %d)\n", event.EventData)

    case 1005: // FlightSaved
        fmt.Printf("💾 Flight saved (ID: %d)\n", event.EventData)

    case 1006: // FlightPlanActivated
        fmt.Printf("🎯 Flight plan activated (ID: %d)\n", event.EventData)

    case 1007: // FlightPlanDeactivated
        fmt.Printf("❌ Flight plan deactivated (ID: %d)\n", event.EventData)

    case 1008: // Sim
        fmt.Printf("🔄 Simulator state changed (State: %d)\n", event.EventData)

    case 1011: // WeatherModeChanged
        weatherModes := map[uint32]string{
            0: "User-defined",
            1: "Real weather",
            2: "Live weather",
        }
        mode := weatherModes[event.EventData]
        if mode == "" {
            mode = fmt.Sprintf("Unknown (%d)", event.EventData)
        }
        fmt.Printf("🌦️ Weather mode changed to: %s\n", mode)

    default:
        fmt.Printf("📡 Event: %s (ID: %d, Data: %d)\n", eventName, event.EventID, event.EventData)
    }
}

func (em *EventMonitor) getEventName(eventID uint32) string {
    eventNames := map[uint32]string{
        1001: "Pause", 1002: "Crashed", 1003: "AircraftLoaded",
        1004: "FlightLoaded", 1005: "FlightSaved", 1006: "FlightPlanActivated",
        1007: "FlightPlanDeactivated", 1008: "Sim", 1009: "Sound",
        1010: "View", 1011: "WeatherModeChanged", 1012: "AirportListReceived",
        1013: "VorListReceived", 1014: "NdbListReceived", 1015: "WaypointListReceived",
    }
    return eventNames[eventID]
}

func (em *EventMonitor) displayStats() {
    if len(em.eventCounts) == 0 {
        fmt.Println("📊 No events received yet...")
        return
    }

    fmt.Println("\n📊 Event Statistics (Last 30 seconds):")
    for eventName, count := range em.eventCounts {
        lastSeen := em.lastEventTime[eventName]
        timeSince := time.Since(lastSeen)
        fmt.Printf("  %s: %d events (last: %v ago)\n", eventName, count, timeSince.Truncate(time.Second))
    }
    fmt.Println()
}

func (em *EventMonitor) displayFinalStats() {
    fmt.Println("\n📊 Final Event Statistics:")
    if len(em.eventCounts) == 0 {
        fmt.Println("  No events were received during this session")
        return
    }

    totalEvents := 0
    for eventName, count := range em.eventCounts {
        totalEvents += count
        fmt.Printf("  %s: %d events\n", eventName, count)
    }
    fmt.Printf("\nTotal events processed: %d\n", totalEvents)
}

func main() {
    monitor := NewEventMonitor()
    defer monitor.sdk.Close()

    if err := monitor.Initialize(); err != nil {
        log.Fatalf("❌ Failed to initialize event monitor: %v", err)
    }

    monitor.Run()
}
```

## Advanced Patterns

### Concurrent Data Processing with Fan-Out

Proper concurrent processing using the fan-out pattern:

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

func main() {
    sdk := client.New("ConcurrentProcessor")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        log.Fatalf("Connection failed: %v", err)
    }

    // Register diverse data sources
    dataSources := []struct {
        id     uint32
        name   string
        units  string
        category string
    }{
        {1, "PLANE ALTITUDE", "feet", "flight"},
        {2, "AIRSPEED INDICATED", "knots", "flight"},
        {3, "HEADING INDICATOR", "degrees", "navigation"},
        {4, "GPS GROUND SPEED", "knots", "navigation"},
        {5, "ELECTRICAL MASTER BATTERY", "Bool", "electrical"},
        {6, "LIGHT BEACON", "Bool", "electrical"},
        {7, "ENG RPM", "rpm", "engine"},
        {8, "FUEL TOTAL QUANTITY", "gallons", "engine"},
    }

    for _, ds := range dataSources {
        dataType := types.SIMCONNECT_DATATYPE_FLOAT32
        if ds.units == "Bool" {
            dataType = types.SIMCONNECT_DATATYPE_INT32
        }

        sdk.RegisterSimVarDefinition(ds.id, ds.name, ds.units, dataType)
        sdk.RequestSimVarDataPeriodic(ds.id, ds.id*100, types.SIMCONNECT_PERIOD_SECOND)
    }

    // Single Listen() call - critical for proper message handling
    messages := sdk.Listen()
    
    // Create specialized channels for different data categories
    flightDataCh := make(chan interface{}, 100)
    navigationCh := make(chan interface{}, 100)
    electricalCh := make(chan interface{}, 100)
    engineCh := make(chan interface{}, 100)
    
    var wg sync.WaitGroup
    
    // Message distributor - ensures all relevant messages reach specialized processors
    wg.Add(1)
    go func() {
        defer wg.Done()
        defer close(flightDataCh)
        defer close(navigationCh)
        defer close(electricalCh)
        defer close(engineCh)
        
        for msg := range messages {
            if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        // Route messages based on data category
                        switch simVar.DefineID {
                        case 1, 2: // Flight data
                            select {
                            case flightDataCh <- msg:
                            default: // Don't block
                            }
                        case 3, 4: // Navigation data
                            select {
                            case navigationCh <- msg:
                            default:
                            }
                        case 5, 6: // Electrical data
                            select {
                            case electricalCh <- msg:
                            default:
                            }
                        case 7, 8: // Engine data
                            select {
                            case engineCh <- msg:
                            default:
                            }
                        }
                    }
                }
            }
        }
    }()
    
    // Specialized processors - each handles specific data types
    wg.Add(1)
    go flightDataProcessor(flightDataCh, &wg)
    
    wg.Add(1)
    go navigationProcessor(navigationCh, &wg)
    
    wg.Add(1)
    go electricalProcessor(electricalCh, &wg)
    
    wg.Add(1)
    go engineProcessor(engineCh, &wg)

    fmt.Println("🎯 Concurrent processors started - running for 30 seconds...")
    time.Sleep(30 * time.Second)
    
    // Cleanup
    for _, ds := range dataSources {
        sdk.StopPeriodicRequest(ds.id * 100)
    }
    
    fmt.Println("🛑 Stopping processors...")
    wg.Wait()
    fmt.Println("✅ All processors stopped")
}

func flightDataProcessor(ch <-chan interface{}, wg *sync.WaitGroup) {
    defer wg.Done()
    
    flightState := make(map[string]float64)
    
    for msg := range ch {
        if msgMap, ok := msg.(map[string]any); ok {
            if data, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := data.(*client.SimVarData); ok {
                    switch simVar.DefineID {
                    case 1:
                        flightState["altitude"] = simVar.Value.(float64)
                    case 2:
                        flightState["airspeed"] = simVar.Value.(float64)
                    }
                    
                    // Process flight envelope analysis
                    if alt, hasAlt := flightState["altitude"]; hasAlt {
                        if spd, hasSpd := flightState["airspeed"]; hasSpd {
                            analyzeFlightEnvelope(alt, spd)
                        }
                    }
                }
            }
        }
    }
    fmt.Println("✅ Flight data processor stopped")
}

func navigationProcessor(ch <-chan interface{}, wg *sync.WaitGroup) {
    defer wg.Done()
    
    for msg := range ch {
        if msgMap, ok := msg.(map[string]any); ok {
            if data, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := data.(*client.SimVarData); ok {
                    switch simVar.DefineID {
                    case 3:
                        fmt.Printf("🧭 Navigation: Heading %.0f°\n", simVar.Value)
                    case 4:
                        fmt.Printf("🧭 Navigation: Ground Speed %.0f kts\n", simVar.Value)
                    }
                }
            }
        }
    }
    fmt.Println("✅ Navigation processor stopped")
}

func electricalProcessor(ch <-chan interface{}, wg *sync.WaitGroup) {
    defer wg.Done()
    
    electricalState := make(map[string]bool)
    
    for msg := range ch {
        if msgMap, ok := msg.(map[string]any); ok {
            if data, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := data.(*client.SimVarData); ok {
                    switch simVar.DefineID {
                    case 5:
                        electricalState["battery"] = simVar.Value.(int32) != 0
                        fmt.Printf("⚡ Electrical: Battery %s\n", formatOnOff(simVar.Value.(int32) != 0))
                    case 6:
                        electricalState["beacon"] = simVar.Value.(int32) != 0
                        fmt.Printf("⚡ Electrical: Beacon %s\n", formatOnOff(simVar.Value.(int32) != 0))
                    }
                    
                    // Check electrical system status
                    checkElectricalWarnings(electricalState)
                }
            }
        }
    }
    fmt.Println("✅ Electrical processor stopped")
}

func engineProcessor(ch <-chan interface{}, wg *sync.WaitGroup) {
    defer wg.Done()
    
    engineState := make(map[string]float64)
    
    for msg := range ch {
        if msgMap, ok := msg.(map[string]any); ok {
            if data, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := data.(*client.SimVarData); ok {
                    switch simVar.DefineID {
                    case 7:
                        engineState["rpm"] = simVar.Value.(float64)
                        fmt.Printf("🔧 Engine: RPM %.0f\n", simVar.Value)
                    case 8:
                        engineState["fuel"] = simVar.Value.(float64)
                        fmt.Printf("🔧 Engine: Fuel %.1f gal\n", simVar.Value)
                    }
                    
                    // Monitor engine health
                    monitorEngineHealth(engineState)
                }
            }
        }
    }
    fmt.Println("✅ Engine processor stopped")
}

func analyzeFlightEnvelope(altitude, airspeed float64) {
    if altitude > 10000 && airspeed < 100 {
        fmt.Printf("⚠️ Flight Analysis: High altitude (%.0fft) with low speed (%.0fkts)\n", altitude, airspeed)
    }
}

func formatOnOff(state bool) string {
    if state {
        return "ON"
    }
    return "OFF"
}

func checkElectricalWarnings(state map[string]bool) {
    if beacon, hasBeacon := state["beacon"]; hasBeacon {
        if battery, hasBattery := state["battery"]; hasBattery {
            if beacon && !battery {
                fmt.Println("⚠️ Electrical Warning: Beacon on without battery!")
            }
        }
    }
}

func monitorEngineHealth(state map[string]float64) {
    if rpm, hasRPM := state["rpm"]; hasRPM {
        if fuel, hasFuel := state["fuel"]; hasFuel {
            if rpm > 2000 && fuel < 5 {
                fmt.Printf("⚠️ Engine Warning: High RPM (%.0f) with low fuel (%.1f gal)\n", rpm, fuel)
            }
        }
    }
}
```

## Real-World Applications

### Data Logger with CSV Export

A production-ready flight data logger:

```go
package main

import (
    "encoding/csv"
    "fmt"
    "log"
    "os"
    "strconv"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

type FlightLogger struct {
    sdk        client.Connection
    csvWriter  *csv.Writer
    logFile    *os.File
    startTime  time.Time
    recordCount int
}

type LogRecord struct {
    Timestamp    time.Time
    Altitude     float64
    Airspeed     float64
    Heading      float64
    VerticalSpeed float64
    Pitch        float64
    Bank         float64
    GroundSpeed  float64
    FuelTotal    float64
}

func NewFlightLogger() *FlightLogger {
    return &FlightLogger{
        sdk:       client.New("FlightLogger"),
        startTime: time.Now(),
    }
}

func (fl *FlightLogger) Initialize() error {
    if err := fl.sdk.Open(); err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }

    // Create log file with timestamp
    filename := fmt.Sprintf("flight_log_%s.csv", fl.startTime.Format("2006-01-02_15-04-05"))
    var err error
    fl.logFile, err = os.Create(filename)
    if err != nil {
        return fmt.Errorf("failed to create log file: %v", err)
    }

    fl.csvWriter = csv.NewWriter(fl.logFile)
    
    // Write CSV header
    header := []string{
        "Timestamp", "Elapsed_Seconds", "Altitude_ft", "Airspeed_kts", 
        "Heading_deg", "VerticalSpeed_fpm", "Pitch_deg", "Bank_deg",
        "GroundSpeed_kts", "FuelTotal_gal",
    }
    if err := fl.csvWriter.Write(header); err != nil {
        return fmt.Errorf("failed to write CSV header: %v", err)
    }
    fl.csvWriter.Flush()

    // Register flight data variables
    variables := []struct {
        id    uint32
        name  string
        units string
    }{
        {1, "PLANE ALTITUDE", "feet"},
        {2, "AIRSPEED INDICATED", "knots"},
        {3, "HEADING INDICATOR", "degrees"},
        {4, "VERTICAL SPEED", "feet per minute"},
        {5, "PLANE PITCH DEGREES", "degrees"},
        {6, "PLANE BANK DEGREES", "degrees"},
        {7, "GPS GROUND SPEED", "knots"},
        {8, "FUEL TOTAL QUANTITY", "gallons"},
    }

    for _, v := range variables {
        if err := fl.sdk.RegisterSimVarDefinition(v.id, v.name, v.units, types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
            return fmt.Errorf("failed to register %s: %v", v.name, err)
        }

        // Log at 1 Hz for reasonable file size
        if err := fl.sdk.RequestSimVarDataPeriodic(v.id, v.id*100, types.SIMCONNECT_PERIOD_SECOND); err != nil {
            return fmt.Errorf("failed to request %s: %v", v.name, err)
        }
    }

    fmt.Printf("✅ Flight logger initialized\n")
    fmt.Printf("📝 Logging to: %s\n", filename)
    return nil
}

func (fl *FlightLogger) Run(duration time.Duration) {
    messages := fl.sdk.Listen()
    
    currentRecord := &LogRecord{}
    timeout := time.After(duration)
    logTicker := time.NewTicker(1 * time.Second)
    defer logTicker.Stop()

    fmt.Printf("📊 Logging flight data for %v...\n", duration)

    for {
        select {
        case <-timeout:
            fmt.Println("⏰ Logging duration completed")
            return

        case msg := <-messages:
            fl.processMessage(msg, currentRecord)

        case <-logTicker.C:
            fl.writeRecord(currentRecord)
        }
    }
}

func (fl *FlightLogger) processMessage(msg any, record *LogRecord) {
    msgMap, ok := msg.(map[string]any)
    if !ok {
        return
    }

    if msgMap["type"] == "SIMOBJECT_DATA" {
        if parsedData, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := parsedData.(*client.SimVarData); ok {
                value := simVar.Value.(float64)
                
                switch simVar.DefineID {
                case 1:
                    record.Altitude = value
                case 2:
                    record.Airspeed = value
                case 3:
                    record.Heading = value
                case 4:
                    record.VerticalSpeed = value
                case 5:
                    record.Pitch = value
                case 6:
                    record.Bank = value
                case 7:
                    record.GroundSpeed = value
                case 8:
                    record.FuelTotal = value
                }
            }
        }
    }
}

func (fl *FlightLogger) writeRecord(record *LogRecord) {
    now := time.Now()
    elapsed := now.Sub(fl.startTime).Seconds()
    
    csvRecord := []string{
        now.Format(time.RFC3339),
        fmt.Sprintf("%.1f", elapsed),
        fmt.Sprintf("%.1f", record.Altitude),
        fmt.Sprintf("%.1f", record.Airspeed),
        fmt.Sprintf("%.1f", record.Heading),
        fmt.Sprintf("%.0f", record.VerticalSpeed),
        fmt.Sprintf("%.2f", record.Pitch),
        fmt.Sprintf("%.2f", record.Bank),
        fmt.Sprintf("%.1f", record.GroundSpeed),
        fmt.Sprintf("%.2f", record.FuelTotal),
    }

    if err := fl.csvWriter.Write(csvRecord); err != nil {
        log.Printf("❌ Failed to write CSV record: %v", err)
        return
    }
    
    fl.csvWriter.Flush()
    fl.recordCount++
    
    // Display progress
    fmt.Printf("\r📝 Records logged: %d | Alt: %.0fft | Spd: %.0fkts | Fuel: %.1fgal",
        fl.recordCount, record.Altitude, record.Airspeed, record.FuelTotal)
}

func (fl *FlightLogger) Cleanup() {
    // Stop all periodic requests
    for i := uint32(1); i <= 8; i++ {
        if err := fl.sdk.StopPeriodicRequest(i * 100); err != nil {
            log.Printf("⚠️ Failed to stop request %d: %v", i*100, err)
        }
    }

    // Close files
    if fl.csvWriter != nil {
        fl.csvWriter.Flush()
    }
    if fl.logFile != nil {
        fl.logFile.Close()
    }

    duration := time.Since(fl.startTime)
    fmt.Printf("\n✅ Logging complete: %d records in %v\n", fl.recordCount, duration.Truncate(time.Second))
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go <duration_minutes>")
        fmt.Println("Example: go run main.go 5  # Log for 5 minutes")
        os.Exit(1)
    }

    minutes, err := strconv.Atoi(os.Args[1])
    if err != nil || minutes <= 0 {
        fmt.Println("❌ Invalid duration. Please provide a positive number of minutes.")
        os.Exit(1)
    }

    logger := NewFlightLogger()
    defer logger.sdk.Close()
    defer logger.Cleanup()

    if err := logger.Initialize(); err != nil {
        log.Fatalf("❌ Failed to initialize logger: %v", err)
    }

    duration := time.Duration(minutes) * time.Minute
    logger.Run(duration)
}
```

This examples guide provides a comprehensive set of working code samples that demonstrate the key capabilities of the SDK, from simple monitoring to complex concurrent processing and real-world applications. Each example is complete and runnable, with proper error handling and cleanup patterns.
