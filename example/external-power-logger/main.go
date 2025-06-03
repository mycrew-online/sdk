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

// Constants for our monitoring
const (
	EXTERNAL_POWER_DEFINE_ID  = 1   // Define ID for EXTERNAL POWER ON variable
	EXTERNAL_POWER_REQUEST_ID = 100 // Request ID for periodic updates
)

func main() {
	fmt.Println("🔌 External Power Monitor - Starting...")
	fmt.Println("   This script monitors EXTERNAL POWER ON for periodic update issues")
	fmt.Println("   Press Ctrl+C to stop monitoring")
	fmt.Println()

	// Create new SimConnect client
	sdk := client.New("ExternalPowerMonitor")
	defer sdk.Close()

	// Connect to SimConnect
	fmt.Println("🔗 Connecting to Microsoft Flight Simulator...")
	if err := sdk.Open(); err != nil {
		log.Fatalf("❌ Failed to connect to SimConnect: %v", err)
	}
	fmt.Println("✅ Connected to Microsoft Flight Simulator!")

	// Register the EXTERNAL POWER ON variable
	fmt.Println("📝 Registering EXTERNAL POWER ON variable...")
	if err := sdk.RegisterSimVarDefinition(
		EXTERNAL_POWER_DEFINE_ID,
		"EXTERNAL POWER ON",
		"Bool",
		types.SIMCONNECT_DATATYPE_INT32,
	); err != nil {
		log.Fatalf("❌ Failed to register EXTERNAL POWER ON: %v", err)
	}
	fmt.Println("✅ EXTERNAL POWER ON variable registered successfully!")

	// Start periodic monitoring
	fmt.Println("⏰ Starting periodic monitoring (every visual frame)...")
	if err := sdk.RequestSimVarDataPeriodic(
		EXTERNAL_POWER_DEFINE_ID,
		EXTERNAL_POWER_REQUEST_ID,
		types.SIMCONNECT_PERIOD_VISUAL_FRAME,
	); err != nil {
		log.Fatalf("❌ Failed to start periodic monitoring: %v", err)
	}
	fmt.Println("✅ Periodic monitoring started!")
	fmt.Println()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start listening for messages
	messages := sdk.Listen()
	if messages == nil {
		log.Fatal("❌ Failed to start listening for messages")
	}

	// Statistics tracking
	var (
		messageCount   int
		lastValue      *int32
		lastChangeTime time.Time
		startTime      = time.Now()
		duplicateCount int
		changeCount    int
	)

	fmt.Println("👂 Listening for EXTERNAL POWER ON messages...")
	fmt.Println("   📊 Statistics will be shown every 10 seconds")
	fmt.Println("   🔄 State changes will be logged immediately")
	fmt.Println()

	// Statistics ticker
	statsTicker := time.NewTicker(10 * time.Second)
	defer statsTicker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutdown signal received...")
			fmt.Println("📊 Final Statistics:")
			printFinalStats(startTime, messageCount, changeCount, duplicateCount)

			fmt.Println("🔌 Stopping periodic monitoring...")
			if err := sdk.StopPeriodicRequest(EXTERNAL_POWER_REQUEST_ID); err != nil {
				fmt.Printf("⚠️  Warning: Failed to stop periodic request: %v\n", err)
			} else {
				fmt.Println("✅ Periodic monitoring stopped")
			}

			fmt.Println("👋 External Power Monitor stopped")
			return

		case <-statsTicker.C:
			printStats(startTime, messageCount, changeCount, duplicateCount)

		case msg, ok := <-messages:
			if !ok {
				fmt.Println("❌ Message channel closed")
				return
			}

			// Parse the message
			msgMap, ok := msg.(map[string]interface{})
			if !ok {
				continue
			}

			// Only process SIMOBJECT_DATA messages
			msgType, exists := msgMap["type"]
			if !exists || msgType != "SIMOBJECT_DATA" {
				continue
			}

			// Check if we have parsed data
			parsedData, exists := msgMap["parsed_data"]
			if !exists {
				continue
			}

			// Cast to SimVarData
			simVarData, ok := parsedData.(*client.SimVarData)
			if !ok {
				continue
			}

			// Only process our EXTERNAL POWER ON variable
			if simVarData.DefineID != EXTERNAL_POWER_DEFINE_ID {
				continue
			}

			messageCount++

			// Extract the current value
			currentValue, ok := simVarData.Value.(int32)
			if !ok {
				fmt.Printf("⚠️  Warning: Unexpected value type: %T (value: %v)\n", simVarData.Value, simVarData.Value)
				continue
			}

			// Convert to boolean for display
			currentState := currentValue != 0

			// Check for state changes
			if lastValue == nil {
				// First message
				fmt.Printf("🔌 Initial state: EXTERNAL POWER %s (value: %d)\n",
					formatPowerState(currentState), currentValue)
				lastValue = &currentValue
				lastChangeTime = time.Now()
				changeCount++
			} else if *lastValue != currentValue {
				// State changed
				now := time.Now()
				timeSinceLastChange := now.Sub(lastChangeTime)

				fmt.Printf("🔄 EXTERNAL POWER changed: %s -> %s (value: %d -> %d) [after %v]\n",
					formatPowerState(*lastValue != 0),
					formatPowerState(currentState),
					*lastValue,
					currentValue,
					timeSinceLastChange)

				*lastValue = currentValue
				lastChangeTime = now
				changeCount++
			} else {
				// Duplicate value
				duplicateCount++
			}

			// Optional: Log periodic updates every 100 messages to show activity
			if messageCount%100 == 0 {
				fmt.Printf("📈 Received %d messages (current state: EXTERNAL POWER %s)\n",
					messageCount, formatPowerState(currentState))
			}
		}
	}
}

func formatPowerState(isOn bool) string {
	if isOn {
		return "ON 🟢"
	}
	return "OFF 🔴"
}

func printStats(startTime time.Time, messageCount, changeCount, duplicateCount int) {
	elapsed := time.Since(startTime)
	messagesPerSecond := float64(messageCount) / elapsed.Seconds()

	fmt.Printf("📊 Statistics (after %v):\n", elapsed.Truncate(time.Second))
	fmt.Printf("   • Total messages: %d\n", messageCount)
	fmt.Printf("   • State changes: %d\n", changeCount)
	fmt.Printf("   • Duplicate values: %d\n", duplicateCount)
	fmt.Printf("   • Messages/second: %.1f\n", messagesPerSecond)

	if changeCount > 0 {
		avgTimeBetweenChanges := elapsed / time.Duration(changeCount)
		fmt.Printf("   • Avg time between changes: %v\n", avgTimeBetweenChanges.Truncate(time.Millisecond))
	}
	fmt.Println()
}

func printFinalStats(startTime time.Time, messageCount, changeCount, duplicateCount int) {
	elapsed := time.Since(startTime)
	messagesPerSecond := float64(messageCount) / elapsed.Seconds()

	fmt.Printf("   • Total runtime: %v\n", elapsed.Truncate(time.Second))
	fmt.Printf("   • Total messages received: %d\n", messageCount)
	fmt.Printf("   • State changes detected: %d\n", changeCount)
	fmt.Printf("   • Duplicate messages: %d\n", duplicateCount)
	fmt.Printf("   • Average messages per second: %.1f\n", messagesPerSecond)

	if messageCount > 0 {
		duplicatePercentage := float64(duplicateCount) / float64(messageCount) * 100
		fmt.Printf("   • Duplicate message percentage: %.1f%%\n", duplicatePercentage)
	}

	if changeCount > 0 {
		avgTimeBetweenChanges := elapsed / time.Duration(changeCount)
		fmt.Printf("   • Average time between state changes: %v\n", avgTimeBetweenChanges.Truncate(time.Millisecond))
	}
}
