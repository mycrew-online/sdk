# External Power Logger

This example demonstrates how to monitor the "EXTERNAL POWER ON" SimConnect variable for debugging periodic update issues. It's designed to help identify problems with variable monitoring and periodic data requests.

## Purpose

This script focuses specifically on logging messages from the `EXTERNAL POWER ON` variable (DefineID: 1) to help diagnose:

- Periodic update frequency issues
- Missing state change notifications  
- Duplicate message detection
- Message timing analysis

## Features

- **Real-time Monitoring**: Tracks the EXTERNAL POWER ON state using `SIMCONNECT_PERIOD_VISUAL_FRAME` for maximum responsiveness
- **State Change Detection**: Logs every time the external power state changes with timestamps
- **Statistics Tracking**: Shows message counts, update rates, and timing information
- **Duplicate Detection**: Identifies when the same value is received multiple times
- **Graceful Shutdown**: Clean exit with final statistics when interrupted

## Usage

### Prerequisites

1. Microsoft Flight Simulator must be running
2. An aircraft must be loaded
3. SimConnect SDK must be installed

### Running the Logger

#### Option 1: Use the Test Suite (Recommended)

```powershell
# Navigate to the example directory
cd example\external-power-logger

# Run the interactive test suite
.\test-suite.bat
```

This will give you options to run different test modes including period comparisons.

#### Option 2: Manual Build and Run

```powershell
# Build the standard logger
go build -o external-power-logger.exe main.go

# Run the standard logger
.\external-power-logger.exe
```

#### Option 3: Period Testing

```powershell
# Build the period test version
go build -o period-test.exe period-test.go

# Test different update periods
.\period-test.exe visual_frame  # High frequency (30-60 Hz)
.\period-test.exe second        # Once per second
.\period-test.exe on_set        # Only on state changes
```

#### Option 4: Quick Start

```powershell
# Just run the simple batch file
.\run.bat
```

### Testing External Power Changes

To test the monitoring, you can change the external power state in several ways:

1. **Using Aircraft Controls**: 
   - Look for external power switches in the aircraft's electrical panel
   - Toggle the external power connection

2. **Using Flight Simulator Menus**:
   - Go to Aircraft > Fuel and Payload
   - Toggle external power options if available

3. **Using Keyboard Shortcuts**:
   - Some aircraft may have assigned keyboard shortcuts for external power

4. **Using the Web Demo**:
   - If you have the web-demo example running, use its "Toggle External Power" button

## Output Explanation

### Real-time Messages

```
🔌 Initial state: EXTERNAL POWER OFF 🔴 (value: 0)
🔄 EXTERNAL POWER changed: OFF 🔴 -> ON 🟢 (value: 0 -> 1) [after 5.234s]
📈 Received 100 messages (current state: EXTERNAL POWER ON 🟢)
```

### Statistics (every 10 seconds)

```
📊 Statistics (after 30s):
   • Total messages: 1,847
   • State changes: 3  
   • Duplicate values: 1,844
   • Messages/second: 61.6
   • Avg time between changes: 10s
```

### Final Statistics (on exit)

```
📊 Final Statistics:
   • Total runtime: 1m23s
   • Total messages received: 5,089
   • State changes detected: 5
   • Duplicate messages: 5,084
   • Average messages per second: 61.3
   • Duplicate message percentage: 99.9%
   • Average time between state changes: 16.6s
```

## Understanding the Results

### Normal Behavior

- **High message rate**: 30-60 messages per second is normal for `VISUAL_FRAME` period
- **High duplicate percentage**: 99%+ duplicates is expected when state doesn't change frequently
- **Consistent timing**: Messages should arrive at regular intervals

### Potential Issues to Look For

- **Missing state changes**: If you toggle external power but no change is logged
- **Delayed updates**: Long gaps between expected state changes
- **Irregular timing**: Significant variations in message arrival times
- **No messages**: Complete absence of messages indicates connection issues

## Troubleshooting

### No Messages Received

1. Verify Flight Simulator is running
2. Ensure an aircraft is loaded (not in main menu)
3. Check SimConnect connection in Flight Simulator settings
4. Try restarting the application

### Missing State Changes

1. Verify the aircraft supports external power simulation
2. Try different aircraft (some may not fully implement electrical systems)
3. Check aircraft documentation for proper external power operation
4. Use aircraft with detailed electrical systems (e.g., study-level aircraft)

### High CPU Usage

This is normal when using `VISUAL_FRAME` period. The script processes 30-60 messages per second to catch all state changes immediately.

## Technical Details

- **Variable**: `EXTERNAL POWER ON`
- **Units**: `Bool`
- **Data Type**: `SIMCONNECT_DATATYPE_INT32`
- **Update Period**: `SIMCONNECT_PERIOD_VISUAL_FRAME`
- **Define ID**: 1
- **Request ID**: 100

## Files

- `main.go` - Main logger application (uses VISUAL_FRAME period)
- `period-test.go` - Period comparison tester (supports different update periods)
- `go.mod` - Go module configuration
- `README.md` - This documentation
- `run.bat` - Simple batch file to build and run standard logger
- `test-suite.bat` - Interactive test suite with multiple options

## Period Testing

The `period-test.go` application allows you to test different update periods to diagnose periodic update issues:

### Available Periods

1. **visual_frame**: Updates every visual frame (30-60 Hz)
   - Pros: Immediate state change detection
   - Cons: High CPU usage, many duplicate messages
   - Best for: Real-time monitoring, debugging timing issues

2. **second**: Updates once per second (1 Hz)
   - Pros: Lower CPU usage, more predictable timing
   - Cons: State changes may be delayed up to 1 second
   - Best for: General monitoring, reduced resource usage

3. **on_set**: Updates only when the value changes
   - Pros: Minimal CPU usage, no duplicate messages
   - Cons: May miss rapid state changes, debugging periodic issues
   - Best for: Efficient change-only monitoring

### Period Comparison Results

Use the period test to identify which update method works best for your use case:

```powershell
# Test high-frequency updates
.\period-test.exe visual_frame

# Test low-frequency updates  
.\period-test.exe second

# Test change-only updates
.\period-test.exe on_set
```

Each test will show timing statistics to help identify periodic update issues.

## Related Examples

- `../web-demo/` - Full web interface with external power control
- `../cmd/main.go` - General SimConnect testing with multiple variables
