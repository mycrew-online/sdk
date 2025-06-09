# SimVars Reference Guide

Guide for using Microsoft Flight Simulator simulation variables with the mycrew-online/sdk.

## Official Documentation

For the complete and authoritative reference of all available simulation variables, please refer to the official Microsoft Flight Simulator documentation:

**📚 [Official SimVars Documentation](https://docs.flightsimulator.com/html/Programming_Tools/SimVars/Simulation_Variables.htm)**

This documentation provides:
- Complete list of all simulation variables
- Proper units for each variable
- Variable descriptions and usage notes
- Categorized organization (Aircraft, Environment, etc.)

## Table of Contents

- [Using SimVars with the SDK](#using-simvars-with-the-sdk)
- [Data Type Guidelines](#data-type-guidelines)
- [Common Examples](#common-examples)
- [Best Practices](#best-practices)

## Using SimVars with the SDK

### Basic Workflow

1. **Find the Variable**: Look up the variable name and units in the [official documentation](https://docs.flightsimulator.com/html/Programming_Tools/SimVars/Simulation_Variables.htm)
2. **Choose Data Type**: Select appropriate `types.SimConnectDataType` based on the variable's expected values
3. **Register Definition**: Use `RegisterSimVarDefinition()` with the variable name, units, and data type
4. **Request Data**: Use `RequestSimVarData()` or `RequestSimVarDataPeriodic()` to get values
5. **Process Results**: Handle incoming data through the `Listen()` channel

### Quick Example

```go
// Register altitude variable (from official docs: "PLANE ALTITUDE", units: "feet")
sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)

// Request periodic updates
sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_SECOND)

// Process data
for msg := range sdk.Listen() {
    if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
        if data, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := data.(*client.SimVarData); ok && simVar.DefineID == 1 {
                altitude := simVar.Value.(float64)
                fmt.Printf("Altitude: %.0f feet\n", altitude)
            }
        }
    }
}
```

## Data Type Guidelines

Choose the appropriate SimConnect data type based on the variable's expected values:

| Variable Type | Recommended Data Type | Go Type After Parsing |
|---------------|----------------------|----------------------|
| Altitude, Speed, Temperature | `SIMCONNECT_DATATYPE_FLOAT32` | `float64` |
| Switch States (On/Off) | `SIMCONNECT_DATATYPE_INT32` | `int32` |
| Aircraft Title, Airport Names | `SIMCONNECT_DATATYPE_STRINGV` | `string` |
| Large Numbers (64-bit) | `SIMCONNECT_DATATYPE_FLOAT64` | `float64` |
| Precise Integers | `SIMCONNECT_DATATYPE_INT64` | `int64` |

### Boolean Variables

Many SimConnect variables use "Bool" units but should be registered with `INT32` data type:

```go
// ✅ Correct
sdk.RegisterSimVarDefinition(2, "ELECTRICAL MASTER BATTERY", "Bool", types.SIMCONNECT_DATATYPE_INT32)

// ❌ Incorrect - there's no boolean data type in SimConnect
sdk.RegisterSimVarDefinition(2, "ELECTRICAL MASTER BATTERY", "Bool", types.SIMCONNECT_DATATYPE_BOOL)
```

## Common Examples

### Flight Instruments
```go
// From official docs - Aircraft SimVars section
sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(2, "AIRSPEED INDICATED", "knots", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(3, "VERTICAL SPEED", "feet per minute", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(4, "PLANE HEADING DEGREES TRUE", "degrees", types.SIMCONNECT_DATATYPE_FLOAT32)
```

### Aircraft Systems
```go
// From official docs - Aircraft Systems section
sdk.RegisterSimVarDefinition(10, "ELECTRICAL MASTER BATTERY", "Bool", types.SIMCONNECT_DATATYPE_INT32)
sdk.RegisterSimVarDefinition(11, "LIGHT BEACON", "Bool", types.SIMCONNECT_DATATYPE_INT32)
sdk.RegisterSimVarDefinition(12, "GEAR HANDLE POSITION", "Bool", types.SIMCONNECT_DATATYPE_INT32)
sdk.RegisterSimVarDefinition(13, "GENERAL ENG RPM:1", "rpm", types.SIMCONNECT_DATATYPE_FLOAT32)
```

### Aircraft Information
```go
// From official docs - Aircraft Information section
sdk.RegisterSimVarDefinition(20, "TITLE", "", types.SIMCONNECT_DATATYPE_STRINGV)
sdk.RegisterSimVarDefinition(21, "ATC ID", "", types.SIMCONNECT_DATATYPE_STRINGV)
```

## Best Practices

### Update Frequencies

Choose appropriate update periods based on how the data changes:

- **`SIMCONNECT_PERIOD_VISUAL_FRAME`** - Use sparingly, only for rapidly changing display values
- **`SIMCONNECT_PERIOD_SECOND`** - Good for most analog instruments (altitude, speed, etc.)
- **`SIMCONNECT_PERIOD_ON_SET`** - Perfect for switches and states that only change when toggled

```go
// High frequency for smooth attitude display
sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)

// Medium frequency for instruments
sdk.RequestSimVarDataPeriodic(2, 200, types.SIMCONNECT_PERIOD_SECOND)

// Only when changed for switches
sdk.RequestSimVarDataPeriodic(3, 300, types.SIMCONNECT_PERIOD_ON_SET)
```

### Variable Discovery

1. **Check Official Docs First**: Always start with the [official SimVars documentation](https://docs.flightsimulator.com/html/Programming_Tools/SimVars/Simulation_Variables.htm)
2. **Use Exact Names**: Variable names are case-sensitive and must match exactly
3. **Verify Units**: Use the exact unit strings from the documentation
4. **Test Small**: Start with a few variables and expand gradually

### Error Handling

```go
// Always check for exceptions when registering variables
for msg := range sdk.Listen() {
    if msgMap, ok := msg.(map[string]any); ok {
        switch msgMap["type"] {
        case "EXCEPTION":
            if exception, exists := msgMap["exception"]; exists {
                if exc, ok := exception.(*types.ExceptionData); ok {
                    if exc.ExceptionName == "SIMCONNECT_EXCEPTION_NAME_UNRECOGNIZED" {
                        fmt.Printf("Variable name not recognized: check official docs\n")
                    }
                }
            }
        }
    }
}
```

## Resources

- **[Official SimVars Documentation](https://docs.flightsimulator.com/html/Programming_Tools/SimVars/Simulation_Variables.htm)** - Complete variable reference
- **[SimConnect SDK Reference](https://docs.flightsimulator.com/html/Programming_Tools/SimConnect/SimConnect_SDK.htm)** - Full SimConnect documentation
- **[API Reference](API.md)** - This SDK's API documentation
- **[Examples Guide](EXAMPLES.md)** - Working code examples

## See Also

- **[Getting Started](GETTING_STARTED.md)** - Setup and installation
- **[API Reference](API.md)** - Complete API documentation
- **[Examples Guide](EXAMPLES.md)** - Complete working examples
- **[Error Handling](ERROR_HANDLING.md)** - Exception management

## Aircraft-Specific Variables

### Standard vs. Aircraft-Specific SimVars

**Important Note**: Not all simulation variables work across all aircraft. The standard SimVars in the official documentation work for most general aviation aircraft, but modern complex aircraft (especially third-party add-ons) may have:

1. **Different variable names** for the same functionality
2. **Custom variables** not found in the standard documentation
3. **Non-functional standard variables** that don't reflect the actual aircraft state

### Local Variables (L:Vars)

For aircraft-specific functionality, many developers use **Local Variables (L:Vars)**. These are custom variables defined by the aircraft developer and are not part of the standard SimConnect documentation.

**L:Var Format**: `L:VariableName`

#### Using L:Vars with the SDK

```go
// Example: Aircraft-specific autopilot engagement
sdk.RegisterSimVarDefinition(100, "L:AP_Master_Switch", "Bool", types.SIMCONNECT_DATATYPE_INT32)

// Example: Custom fuel pump states  
sdk.RegisterSimVarDefinition(101, "L:Fuel_Pump_Left", "Bool", types.SIMCONNECT_DATATYPE_INT32)
sdk.RegisterSimVarDefinition(102, "L:Fuel_Pump_Right", "Bool", types.SIMCONNECT_DATATYPE_INT32)

// Example: Engine start sequence state
sdk.RegisterSimVarDefinition(103, "L:Engine_Start_Sequence", "Enum", types.SIMCONNECT_DATATYPE_INT32)
```

#### Finding Aircraft-Specific Variables

1. **Aircraft Documentation**: Check the aircraft manual or developer documentation
2. **Community Resources**: Look for aircraft-specific variable lists on forums
3. **SimConnect Variable Explorer**: Use tools like FSUIPC or SimConnect Variable Inspector
4. **Trial and Error**: Test variables to see if they respond correctly

#### L:Var Discovery Example

```go
// Test if a potential L:Var exists and responds
sdk.RegisterSimVarDefinition(200, "L:Custom_Switch_State", "Bool", types.SIMCONNECT_DATATYPE_INT32)
sdk.RequestSimVarData(200, 1000)

// In your message processing loop, watch for exceptions:
for msg := range sdk.Listen() {
    if msgMap, ok := msg.(map[string]any); ok {
        switch msgMap["type"] {
        case "EXCEPTION":
            if exception, exists := msgMap["exception"]; exists {
                if exc, ok := exception.(*types.ExceptionData); ok {
                    if exc.ExceptionName == "SIMCONNECT_EXCEPTION_NAME_UNRECOGNIZED" {
                        fmt.Printf("L:Var 'L:Custom_Switch_State' not available in this aircraft\n")
                    }
                }
            }
        case "SIMOBJECT_DATA":
            // L:Var found and working
            fmt.Printf("L:Var responded successfully\n")
        }
    }
}
```

### Best Practices for Aircraft Compatibility

1. **Test with Target Aircraft**: Always test your variables with the specific aircraft you're targeting
2. **Fallback Strategies**: Implement fallbacks to standard SimVars when L:Vars aren't available
3. **Aircraft Detection**: Use the `TITLE` SimVar to detect aircraft type and adjust variable usage accordingly
4. **Error Handling**: Always handle `SIMCONNECT_EXCEPTION_NAME_UNRECOGNIZED` exceptions gracefully

#### Aircraft Detection Pattern

```go
// Detect aircraft type first
sdk.RegisterSimVarDefinition(1, "TITLE", "", types.SIMCONNECT_DATATYPE_STRINGV)
sdk.RequestSimVarData(1, 100)

var aircraftTitle string
for msg := range sdk.Listen() {
    if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
        if data, exists := msgMap["parsed_data"]; exists {
            if simVar, ok := data.(*client.SimVarData); ok && simVar.DefineID == 1 {
                aircraftTitle = simVar.Value.(string)
                
                // Use different variables based on aircraft
                if strings.Contains(aircraftTitle, "A320") {
                    // Use A320-specific L:Vars
                    sdk.RegisterSimVarDefinition(10, "L:A32NX_ENGINE_STATE", "Enum", types.SIMCONNECT_DATATYPE_INT32)
                } else {
                    // Use standard SimVars
                    sdk.RegisterSimVarDefinition(10, "ENG COMBUSTION:1", "Bool", types.SIMCONNECT_DATATYPE_INT32)
                }
            }
        }
    }
}
```

### Additional Resources

- **[SimConnect AddToDataDefinition Reference](https://docs.flightsimulator.com/html/Programming_Tools/SimConnect/API_Reference/Events_And_Data/SimConnect_AddToDataDefinition.htm)** - Official documentation on data definitions including L:Vars
- **Aircraft-specific documentation** from the aircraft developer
- **Community forums** and variable databases for specific aircraft models
