package main

import (
	"fmt"
	"time"

	"github.com/mycrew-online/sdk/pkg/client"
	"github.com/mycrew-online/sdk/pkg/types"
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

	fmt.Println("✅ Connected successfully!") // Test 1: Add a simple sim variable
	fmt.Println("🧪 Testing AddSimVar...")
	if err := sdk.AddSimVar(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
		fmt.Printf("❌ AddSimVar failed: %v\n", err)
	} else {
		fmt.Println("✅ AddSimVar succeeded - variable registered!")

		// Test 2: Request data for the registered variable
		fmt.Println("🧪 Testing RequestSimVarData...")
		if err := sdk.RequestSimVarData(1, 100); err != nil {
			fmt.Printf("❌ RequestSimVarData failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData succeeded - data requested!")
		}
	}

	// Test 3: Add a second sim variable (different data type)
	fmt.Println("🧪 Testing AddSimVar for CAMERA STATE...")
	if err := sdk.AddSimVar(2, "CAMERA STATE", "Enum", types.SIMCONNECT_DATATYPE_INT32); err != nil {
		fmt.Printf("❌ AddSimVar (camera) failed: %v\n", err)
	} else {
		fmt.Println("✅ AddSimVar (camera) succeeded!")
		// Request camera data
		fmt.Println("🧪 Testing RequestSimVarData for camera...")
		if err := sdk.RequestSimVarData(2, 200); err != nil {
			fmt.Printf("❌ RequestSimVarData (camera) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData (camera) succeeded!")
		}
		// Test 3B: SetSimVar with CAMERA STATE (Baby Step 3B) - Cycle through different camera states
		fmt.Println("🧪 Testing SetSimVar for CAMERA STATE - cycling through values...")
		cameraStates := []int32{2, 3, 4, 5, 6}

		for i, cameraState := range cameraStates {
			fmt.Printf("🎥 Setting camera state to %d (test %d/5)...\n", cameraState, i+1)

			if err := sdk.SetSimVar(2, cameraState); err != nil {
				fmt.Printf("❌ SetSimVar (camera state %d) failed: %v\n", cameraState, err)
			} else {
				fmt.Printf("✅ SetSimVar (camera) succeeded! Set value to %d\n", cameraState)

				// Small pause to allow the simulator to process the change
				time.Sleep(500 * time.Millisecond)

				// Request the data back to verify the set operation
				fmt.Printf("🧪 Verifying camera state %d by requesting data...\n", cameraState)
				if err := sdk.RequestSimVarData(2, uint32(201+i)); err != nil {
					fmt.Printf("❌ RequestSimVarData (verification %d) failed: %v\n", cameraState, err)
				} else {
					fmt.Printf("✅ RequestSimVarData (verification %d) succeeded!\n", cameraState)
				}

				// Pause between different camera state changes
				if i < len(cameraStates)-1 {
					fmt.Println("⏱️  Pausing 1 second before next camera state change...")
					time.Sleep(1 * time.Second)
				}
			}
		}
	} // Test 4: Add a string variable (ATC TYPE) with STRINGV and empty units
	fmt.Println("🧪 Testing AddSimVar for ATC TYPE...")
	if err := sdk.AddSimVar(3, "ATC TYPE", "", types.SIMCONNECT_DATATYPE_STRINGV); err != nil {
		fmt.Printf("❌ AddSimVar (ATC TYPE) failed: %v\n", err)
	} else {
		fmt.Println("✅ AddSimVar (ATC TYPE) succeeded!")

		// Request ATC TYPE data
		fmt.Println("🧪 Testing RequestSimVarData for ATC TYPE...")
		if err := sdk.RequestSimVarData(3, 300); err != nil {
			fmt.Printf("❌ RequestSimVarData (ATC TYPE) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData (ATC TYPE) succeeded!")
		}
	}

	// Test 5: Add TITLE variable with proper empty units
	fmt.Println("🧪 Testing AddSimVar for TITLE...")
	if err := sdk.AddSimVar(4, "TITLE", "", types.SIMCONNECT_DATATYPE_STRINGV); err != nil {
		fmt.Printf("❌ AddSimVar (TITLE) failed: %v\n", err)
	} else {
		fmt.Println("✅ AddSimVar (TITLE) succeeded!")

		// Request TITLE data
		fmt.Println("🧪 Testing RequestSimVarData for TITLE...")
		if err := sdk.RequestSimVarData(4, 400); err != nil {
			fmt.Printf("❌ RequestSimVarData (TITLE) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData (TITLE) succeeded!")
		}
	}

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
