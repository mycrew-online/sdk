# API Reference

Complete reference for the SimConnect Go SDK API.

## Table of Contents

- [Client Creation](#client-creation)
- [Connection Management](#connection-management)
- [SimVar Operations](#simvar-operations)
- [Event Management](#event-management)
- [Data Types](#data-types)
- [Error Handling](#error-handling)

## Client Creation

### `client.New(name string) Connection`

Creates a new SimConnect client with the default DLL path.

**Parameters:**
- `name` (string): Application name that appears in SimConnect

**Returns:**
- `Connection`: Interface for interacting with SimConnect

**Example:**
```go
sdk := client.New("MyFlightApp")
```

### `client.NewWithCustomDLL(name string, path string) Connection`

Creates a new SimConnect client with a custom DLL path.

**Parameters:**
- `name` (string): Application name
- `path` (string): Full path to SimConnect.dll

**Returns:**
- `Connection`: Interface for interacting with SimConnect

**Example:**
```go
sdk := client.NewWithCustomDLL("MyApp", "D:/Custom/SimConnect.dll")
```

## Connection Management

### `Open() error`

Establishes connection to Microsoft Flight Simulator.

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
if err := sdk.Open(); err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
```

### `Close() error`

Closes the SimConnect connection and cleans up resources.

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
defer func() {
    if err := sdk.Close(); err != nil {
        log.Printf("Error closing: %v", err)
    }
}()
```

### `Listen() <-chan any`

Returns a read-only channel for receiving SimConnect messages.

**Returns:**
- `<-chan any`: Channel containing parsed message data

**Example:**
```go
messages := sdk.Listen()
for msg := range messages {
    // Process messages
}
```

## SimVar Operations

### `RegisterSimVarDefinition(defID uint32, varName string, units string, dataType types.SimConnectDataType) error`

Registers a simulation variable for data requests.

**Parameters:**
- `defID` (uint32): Unique definition identifier
- `varName` (string): SimConnect variable name (e.g., "PLANE ALTITUDE")
- `units` (string): Variable units (e.g., "feet", "knots") or empty string
- `dataType` (types.SimConnectDataType): Expected data type

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
err := sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
```

### `RequestSimVarData(defID uint32, requestID uint32) error`

Requests a one-time data snapshot for a registered variable.

**Parameters:**
- `defID` (uint32): Previously registered definition ID
- `requestID` (uint32): Unique request identifier

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
err := sdk.RequestSimVarData(1, 100)
```

### `RequestSimVarDataPeriodic(defID uint32, requestID uint32, period types.SimConnectPeriod) error`

Requests continuous data updates at specified intervals.

**Parameters:**
- `defID` (uint32): Previously registered definition ID
- `requestID` (uint32): Unique request identifier
- `period` (types.SimConnectPeriod): Update frequency

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
// Update every visual frame
err := sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
```

### `StopPeriodicRequest(requestID uint32) error`

Stops a previously started periodic data request.

**Parameters:**
- `requestID` (uint32): Request ID to stop

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
err := sdk.StopPeriodicRequest(100)
```

### `SetSimVar(defID uint32, value interface{}) error`

Sets a simulation variable value in the simulator.

**Parameters:**
- `defID` (uint32): Previously registered definition ID  
- `value` (interface{}): Value to set (int32, float32/64, string)

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
// Set camera state
err := sdk.SetSimVar(3, int32(2))
```

## Event Management

### `SubscribeToSystemEvent(eventID uint32, eventName string) error`

Subscribes to system events from the simulator.

**Parameters:**
- `eventID` (uint32): Unique event identifier
- `eventName` (string): System event name (e.g., "Pause", "Sim", "AircraftLoaded")

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
// Subscribe to events that may carry data values
err := sdk.SubscribeToSystemEvent(1010, "Pause")        // May carry pause reason
err := sdk.SubscribeToSystemEvent(1020, "AircraftLoaded") // May carry aircraft ID
err := sdk.SubscribeToSystemEvent(1030, "Crashed")      // May carry crash type

// Process events with their data values
for msg := range sdk.Listen() {
    if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "EVENT" {
        if eventData, exists := msgMap["event"]; exists {
            if event, ok := eventData.(*types.EventData); ok {
                fmt.Printf("Event: %s, Value: %d\n", event.EventName, event.EventData)
            }
        }
    }
}
```

### `MapClientEventToSimEvent(eventID types.ClientEventID, eventName string) error`

Maps a client event ID to a simulator event name.

**Parameters:**
- `eventID` (types.ClientEventID): Client-defined event identifier
- `eventName` (string): Simulator event name

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
err := sdk.MapClientEventToSimEvent(10011511, "TOGGLE_EXTERNAL_POWER")
```

### `AddClientEventToNotificationGroup(groupID types.NotificationGroupID, eventID types.ClientEventID, maskable bool) error`

Adds a client event to a notification group.

**Parameters:**
- `groupID` (types.NotificationGroupID): Notification group identifier
- `eventID` (types.ClientEventID): Client event ID
- `maskable` (bool): Whether event can be masked

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
err := sdk.AddClientEventToNotificationGroup(2000, 10011511, false)
```

### `SetNotificationGroupPriority(groupID types.NotificationGroupID, priority uint32) error`

Sets the priority level for a notification group.

**Parameters:**
- `groupID` (types.NotificationGroupID): Notification group identifier
- `priority` (uint32): Priority level (use types.SIMCONNECT_GROUP_PRIORITY_* constants)

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
err := sdk.SetNotificationGroupPriority(2000, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST)
```

### `TransmitClientEvent(objectID uint32, eventID types.ClientEventID, data uint32, groupID types.NotificationGroupID, flags uint32) error`

Transmits a client event to the simulator.

**Parameters:**
- `objectID` (uint32): Target object (use types.SIMCONNECT_OBJECT_ID_USER for user aircraft)
- `eventID` (types.ClientEventID): Client event identifier
- `data` (uint32): Event data payload
- `groupID` (types.NotificationGroupID): Notification group
- `flags` (uint32): Event flags

**Returns:**
- `error`: nil on success, error details on failure

**Example:**
```go
// Simple event transmission
err := sdk.TransmitClientEvent(
    types.SIMCONNECT_OBJECT_ID_USER,
    10011511,
    0, // No data value
    2000,
    types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY,
)

// Event with data value
err := sdk.TransmitClientEvent(
    types.SIMCONNECT_OBJECT_ID_USER,
    EVENT_ID_SET_FREQUENCY,
    123450, // Frequency value in Hz * 10
    2000,
    types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY,
)
```

## Data Types

### SimConnect Data Types

```go
const (
    SIMCONNECT_DATATYPE_INVALID      // Invalid data type
    SIMCONNECT_DATATYPE_INT32        // 32-bit signed integer
    SIMCONNECT_DATATYPE_INT64        // 64-bit signed integer
    SIMCONNECT_DATATYPE_FLOAT32      // 32-bit floating point
    SIMCONNECT_DATATYPE_FLOAT64      // 64-bit floating point
    SIMCONNECT_DATATYPE_STRING8      // 8-byte fixed string
    SIMCONNECT_DATATYPE_STRING32     // 32-byte fixed string
    SIMCONNECT_DATATYPE_STRING64     // 64-byte fixed string
    SIMCONNECT_DATATYPE_STRING128    // 128-byte fixed string
    SIMCONNECT_DATATYPE_STRING256    // 256-byte fixed string
    SIMCONNECT_DATATYPE_STRING260    // 260-byte fixed string
    SIMCONNECT_DATATYPE_STRINGV      // Variable-length string
)
```

### Update Periods

```go
const (
    SIMCONNECT_PERIOD_NEVER        // Never send data
    SIMCONNECT_PERIOD_ONCE         // Send once only
    SIMCONNECT_PERIOD_VISUAL_FRAME // Every visual frame
    SIMCONNECT_PERIOD_ON_SET       // When value changes
    SIMCONNECT_PERIOD_SECOND       // Once per second
)
```

### Message Structures

#### SimVarData
```go
type SimVarData struct {
    RequestID uint32      // Request identifier
    DefineID  uint32      // Variable definition ID
    Value     interface{} // Parsed value (int32, float64, string)
}
```

#### EventData
```go
type EventData struct {
    GroupID   uint32  // Notification group ID
    EventID   uint32  // Event identifier
    EventData uint32  // Event data payload - contains the actual event value
    EventType string  // "system" or "client"
    EventName string  // Human-readable event name
}
```

**Event Data Values:**
- System events may carry state information (pause/unpause, crash types, etc.)
- Client events can include custom data values (frequencies, settings, etc.)
- Value of 0 typically indicates a simple trigger event with no additional data
- Non-zero values contain meaningful state or parameter information
    EventType string  // "system" or "client"
    EventName string  // Human-readable event name
}
```

#### ExceptionData
```go
type ExceptionData struct {
    ExceptionCode types.SimConnectException
    ExceptionName string
    Description   string
    SendID        uint32
    Index         uint32
    Severity      string
}
```

## Error Handling

### Exception Types

The SDK handles all SimConnect exceptions and provides detailed error information:

- `SIMCONNECT_EXCEPTION_NONE` - No error
- `SIMCONNECT_EXCEPTION_ERROR` - General error
- `SIMCONNECT_EXCEPTION_SIZE_MISMATCH` - Data size mismatch
- `SIMCONNECT_EXCEPTION_UNRECOGNIZED_ID` - Invalid ID
- `SIMCONNECT_EXCEPTION_UNOPENED` - Connection not open
- `SIMCONNECT_EXCEPTION_VERSION_MISMATCH` - Version incompatibility
- `SIMCONNECT_EXCEPTION_TOO_MANY_GROUPS` - Too many groups (max 20)
- `SIMCONNECT_EXCEPTION_NAME_UNRECOGNIZED` - Invalid event name
- `SIMCONNECT_EXCEPTION_TOO_MANY_EVENT_NAMES` - Too many events (max 1000)
- `SIMCONNECT_EXCEPTION_EVENT_ID_DUPLICATE` - Duplicate event ID
- `SIMCONNECT_EXCEPTION_TOO_MANY_MAPS` - Too many mappings (max 20)
- `SIMCONNECT_EXCEPTION_TOO_MANY_OBJECTS` - Too many objects (max 1000)
- `SIMCONNECT_EXCEPTION_TOO_MANY_REQUESTS` - Too many requests (max 1000)
- `SIMCONNECT_EXCEPTION_INVALID_DATA_TYPE` - Invalid data type
- `SIMCONNECT_EXCEPTION_INVALID_DATA_SIZE` - Invalid data size
- `SIMCONNECT_EXCEPTION_DATA_ERROR` - Generic data error
- `SIMCONNECT_EXCEPTION_INVALID_ARRAY` - Invalid array
- `SIMCONNECT_EXCEPTION_CREATE_OBJECT_FAILED` - Object creation failed
- `SIMCONNECT_EXCEPTION_LOAD_FLIGHTPLAN_FAILED` - Flight plan load failed
- `SIMCONNECT_EXCEPTION_OPERATION_INVALID_FOR_OBJECT_TYPE` - Invalid operation
- `SIMCONNECT_EXCEPTION_ILLEGAL_OPERATION` - Illegal operation
- `SIMCONNECT_EXCEPTION_ALREADY_SUBSCRIBED` - Already subscribed
- `SIMCONNECT_EXCEPTION_INVALID_ENUM` - Invalid enumeration
- `SIMCONNECT_EXCEPTION_DEFINITION_ERROR` - Definition error
- `SIMCONNECT_EXCEPTION_DUPLICATE_ID` - Duplicate ID
- `SIMCONNECT_EXCEPTION_DATUM_ID` - Invalid datum ID
- `SIMCONNECT_EXCEPTION_OUT_OF_BOUNDS` - Value out of bounds
- `SIMCONNECT_EXCEPTION_ALREADY_CREATED` - Already created
- `SIMCONNECT_EXCEPTION_OBJECT_OUTSIDE_REALITY_BUBBLE` - Object outside reality bubble

### Exception Handling Example

```go
for msg := range messages {
    if msgMap, ok := msg.(map[string]any); ok {
        switch msgMap["type"] {
        case "EXCEPTION":
            if exception, exists := msgMap["exception"]; exists {
                if exc, ok := exception.(*types.ExceptionData); ok {
                    fmt.Printf("SimConnect Exception: %s - %s\n", 
                        exc.ExceptionName, exc.Description)
                }
            }
        case "SIMOBJECT_DATA":
            // Handle data normally
        }
    }
}
```

### Error Recovery

```go
// Robust connection handling
func connectWithRetry(sdk client.Connection, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := sdk.Open(); err == nil {
            return nil
        }
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return fmt.Errorf("failed to connect after %d retries", maxRetries)
}
```
