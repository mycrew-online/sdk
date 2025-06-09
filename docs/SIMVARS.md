# SimVars Reference Guide

Comprehensive reference for Microsoft Flight Simulator SimConnect variables with units, data types, and usage examples.

## Table of Contents

- [Flight Instruments](#flight-instruments)
- [Aircraft Systems](#aircraft-systems)
- [Engine Parameters](#engine-parameters)
- [Navigation Data](#navigation-data)
- [Environmental Data](#environmental-data)
- [Autopilot Variables](#autopilot-variables)
- [Communication Systems](#communication-systems)
- [Data Type Guidelines](#data-type-guidelines)

## Flight Instruments

### Primary Flight Display

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `PLANE ALTITUDE` | feet | FLOAT32 | Current altitude above sea level | VISUAL_FRAME |
| `INDICATED ALTITUDE` | feet | FLOAT32 | Altimeter reading | VISUAL_FRAME |
| `AIRSPEED INDICATED` | knots | FLOAT32 | Indicated airspeed from pitot tube | VISUAL_FRAME |
| `AIRSPEED TRUE` | knots | FLOAT32 | True airspeed corrected for density | SECOND |
| `GROUND VELOCITY` | knots | FLOAT32 | Speed over ground | SECOND |
| `VERTICAL SPEED` | feet per minute | FLOAT32 | Rate of climb/descent | VISUAL_FRAME |

**Example Usage:**
```go
// Register primary flight instruments
sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(2, "AIRSPEED INDICATED", "knots", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(3, "VERTICAL SPEED", "feet per minute", types.SIMCONNECT_DATATYPE_FLOAT32)

// Request high-frequency updates for smooth display
sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
sdk.RequestSimVarDataPeriodic(2, 200, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
sdk.RequestSimVarDataPeriodic(3, 300, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
```

### Attitude and Heading

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `PLANE PITCH DEGREES` | degrees | FLOAT32 | Pitch angle (-90 to +90) | VISUAL_FRAME |
| `PLANE BANK DEGREES` | degrees | FLOAT32 | Bank angle (-180 to +180) | VISUAL_FRAME |
| `PLANE HEADING DEGREES TRUE` | degrees | FLOAT32 | True heading (0-360) | VISUAL_FRAME |
| `PLANE HEADING DEGREES MAGNETIC` | degrees | FLOAT32 | Magnetic heading (0-360) | VISUAL_FRAME |
| `HEADING INDICATOR` | degrees | FLOAT32 | Gyro compass heading | VISUAL_FRAME |
| `ATTITUDE INDICATOR PITCH DEGREES` | degrees | FLOAT32 | Artificial horizon pitch | VISUAL_FRAME |
| `ATTITUDE INDICATOR BANK DEGREES` | degrees | FLOAT32 | Artificial horizon bank | VISUAL_FRAME |

**Example Usage:**
```go
// Register attitude indicators
sdk.RegisterSimVarDefinition(10, "PLANE PITCH DEGREES", "degrees", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(11, "PLANE BANK DEGREES", "degrees", types.SIMCONNECT_DATATYPE_FLOAT32)
sdk.RegisterSimVarDefinition(12, "HEADING INDICATOR", "degrees", types.SIMCONNECT_DATATYPE_FLOAT32)

// Request updates for smooth attitude display
for id := uint32(10); id <= 12; id++ {
    sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_VISUAL_FRAME)
}
```

## Aircraft Systems

### Electrical System

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `ELECTRICAL MASTER BATTERY` | Bool | INT32 | Master battery switch state | ON_SET |
| `ELECTRICAL MAIN BUS VOLTAGE` | Volts | FLOAT32 | Main electrical bus voltage | SECOND |
| `ELECTRICAL MAIN BUS AMPS` | Amperes | FLOAT32 | Current draw on main bus | SECOND |
| `ELECTRICAL BATTERY VOLTAGE` | Volts | FLOAT32 | Battery voltage | SECOND |
| `ELECTRICAL BATTERY LOAD` | Amperes | FLOAT32 | Battery load current | SECOND |
| `EXTERNAL POWER ON` | Bool | INT32 | External power connected | ON_SET |
| `GENERAL ENG MASTER ALTERNATOR` | Bool | INT32 | Alternator switch state | ON_SET |

**Example Usage:**
```go
// Register electrical system monitoring
electricalVars := []struct {
    id    uint32
    name  string
    units string
    dataType types.SimConnectDataType
}{
    {20, "ELECTRICAL MASTER BATTERY", "Bool", types.SIMCONNECT_DATATYPE_INT32},
    {21, "ELECTRICAL MAIN BUS VOLTAGE", "Volts", types.SIMCONNECT_DATATYPE_FLOAT32},
    {22, "EXTERNAL POWER ON", "Bool", types.SIMCONNECT_DATATYPE_INT32},
}

for _, v := range electricalVars {
    sdk.RegisterSimVarDefinition(v.id, v.name, v.units, v.dataType)
    // Use ON_SET for switches, SECOND for analog values
    period := types.SIMCONNECT_PERIOD_ON_SET
    if v.dataType == types.SIMCONNECT_DATATYPE_FLOAT32 {
        period = types.SIMCONNECT_PERIOD_SECOND
    }
    sdk.RequestSimVarDataPeriodic(v.id, v.id*100, period)
}
```

### Lighting Systems

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `LIGHT BEACON` | Bool | INT32 | Beacon light state | ON_SET |
| `LIGHT NAV` | Bool | INT32 | Navigation lights state | ON_SET |
| `LIGHT STROBE` | Bool | INT32 | Strobe lights state | ON_SET |
| `LIGHT TAXI` | Bool | INT32 | Taxi lights state | ON_SET |
| `LIGHT LANDING` | Bool | INT32 | Landing lights state | ON_SET |
| `LIGHT PANEL` | Bool | INT32 | Panel lights state | ON_SET |
| `LIGHT CABIN` | Bool | INT32 | Cabin lights state | ON_SET |

**Example Usage:**
```go
// Register all lighting systems
lightVars := []string{
    "LIGHT BEACON", "LIGHT NAV", "LIGHT STROBE", 
    "LIGHT TAXI", "LIGHT LANDING", "LIGHT PANEL",
}

for i, lightVar := range lightVars {
    id := uint32(30 + i)
    sdk.RegisterSimVarDefinition(id, lightVar, "Bool", types.SIMCONNECT_DATATYPE_INT32)
    sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_ON_SET)
}
```

### Landing Gear

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `GEAR HANDLE POSITION` | Bool | INT32 | Gear lever position (0=up, 1=down) | ON_SET |
| `GEAR TOTAL PCT EXTENDED` | Percent | FLOAT32 | Overall gear extension (0-100) | VISUAL_FRAME |
| `GEAR LEFT POSITION` | Percent | FLOAT32 | Left gear position (0-100) | SECOND |
| `GEAR CENTER POSITION` | Percent | FLOAT32 | Center gear position (0-100) | SECOND |
| `GEAR RIGHT POSITION` | Percent | FLOAT32 | Right gear position (0-100) | SECOND |
| `GEAR SPEED EXCEEDED` | Bool | INT32 | Gear speed limit exceeded | ON_SET |

## Engine Parameters

### Engine State

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `GENERAL ENG RPM:1` | rpm | FLOAT32 | Engine 1 RPM | SECOND |
| `GENERAL ENG PCT MAX RPM:1` | Percent | FLOAT32 | Engine 1 RPM percentage | SECOND |
| `GENERAL ENG THROTTLE LEVER POSITION:1` | Percent | FLOAT32 | Throttle position (0-100) | VISUAL_FRAME |
| `GENERAL ENG MIXTURE LEVER POSITION:1` | Percent | FLOAT32 | Mixture lever position | SECOND |
| `GENERAL ENG PROPELLER LEVER POSITION:1` | Percent | FLOAT32 | Prop lever position | SECOND |
| `ENG COMBUSTION:1` | Bool | INT32 | Engine combustion state | ON_SET |
| `ENG FAILED:1` | Bool | INT32 | Engine failure state | ON_SET |

**Example Usage:**
```go
// Register engine parameters for engine 1
engineVars := []struct {
    id     uint32
    name   string
    units  string
    period types.SimConnectPeriod
}{
    {40, "GENERAL ENG RPM:1", "rpm", types.SIMCONNECT_PERIOD_SECOND},
    {41, "GENERAL ENG THROTTLE LEVER POSITION:1", "Percent", types.SIMCONNECT_PERIOD_VISUAL_FRAME},
    {42, "ENG COMBUSTION:1", "Bool", types.SIMCONNECT_PERIOD_ON_SET},
}

for _, v := range engineVars {
    dataType := types.SIMCONNECT_DATATYPE_FLOAT32
    if v.units == "Bool" {
        dataType = types.SIMCONNECT_DATATYPE_INT32
    }
    
    sdk.RegisterSimVarDefinition(v.id, v.name, v.units, dataType)
    sdk.RequestSimVarDataPeriodic(v.id, v.id*100, v.period)
}
```

### Fuel System

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `FUEL TOTAL QUANTITY` | Gallons | FLOAT32 | Total fuel in all tanks | SECOND |
| `FUEL TOTAL CAPACITY` | Gallons | FLOAT32 | Maximum fuel capacity | ONCE |
| `FUEL LEFT QUANTITY` | Gallons | FLOAT32 | Left tank fuel quantity | SECOND |
| `FUEL RIGHT QUANTITY` | Gallons | FLOAT32 | Right tank fuel quantity | SECOND |
| `FUEL TOTAL QUANTITY WEIGHT` | Pounds | FLOAT32 | Total fuel weight | SECOND |
| `ESTIMATED FUEL FLOW` | Gallons per hour | FLOAT32 | Current fuel consumption rate | SECOND |

## Navigation Data

### GPS and Position

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `PLANE LATITUDE` | Degrees | FLOAT64 | Current latitude (-90 to +90) | SECOND |
| `PLANE LONGITUDE` | Degrees | FLOAT64 | Current longitude (-180 to +180) | SECOND |
| `GPS GROUND SPEED` | Knots | FLOAT32 | GPS ground speed | SECOND |
| `GPS GROUND TRUE TRACK` | Degrees | FLOAT32 | GPS track over ground | SECOND |
| `GPS WP DISTANCE` | Meters | FLOAT32 | Distance to active waypoint | SECOND |
| `GPS WP BEARING` | Degrees | FLOAT32 | Bearing to active waypoint | SECOND |
| `GPS WP ETE` | Seconds | FLOAT32 | Estimated time to waypoint | SECOND |

**Example Usage:**
```go
// Register GPS navigation data
sdk.RegisterSimVarDefinition(50, "PLANE LATITUDE", "Degrees", types.SIMCONNECT_DATATYPE_FLOAT64)
sdk.RegisterSimVarDefinition(51, "PLANE LONGITUDE", "Degrees", types.SIMCONNECT_DATATYPE_FLOAT64)
sdk.RegisterSimVarDefinition(52, "GPS GROUND SPEED", "Knots", types.SIMCONNECT_DATATYPE_FLOAT32)

// Request position updates at 1Hz
for id := uint32(50); id <= 52; id++ {
    sdk.RequestSimVarDataPeriodic(id, id*100, types.SIMCONNECT_PERIOD_SECOND)
}
```

### Radio Navigation

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `NAV OBS:1` | Degrees | FLOAT32 | VOR OBS setting | ON_SET |
| `NAV RADIAL:1` | Degrees | FLOAT32 | Current radial from VOR | SECOND |
| `NAV CDI:1` | Number | FLOAT32 | Course deviation (-127 to +127) | SECOND |
| `NAV GSI:1` | Number | FLOAT32 | Glideslope deviation | SECOND |
| `NAV HAS NAV:1` | Bool | INT32 | Valid navigation signal | ON_SET |
| `NAV HAS LOCALIZER:1` | Bool | INT32 | Valid localizer signal | ON_SET |
| `NAV HAS GLIDE SLOPE:1` | Bool | INT32 | Valid glideslope signal | ON_SET |

## Environmental Data

### Weather and Atmosphere

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `AMBIENT WIND VELOCITY` | Knots | FLOAT32 | Wind speed at aircraft | SECOND |
| `AMBIENT WIND DIRECTION` | Degrees | FLOAT32 | Wind direction | SECOND |
| `AMBIENT TEMPERATURE` | Celsius | FLOAT32 | Outside air temperature | SECOND |
| `AMBIENT PRESSURE` | inHg | FLOAT32 | Barometric pressure | SECOND |
| `AMBIENT DENSITY` | Slugs per cubic feet | FLOAT32 | Air density | SECOND |
| `AMBIENT VISIBILITY` | Meters | FLOAT32 | Visibility distance | SECOND |
| `AMBIENT CLOUD STATE` | Enum | INT32 | Cloud condition (0=clear, 1=few, etc.) | SECOND |

**Example Usage:**
```go
// Register weather monitoring
weatherVars := []struct {
    id   uint32
    name string
    units string
}{
    {60, "AMBIENT WIND VELOCITY", "Knots"},
    {61, "AMBIENT WIND DIRECTION", "Degrees"},
    {62, "AMBIENT TEMPERATURE", "Celsius"},
    {63, "AMBIENT PRESSURE", "inHg"},
}

for _, v := range weatherVars {
    sdk.RegisterSimVarDefinition(v.id, v.name, v.units, types.SIMCONNECT_DATATYPE_FLOAT32)
    sdk.RequestSimVarDataPeriodic(v.id, v.id*100, types.SIMCONNECT_PERIOD_SECOND)
}
```

## Autopilot Variables

### Autopilot State

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `AUTOPILOT MASTER` | Bool | INT32 | Autopilot master switch | ON_SET |
| `AUTOPILOT HEADING LOCK` | Bool | INT32 | Heading hold mode | ON_SET |
| `AUTOPILOT ALTITUDE LOCK` | Bool | INT32 | Altitude hold mode | ON_SET |
| `AUTOPILOT ATTITUDE HOLD` | Bool | INT32 | Attitude hold mode | ON_SET |
| `AUTOPILOT NAV1 LOCK` | Bool | INT32 | Nav1 tracking mode | ON_SET |
| `AUTOPILOT APPROACH HOLD` | Bool | INT32 | Approach mode | ON_SET |
| `AUTOPILOT BACKCOURSE HOLD` | Bool | INT32 | Backcourse mode | ON_SET |

### Autopilot Settings

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `AUTOPILOT HEADING LOCK DIR` | Degrees | FLOAT32 | Target heading | ON_SET |
| `AUTOPILOT ALTITUDE LOCK VAR` | Feet | FLOAT32 | Target altitude | ON_SET |
| `AUTOPILOT AIRSPEED HOLD VAR` | Knots | FLOAT32 | Target airspeed | ON_SET |
| `AUTOPILOT VERTICAL SPEED HOLD VAR` | Feet per minute | FLOAT32 | Target vertical speed | ON_SET |

## Communication Systems

### Radio Equipment

| Variable Name | Units | Data Type | Description | Update Rate |
|---------------|--------|-----------|-------------|-------------|
| `COM ACTIVE FREQUENCY:1` | MHz | FLOAT32 | Active COM1 frequency | ON_SET |
| `COM STANDBY FREQUENCY:1` | MHz | FLOAT32 | Standby COM1 frequency | ON_SET |
| `NAV ACTIVE FREQUENCY:1` | MHz | FLOAT32 | Active NAV1 frequency | ON_SET |
| `NAV STANDBY FREQUENCY:1` | MHz | FLOAT32 | Standby NAV1 frequency | ON_SET |
| `TRANSPONDER CODE:1` | Number | INT32 | Transponder squawk code | ON_SET |
| `COM RECEIVE ALL` | Bool | INT32 | Receive on all COM radios | ON_SET |

## Data Type Guidelines

### Choosing the Right Data Type

| SimConnect Type | Go Type | Use Case | Memory | Notes |
|-----------------|---------|----------|--------|-------|
| `SIMCONNECT_DATATYPE_INT32` | `int32` | Boolean states, discrete values | 4 bytes | Use for on/off switches |
| `SIMCONNECT_DATATYPE_FLOAT32` | `float64` | Most analog values | 4 bytes | Standard for flight data |
| `SIMCONNECT_DATATYPE_FLOAT64` | `float64` | High precision (lat/lon) | 8 bytes | GPS coordinates |
| `SIMCONNECT_DATATYPE_STRINGV` | `string` | Variable text | Variable | Aircraft names, IDs |

### Boolean Values

```go
// Boolean variables use INT32 type with Bool units
sdk.RegisterSimVarDefinition(1, "ELECTRICAL MASTER BATTERY", "Bool", types.SIMCONNECT_DATATYPE_INT32)

// Check boolean state
if msgMap["type"] == "SIMOBJECT_DATA" {
    if data, exists := msgMap["parsed_data"]; exists {
        if simVar, ok := data.(*client.SimVarData); ok {
            isOn := simVar.Value.(int32) != 0
            fmt.Printf("Battery: %s\n", map[bool]string{true: "ON", false: "OFF"}[isOn])
        }
    }
}
```

### High-Precision Values

```go
// Use FLOAT64 for coordinates requiring high precision
sdk.RegisterSimVarDefinition(10, "PLANE LATITUDE", "Degrees", types.SIMCONNECT_DATATYPE_FLOAT64)
sdk.RegisterSimVarDefinition(11, "PLANE LONGITUDE", "Degrees", types.SIMCONNECT_DATATYPE_FLOAT64)

// Access high-precision values
if simVar.DefineID == 10 || simVar.DefineID == 11 {
    coordinate := simVar.Value.(float64)
    fmt.Printf("Coordinate: %.6f degrees\n", coordinate)
}
```

### String Values

```go
// Use STRINGV for variable-length strings
sdk.RegisterSimVarDefinition(20, "TITLE", "String", types.SIMCONNECT_DATATYPE_STRINGV)

// Process string data
if simVar.DefineID == 20 {
    aircraftTitle := simVar.Value.(string)
    fmt.Printf("Aircraft: %s\n", aircraftTitle)
}
```

### Performance Tips

1. **Use appropriate update rates** - Don't request VISUAL_FRAME updates for slowly changing data
2. **Group related variables** - Register similar variables with consecutive IDs for easier management
3. **Use ON_SET for switches** - Boolean states typically only need updates when they change
4. **Choose minimal data types** - Use INT32 for booleans, FLOAT32 for most analog values
5. **Stop unused requests** - Always clean up periodic requests to free resources

### Common Units Reference

| Measurement | Common Units | Notes |
|-------------|--------------|-------|
| Distance | feet, meters, nautical miles | |
| Speed | knots, feet per minute, meters per second | |
| Angle | degrees, radians | Most angles use degrees |
| Temperature | celsius, fahrenheit | |
| Pressure | inHg, millibars, PSI | |
| Frequency | MHz, KHz | Radio frequencies |
| Fuel | gallons, liters, pounds | |
| Boolean | Bool | Always use with INT32 |
| Percentage | Percent | 0-100 scale |

This reference provides the most commonly used SimConnect variables with their proper units and data types. For a complete list of all available variables, consult the official SimConnect SDK documentation.
