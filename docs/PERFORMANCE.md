# Performance Guide

Optimization techniques and best practices for high-performance SimConnect applications.

## Table of Contents

- [Update Frequency Guidelines](#update-frequency-guidelines)
- [Memory Optimization](#memory-optimization)
- [Concurrent Processing](#concurrent-processing)
- [Resource Management](#resource-management)
- [Profiling and Monitoring](#profiling-and-monitoring)
- [Common Performance Issues](#common-performance-issues)

## Update Frequency Guidelines

Choosing the right update frequency is crucial for performance. Use the minimum frequency needed for your application.

### Update Period Reference

| Period | Frequency | Use Case | CPU Impact |
|--------|-----------|----------|------------|
| `SIMCONNECT_PERIOD_VISUAL_FRAME` | ~30-60 FPS | Critical flight instruments | High |
| `SIMCONNECT_PERIOD_SECOND` | 1 Hz | Navigation, fuel, systems | Low |
| `SIMCONNECT_PERIOD_ON_SET` | When changed | User settings, switches | Minimal |
| `SIMCONNECT_PERIOD_ONCE` | Single request | Static data, aircraft info | None |

### Recommended Frequencies by Data Type

```go
// Critical flight instruments - high frequency for smooth display
sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_VISUAL_FRAME) // Altitude
sdk.RequestSimVarDataPeriodic(2, 200, types.SIMCONNECT_PERIOD_VISUAL_FRAME) // Airspeed
sdk.RequestSimVarDataPeriodic(3, 300, types.SIMCONNECT_PERIOD_VISUAL_FRAME) // Attitude

// Navigation data - medium frequency sufficient
sdk.RequestSimVarDataPeriodic(10, 1000, types.SIMCONNECT_PERIOD_SECOND) // GPS position
sdk.RequestSimVarDataPeriodic(11, 1100, types.SIMCONNECT_PERIOD_SECOND) // Ground speed

// System states - only when changed
sdk.RequestSimVarDataPeriodic(20, 2000, types.SIMCONNECT_PERIOD_ON_SET) // Autopilot
sdk.RequestSimVarDataPeriodic(21, 2100, types.SIMCONNECT_PERIOD_ON_SET) // Gear position

// Static information - one-time request
sdk.RequestSimVarData(30, 3000) // Aircraft type, max altitude, etc.
```

### Frequency Optimization Example

```go
type InstrumentManager struct {
    sdk           client.Connection
    criticalData  map[uint32]time.Time
    normalData    map[uint32]time.Time
}

func (im *InstrumentManager) OptimizeUpdateRates() {
    // Start with high frequency for all
    for id := uint32(1); id <= 10; id++ {
        im.sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
    }

    // Monitor for 30 seconds, then optimize
    time.Sleep(30 * time.Second)
    
    // Reduce frequency for stable data
    for id, lastUpdate := range im.criticalData {
        if time.Since(lastUpdate) > 5*time.Second {
            // Data hasn't changed much, reduce to 1Hz
            im.sdk.StopPeriodicRequest(id * 100)
            im.sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
        }
    }
}
```

## Memory Optimization

### Efficient Message Processing

```go
// ✅ GOOD: Minimal allocations, fast processing
func processMessageEfficient(msg any) {
    // Fast type assertion
    msgMap, ok := msg.(map[string]any)
    if !ok {
        return
    }

    // Early return for unneeded message types
    msgType, exists := msgMap["type"]
    if !exists || msgType != "SIMOBJECT_DATA" {
        return
    }

    // Direct field access without intermediate variables
    if data, exists := msgMap["parsed_data"]; exists {
        if simVar, ok := data.(*client.SimVarData); ok {
            // Process immediately without storing
            updateInstrumentDisplay(simVar.DefineID, simVar.Value)
        }
    }
}

// ❌ BAD: Excessive allocations, complex processing
func processMessageInefficient(msg any) {
    // Creates unnecessary intermediate structures
    msgMap := msg.(map[string]any)
    msgType := fmt.Sprintf("%v", msgMap["type"])
    
    // String formatting on hot path
    log.Printf("Processing message type: %s", msgType)
    
    if msgType == "SIMOBJECT_DATA" {
        data := msgMap["parsed_data"]
        simVar := data.(*client.SimVarData)
        
        // Unnecessary map allocation
        values := make(map[string]interface{})
        values["id"] = simVar.DefineID
        values["value"] = simVar.Value
        
        // JSON marshaling on hot path
        jsonData, _ := json.Marshal(values)
        processJSONData(jsonData)
    }
}
```

### Memory Pool Pattern for High-Throughput

```go
import "sync"

type DataPointPool struct {
    pool sync.Pool
}

func NewDataPointPool() *DataPointPool {
    return &DataPointPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &DataPoint{
                    Values: make(map[uint32]float64, 10),
                }
            },
        },
    }
}

type DataPoint struct {
    Timestamp time.Time
    Values    map[uint32]float64
}

func (dp *DataPointPool) Get() *DataPoint {
    point := dp.pool.Get().(*DataPoint)
    point.Timestamp = time.Now()
    // Clear map but keep allocated memory
    for k := range point.Values {
        delete(point.Values, k)
    }
    return point
}

func (dp *DataPointPool) Put(point *DataPoint) {
    dp.pool.Put(point)
}

// Usage in high-throughput scenario
func highThroughputProcessor() {
    pool := NewDataPointPool()
    
    for msg := range messages {
        // Reuse allocated memory
        dataPoint := pool.Get()
        defer pool.Put(dataPoint)
        
        // Process with minimal allocations
        processDataPoint(dataPoint, msg)
    }
}
```

### String Interning for Repeated Values

```go
type StringInterner struct {
    cache map[string]string
    mu    sync.RWMutex
}

func NewStringInterner() *StringInterner {
    return &StringInterner{
        cache: make(map[string]string),
    }
}

func (si *StringInterner) Intern(s string) string {
    si.mu.RLock()
    if interned, exists := si.cache[s]; exists {
        si.mu.RUnlock()
        return interned
    }
    si.mu.RUnlock()

    si.mu.Lock()
    defer si.mu.Unlock()
    
    // Double-check after acquiring write lock
    if interned, exists := si.cache[s]; exists {
        return interned
    }
    
    si.cache[s] = s
    return s
}

// Use for repeated string values like event names
var eventNameInterner = NewStringInterner()

func processEvent(eventName string) {
    // Reduces string allocation for repeated event names
    internedName := eventNameInterner.Intern(eventName)
    handleEventByName(internedName)
}
```

## Concurrent Processing

### Optimal Fan-Out Pattern

```go
func optimalFanOutProcessor() {
    sdk := client.New("OptimalProcessor")
    defer sdk.Close()

    if err := sdk.Open(); err != nil {
        log.Fatal(err)
    }

    // Single listen call
    messages := sdk.Listen()
    
    // Buffered channels sized for expected throughput
    const bufferSize = 1000
    flightCh := make(chan *client.SimVarData, bufferSize)
    engineCh := make(chan *client.SimVarData, bufferSize)
    navigationCh := make(chan *client.SimVarData, bufferSize)
    
    var wg sync.WaitGroup
    
    // Single dispatcher - no data duplication
    wg.Add(1)
    go func() {
        defer wg.Done()
        defer close(flightCh)
        defer close(engineCh)
        defer close(navigationCh)
        
        for msg := range messages {
            if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        // Route to appropriate processor based on data type
                        switch {
                        case simVar.DefineID <= 10: // Flight instruments
                            select {
                            case flightCh <- simVar:
                            default: // Drop if processor is behind
                                dropCounter.Add(1)
                            }
                        case simVar.DefineID <= 20: // Engine data
                            select {
                            case engineCh <- simVar:
                            default:
                                dropCounter.Add(1)
                            }
                        default: // Navigation
                            select {
                            case navigationCh <- simVar:
                            default:
                                dropCounter.Add(1)
                            }
                        }
                    }
                }
            }
        }
    }()
    
    // Specialized processors
    wg.Add(1)
    go flightProcessor(flightCh, &wg)
    
    wg.Add(1)
    go engineProcessor(engineCh, &wg)
    
    wg.Add(1)
    go navigationProcessor(navigationCh, &wg)
    
    wg.Wait()
}
```

### Worker Pool Pattern

```go
type WorkerPool struct {
    workers   int
    taskQueue chan Task
    wg        sync.WaitGroup
}

type Task struct {
    ID   uint32
    Data interface{}
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
    return &WorkerPool{
        workers:   workers,
        taskQueue: make(chan Task, queueSize),
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
}

func (wp *WorkerPool) Submit(task Task) bool {
    select {
    case wp.taskQueue <- task:
        return true
    default:
        return false // Queue full
    }
}

func (wp *WorkerPool) Stop() {
    close(wp.taskQueue)
    wp.wg.Wait()
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    
    for task := range wp.taskQueue {
        // Process task with minimal allocations
        processTaskEfficient(task)
    }
}

// Usage with SimConnect data
func useWorkerPool() {
    pool := NewWorkerPool(4, 1000) // 4 workers, 1000 task buffer
    pool.Start()
    defer pool.Stop()
    
    messages := sdk.Listen()
    for msg := range messages {
        if msgMap, ok := msg.(map[string]any); ok {
            if data, exists := msgMap["parsed_data"]; exists {
                if simVar, ok := data.(*client.SimVarData); ok {
                    task := Task{
                        ID:   simVar.DefineID,
                        Data: simVar.Value,
                    }
                    
                    if !pool.Submit(task) {
                        // Handle queue full condition
                        queueFullCounter.Add(1)
                    }
                }
            }
        }
    }
}
```

## Resource Management

### Automatic Cleanup with Context

```go
import "context"

type ResourceManager struct {
    sdk            client.Connection
    activeRequests []uint32
    ctx            context.Context
    cancel         context.CancelFunc
}

func NewResourceManager() *ResourceManager {
    ctx, cancel := context.WithCancel(context.Background())
    return &ResourceManager{
        sdk:            client.New("ResourceManager"),
        activeRequests: make([]uint32, 0, 100),
        ctx:            ctx,
        cancel:         cancel,
    }
}

func (rm *ResourceManager) StartMonitoring(variables []Variable) error {
    if err := rm.sdk.Open(); err != nil {
        return err
    }

    for _, v := range variables {
        if err := rm.sdk.RegisterSimVarDefinition(v.ID, v.Name, v.Units, v.DataType); err != nil {
            return err
        }

        requestID := v.ID * 100
        if err := rm.sdk.RequestSimVarDataPeriodic(v.ID, requestID, v.Period); err != nil {
            return err
        }

        rm.activeRequests = append(rm.activeRequests, requestID)
    }

    // Start automatic cleanup on context cancellation
    go rm.monitorContext()
    
    return nil
}

func (rm *ResourceManager) monitorContext() {
    <-rm.ctx.Done()
    rm.cleanup()
}

func (rm *ResourceManager) cleanup() {
    for _, requestID := range rm.activeRequests {
        if err := rm.sdk.StopPeriodicRequest(requestID); err != nil {
            log.Printf("Failed to stop request %d: %v", requestID, err)
        }
    }
    rm.activeRequests = rm.activeRequests[:0]
}

func (rm *ResourceManager) Stop() {
    rm.cancel() // Triggers automatic cleanup
    rm.sdk.Close()
}

// Usage with automatic cleanup
func managedMonitoring() {
    rm := NewResourceManager()
    defer rm.Stop() // Ensures cleanup even on panic

    variables := []Variable{
        {ID: 1, Name: "PLANE ALTITUDE", Units: "feet", DataType: types.SIMCONNECT_DATATYPE_FLOAT32, Period: types.SIMCONNECT_PERIOD_SECOND},
        // ... more variables
    }

    if err := rm.StartMonitoring(variables); err != nil {
        log.Fatal(err)
    }

    // Do work...
    time.Sleep(60 * time.Second)
    
    // rm.Stop() called automatically by defer
}
```

### Connection Pool for Multiple Clients

```go
type ConnectionPool struct {
    connections chan client.Connection
    maxSize     int
    appName     string
}

func NewConnectionPool(appName string, maxSize int) *ConnectionPool {
    return &ConnectionPool{
        connections: make(chan client.Connection, maxSize),
        maxSize:     maxSize,
        appName:     appName,
    }
}

func (cp *ConnectionPool) Get() (client.Connection, error) {
    select {
    case conn := <-cp.connections:
        return conn, nil
    default:
        // Create new connection if pool is empty
        conn := client.New(fmt.Sprintf("%s_%d", cp.appName, time.Now().UnixNano()))
        if err := conn.Open(); err != nil {
            return nil, err
        }
        return conn, nil
    }
}

func (cp *ConnectionPool) Put(conn client.Connection) {
    select {
    case cp.connections <- conn:
        // Connection returned to pool
    default:
        // Pool is full, close connection
        conn.Close()
    }
}

func (cp *ConnectionPool) Close() {
    close(cp.connections)
    for conn := range cp.connections {
        conn.Close()
    }
}

// Usage for high-concurrency scenarios
func useConnectionPool() {
    pool := NewConnectionPool("HighConcurrency", 10)
    defer pool.Close()

    var wg sync.WaitGroup
    
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            
            conn, err := pool.Get()
            if err != nil {
                log.Printf("Worker %d: Failed to get connection: %v", workerID, err)
                return
            }
            defer pool.Put(conn)
            
            // Use connection for work
            doWork(conn, workerID)
        }(i)
    }
    
    wg.Wait()
}
```

## Profiling and Monitoring

### Performance Metrics Collection

```go
import (
    "expvar"
    "runtime"
    "time"
)

type PerformanceMonitor struct {
    messagesProcessed *expvar.Int
    processingTime    *expvar.Int
    memoryUsage       *expvar.Int
    dropCounter       *expvar.Int
    lastGC            time.Time
}

func NewPerformanceMonitor() *PerformanceMonitor {
    pm := &PerformanceMonitor{
        messagesProcessed: expvar.NewInt("messages_processed_total"),
        processingTime:    expvar.NewInt("processing_time_microseconds"),
        memoryUsage:       expvar.NewInt("memory_usage_bytes"),
        dropCounter:       expvar.NewInt("messages_dropped_total"),
        lastGC:            time.Now(),
    }
    
    // Start metrics collection
    go pm.collectMetrics()
    
    return pm
}

func (pm *PerformanceMonitor) collectMetrics() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        pm.memoryUsage.Set(int64(m.Alloc))
        
        // Check if GC occurred
        if m.LastGC > 0 && time.Unix(0, int64(m.LastGC)).After(pm.lastGC) {
            pm.lastGC = time.Unix(0, int64(m.LastGC))
            log.Printf("GC occurred: %d KB freed", m.Frees*8/1024)
        }
    }
}

func (pm *PerformanceMonitor) TrackMessage(processingTime time.Duration) {
    pm.messagesProcessed.Add(1)
    pm.processingTime.Add(processingTime.Microseconds())
}

func (pm *PerformanceMonitor) TrackDrop() {
    pm.dropCounter.Add(1)
}

// Usage in message processing
func monitoredMessageProcessor(messages <-chan interface{}, pm *PerformanceMonitor) {
    for msg := range messages {
        start := time.Now()
        
        // Process message
        processMessage(msg)
        
        // Track performance
        pm.TrackMessage(time.Since(start))
    }
}
```

### CPU Profiling Integration

```go
import (
    "log"
    "os"
    "runtime/pprof"
)

func profiledExecution() {
    // CPU profiling
    cpuProfile, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer cpuProfile.Close()
    
    if err := pprof.StartCPUProfile(cpuProfile); err != nil {
        log.Fatal(err)
    }
    defer pprof.StopCPUProfile()
    
    // Your SimConnect processing code here
    runSimConnectProcessing()
    
    // Memory profiling
    memProfile, err := os.Create("mem.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer memProfile.Close()
    
    if err := pprof.WriteHeapProfile(memProfile); err != nil {
        log.Fatal(err)
    }
}

// Analyze profiles with:
// go tool pprof cpu.prof
// go tool pprof mem.prof
```

## Common Performance Issues

### Issue 1: Excessive Update Rates

**Problem:** Requesting high-frequency updates for non-critical data.

```go
// ❌ BAD: Everything at maximum frequency
for id := uint32(1); id <= 100; id++ {
    sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
}
```

**Solution:** Use appropriate frequencies based on data importance.

```go
// ✅ GOOD: Frequency matches requirements
criticalVars := []uint32{1, 2, 3} // Altitude, speed, attitude
for _, id := range criticalVars {
    sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
}

normalVars := []uint32{10, 11, 12} // Navigation, fuel, systems
for _, id := range normalVars {
    sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
}
```

### Issue 2: Memory Leaks from Unreleased Resources

**Problem:** Not stopping periodic requests or closing connections.

```go
// ❌ BAD: Resources not cleaned up
func leakyFunction() {
    sdk := client.New("LeakyApp")
    sdk.Open()
    
    sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
    
    // Function returns without cleanup
    // sdk.StopPeriodicRequest(100) - MISSING
    // sdk.Close() - MISSING
}
```

**Solution:** Always clean up resources.

```go
// ✅ GOOD: Proper resource management
func properFunction() {
    sdk := client.New("ProperApp")
    defer sdk.Close()
    
    if err := sdk.Open(); err != nil {
        return
    }
    
    sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
    defer sdk.StopPeriodicRequest(100)
    
    // Do work...
}
```

### Issue 3: Blocking Message Processing

**Problem:** Slow processing blocks the message channel.

```go
// ❌ BAD: Slow processing blocks channel
for msg := range messages {
    // Slow operations on the main message loop
    time.Sleep(100 * time.Millisecond) // Simulating slow work
    processMessage(msg)
}
```

**Solution:** Use buffered channels and worker pools.

```go
// ✅ GOOD: Non-blocking with worker pool
workQueue := make(chan interface{}, 1000)

// Fast message dispatcher
go func() {
    for msg := range messages {
        select {
        case workQueue <- msg:
        default:
            dropCounter.Add(1) // Track drops instead of blocking
        }
    }
}()

// Separate workers for slow processing
for i := 0; i < 4; i++ {
    go func() {
        for msg := range workQueue {
            processMessage(msg) // Can be slow without blocking main loop
        }
    }()
}
```

### Issue 4: String Concatenation in Hot Paths

**Problem:** Creating strings on every message.

```go
// ❌ BAD: String allocation on every message
for msg := range messages {
    logMessage := fmt.Sprintf("Received message type: %s at %s", 
        msg.Type, time.Now().Format(time.RFC3339))
    log.Println(logMessage)
}
```

**Solution:** Pre-allocate strings and use efficient formatting.

```go
// ✅ GOOD: Minimal allocations
var buf strings.Builder
for msg := range messages {
    if shouldLog {
        buf.Reset()
        buf.WriteString("Received message type: ")
        buf.WriteString(msg.Type)
        log.Println(buf.String())
    }
}
```

### Performance Checklist

✅ **Update Frequencies**
- [ ] Use `VISUAL_FRAME` only for critical flight instruments
- [ ] Use `SECOND` for navigation and system data
- [ ] Use `ON_SET` for rarely changing settings
- [ ] Stop unused periodic requests

✅ **Memory Management**
- [ ] Use `defer` for cleanup functions
- [ ] Stop all periodic requests on shutdown
- [ ] Close SDK connections properly
- [ ] Use object pools for high-frequency allocations

✅ **Concurrency**
- [ ] Single `Listen()` call per client
- [ ] Use fan-out pattern for multiple processors
- [ ] Use buffered channels appropriately
- [ ] Avoid shared state without synchronization

✅ **Processing Efficiency**
- [ ] Fast message dispatching
- [ ] Minimize allocations in hot paths
- [ ] Use efficient data structures
- [ ] Profile bottlenecks with pprof

✅ **Resource Monitoring**
- [ ] Track message processing rates
- [ ] Monitor memory usage
- [ ] Log dropped messages
- [ ] Use performance metrics for optimization

Following these performance guidelines will ensure your SimConnect applications run efficiently and can handle high-throughput scenarios without impacting the simulator or system performance.
