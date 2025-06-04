# Advanced Usage Guide

This guide covers advanced usage patterns for the SimConnect Go SDK, including concurrent monitoring, multiple client scenarios, and best practices for production applications.

## Table of Contents

- [Multiple Listeners and Goroutines](#multiple-listeners-and-goroutines)
- [Concurrent Monitoring Patterns](#concurrent-monitoring-patterns)
- [Multiple Client Architecture](#multiple-client-architecture)
- [Production Patterns](#production-patterns)
- [Performance Optimization](#performance-optimization)
- [Error Handling and Recovery](#error-handling-and-recovery)
- [Best Practices](#best-practices)

## Multiple Listeners and Goroutines

### Important: Single Listen() Call Per Client

The SDK's `Listen()` method is designed to be called **only once per client instance**. It uses internal synchronization (`sync.Once`) to ensure thread safety and prevent multiple dispatch goroutines.

**Critical**: When multiple goroutines read from the same channel, each message is delivered to only ONE goroutine (load balancing), not all of them. This causes message loss and unpredictable behavior.

```go
// ❌ INCORRECT: Multiple Listen() calls on the same client
sdk := client.New("MyApp")
messages1 := sdk.Listen() // This works
messages2 := sdk.Listen() // This returns the SAME channel as messages1
```

```go
// ❌ INCORRECT: Multiple goroutines reading from same channel
// This pattern causes message loss - each message goes to only ONE goroutine
sdk := client.New("MyApp")
messages := sdk.Listen()

// DON'T DO THIS - messages will be randomly distributed between goroutines
go func() {
    for msg := range messages {
        // This goroutine will miss messages that go to the other goroutine
        processMessage(msg, "Worker-1")
    }
}()

go func() {
    for msg := range messages {
        // This goroutine will also miss messages that go to the other goroutine
        processMessage(msg, "Worker-2")
    }
}()

// ❌ This is problematic because:
// 1. Each message is delivered to only ONE goroutine (not both)
// 2. You cannot predict which goroutine will receive which message
// 3. Critical messages may be missed by the intended processor
// 4. This should only be used for compute-heavy processing where 
//    missing some messages is acceptable
```

### Fan-Out Pattern for Message Distribution

**Important**: In Go channels, when multiple goroutines read from the same channel, each message goes to **only ONE goroutine** (load balancing), not all of them. If you need all goroutines to process all messages, or want specific message types to go to specific processors, use a fan-out pattern:

```go
package main

import (
    "fmt"
    "sync"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
)

type MessageDistributor struct {
    simVarChan   chan interface{}
    eventChan    chan interface{}
    exceptionChan chan interface{}
    wg           sync.WaitGroup
}

func NewMessageDistributor() *MessageDistributor {
    return &MessageDistributor{
        simVarChan:   make(chan interface{}, 100),
        eventChan:    make(chan interface{}, 100),
        exceptionChan: make(chan interface{}, 100),
    }
}

func (md *MessageDistributor) Start(messages <-chan any) {
    // Single goroutine reads from SDK and distributes messages
    go func() {
        defer func() {
            close(md.simVarChan)
            close(md.eventChan)
            close(md.exceptionChan)
        }()
        
        for msg := range messages {
            if msgMap, ok := msg.(map[string]any); ok {
                switch msgMap["type"] {
                case "SIMOBJECT_DATA":
                    select {
                    case md.simVarChan <- msg:
                    default:
                        fmt.Println("⚠️ SimVar channel full, dropping message")
                    }
                case "EVENT":
                    select {
                    case md.eventChan <- msg:
                    default:
                        fmt.Println("⚠️ Event channel full, dropping message")
                    }
                case "EXCEPTION":
                    select {
                    case md.exceptionChan <- msg:
                    default:
                        fmt.Println("⚠️ Exception channel full, dropping message")
                    }
                }
            }
        }
    }()
}

func (md *MessageDistributor) StartWorkers() {
    // Flight data processor
    md.wg.Add(1)
    go func() {
        defer md.wg.Done()
        for msg := range md.simVarChan {
            md.processFlightData(msg)
        }
    }()

    // Event processor
    md.wg.Add(1)
    go func() {
        defer md.wg.Done()
        for msg := range md.eventChan {
            md.processEvents(msg)
        }
    }()

    // Exception handler
    md.wg.Add(1)
    go func() {
        defer md.wg.Done()
        for msg := range md.exceptionChan {
            md.handleExceptions(msg)
        }
    }()
}

func (md *MessageDistributor) Wait() {
    md.wg.Wait()
}

func (md *MessageDistributor) processFlightData(msg interface{}) {
    if msgMap, ok := msg.(map[string]any); ok {
        if parsedData, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := parsedData.(*client.SimVarData); ok {
                fmt.Printf("📊 Flight Data - DefineID: %d, Value: %v\n", 
                    simVar.DefineID, simVar.Value)
            }
        }
    }
}

func (md *MessageDistributor) processEvents(msg interface{}) {
    if msgMap, ok := msg.(map[string]any); ok {
        if eventData, exists := msgMap["event"]; exists {
            if event, ok := eventData.(*types.EventData); ok {
                fmt.Printf("📡 Event - %s: %d\n", event.EventName, event.EventData)
            }
        }
    }
}

func (md *MessageDistributor) handleExceptions(msg interface{}) {
    if msgMap, ok := msg.(map[string]any); ok {
        if exceptionData, exists := msgMap["exception"]; exists {
            fmt.Printf("❌ Exception: %+v\n", exceptionData)
        }
    }
}

// Example usage
func main() {
    sdk := client.New("AdvancedMonitor")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        panic(err)
    }

    // Register multiple variables
    variables := map[uint32]string{
        1: "PLANE ALTITUDE",
        2: "AIRSPEED INDICATED",
        3: "HEADING INDICATOR",
    }

    for id, name := range variables {
        sdk.RegisterSimVarDefinition(id, name, "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
        sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
    }

    // Subscribe to events
    sdk.SubscribeToSystemEvent(1001, "Pause")
    sdk.SubscribeToSystemEvent(1002, "Sim")

    // Set up message distribution
    distributor := NewMessageDistributor()
    messages := sdk.Listen()
    distributor.Start(messages)
    distributor.StartWorkers()

    // Run for demonstration
    time.Sleep(30 * time.Second)
    
    // Cleanup
    for id := range variables {
        sdk.StopPeriodicRequest(id * 100)
    }
    
    distributor.Wait()
}
```

## Concurrent Monitoring Patterns

### Pattern 1: Specialized Monitoring Services

Create specialized services for different aircraft systems:

```go
// Electrical system monitor
type ElectricalMonitor struct {
    sdk    client.Connection
    data   map[string]interface{}
    mu     sync.RWMutex
    stopCh chan struct{}
}

func NewElectricalMonitor(sdk client.Connection) *ElectricalMonitor {
    return &ElectricalMonitor{
        sdk:    sdk,
        data:   make(map[string]interface{}),
        stopCh: make(chan struct{}),
    }
}

func (em *ElectricalMonitor) Start(messages <-chan any) {
    // Register electrical system variables
    electricalVars := map[uint32]string{
        10: "EXTERNAL POWER ON",
        11: "ELECTRICAL MASTER BATTERY",
        12: "GENERAL ENG MASTER ALTERNATOR",
        13: "ELECTRICAL MAIN BUS VOLTAGE",
        14: "ELECTRICAL BATTERY LOAD",
    }

    for id, name := range electricalVars {
        dataType := types.SIMCONNECT_DATATYPE_INT32
        units := "Bool"
        if id >= 13 { // Voltage and current are floats
            dataType = types.SIMCONNECT_DATATYPE_FLOAT32
            if id == 13 {
                units = "Volts"
            } else {
                units = "Amperes"
            }
        }
        
        em.sdk.RegisterSimVarDefinition(id, name, units, dataType)
        em.sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
    }

    go em.processMessages(messages)
}

func (em *ElectricalMonitor) processMessages(messages <-chan any) {
    for {
        select {
        case msg := <-messages:
            if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
                if parsedData, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := parsedData.(*client.SimVarData); ok {
                        // Only process electrical system data (DefineID 10-14)
                        if simVar.DefineID >= 10 && simVar.DefineID <= 14 {
                            em.updateData(simVar)
                        }
                    }
                }
            }
        case <-em.stopCh:
            return
        }
    }
}

func (em *ElectricalMonitor) updateData(simVar *client.SimVarData) {
    em.mu.Lock()
    defer em.mu.Unlock()
    
    var fieldName string
    switch simVar.DefineID {
    case 10:
        fieldName = "ExternalPower"
    case 11:
        fieldName = "MasterBattery"
    case 12:
        fieldName = "Alternator"
    case 13:
        fieldName = "BusVoltage"
    case 14:
        fieldName = "BatteryLoad"
    }
    
    em.data[fieldName] = simVar.Value
    fmt.Printf("⚡ Electrical: %s = %v\n", fieldName, simVar.Value)
}

func (em *ElectricalMonitor) GetCurrentData() map[string]interface{} {
    em.mu.RLock()
    defer em.mu.RUnlock()
    
    result := make(map[string]interface{})
    for k, v := range em.data {
        result[k] = v
    }
    return result
}

func (em *ElectricalMonitor) Stop() {
    close(em.stopCh)
    
    // Stop periodic requests
    for id := uint32(10); id <= 14; id++ {
        em.sdk.StopPeriodicRequest(id * 100)
    }
}
```

### Pattern 2: Multi-System Dashboard

```go
type FlightDashboard struct {
    sdk              client.Connection
    electricalMonitor *ElectricalMonitor
    flightDataMonitor *FlightDataMonitor
    eventMonitor     *EventMonitor
    wg               sync.WaitGroup
}

func NewFlightDashboard() *FlightDashboard {
    sdk := client.New("FlightDashboard")
    
    return &FlightDashboard{
        sdk:              sdk,
        electricalMonitor: NewElectricalMonitor(sdk),
        flightDataMonitor: NewFlightDataMonitor(sdk),
        eventMonitor:     NewEventMonitor(sdk),
    }
}

func (fd *FlightDashboard) Start() error {
    if err := fd.sdk.Open(); err != nil {
        return err
    }

    // Single Listen() call - all monitors share this channel
    messages := fd.sdk.Listen()
    if messages == nil {
        return fmt.Errorf("failed to start listening")
    }

    // Start all monitors with the same message channel
    fd.electricalMonitor.Start(messages)
    fd.flightDataMonitor.Start(messages)
    fd.eventMonitor.Start(messages)

    // Start web server for dashboard
    fd.startWebServer()

    return nil
}

func (fd *FlightDashboard) startWebServer() {
    http.HandleFunc("/api/electrical", func(w http.ResponseWriter, r *http.Request) {
        data := fd.electricalMonitor.GetCurrentData()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(data)
    })

    http.HandleFunc("/api/flight", func(w http.ResponseWriter, r *http.Request) {
        data := fd.flightDataMonitor.GetCurrentData()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(data)
    })

    go func() {
        log.Println("Dashboard server starting on :8080")
        if err := http.ListenAndServe(":8080", nil); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()
}
```

## Multiple Client Architecture

### When to Use Multiple Clients

Use multiple SDK clients when you need:

1. **Different Update Frequencies**: Critical vs. non-critical data
2. **System Isolation**: Separate failure domains
3. **Permissions**: Different access levels
4. **Load Distribution**: Heavy monitoring vs. light control

```go
type MultiClientArchitecture struct {
    criticalClient  client.Connection  // High-frequency flight data
    environmentClient client.Connection  // Environmental data
    controlClient   client.Connection  // Aircraft control
    
    shutdown chan struct{}
    wg       sync.WaitGroup
}

func NewMultiClientArchitecture() *MultiClientArchitecture {
    return &MultiClientArchitecture{
        criticalClient:   client.New("CriticalSystems"),
        environmentClient: client.New("EnvironmentalData"), 
        controlClient:    client.New("AircraftControl"),
        shutdown:         make(chan struct{}),
    }
}

func (mca *MultiClientArchitecture) Start() error {
    // Connect all clients
    clients := []struct {
        name   string
        client client.Connection
    }{
        {"Critical", mca.criticalClient},
        {"Environment", mca.environmentClient},
        {"Control", mca.controlClient},
    }

    for _, c := range clients {
        if err := c.client.Open(); err != nil {
            return fmt.Errorf("failed to connect %s client: %v", c.name, err)
        }
    }

    // Start critical systems monitoring (high frequency)
    mca.wg.Add(1)
    go mca.monitorCriticalSystems()

    // Start environmental monitoring (low frequency)
    mca.wg.Add(1)
    go mca.monitorEnvironmentalData()

    // Start control event handling
    mca.wg.Add(1)
    go mca.handleControlEvents()

    return nil
}

func (mca *MultiClientArchitecture) monitorCriticalSystems() {
    defer mca.wg.Done()
    
    // Register critical flight data
    criticalVars := map[uint32]string{
        1: "PLANE ALTITUDE",
        2: "AIRSPEED INDICATED",
        3: "VERTICAL SPEED",
        4: "HEADING INDICATOR",
    }

    for id, name := range criticalVars {
        mca.criticalClient.RegisterSimVarDefinition(id, name, "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
        // High frequency updates for critical data
        mca.criticalClient.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
    }

    messages := mca.criticalClient.Listen()
    for {
        select {
        case msg := <-messages:
            mca.processCriticalData(msg)
        case <-mca.shutdown:
            // Cleanup
            for id := range criticalVars {
                mca.criticalClient.StopPeriodicRequest(id * 100)
            }
            return
        }
    }
}

func (mca *MultiClientArchitecture) monitorEnvironmentalData() {
    defer mca.wg.Done()
    
    // Register environmental variables
    envVars := map[uint32]string{
        20: "AMBIENT TEMPERATURE",
        21: "AMBIENT PRESSURE",
        22: "AMBIENT WIND VELOCITY",
        23: "AMBIENT WIND DIRECTION",
    }

    for id, name := range envVars {
        mca.environmentClient.RegisterSimVarDefinition(id, name, "celsius", types.SIMCONNECT_DATATYPE_FLOAT32)
        // Lower frequency for environmental data
        mca.environmentClient.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
    }

    messages := mca.environmentClient.Listen()
    for {
        select {
        case msg := <-messages:
            mca.processEnvironmentalData(msg)
        case <-mca.shutdown:
            // Cleanup
            for id := range envVars {
                mca.environmentClient.StopPeriodicRequest(id * 100)
            }
            return
        }
    }
}

func (mca *MultiClientArchitecture) handleControlEvents() {
    defer mca.wg.Done()
    
    // Set up control events
    controlEvents := map[types.ClientEventID]string{
        1001: "TOGGLE_MASTER_BATTERY",
        1002: "TOGGLE_EXTERNAL_POWER",
        1003: "TOGGLE_BEACON_LIGHTS",
    }

    for id, name := range controlEvents {
        mca.controlClient.MapClientEventToSimEvent(id, name)
        mca.controlClient.AddClientEventToNotificationGroup(3000, id, false)
    }
    
    mca.controlClient.SetNotificationGroupPriority(3000, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST)

    messages := mca.controlClient.Listen()
    for {
        select {
        case msg := <-messages:
            mca.processControlEvents(msg)
        case <-mca.shutdown:
            return
        }
    }
}

func (mca *MultiClientArchitecture) processCriticalData(msg any) {
    // Process high-frequency critical flight data
    if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
        if parsedData, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := parsedData.(*client.SimVarData); ok {
                // Critical data processing with immediate response
                fmt.Printf("🔴 CRITICAL: DefineID %d = %v\n", simVar.DefineID, simVar.Value)
            }
        }
    }
}

func (mca *MultiClientArchitecture) processEnvironmentalData(msg any) {
    // Process lower-frequency environmental data
    if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
        if parsedData, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := parsedData.(*client.SimVarData); ok {
                fmt.Printf("🌍 ENV: DefineID %d = %v\n", simVar.DefineID, simVar.Value)
            }
        }
    }
}

func (mca *MultiClientArchitecture) processControlEvents(msg any) {
    // Process control events and responses
    if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "EVENT" {
        if eventData, exists := msgMap["event"]; exists {
            if event, ok := eventData.(*types.EventData); ok {
                fmt.Printf("🎮 CONTROL: %s executed\n", event.EventName)
            }
        }
    }
}

func (mca *MultiClientArchitecture) Stop() {
    close(mca.shutdown)
    mca.wg.Wait()
    
    mca.criticalClient.Close()
    mca.environmentClient.Close()
    mca.controlClient.Close()
}
```

## Production Patterns

### Pattern 1: Resilient Monitor with Automatic Recovery

```go
type ResilientMonitor struct {
    sdk           client.Connection
    config        MonitorConfig
    restartCh     chan struct{}
    running       atomic.Bool
    healthCheck   *time.Ticker
    lastMessage   atomic.Value // stores time.Time
}

type MonitorConfig struct {
    AppName           string
    ReconnectInterval time.Duration
    HealthCheckInterval time.Duration
    MessageTimeout    time.Duration
    MaxRetries        int
}

func NewResilientMonitor(config MonitorConfig) *ResilientMonitor {
    return &ResilientMonitor{
        config:    config,
        restartCh: make(chan struct{}, 1),
    }
}

func (rm *ResilientMonitor) Start() error {
    rm.running.Store(true)
    
    // Start health check
    rm.healthCheck = time.NewTicker(rm.config.HealthCheckInterval)
    go rm.healthCheckLoop()
    
    // Start main monitoring loop with automatic recovery
    go rm.monitoringLoop()
    
    return nil
}

func (rm *ResilientMonitor) monitoringLoop() {
    retries := 0
    
    for rm.running.Load() {
        if err := rm.runMonitoring(); err != nil {
            log.Printf("Monitoring error: %v", err)
            retries++
            
            if retries >= rm.config.MaxRetries {
                log.Printf("Max retries (%d) reached, stopping", rm.config.MaxRetries)
                break
            }
            
            log.Printf("Retrying in %v (attempt %d/%d)", 
                rm.config.ReconnectInterval, retries, rm.config.MaxRetries)
            time.Sleep(rm.config.ReconnectInterval)
            continue
        }
        
        retries = 0 // Reset on successful run
    }
}

func (rm *ResilientMonitor) runMonitoring() error {
    // Create new client for this monitoring session
    rm.sdk = client.New(rm.config.AppName)
    defer rm.sdk.Close()
    
    if err := rm.sdk.Open(); err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }
    
    // Set up monitoring
    if err := rm.setupMonitoring(); err != nil {
        return fmt.Errorf("failed to setup monitoring: %v", err)
    }
    
    messages := rm.sdk.Listen()
    if messages == nil {
        return fmt.Errorf("failed to start listening")
    }
    
    // Process messages with timeout handling
    for rm.running.Load() {
        select {
        case msg, ok := <-messages:
            if !ok {
                return fmt.Errorf("message channel closed")
            }
            rm.processMessage(msg)
            rm.lastMessage.Store(time.Now())
            
        case <-rm.restartCh:
            log.Println("Manual restart requested")
            return nil
            
        case <-time.After(rm.config.MessageTimeout):
            return fmt.Errorf("message timeout - no data received")
        }
    }
    
    return nil
}

func (rm *ResilientMonitor) healthCheckLoop() {
    defer rm.healthCheck.Stop()
    
    for range rm.healthCheck.C {
        if !rm.running.Load() {
            return
        }
        
        lastMsg := rm.lastMessage.Load()
        if lastMsg != nil {
            if time.Since(lastMsg.(time.Time)) > rm.config.MessageTimeout {
                log.Println("Health check failed - triggering restart")
                select {
                case rm.restartCh <- struct{}{}:
                default: // Don't block if channel is full
                }
            }
        }
    }
}

func (rm *ResilientMonitor) setupMonitoring() error {
    // Register critical variables
    variables := []struct {
        id   uint32
        name string
        units string
        period types.SimConnectPeriod
    }{
        {1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_PERIOD_SECOND},
        {2, "AIRSPEED INDICATED", "knots", types.SIMCONNECT_PERIOD_SECOND},
        {3, "HEADING INDICATOR", "degrees", types.SIMCONNECT_PERIOD_SECOND},
    }
    
    for _, v := range variables {
        if err := rm.sdk.RegisterSimVarDefinition(v.id, v.name, v.units, types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
            return err
        }
        if err := rm.sdk.RequestSimVarDataPeriodic(v.id, v.id*100, v.period); err != nil {
            return err
        }
    }
    
    return nil
}

func (rm *ResilientMonitor) processMessage(msg any) {
    // Your message processing logic here
    if msgMap, ok := msg.(map[string]any); ok {
        switch msgMap["type"] {
        case "SIMOBJECT_DATA":
            // Process flight data
        case "EVENT":
            // Process events
        case "EXCEPTION":
            // Handle exceptions
            log.Printf("SimConnect exception: %+v", msgMap["exception"])
        }
    }
}

func (rm *ResilientMonitor) Stop() {
    rm.running.Store(false)
    
    // Trigger restart to break out of monitoring loop
    select {
    case rm.restartCh <- struct{}{}:
    default:
    }
}
```

## Performance Optimization

### 1. Efficient Message Processing

```go
// Use buffered channels for high-throughput scenarios
type HighThroughputMonitor struct {
    processingQueue chan interface{}
    workers         int
    wg              sync.WaitGroup
}

func NewHighThroughputMonitor(workers int) *HighThroughputMonitor {
    return &HighThroughputMonitor{
        processingQueue: make(chan interface{}, 1000), // Large buffer
        workers:         workers,
    }
}

func (htm *HighThroughputMonitor) Start(messages <-chan any) {
    // Start worker pool
    for i := 0; i < htm.workers; i++ {
        htm.wg.Add(1)
        go htm.worker(i)
    }
    
    // Distribute messages to workers
    go func() {
        defer close(htm.processingQueue)
        for msg := range messages {
            select {
            case htm.processingQueue <- msg:
            default:
                // Queue full - could implement backpressure or dropping strategy
                log.Println("Processing queue full, dropping message")
            }
        }
    }()
}

func (htm *HighThroughputMonitor) worker(id int) {
    defer htm.wg.Done()
    
    for msg := range htm.processingQueue {
        htm.processMessage(msg, id)
    }
}

func (htm *HighThroughputMonitor) processMessage(msg interface{}, workerID int) {
    // Fast message processing
    start := time.Now()
    defer func() {
        if duration := time.Since(start); duration > 10*time.Millisecond {
            log.Printf("Slow processing in worker %d: %v", workerID, duration)
        }
    }()
    
    // Your processing logic here
}
```

### 2. Memory-Efficient Data Structures

```go
// Use sync.Pool for frequently allocated objects
type DataProcessor struct {
    messagePool sync.Pool
}

func NewDataProcessor() *DataProcessor {
    return &DataProcessor{
        messagePool: sync.Pool{
            New: func() interface{} {
                return &ProcessedMessage{
                    Data: make(map[string]interface{}, 10),
                }
            },
        },
    }
}

type ProcessedMessage struct {
    Timestamp time.Time
    Type      string
    Data      map[string]interface{}
}

func (dp *DataProcessor) ProcessMessage(msg any) {
    processed := dp.messagePool.Get().(*ProcessedMessage)
    defer dp.messagePool.Put(processed)
    
    // Reset for reuse
    processed.Timestamp = time.Now()
    for k := range processed.Data {
        delete(processed.Data, k)
    }
    
    // Process message
    if msgMap, ok := msg.(map[string]any); ok {
        processed.Type = msgMap["type"].(string)
        // Fill processed.Data
    }
    
    // Use processed message
    dp.handleProcessedMessage(processed)
}
```

## Best Practices

### 1. Connection Management

```go
// Always use defer for cleanup
func setupConnection() (client.Connection, error) {
    sdk := client.New("MyApp")
    
    if err := sdk.Open(); err != nil {
        return nil, err
    }
    
    // Register cleanup on success
    runtime.SetFinalizer(sdk, (*client.Engine).Close)
    
    return sdk, nil
}

// Graceful shutdown pattern
func gracefulShutdown(sdk client.Connection, stopRequests []uint32) {
    // Stop all periodic requests
    for _, requestID := range stopRequests {
        if err := sdk.StopPeriodicRequest(requestID); err != nil {
            log.Printf("Warning: Failed to stop request %d: %v", requestID, err)
        }
    }
    
    // Allow time for cleanup
    time.Sleep(100 * time.Millisecond)
    
    // Close connection
    if err := sdk.Close(); err != nil {
        log.Printf("Warning: Error closing connection: %v", err)
    }
}
```

### 2. Error Handling

```go
func robustMessageHandling(messages <-chan any) {
    for msg := range messages {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("Panic in message handling: %v", r)
                    // Continue processing other messages
                }
            }()
            
            processMessage(msg)
        }()
    }
}
```

### 3. Resource Monitoring

```go
type ResourceMonitor struct {
    memStats     runtime.MemStats
    lastGC       time.Time
    messageCount uint64
}

func (rm *ResourceMonitor) logStats() {
    runtime.ReadMemStats(&rm.memStats)
    
    log.Printf("Memory: Alloc=%d KB, Sys=%d KB, NumGC=%d, Messages=%d",
        rm.memStats.Alloc/1024,
        rm.memStats.Sys/1024,
        rm.memStats.NumGC,
        atomic.LoadUint64(&rm.messageCount))
        
    if time.Since(rm.lastGC) > time.Minute {
        runtime.GC()
        rm.lastGC = time.Now()
    }
}
```

This advanced usage guide demonstrates how to build robust, concurrent applications with the SimConnect Go SDK while respecting its design constraints and leveraging Go's concurrency primitives effectively.
