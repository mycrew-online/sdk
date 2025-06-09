# Error Handling and Recovery Guide

Comprehensive guide for handling errors, implementing recovery strategies, and building resilient SimConnect applications.

## Table of Contents

- [Error Types](#error-types)
- [Exception Handling](#exception-handling)
- [Connection Recovery](#connection-recovery)
- [Graceful Degradation](#graceful-degradation)
- [Logging and Monitoring](#logging-and-monitoring)
- [Production Patterns](#production-patterns)

## Error Types

### Connection Errors

**Common causes:**
- MSFS not running
- SimConnect DLL not found
- Permission issues
- Network connectivity problems

```go
func handleConnectionError(err error) error {
    switch {
    case strings.Contains(err.Error(), "SimConnect.dll"):
        return fmt.Errorf("SimConnect DLL not found - ensure MSFS SDK is installed: %w", err)
    case strings.Contains(err.Error(), "connection refused"):
        return fmt.Errorf("MSFS not running - please start the simulator: %w", err)
    case strings.Contains(err.Error(), "access denied"):
        return fmt.Errorf("permission denied - try running as administrator: %w", err)
    default:
        return fmt.Errorf("connection failed: %w", err)
    }
}

func connectWithRetry(sdk client.Connection, maxRetries int) error {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        if err := sdk.Open(); err != nil {
            wrappedErr := handleConnectionError(err)
            
            if attempt == maxRetries {
                return fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, wrappedErr)
            }
            
            backoff := time.Duration(attempt) * time.Second
            log.Printf("Connection attempt %d failed: %v. Retrying in %v...", attempt, wrappedErr, backoff)
            time.Sleep(backoff)
            continue
        }
        
        log.Printf("✅ Connected successfully on attempt %d", attempt)
        return nil
    }
    
    return fmt.Errorf("unexpected error: retry loop completed without success")
}
```

### SimConnect Exceptions

SimConnect exceptions provide detailed error information:

```go
type ExceptionHandler struct {
    recoverableExceptions map[types.SimConnectException]bool
    exceptionCounts      map[types.SimConnectException]int
    lastException        map[types.SimConnectException]time.Time
    mu                   sync.RWMutex
}

func NewExceptionHandler() *ExceptionHandler {
    return &ExceptionHandler{
        recoverableExceptions: map[types.SimConnectException]bool{
            types.SIMCONNECT_EXCEPTION_SIZE_MISMATCH:        true,
            types.SIMCONNECT_EXCEPTION_UNRECOGNIZED_ID:      true,
            types.SIMCONNECT_EXCEPTION_INVALID_DATA_TYPE:    true,
            types.SIMCONNECT_EXCEPTION_INVALID_DATA_SIZE:    true,
            types.SIMCONNECT_EXCEPTION_DATA_ERROR:           true,
            // Non-recoverable exceptions
            types.SIMCONNECT_EXCEPTION_UNOPENED:             false,
            types.SIMCONNECT_EXCEPTION_VERSION_MISMATCH:     false,
            types.SIMCONNECT_EXCEPTION_TOO_MANY_GROUPS:      false,
            types.SIMCONNECT_EXCEPTION_TOO_MANY_REQUESTS:    false,
        },
        exceptionCounts: make(map[types.SimConnectException]int),
        lastException:   make(map[types.SimConnectException]time.Time),
    }
}

func (eh *ExceptionHandler) HandleException(exception *types.ExceptionData) error {
    eh.mu.Lock()
    defer eh.mu.Unlock()

    exceptionCode := exception.ExceptionCode
    eh.exceptionCounts[exceptionCode]++
    eh.lastException[exceptionCode] = time.Now()

    // Log exception details
    log.Printf("❌ SimConnect Exception: %s (%s)", 
        exception.ExceptionName, exception.Description)
    log.Printf("   SendID: %d, Index: %d, Severity: %s", 
        exception.SendID, exception.Index, exception.Severity)

    // Check if exception is recoverable
    recoverable, exists := eh.recoverableExceptions[exceptionCode]
    if !exists {
        recoverable = false // Unknown exceptions are not recoverable
    }

    if !recoverable {
        return fmt.Errorf("non-recoverable SimConnect exception: %s", exception.ExceptionName)
    }

    // Check for repeated exceptions (possible infinite loop)
    if eh.exceptionCounts[exceptionCode] > 10 {
        return fmt.Errorf("too many repeated exceptions: %s (count: %d)", 
            exception.ExceptionName, eh.exceptionCounts[exceptionCode])
    }

    // Suggest recovery action
    eh.suggestRecoveryAction(exception)
    return nil // Recoverable exception
}

func (eh *ExceptionHandler) suggestRecoveryAction(exception *types.ExceptionData) {
    switch exception.ExceptionCode {
    case types.SIMCONNECT_EXCEPTION_UNRECOGNIZED_ID:
        log.Printf("💡 Recovery: Check DefineID %d - may need to re-register variable", exception.Index)
    case types.SIMCONNECT_EXCEPTION_INVALID_DATA_TYPE:
        log.Printf("💡 Recovery: Verify data type for variable at index %d", exception.Index)
    case types.SIMCONNECT_EXCEPTION_SIZE_MISMATCH:
        log.Printf("💡 Recovery: Check data size expectations for request %d", exception.SendID)
    case types.SIMCONNECT_EXCEPTION_DATA_ERROR:
        log.Printf("💡 Recovery: Data validation error - check variable name and units")
    }
}

func (eh *ExceptionHandler) GetExceptionStats() map[types.SimConnectException]int {
    eh.mu.RLock()
    defer eh.mu.RUnlock()
    
    stats := make(map[types.SimConnectException]int)
    for exc, count := range eh.exceptionCounts {
        stats[exc] = count
    }
    return stats
}
```

### Data Validation Errors

```go
type DataValidator struct {
    expectedRanges map[uint32]ValueRange
    validationErrors map[uint32]int
}

type ValueRange struct {
    Min, Max float64
    Units    string
}

func NewDataValidator() *DataValidator {
    return &DataValidator{
        expectedRanges: map[uint32]ValueRange{
            1: {Min: -1000, Max: 60000, Units: "feet"},    // Altitude
            2: {Min: 0, Max: 500, Units: "knots"},          // Airspeed
            3: {Min: 0, Max: 360, Units: "degrees"},        // Heading
            4: {Min: -6000, Max: 6000, Units: "fpm"},       // Vertical speed
        },
        validationErrors: make(map[uint32]int),
    }
}

func (dv *DataValidator) ValidateData(simVar *client.SimVarData) error {
    expectedRange, exists := dv.expectedRanges[simVar.DefineID]
    if !exists {
        return nil // No validation rule for this variable
    }

    value, ok := simVar.Value.(float64)
    if !ok {
        if floatVal, ok := simVar.Value.(float32); ok {
            value = float64(floatVal)
        } else {
            return fmt.Errorf("unexpected data type for DefineID %d: %T", simVar.DefineID, simVar.Value)
        }
    }

    if value < expectedRange.Min || value > expectedRange.Max {
        dv.validationErrors[simVar.DefineID]++
        
        // Log validation error but don't fail immediately
        log.Printf("⚠️ Data validation warning: DefineID %d value %.2f %s outside expected range [%.2f, %.2f]",
            simVar.DefineID, value, expectedRange.Units, expectedRange.Min, expectedRange.Max)
        
        // If too many validation errors, something is seriously wrong
        if dv.validationErrors[simVar.DefineID] > 100 {
            return fmt.Errorf("too many validation errors for DefineID %d: value %.2f %s", 
                simVar.DefineID, value, expectedRange.Units)
        }
    }

    return nil
}
```

## Exception Handling

### Comprehensive Exception Processor

```go
type ResilientMessageProcessor struct {
    sdk              client.Connection
    exceptionHandler *ExceptionHandler
    dataValidator    *DataValidator
    reconnectChan    chan struct{}
    stopChan         chan struct{}
    wg              sync.WaitGroup
}

func NewResilientMessageProcessor(appName string) *ResilientMessageProcessor {
    return &ResilientMessageProcessor{
        sdk:              client.New(appName),
        exceptionHandler: NewExceptionHandler(),
        dataValidator:    NewDataValidator(),
        reconnectChan:    make(chan struct{}, 1),
        stopChan:         make(chan struct{}),
    }
}

func (rmp *ResilientMessageProcessor) Start() error {
    if err := connectWithRetry(rmp.sdk, 5); err != nil {
        return err
    }

    // Start message processing
    rmp.wg.Add(1)
    go rmp.processMessages()

    // Start reconnection monitor
    rmp.wg.Add(1)
    go rmp.reconnectionMonitor()

    return nil
}

func (rmp *ResilientMessageProcessor) processMessages() {
    defer rmp.wg.Done()

    for {
        select {
        case <-rmp.stopChan:
            return
        case <-rmp.reconnectChan:
            // Reconnection requested
            if err := rmp.handleReconnection(); err != nil {
                log.Printf("❌ Reconnection failed: %v", err)
                time.Sleep(5 * time.Second) // Wait before next attempt
                continue
            }
        default:
            // Normal message processing
            rmp.processMessageBatch()
        }
    }
}

func (rmp *ResilientMessageProcessor) processMessageBatch() {
    messages := rmp.sdk.Listen()
    if messages == nil {
        log.Println("⚠️ Message channel is nil, requesting reconnection")
        select {
        case rmp.reconnectChan <- struct{}{}:
        default:
        }
        return
    }

    // Process messages with timeout
    timeout := time.After(10 * time.Second)
    messageCount := 0
    
    for {
        select {
        case msg, ok := <-messages:
            if !ok {
                log.Println("⚠️ Message channel closed, requesting reconnection")
                select {
                case rmp.reconnectChan <- struct{}{}:
                default:
                }
                return
            }
            
            if err := rmp.processMessage(msg); err != nil {
                log.Printf("❌ Message processing error: %v", err)
                if strings.Contains(err.Error(), "non-recoverable") {
                    select {
                    case rmp.reconnectChan <- struct{}{}:
                    default:
                    }
                    return
                }
            }
            
            messageCount++
            
        case <-timeout:
            // No message timeout - this might indicate connection issues
            if messageCount == 0 {
                log.Println("⚠️ No messages received in 10 seconds, checking connection")
                select {
                case rmp.reconnectChan <- struct{}{}:
                default:
                }
            }
            return
            
        case <-rmp.stopChan:
            return
        }
    }
}

func (rmp *ResilientMessageProcessor) processMessage(msg any) error {
    msgMap, ok := msg.(map[string]any)
    if !ok {
        return fmt.Errorf("invalid message format: %T", msg)
    }

    switch msgMap["type"] {
    case "SIMOBJECT_DATA":
        return rmp.processSimObjectData(msgMap)
    case "EXCEPTION":
        return rmp.processException(msgMap)
    case "EVENT":
        return rmp.processEvent(msgMap)
    default:
        // Unknown message type - not an error
        return nil
    }
}

func (rmp *ResilientMessageProcessor) processSimObjectData(msgMap map[string]any) error {
    if data, exists := msgMap["parsed_data"]; exists {
        if simVar, ok := data.(*client.SimVarData); ok {
            // Validate data
            if err := rmp.dataValidator.ValidateData(simVar); err != nil {
                return fmt.Errorf("data validation failed: %w", err)
            }
            
            // Process valid data
            rmp.handleValidData(simVar)
        }
    }
    return nil
}

func (rmp *ResilientMessageProcessor) processException(msgMap map[string]any) error {
    if exception, exists := msgMap["exception"]; exists {
        if exceptionData, ok := exception.(*types.ExceptionData); ok {
            return rmp.exceptionHandler.HandleException(exceptionData)
        }
    }
    return nil
}

func (rmp *ResilientMessageProcessor) handleReconnection() error {
    log.Println("🔄 Starting reconnection process...")
    
    // Close existing connection
    if err := rmp.sdk.Close(); err != nil {
        log.Printf("⚠️ Error closing connection: %v", err)
    }
    
    // Wait before reconnecting
    time.Sleep(2 * time.Second)
    
    // Attempt reconnection
    if err := connectWithRetry(rmp.sdk, 5); err != nil {
        return fmt.Errorf("reconnection failed: %w", err)
    }
    
    // Re-register all variables and requests
    if err := rmp.reregisterVariables(); err != nil {
        return fmt.Errorf("failed to re-register variables: %w", err)
    }
    
    log.Println("✅ Reconnection successful")
    return nil
}

func (rmp *ResilientMessageProcessor) Stop() {
    close(rmp.stopChan)
    rmp.wg.Wait()
    rmp.sdk.Close()
}
```

## Connection Recovery

### Automatic Reconnection Manager

```go
type ReconnectionManager struct {
    sdk                client.Connection
    appName           string
    reconnectInterval time.Duration
    maxReconnectAttempts int
    variablesToRegister []VariableDefinition
    activeRequests     []RequestDefinition
    isConnected       bool
    mu                sync.RWMutex
    stopChan          chan struct{}
    wg                sync.WaitGroup
}

type VariableDefinition struct {
    DefineID uint32
    VarName  string
    Units    string
    DataType types.SimConnectDataType
}

type RequestDefinition struct {
    DefineID  uint32
    RequestID uint32
    Period    types.SimConnectPeriod
}

func NewReconnectionManager(appName string) *ReconnectionManager {
    return &ReconnectionManager{
        appName:               appName,
        reconnectInterval:     5 * time.Second,
        maxReconnectAttempts: 10,
        stopChan:             make(chan struct{}),
    }
}

func (rm *ReconnectionManager) RegisterVariable(defID uint32, varName, units string, dataType types.SimConnectDataType) error {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    // Store variable definition for re-registration
    rm.variablesToRegister = append(rm.variablesToRegister, VariableDefinition{
        DefineID: defID,
        VarName:  varName,
        Units:    units,
        DataType: dataType,
    })

    // Register immediately if connected
    if rm.isConnected {
        return rm.sdk.RegisterSimVarDefinition(defID, varName, units, dataType)
    }

    return nil
}

func (rm *ReconnectionManager) RequestDataPeriodic(defID, reqID uint32, period types.SimConnectPeriod) error {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    // Store request definition for re-registration
    rm.activeRequests = append(rm.activeRequests, RequestDefinition{
        DefineID:  defID,
        RequestID: reqID,
        Period:    period,
    })

    // Request immediately if connected
    if rm.isConnected {
        return rm.sdk.RequestSimVarDataPeriodic(defID, reqID, period)
    }

    return nil
}

func (rm *ReconnectionManager) Start() error {
    rm.sdk = client.New(rm.appName)
    
    if err := rm.connect(); err != nil {
        return err
    }

    // Start connection monitor
    rm.wg.Add(1)
    go rm.connectionMonitor()

    return nil
}

func (rm *ReconnectionManager) connect() error {
    if err := connectWithRetry(rm.sdk, 3); err != nil {
        return err
    }

    rm.mu.Lock()
    rm.isConnected = true
    rm.mu.Unlock()

    // Re-register all variables and requests
    return rm.reregisterAll()
}

func (rm *ReconnectionManager) reregisterAll() error {
    rm.mu.RLock()
    variables := make([]VariableDefinition, len(rm.variablesToRegister))
    copy(variables, rm.variablesToRegister)
    requests := make([]RequestDefinition, len(rm.activeRequests))
    copy(requests, rm.activeRequests)
    rm.mu.RUnlock()

    // Re-register variables
    for _, variable := range variables {
        if err := rm.sdk.RegisterSimVarDefinition(variable.DefineID, variable.VarName, variable.Units, variable.DataType); err != nil {
            return fmt.Errorf("failed to re-register variable %s: %w", variable.VarName, err)
        }
    }

    // Re-register requests
    for _, request := range requests {
        if err := rm.sdk.RequestSimVarDataPeriodic(request.DefineID, request.RequestID, request.Period); err != nil {
            return fmt.Errorf("failed to re-register request %d: %w", request.RequestID, err)
        }
    }

    log.Printf("✅ Re-registered %d variables and %d requests", len(variables), len(requests))
    return nil
}

func (rm *ReconnectionManager) connectionMonitor() {
    defer rm.wg.Done()

    ticker := time.NewTicker(rm.reconnectInterval)
    defer ticker.Stop()

    for {
        select {
        case <-rm.stopChan:
            return
        case <-ticker.C:
            rm.mu.RLock()
            connected := rm.isConnected
            rm.mu.RUnlock()

            if !connected {
                rm.attemptReconnection()
            }
        }
    }
}

func (rm *ReconnectionManager) attemptReconnection() {
    log.Println("🔄 Attempting reconnection...")

    for attempt := 1; attempt <= rm.maxReconnectAttempts; attempt++ {
        // Close existing connection
        if rm.sdk != nil {
            rm.sdk.Close()
        }

        // Create new connection
        rm.sdk = client.New(fmt.Sprintf("%s_reconnect_%d", rm.appName, attempt))

        if err := rm.connect(); err != nil {
            log.Printf("Reconnection attempt %d failed: %v", attempt, err)
            if attempt < rm.maxReconnectAttempts {
                time.Sleep(time.Duration(attempt) * rm.reconnectInterval)
                continue
            }
            log.Printf("❌ All reconnection attempts failed")
            return
        }

        log.Printf("✅ Reconnected successfully on attempt %d", attempt)
        return
    }
}

func (rm *ReconnectionManager) IsConnected() bool {
    rm.mu.RLock()
    defer rm.mu.RUnlock()
    return rm.isConnected
}

func (rm *ReconnectionManager) Listen() <-chan any {
    return rm.sdk.Listen()
}

func (rm *ReconnectionManager) Stop() {
    close(rm.stopChan)
    rm.wg.Wait()
    
    rm.mu.Lock()
    rm.isConnected = false
    rm.mu.Unlock()
    
    if rm.sdk != nil {
        rm.sdk.Close()
    }
}
```

## Graceful Degradation

### Fallback Data Sources

```go
type FallbackDataManager struct {
    primary     client.Connection
    secondary   client.Connection
    cacheData   map[uint32]CachedValue
    primaryFailed bool
    mu          sync.RWMutex
}

type CachedValue struct {
    Value     interface{}
    Timestamp time.Time
    Source    string
}

func NewFallbackDataManager() *FallbackDataManager {
    return &FallbackDataManager{
        primary:   client.New("Primary"),
        secondary: client.New("Secondary"),
        cacheData: make(map[uint32]CachedValue),
    }
}

func (fdm *FallbackDataManager) Initialize() error {
    // Try primary connection
    if err := fdm.primary.Open(); err != nil {
        log.Printf("⚠️ Primary connection failed, trying secondary: %v", err)
        fdm.primaryFailed = true
        
        // Fallback to secondary
        if err := fdm.secondary.Open(); err != nil {
            return fmt.Errorf("both primary and secondary connections failed: %w", err)
        }
        
        return fdm.setupDataSources(fdm.secondary, "secondary")
    }

    return fdm.setupDataSources(fdm.primary, "primary")
}

func (fdm *FallbackDataManager) GetValue(defineID uint32) (interface{}, bool) {
    fdm.mu.RLock()
    defer fdm.mu.RUnlock()

    cached, exists := fdm.cacheData[defineID]
    if !exists {
        return nil, false
    }

    // Check if cached value is too old
    if time.Since(cached.Timestamp) > 30*time.Second {
        log.Printf("⚠️ Cached value for DefineID %d is stale (age: %v)", 
            defineID, time.Since(cached.Timestamp))
        return cached.Value, false // Return value but indicate it's stale
    }

    return cached.Value, true
}

func (fdm *FallbackDataManager) updateCache(defineID uint32, value interface{}, source string) {
    fdm.mu.Lock()
    defer fdm.mu.Unlock()

    fdm.cacheData[defineID] = CachedValue{
        Value:     value,
        Timestamp: time.Now(),
        Source:    source,
    }
}
```

### Degraded Mode Operation

```go
type DegradedModeManager struct {
    normalMode    bool
    degradedMode  bool
    criticalVars  map[uint32]bool
    updateRates   map[uint32]types.SimConnectPeriod
    originalRates map[uint32]types.SimConnectPeriod
    sdk          client.Connection
    mu           sync.RWMutex
}

func NewDegradedModeManager(sdk client.Connection) *DegradedModeManager {
    return &DegradedModeManager{
        sdk:           sdk,
        normalMode:    true,
        criticalVars:  make(map[uint32]bool),
        updateRates:   make(map[uint32]types.SimConnectPeriod),
        originalRates: make(map[uint32]types.SimConnectPeriod),
    }
}

func (dmm *DegradedModeManager) MarkCritical(defineID uint32) {
    dmm.mu.Lock()
    defer dmm.mu.Unlock()
    dmm.criticalVars[defineID] = true
}

func (dmm *DegradedModeManager) EnterDegradedMode() error {
    dmm.mu.Lock()
    defer dmm.mu.Unlock()

    if dmm.degradedMode {
        return nil // Already in degraded mode
    }

    log.Println("⚠️ Entering degraded mode - reducing update frequencies")

    // Reduce update frequencies for non-critical variables
    for defineID, currentRate := range dmm.updateRates {
        dmm.originalRates[defineID] = currentRate

        if !dmm.criticalVars[defineID] {
            // Reduce frequency for non-critical variables
            degradedRate := dmm.getDegradedRate(currentRate)
            
            // Stop current request
            if err := dmm.sdk.StopPeriodicRequest(defineID * 100); err != nil {
                log.Printf("⚠️ Failed to stop request %d: %v", defineID*100, err)
            }
            
            // Restart with degraded rate
            if err := dmm.sdk.RequestSimVarDataPeriodic(defineID, defineID*100, degradedRate); err != nil {
                log.Printf("⚠️ Failed to restart request %d with degraded rate: %v", defineID*100, err)
            } else {
                dmm.updateRates[defineID] = degradedRate
            }
        }
    }

    dmm.normalMode = false
    dmm.degradedMode = true
    return nil
}

func (dmm *DegradedModeManager) ExitDegradedMode() error {
    dmm.mu.Lock()
    defer dmm.mu.Unlock()

    if !dmm.degradedMode {
        return nil // Not in degraded mode
    }

    log.Println("✅ Exiting degraded mode - restoring normal frequencies")

    // Restore original update frequencies
    for defineID, originalRate := range dmm.originalRates {
        // Stop current request
        if err := dmm.sdk.StopPeriodicRequest(defineID * 100); err != nil {
            log.Printf("⚠️ Failed to stop request %d: %v", defineID*100, err)
        }
        
        // Restart with original rate
        if err := dmm.sdk.RequestSimVarDataPeriodic(defineID, defineID*100, originalRate); err != nil {
            log.Printf("⚠️ Failed to restart request %d with original rate: %v", defineID*100, err)
        } else {
            dmm.updateRates[defineID] = originalRate
        }
    }

    dmm.normalMode = true
    dmm.degradedMode = false
    return nil
}

func (dmm *DegradedModeManager) getDegradedRate(currentRate types.SimConnectPeriod) types.SimConnectPeriod {
    switch currentRate {
    case types.SIMCONNECT_PERIOD_VISUAL_FRAME:
        return types.SIMCONNECT_PERIOD_SECOND
    case types.SIMCONNECT_PERIOD_SECOND:
        return types.SIMCONNECT_PERIOD_SECOND // Keep same for already slow rates
    default:
        return currentRate
    }
}
```

## Logging and Monitoring

### Comprehensive Error Logger

```go
type ErrorLogger struct {
    logFile        *os.File
    errorCounts    map[string]int
    lastErrors     map[string]time.Time
    alertThresholds map[string]int
    mu             sync.RWMutex
}

func NewErrorLogger(filename string) (*ErrorLogger, error) {
    logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to open log file: %w", err)
    }

    return &ErrorLogger{
        logFile:         logFile,
        errorCounts:     make(map[string]int),
        lastErrors:      make(map[string]time.Time),
        alertThresholds: map[string]int{
            "connection_error":  3,
            "exception_error":   10,
            "validation_error":  100,
            "processing_error":  50,
        },
    }, nil
}

func (el *ErrorLogger) LogError(errorType string, err error) {
    el.mu.Lock()
    defer el.mu.Unlock()

    timestamp := time.Now()
    el.errorCounts[errorType]++
    el.lastErrors[errorType] = timestamp

    // Write to log file
    logEntry := fmt.Sprintf("[%s] %s: %v\n", 
        timestamp.Format("2006-01-02 15:04:05"), errorType, err)
    
    if _, writeErr := el.logFile.WriteString(logEntry); writeErr != nil {
        log.Printf("Failed to write to error log: %v", writeErr)
    }
    el.logFile.Sync()

    // Check alert thresholds
    if threshold, exists := el.alertThresholds[errorType]; exists {
        if el.errorCounts[errorType] >= threshold {
            el.triggerAlert(errorType, el.errorCounts[errorType])
        }
    }

    // Also log to console for immediate visibility
    log.Printf("❌ %s: %v", errorType, err)
}

func (el *ErrorLogger) triggerAlert(errorType string, count int) {
    alertMsg := fmt.Sprintf("🚨 ALERT: %s has occurred %d times", errorType, count)
    log.Println(alertMsg)
    
    // Write alert to log file
    alertEntry := fmt.Sprintf("[%s] ALERT: %s count reached %d\n", 
        time.Now().Format("2006-01-02 15:04:05"), errorType, count)
    el.logFile.WriteString(alertEntry)
    el.logFile.Sync()
}

func (el *ErrorLogger) GetErrorSummary() map[string]ErrorSummary {
    el.mu.RLock()
    defer el.mu.RUnlock()

    summary := make(map[string]ErrorSummary)
    for errorType, count := range el.errorCounts {
        summary[errorType] = ErrorSummary{
            Count:    count,
            LastSeen: el.lastErrors[errorType],
        }
    }
    return summary
}

type ErrorSummary struct {
    Count    int
    LastSeen time.Time
}

func (el *ErrorLogger) Close() error {
    if el.logFile != nil {
        return el.logFile.Close()
    }
    return nil
}
```

## Production Patterns

### Production-Ready Error Handler

```go
type ProductionErrorHandler struct {
    appName         string
    errorLogger     *ErrorLogger
    healthChecker   *HealthChecker
    alerter         *Alerter
    reconnectMgr    *ReconnectionManager
    degradedMgr     *DegradedModeManager
    circuitBreaker  *CircuitBreaker
}

func NewProductionErrorHandler(appName string) (*ProductionErrorHandler, error) {
    errorLogger, err := NewErrorLogger(fmt.Sprintf("%s_errors.log", appName))
    if err != nil {
        return nil, err
    }

    return &ProductionErrorHandler{
        appName:        appName,
        errorLogger:    errorLogger,
        healthChecker:  NewHealthChecker(),
        alerter:        NewAlerter(),
        reconnectMgr:   NewReconnectionManager(appName),
        circuitBreaker: NewCircuitBreaker(),
    }, nil
}

func (peh *ProductionErrorHandler) HandleError(errorType string, err error, severity ErrorSeverity) error {
    // Log all errors
    peh.errorLogger.LogError(errorType, err)
    
    // Update health status
    peh.healthChecker.RecordError(errorType, severity)

    // Handle based on severity
    switch severity {
    case SeverityCritical:
        return peh.handleCriticalError(errorType, err)
    case SeverityHigh:
        return peh.handleHighSeverityError(errorType, err)
    case SeverityMedium:
        return peh.handleMediumSeverityError(errorType, err)
    case SeverityLow:
        return peh.handleLowSeverityError(errorType, err)
    }

    return nil
}

func (peh *ProductionErrorHandler) handleCriticalError(errorType string, err error) error {
    // Send immediate alert
    peh.alerter.SendCriticalAlert(errorType, err)
    
    // Enter degraded mode if available
    if peh.degradedMgr != nil {
        peh.degradedMgr.EnterDegradedMode()
    }
    
    // Attempt immediate reconnection
    if strings.Contains(errorType, "connection") {
        go peh.reconnectMgr.attemptReconnection()
    }
    
    return fmt.Errorf("critical error requires immediate attention: %w", err)
}

type ErrorSeverity int

const (
    SeverityLow ErrorSeverity = iota
    SeverityMedium
    SeverityHigh
    SeverityCritical
)

// Circuit breaker pattern for repeated failures
type CircuitBreaker struct {
    failureCount    int
    lastFailureTime time.Time
    state          CircuitState
    threshold      int
    timeout        time.Duration
    mu             sync.RWMutex
}

type CircuitState int

const (
    CircuitClosed CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)

func NewCircuitBreaker() *CircuitBreaker {
    return &CircuitBreaker{
        threshold: 5,
        timeout:   30 * time.Second,
        state:     CircuitClosed,
    }
}

func (cb *CircuitBreaker) Call(operation func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case CircuitOpen:
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = CircuitHalfOpen
            cb.failureCount = 0
        } else {
            return fmt.Errorf("circuit breaker is open, refusing call")
        }
    case CircuitHalfOpen:
        // Allow limited calls to test if service recovered
    case CircuitClosed:
        // Normal operation
    }

    // Execute operation
    err := operation()
    
    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()
        
        if cb.failureCount >= cb.threshold {
            cb.state = CircuitOpen
            log.Printf("🔴 Circuit breaker opened after %d failures", cb.failureCount)
        }
        
        return err
    }

    // Success - reset circuit breaker
    if cb.state == CircuitHalfOpen {
        cb.state = CircuitClosed
        log.Println("🟢 Circuit breaker closed - service recovered")
    }
    cb.failureCount = 0
    
    return nil
}
```

This comprehensive error handling guide provides the patterns and tools needed to build resilient SimConnect applications that can handle errors gracefully, recover from failures automatically, and maintain operational visibility in production environments.
