# Multiple Clients Architecture Guide

Guide for using multiple SDK instances effectively, including when to use multiple clients and how to implement them properly.

## Table of Contents

- [When to Use Multiple Clients](#when-to-use-multiple-clients)
- [Single vs Multiple Client Decision](#single-vs-multiple-client-decision)
- [Architecture Patterns](#architecture-patterns)
- [Implementation Examples](#implementation-examples)
- [Performance Considerations](#performance-considerations)
- [Best Practices](#best-practices)

## When to Use Multiple Clients

### Valid Use Cases

✅ **Functional Separation**
- One client for flight data monitoring
- Another client for aircraft control
- Separate client for system diagnostics

✅ **Different Update Requirements**
- High-frequency client for critical instruments (60 FPS)
- Low-frequency client for system states (1 Hz)
- Event-only client for system notifications

✅ **Isolation Requirements**
- Separate client per aircraft in multi-aircraft scenarios
- Isolated clients for different subsystems
- Independent clients for different user interfaces

✅ **Load Distribution**
- Distribute processing across multiple threads
- Separate clients for heavy computational tasks
- Balance SimConnect load across multiple connections

### Invalid Use Cases

❌ **Don't Use Multiple Clients For**
- Working around message distribution issues
- Trying to get multiple copies of the same data
- Avoiding proper concurrent programming patterns
- Splitting a single logical application unnecessarily

## Single vs Multiple Client Decision

### Decision Matrix

| Scenario | Single Client | Multiple Clients | Reason |
|----------|---------------|------------------|---------|
| Basic flight monitoring | ✅ | ❌ | Simple fan-out pattern sufficient |
| Real-time + logging | ✅ | ❌ | Same data, different processing |
| Flight data + aircraft control | ❌ | ✅ | Different functional domains |
| Multiple aircraft tracking | ❌ | ✅ | Logical separation needed |
| Web dashboard + background logging | ❌ | ✅ | Different update frequencies |
| High-freq instruments + low-freq systems | ❌ | ✅ | Performance optimization |

### Single Client Pattern (Recommended for Most Cases)

```go
func singleClientPattern() {
    sdk := client.New("UnifiedApp")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        log.Fatal(err)
    }

    // Register all data needs
    registerFlightInstruments(sdk)
    registerSystemStates(sdk)
    registerEngineData(sdk)

    // Single message stream
    messages := sdk.Listen()
    
    // Fan-out to specialized processors
    flightCh := make(chan interface{}, 100)
    systemCh := make(chan interface{}, 100)
    engineCh := make(chan interface{}, 100)

    // Single dispatcher
    go distributeMessages(messages, flightCh, systemCh, engineCh)
    
    // Specialized processors
    go processFlightData(flightCh)
    go processSystemData(systemCh)
    go processEngineData(engineCh)
}
```

### Multiple Client Pattern (When Justified)

```go
func multipleClientPattern() {
    // Client 1: High-frequency flight instruments
    flightClient := client.New("FlightInstruments")
    defer flightClient.Close()
    
    // Client 2: Low-frequency system monitoring
    systemClient := client.New("SystemMonitor")
    defer systemClient.Close()
    
    // Client 3: Aircraft control
    controlClient := client.New("AircraftControl")
    defer controlClient.Close()

    // Each client handles its specific domain
    go runFlightInstruments(flightClient)
    go runSystemMonitor(systemClient)
    go runAircraftControl(controlClient)
}
```

## Architecture Patterns

### Pattern 1: Functional Separation

Separate clients based on functional boundaries:

```go
package main

import (
    "log"
    "sync"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

type AircraftManager struct {
    flightMonitor *FlightMonitor
    systemController *SystemController
    dataLogger *DataLogger
    wg sync.WaitGroup
}

type FlightMonitor struct {
    sdk client.Connection
}

type SystemController struct {
    sdk client.Connection
}

type DataLogger struct {
    sdk client.Connection
}

func NewAircraftManager() *AircraftManager {
    return &AircraftManager{
        flightMonitor: &FlightMonitor{
            sdk: client.New("FlightMonitor"),
        },
        systemController: &SystemController{
            sdk: client.New("SystemController"),
        },
        dataLogger: &DataLogger{
            sdk: client.New("DataLogger"),
        },
    }
}

func (am *AircraftManager) Start() error {
    // Initialize each subsystem
    clients := []struct {
        name string
        sdk  client.Connection
        init func() error
    }{
        {"FlightMonitor", am.flightMonitor.sdk, am.flightMonitor.Initialize},
        {"SystemController", am.systemController.sdk, am.systemController.Initialize},
        {"DataLogger", am.dataLogger.sdk, am.dataLogger.Initialize},
    }

    for _, c := range clients {
        if err := c.sdk.Open(); err != nil {
            return fmt.Errorf("failed to open %s: %v", c.name, err)
        }
        
        if err := c.init(); err != nil {
            return fmt.Errorf("failed to initialize %s: %v", c.name, err)
        }
    }

    // Start each subsystem in its own goroutine
    am.wg.Add(3)
    go am.flightMonitor.Run(&am.wg)
    go am.systemController.Run(&am.wg)
    go am.dataLogger.Run(&am.wg)

    return nil
}

func (am *AircraftManager) Stop() {
    // Close all clients
    am.flightMonitor.sdk.Close()
    am.systemController.sdk.Close()
    am.dataLogger.sdk.Close()
    
    // Wait for all goroutines to finish
    am.wg.Wait()
}

func (fm *FlightMonitor) Initialize() error {
    // Register flight-critical variables with high frequency
    flightVars := []struct {
        id   uint32
        name string
    }{
        {1, "PLANE ALTITUDE"},
        {2, "AIRSPEED INDICATED"},
        {3, "HEADING INDICATOR"},
        {4, "VERTICAL SPEED"},
    }

    for _, v := range flightVars {
        if err := fm.sdk.RegisterSimVarDefinition(v.id, v.name, "feet", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
            return err
        }
        // High frequency for smooth display
        if err := fm.sdk.RequestSimVarDataPeriodic(v.id, v.id*100, types.SIMCONNECT_PERIOD_VISUAL_FRAME); err != nil {
            return err
        }
    }
    return nil
}

func (fm *FlightMonitor) Run(wg *sync.WaitGroup) {
    defer wg.Done()
    
    messages := fm.sdk.Listen()
    for msg := range messages {
        fm.processFlightData(msg)
    }
}

func (sc *SystemController) Initialize() error {
    // Register system control events
    controlEvents := map[types.ClientEventID]string{
        1001: "TOGGLE_MASTER_BATTERY",
        1002: "TOGGLE_BEACON_LIGHTS",
        1003: "TOGGLE_NAV_LIGHTS",
    }

    for eventID, eventName := range controlEvents {
        if err := sc.sdk.MapClientEventToSimEvent(eventID, eventName); err != nil {
            return err
        }
        if err := sc.sdk.AddClientEventToNotificationGroup(3000, eventID, false); err != nil {
            return err
        }
    }

    return sc.sdk.SetNotificationGroupPriority(3000, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST)
}

func (sc *SystemController) Run(wg *sync.WaitGroup) {
    defer wg.Done()
    
    messages := sc.sdk.Listen()
    for msg := range messages {
        sc.processControlData(msg)
    }
}

func (dl *DataLogger) Initialize() error {
    // Register comprehensive data set with lower frequency
    logVars := []struct {
        id   uint32
        name string
    }{
        {100, "PLANE ALTITUDE"},
        {101, "AIRSPEED INDICATED"},
        {102, "FUEL TOTAL QUANTITY"},
        {103, "ENG RPM:1"},
        {104, "ELECTRICAL MAIN BUS VOLTAGE"},
    }

    for _, v := range logVars {
        if err := dl.sdk.RegisterSimVarDefinition(v.id, v.name, "feet", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
            return err
        }
        // Lower frequency for logging
        if err := dl.sdk.RequestSimVarDataPeriodic(v.id, v.id*10, types.SIMCONNECT_PERIOD_SECOND); err != nil {
            return err
        }
    }
    return nil
}

func (dl *DataLogger) Run(wg *sync.WaitGroup) {
    defer wg.Done()
    
    messages := dl.sdk.Listen()
    for msg := range messages {
        dl.logData(msg)
    }
}

func main() {
    manager := NewAircraftManager()
    defer manager.Stop()

    if err := manager.Start(); err != nil {
        log.Fatalf("Failed to start aircraft manager: %v", err)
    }

    // Run for demonstration
    time.Sleep(60 * time.Second)
}
```

### Pattern 2: Performance-Based Separation

Separate clients based on performance requirements:

```go
type PerformanceBasedManager struct {
    highFrequencyClient client.Connection
    lowFrequencyClient  client.Connection
    eventClient         client.Connection
}

func NewPerformanceBasedManager() *PerformanceBasedManager {
    return &PerformanceBasedManager{
        highFrequencyClient: client.New("HighFrequency"),
        lowFrequencyClient:  client.New("LowFrequency"),
        eventClient:         client.New("Events"),
    }
}

func (pbm *PerformanceBasedManager) Initialize() error {
    // High-frequency client: Critical flight instruments
    if err := pbm.highFrequencyClient.Open(); err != nil {
        return err
    }
    
    criticalVars := []string{"PLANE ALTITUDE", "AIRSPEED INDICATED", "HEADING INDICATOR"}
    for i, varName := range criticalVars {
        id := uint32(i + 1)
        pbm.highFrequencyClient.RegisterSimVarDefinition(id, varName, "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
        pbm.highFrequencyClient.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
    }

    // Low-frequency client: System states and navigation
    if err := pbm.lowFrequencyClient.Open(); err != nil {
        return err
    }
    
    systemVars := []string{"FUEL TOTAL QUANTITY", "ENG RPM:1", "GPS GROUND SPEED"}
    for i, varName := range systemVars {
        id := uint32(i + 10)
        pbm.lowFrequencyClient.RegisterSimVarDefinition(id, varName, "gallons", types.SIMCONNECT_DATATYPE_FLOAT32)
        pbm.lowFrequencyClient.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
    }

    // Event client: System events only
    if err := pbm.eventClient.Open(); err != nil {
        return err
    }
    
    events := []struct {
        id   uint32
        name string
    }{
        {2001, "Pause"},
        {2002, "Crashed"},
        {2003, "AircraftLoaded"},
    }
    
    for _, event := range events {
        pbm.eventClient.SubscribeToSystemEvent(event.id, event.name)
    }

    return nil
}

func (pbm *PerformanceBasedManager) Run() {
    var wg sync.WaitGroup

    // High-frequency processor
    wg.Add(1)
    go func() {
        defer wg.Done()
        messages := pbm.highFrequencyClient.Listen()
        for msg := range messages {
            pbm.processHighFrequencyData(msg)
        }
    }()

    // Low-frequency processor
    wg.Add(1)
    go func() {
        defer wg.Done()
        messages := pbm.lowFrequencyClient.Listen()
        for msg := range messages {
            pbm.processLowFrequencyData(msg)
        }
    }()

    // Event processor
    wg.Add(1)
    go func() {
        defer wg.Done()
        messages := pbm.eventClient.Listen()
        for msg := range messages {
            pbm.processEvents(msg)
        }
    }()

    wg.Wait()
}
```

### Pattern 3: Multi-Aircraft Management

Separate clients for each aircraft in multi-aircraft scenarios:

```go
type MultiAircraftManager struct {
    aircraftClients map[string]client.Connection
    mu             sync.RWMutex
}

func NewMultiAircraftManager() *MultiAircraftManager {
    return &MultiAircraftManager{
        aircraftClients: make(map[string]client.Connection),
    }
}

func (mam *MultiAircraftManager) AddAircraft(aircraftID string) error {
    mam.mu.Lock()
    defer mam.mu.Unlock()

    if _, exists := mam.aircraftClients[aircraftID]; exists {
        return fmt.Errorf("aircraft %s already exists", aircraftID)
    }

    // Create dedicated client for this aircraft
    aircraftClient := client.New(fmt.Sprintf("Aircraft_%s", aircraftID))
    if err := aircraftClient.Open(); err != nil {
        return fmt.Errorf("failed to open client for aircraft %s: %v", aircraftID, err)
    }

    // Register aircraft-specific data
    if err := mam.setupAircraftMonitoring(aircraftClient); err != nil {
        aircraftClient.Close()
        return fmt.Errorf("failed to setup monitoring for aircraft %s: %v", aircraftID, err)
    }

    mam.aircraftClients[aircraftID] = aircraftClient

    // Start monitoring this aircraft
    go mam.monitorAircraft(aircraftID, aircraftClient)

    return nil
}

func (mam *MultiAircraftManager) RemoveAircraft(aircraftID string) error {
    mam.mu.Lock()
    defer mam.mu.Unlock()

    client, exists := mam.aircraftClients[aircraftID]
    if !exists {
        return fmt.Errorf("aircraft %s not found", aircraftID)
    }

    client.Close()
    delete(mam.aircraftClients, aircraftID)
    return nil
}

func (mam *MultiAircraftManager) setupAircraftMonitoring(aircraftClient client.Connection) error {
    // Register standard flight data for each aircraft
    flightVars := []struct {
        id   uint32
        name string
    }{
        {1, "PLANE ALTITUDE"},
        {2, "AIRSPEED INDICATED"},
        {3, "PLANE LATITUDE"},
        {4, "PLANE LONGITUDE"},
    }

    for _, v := range flightVars {
        if err := aircraftClient.RegisterSimVarDefinition(v.id, v.name, "feet", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
            return err
        }
        if err := aircraftClient.RequestSimVarDataPeriodic(v.id, v.id*100, types.SIMCONNECT_PERIOD_SECOND); err != nil {
            return err
        }
    }

    return nil
}

func (mam *MultiAircraftManager) monitorAircraft(aircraftID string, aircraftClient client.Connection) {
    messages := aircraftClient.Listen()
    
    for msg := range messages {
        if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
            if data, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := data.(*client.SimVarData); ok {
                    mam.processAircraftData(aircraftID, simVar)
                }
            }
        }
    }
}

func (mam *MultiAircraftManager) processAircraftData(aircraftID string, simVar *client.SimVarData) {
    fmt.Printf("Aircraft %s - Variable %d: %v\n", aircraftID, simVar.DefineID, simVar.Value)
}

func (mam *MultiAircraftManager) GetAircraftList() []string {
    mam.mu.RLock()
    defer mam.mu.RUnlock()

    var aircraftList []string
    for aircraftID := range mam.aircraftClients {
        aircraftList = append(aircraftList, aircraftID)
    }
    return aircraftList
}

func (mam *MultiAircraftManager) Shutdown() {
    mam.mu.Lock()
    defer mam.mu.Unlock()

    for aircraftID, client := range mam.aircraftClients {
        client.Close()
        fmt.Printf("Closed client for aircraft %s\n", aircraftID)
    }
    
    mam.aircraftClients = make(map[string]client.Connection)
}
```

## Performance Considerations

### Resource Usage per Client

Each SimConnect client consumes:
- **Memory**: ~2-4 MB per client
- **CPU**: Dedicated message processing thread
- **Network**: Separate connection to SimConnect
- **SimConnect slots**: Limited number of simultaneous connections

### Optimization Strategies

```go
type OptimizedClientManager struct {
    clients      []client.Connection
    clientPool   chan client.Connection
    maxClients   int
    activeCount  int32
    mu          sync.Mutex
}

func NewOptimizedClientManager(maxClients int) *OptimizedClientManager {
    return &OptimizedClientManager{
        clientPool: make(chan client.Connection, maxClients),
        maxClients: maxClients,
    }
}

func (ocm *OptimizedClientManager) GetClient() (client.Connection, error) {
    // Try to get existing client from pool
    select {
    case client := <-ocm.clientPool:
        return client, nil
    default:
        // Create new client if under limit
        ocm.mu.Lock()
        defer ocm.mu.Unlock()
        
        if int(atomic.LoadInt32(&ocm.activeCount)) >= ocm.maxClients {
            return nil, fmt.Errorf("maximum client limit reached")
        }
        
        clientName := fmt.Sprintf("PooledClient_%d", atomic.AddInt32(&ocm.activeCount, 1))
        newClient := client.New(clientName)
        
        if err := newClient.Open(); err != nil {
            atomic.AddInt32(&ocm.activeCount, -1)
            return nil, err
        }
        
        return newClient, nil
    }
}

func (ocm *OptimizedClientManager) ReturnClient(client client.Connection) {
    select {
    case ocm.clientPool <- client:
        // Client returned to pool
    default:
        // Pool is full, close client
        client.Close()
        atomic.AddInt32(&ocm.activeCount, -1)
    }
}

func (ocm *OptimizedClientManager) Shutdown() {
    close(ocm.clientPool)
    for client := range ocm.clientPool {
        client.Close()
    }
}
```

## Best Practices

### Do's ✅

1. **Use functional separation** - Separate clients based on logical boundaries
2. **Consider performance needs** - Use multiple clients for different update frequencies
3. **Implement proper cleanup** - Always close all clients on shutdown
4. **Use connection pooling** - Reuse clients when possible
5. **Monitor resource usage** - Track memory and CPU usage per client
6. **Handle errors gracefully** - One client failure shouldn't affect others

### Don'ts ❌

1. **Don't create clients unnecessarily** - Single client with fan-out is often better
2. **Don't ignore resource limits** - SimConnect has connection limits
3. **Don't share clients between goroutines** - Each client should have a single owner
4. **Don't forget error handling** - Each client needs proper error handling
5. **Don't duplicate data requests** - Avoid requesting same data from multiple clients
6. **Don't ignore connection failures** - Implement reconnection logic

### Monitoring Multiple Clients

```go
type ClientHealthMonitor struct {
    clients map[string]ClientHealth
    mu      sync.RWMutex
}

type ClientHealth struct {
    Name            string
    Connected       bool
    LastMessage     time.Time
    MessageCount    int64
    ErrorCount      int64
    ReconnectCount  int64
}

func (chm *ClientHealthMonitor) UpdateHealth(clientName string, connected bool, messageReceived bool) {
    chm.mu.Lock()
    defer chm.mu.Unlock()

    health := chm.clients[clientName]
    health.Name = clientName
    health.Connected = connected
    
    if messageReceived {
        health.LastMessage = time.Now()
        health.MessageCount++
    }
    
    chm.clients[clientName] = health
}

func (chm *ClientHealthMonitor) GetHealthReport() map[string]ClientHealth {
    chm.mu.RLock()
    defer chm.mu.RUnlock()

    report := make(map[string]ClientHealth)
    for name, health := range chm.clients {
        report[name] = health
    }
    return report
}

func (chm *ClientHealthMonitor) StartMonitoring() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        chm.checkClientHealth()
    }
}

func (chm *ClientHealthMonitor) checkClientHealth() {
    chm.mu.RLock()
    defer chm.mu.RUnlock()

    for name, health := range chm.clients {
        if health.Connected && time.Since(health.LastMessage) > 30*time.Second {
            log.Printf("⚠️ Client %s hasn't received messages for %v", 
                name, time.Since(health.LastMessage))
        }
    }
}
```

### Decision Checklist

Before creating multiple clients, ask:

- [ ] Can this be solved with a single client and fan-out pattern?
- [ ] Do I need different update frequencies for different data?
- [ ] Are these logically separate functional domains?
- [ ] Will multiple clients improve performance or reliability?
- [ ] Do I have the resources (memory, connections) for multiple clients?
- [ ] Have I implemented proper error handling and cleanup for each client?

**Remember**: Multiple clients add complexity. Use them only when they provide clear benefits over single-client patterns.
