# Getting Started Guide

Complete setup guide for the SimConnect Go SDK with troubleshooting and installation details.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [First Connection](#first-connection)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

## Prerequisites

### Microsoft Flight Simulator

The SDK requires Microsoft Flight Simulator 2024 or 2020 to be installed and running:

- **MSFS 2024** - Recommended, latest features and compatibility
- **MSFS 2020** - Fully supported, extensive testing

### SimConnect SDK

SimConnect is typically installed automatically with MSFS:

**Default Locations:**
- MSFS 2024: `C:/MSFS 2024 SDK/SimConnect SDK/lib/SimConnect.dll`
- MSFS 2020: `C:/MSFS SDK/SimConnect SDK/lib/SimConnect.dll`

**Custom Installations:**
If you have a custom SimConnect installation, note the DLL path for later use.

### Go Environment

- **Go 1.21+** required for module support and generics
- **Windows OS** - SimConnect is Windows-only
- **CGO enabled** - Required for DLL interaction

Verify your Go installation:
```bash
go version  # Should show 1.21 or higher
```

## Installation

### 1. Initialize Your Go Project

```bash
# Create new project
mkdir my-flight-app
cd my-flight-app
go mod init my-flight-app

# Or add to existing project
go mod tidy
```

### 2. Install the SDK

```bash
go get github.com/mycrew-online/sdk
```

### 3. Verify Installation

Create a simple test file `test.go`:

```go
package main

import (
    "fmt"
    "github.com/mycrew-online/sdk/pkg/client"
)

func main() {
    sdk := client.New("TestApp")
    fmt.Println("✅ SDK imported successfully")
    defer sdk.Close()
}
```

Run the test:
```bash
go run test.go
```

If you see "✅ SDK imported successfully", the installation is complete.

## First Connection

### 1. Start Microsoft Flight Simulator

- Launch MSFS 2024 or 2020
- Load any aircraft (default flight is fine)
- Ensure the simulator is not paused

### 2. Create Your First Connection

Create `main.go`:

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
    // Create SDK client
    sdk := client.New("MyFirstApp")
    defer sdk.Close()

    // Attempt connection
    fmt.Println("🔌 Connecting to Microsoft Flight Simulator...")
    if err := sdk.Open(); err != nil {
        log.Fatalf("❌ Connection failed: %v", err)
    }
    
    fmt.Println("✅ Successfully connected to MSFS!")
    
    // Test basic functionality
    err := sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32)
    if err != nil {
        log.Fatalf("❌ Failed to register variable: %v", err)
    }
    
    err = sdk.RequestSimVarData(1, 100)
    if err != nil {
        log.Fatalf("❌ Failed to request data: %v", err)
    }
    
    // Listen for response
    messages := sdk.Listen()
    timeout := time.After(5 * time.Second)
    
    fmt.Println("📊 Requesting altitude data...")
    
    for {
        select {
        case msg := <-messages:
            if msgMap, ok := msg.(map[string]any); ok && msgMap["type"] == "SIMOBJECT_DATA" {
                if data, exists := msgMap["parsed_data"]; exists {
                    if simVar, ok := data.(*client.SimVarData); ok {
                        fmt.Printf("✈️ Current altitude: %.0f feet\n", simVar.Value)
                        fmt.Println("🎉 First connection successful!")
                        return
                    }
                }
            }
        case <-timeout:
            fmt.Println("⏰ Timeout waiting for data - but connection was successful!")
            return
        }
    }
}
```

### 3. Run Your First App

```bash
go run main.go
```

**Expected Output:**
```
🔌 Connecting to Microsoft Flight Simulator...
✅ Successfully connected to MSFS!
📊 Requesting altitude data...
✈️ Current altitude: 2156 feet
🎉 First connection successful!
```

## Troubleshooting

### Connection Issues

#### Error: "Failed to load SimConnect.dll"

**Cause:** SimConnect DLL not found in expected location.

**Solutions:**
1. **Verify MSFS Installation:**
   ```bash
   # Check if file exists
   dir "C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll"
   ```

2. **Use Custom DLL Path:**
   ```go
   sdk := client.NewWithCustomDLL("MyApp", "D:/CustomPath/SimConnect.dll")
   ```

3. **Check Environment Variables:**
   Some installations use environment variables like `MSFS_SDK`.

#### Error: "Connection refused" or "No connection to simulator"

**Cause:** MSFS is not running or SimConnect is disabled.

**Solutions:**
1. **Start MSFS First:** Always start the simulator before running your app
2. **Load Aircraft:** Ensure you're in an active flight, not just the main menu
3. **Check SimConnect Settings:** In MSFS, go to Options > General > Developers and ensure SimConnect is enabled

#### Error: "Access denied" or Permission errors

**Cause:** Windows permissions or antivirus blocking DLL access.

**Solutions:**
1. **Run as Administrator:** Try running your app as administrator
2. **Antivirus Exclusions:** Add your Go workspace to antivirus exclusions
3. **Windows Defender:** Ensure Windows Defender isn't blocking the DLL

### Runtime Issues

#### No Data Received

**Symptoms:** Connection succeeds but no data arrives.

**Debug Steps:**
1. **Check Message Types:**
   ```go
   for msg := range messages {
       if msgMap, ok := msg.(map[string]any); ok {
           fmt.Printf("Message type: %s\n", msgMap["type"])
       }
   }
   ```

2. **Look for Exceptions:**
   ```go
   case "EXCEPTION":
       if exception, exists := msgMap["exception"]; exists {
           fmt.Printf("Exception: %+v\n", exception)
       }
   ```

3. **Verify Variable Names:** Ensure SimVar names are correct (see [SimVars Reference](SIMVARS.md))

#### Memory or Performance Issues

**Symptoms:** High CPU usage, memory leaks, or slowdowns.

**Solutions:**
1. **Limit Update Frequency:**
   ```go
   // Use SECOND instead of VISUAL_FRAME for non-critical data
   sdk.RequestSimVarDataPeriodic(1, 100, types.SIMCONNECT_PERIOD_SECOND)
   ```

2. **Stop Unused Requests:**
   ```go
   defer sdk.StopPeriodicRequest(100)
   ```

3. **Use Appropriate Data Types:**
   ```go
   // Use INT32 for boolean values, not FLOAT32
   sdk.RegisterSimVarDefinition(1, "ELECTRICAL MASTER BATTERY", "Bool", types.SIMCONNECT_DATATYPE_INT32)
   ```

### Build Issues

#### CGO Errors

**Error:** "CGO_ENABLED required" or "gcc not found"

**Solutions:**
1. **Enable CGO:**
   ```bash
   set CGO_ENABLED=1  # Windows CMD
   $env:CGO_ENABLED=1 # PowerShell
   ```

2. **Install GCC:** Install TDM-GCC or MinGW-w64
3. **Alternative:** Use pre-built binaries if available

#### Module Issues

**Error:** "module not found" or version conflicts

**Solutions:**
1. **Update Dependencies:**
   ```bash
   go mod tidy
   go mod download
   ```

2. **Clear Module Cache:**
   ```bash
   go clean -modcache
   go mod download
   ```

3. **Check Go Version:**
   ```bash
   go version  # Ensure 1.21+
   ```

## Next Steps

### Learn the Basics

1. **[Examples](EXAMPLES.md)** - Start with basic monitoring examples
2. **[API Reference](API.md)** - Understand available methods and parameters
3. **[SimVars Reference](SIMVARS.md)** - Common variables and their units

### Explore Advanced Features

1. **[Advanced Usage](ADVANCED_USAGE.md)** - Concurrent processing and multiple clients
2. **[Performance Guide](PERFORMANCE.md)** - Optimization techniques
3. **[Production Deployment](PRODUCTION.md)** - Production-ready patterns

### Common Next Projects

**Beginner:**
- Altitude and speed monitor
- Basic autopilot status display
- Simple system state checker

**Intermediate:**
- Flight data logger
- Custom instrument panel
- Aircraft systems controller

**Advanced:**
- Real-time flight tracking
- Multi-client monitoring system
- Web-based flight dashboard

### Getting Help

1. **Check Logs:** Enable detailed logging for debugging
2. **Review Examples:** Study working examples in the `example/` directory
3. **Community:** Check GitHub issues and discussions
4. **Documentation:** Refer to complete API documentation

**Ready to build something amazing?** Start with the [Examples](EXAMPLES.md) guide!

---

> **Tip:** Keep the simulator running while developing - connecting and disconnecting repeatedly can help identify connection issues early.
