package main

import (
	"fmt"
	"time"

	"github.com/mycrew-online/sdk/pkg/client"
)

func main() {
	fmt.Println("🚀 Testing graceful shutdown with channels...")

	// Create a new SimConnect client
	sdk := client.New("GracefulShutdownTest")
	defer func() {
		fmt.Println("🔄 Closing connection...")
		if err := sdk.Close(); err != nil {
			fmt.Printf("❌ Error closing: %v\n", err)
		} else {
			fmt.Println("✅ Connection closed gracefully")
		}
	}()

	// Try to open connection (this might fail if MSFS is not running)
	fmt.Println("📡 Attempting to connect to SimConnect...")
	if err := sdk.Open(); err != nil {
		fmt.Printf("⚠️  Connection failed (MSFS not running?): %v\n", err)
		fmt.Println("✅ Testing shutdown without connection...")
		return
	}

	fmt.Println("✅ Connected successfully!")

	// Start listening for messages
	messages := sdk.Listen()
	if messages == nil {
		fmt.Println("❌ Failed to start listening")
		return
	}

	fmt.Println("👂 Listening for messages for 5 seconds...")

	// Listen for messages with a timeout
	timeout := time.After(5 * time.Second)
	messageCount := 0

	for {
		select {
		case msg := <-messages:
			if msg != nil {
				messageCount++
				if msgMap, ok := msg.(map[string]any); ok {
					fmt.Printf("📨 Message %d: Type=%v, ID=%v\n",
						messageCount, msgMap["type"], msgMap["id"])
				}
			}
		case <-timeout:
			fmt.Printf("⏰ Timeout reached. Received %d messages.\n", messageCount)
			fmt.Println("🛑 Initiating graceful shutdown...")
			return
		}
	}
}
