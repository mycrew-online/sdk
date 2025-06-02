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
	// Test exception handling by trying to use an invalid SimVar
	fmt.Println("🧪 Testing exception handling with invalid SimVar...")
	if err := sdk.AddSimVar(999, "INVALID_VAR_NAME", "invalid_unit", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
		fmt.Printf("❌ AddSimVar (invalid) failed as expected: %v\n", err)
	} else {
		fmt.Println("✅ AddSimVar (invalid) succeeded - will request to potentially trigger exception...")
		// Try to request data for the invalid variable - this should cause an exception
		if err := sdk.RequestSimVarData(999, 999); err != nil {
			fmt.Printf("❌ RequestSimVarData (invalid) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData (invalid) submitted - watch for exceptions...")
		}
	}

	// Test 6: Periodic data requests
	fmt.Println("🧪 Testing periodic data requests...")
	// Add a variable specifically for periodic testing
	fmt.Println("🧪 Adding AIRSPEED INDICATED for periodic testing...")
	if err := sdk.AddSimVar(5, "AIRSPEED INDICATED", "knots", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
		fmt.Printf("❌ AddSimVar (AIRSPEED) failed: %v\n", err)
	} else {
		fmt.Println("✅ AddSimVar (AIRSPEED) succeeded!")

		// Start periodic request every visual frame
		fmt.Println("🧪 Starting periodic request for AIRSPEED (every visual frame)...")
		if err := sdk.RequestSimVarDataPeriodic(5, 500, types.SIMCONNECT_PERIOD_VISUAL_FRAME); err != nil {
			fmt.Printf("❌ RequestSimVarDataPeriodic (AIRSPEED) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarDataPeriodic (AIRSPEED) succeeded! Data will flow continuously...")
		}
	}
	// Add another variable for periodic testing with different frequency
	fmt.Println("🧪 Adding PLANE LATITUDE for periodic testing...")
	if err := sdk.AddSimVar(6, "PLANE LATITUDE", "radians", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
		fmt.Printf("❌ AddSimVar (LATITUDE) failed: %v\n", err)
	} else {
		fmt.Println("✅ AddSimVar (LATITUDE) succeeded!")

		// Start periodic request every second
		fmt.Println("🧪 Starting periodic request for LATITUDE (every second)...")
		if err := sdk.RequestSimVarDataPeriodic(6, 600, types.SIMCONNECT_PERIOD_SECOND); err != nil {
			fmt.Printf("❌ RequestSimVarDataPeriodic (LATITUDE) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarDataPeriodic (LATITUDE) succeeded! Data will arrive every second...")
		}
	}

	// Baby Step 1: Test system event subscriptions
	fmt.Println("🧪 Testing system event subscriptions...")
	// Test 1: Subscribe to "Pause" system event (gets both pause and unpause notifications)
	fmt.Println("🧪 Subscribing to 'Pause' system event...")
	if err := sdk.SubscribeToSystemEvent(1010, "Pause"); err != nil {
		fmt.Printf("❌ SubscribeToSystemEvent (Pause) failed: %v\n", err)
	} else {
		fmt.Println("✅ SubscribeToSystemEvent (Pause) succeeded!")
	}

	// Test 2: Subscribe to "Sim" system event
	fmt.Println("🧪 Subscribing to 'Sim' system event...")
	if err := sdk.SubscribeToSystemEvent(1020, "Sim"); err != nil {
		fmt.Printf("❌ SubscribeToSystemEvent (Sim) failed: %v\n", err)
	} else {
		fmt.Println("✅ SubscribeToSystemEvent (Sim) succeeded!")
	}

	// Test 3: Subscribe to "AircraftLoaded" system event
	fmt.Println("🧪 Subscribing to 'AircraftLoaded' system event...")
	if err := sdk.SubscribeToSystemEvent(1030, "AircraftLoaded"); err != nil {
		fmt.Printf("❌ SubscribeToSystemEvent (AircraftLoaded) failed: %v\n", err)
	} else {
		fmt.Println("✅ SubscribeToSystemEvent (AircraftLoaded) succeeded!")
	}

	// Start listening for messages
	messages := sdk.Listen()
	if messages == nil {
		fmt.Println("❌ Failed to start listening")
		return
	}
	fmt.Println("👂 Listening for messages for 8 seconds...")
	fmt.Println("   📊 Expect to see periodic data for AIRSPEED (every frame) and LATITUDE (every second)")
	fmt.Println("   📡 Also watching for system events: Pause, Sim, AircraftLoaded")
	fmt.Println("   🛑 Will stop periodic requests after 3 seconds...")

	// Listen for messages with a timeout and periodic stop demonstration
	timeout := time.After(8 * time.Second)
	stopPeriodicTimer := time.After(3 * time.Second)
	messageCount := 0
	periodicStopped := false
	for {
		select {
		case <-stopPeriodicTimer:
			if !periodicStopped {
				fmt.Println("🛑 Stopping periodic data requests...")

				// Stop the airspeed periodic request
				if err := sdk.StopPeriodicRequest(500); err != nil {
					fmt.Printf("❌ Failed to stop AIRSPEED periodic request: %v\n", err)
				} else {
					fmt.Println("✅ AIRSPEED periodic request stopped")
				}

				// Stop the latitude periodic request
				if err := sdk.StopPeriodicRequest(600); err != nil {
					fmt.Printf("❌ Failed to stop LATITUDE periodic request: %v\n", err)
				} else {
					fmt.Println("✅ LATITUDE periodic request stopped")
				}

				periodicStopped = true
				fmt.Println("📊 Continuing to listen for remaining 5 seconds (should see fewer messages now)...")
			}
		case msg := <-messages:
			if msg != nil {
				messageCount++

				// Check for exceptions first using the utility function
				if exception, isExceptionMsg := types.IsException(msg); isExceptionMsg {
					fmt.Printf("📨 Message %d: ❌ EXCEPTION - %s\n", messageCount, exception.ExceptionName)
					fmt.Printf("   🔍 Details: %s\n", exception.Description)
					fmt.Printf("   🎯 SendID: %d, Index: %d, Severity: %s\n",
						exception.SendID, exception.Index, exception.Severity)

					// Check severity and take appropriate action
					if types.IsCriticalException(exception) {
						fmt.Printf("   🚨 CRITICAL EXCEPTION! This may require immediate attention.\n")
					} else if types.IsErrorException(exception) {
						fmt.Printf("   ⚠️  ERROR EXCEPTION! Check your request parameters.\n")
					} else if types.IsWarningException(exception) {
						fmt.Printf("   ℹ️  WARNING EXCEPTION: Non-critical issue.\n")
					}
					continue
				}

				if msgMap, ok := msg.(map[string]any); ok {
					fmt.Printf("📨 Message %d: Type=%v, ID=%v\n",
						messageCount, msgMap["type"], msgMap["id"]) // Handle different message types
					switch msgMap["type"] {
					case "SIMOBJECT_DATA":
						// Check if we have pre-parsed data available
						if parsedData, exists := msgMap["parsed_data"]; exists {
							// Try to cast to SimVarData (we need to import the client package for this)
							fmt.Printf("   📈 PARSED DATA AVAILABLE: %+v (type: %T)\n", parsedData, parsedData)

							// For now, let's access it as a map or try to extract fields
							if simVarData, ok := parsedData.(*client.SimVarData); ok {
								fmt.Printf("   ✨ VALUE: RequestID=%d, DefineID=%d, Value=%v (type: %T)\n",
									simVarData.RequestID, simVarData.DefineID, simVarData.Value, simVarData.Value)
							} else {
								fmt.Printf("   ⚠️  Could not cast parsed_data to SimVarData, got type: %T\n", parsedData)
							}
						} else {
							fmt.Printf("   ⚠️  No parsed_data field found in SIMOBJECT_DATA message\n")
						}

					case "OPEN":
						fmt.Printf("   🔓 SimConnect connection opened successfully\n")

					case "EVENT":
						// Check if we have pre-parsed event data available
						if eventData, exists := msgMap["event"]; exists {
							fmt.Printf("   📡 EVENT DATA AVAILABLE: %+v (type: %T)\n", eventData, eventData)

							// Try to cast to EventData
							if parsedEvent, ok := eventData.(*types.EventData); ok {
								fmt.Printf("   🎯 EVENT: ID=%d, Group=%d, Data=%d, Type=%s, Name=%s\n",
									parsedEvent.EventID, parsedEvent.GroupID, parsedEvent.EventData,
									parsedEvent.EventType, parsedEvent.EventName)

								// Special handling for known events
								switch parsedEvent.EventID {
								case 1010: // Pause event (both pause and unpause notifications)
									if parsedEvent.EventData == 1 {
										fmt.Printf("   ⏸️  Simulator PAUSED\n")
									} else {
										fmt.Printf("   ▶️  Simulator RESUMED\n")
									}
								case 1020: // Sim event
									if parsedEvent.EventData == 1 {
										fmt.Printf("   🚁 Simulator RUNNING\n")
									} else {
										fmt.Printf("   🛑 Simulator STOPPED\n")
									}
								case 1030: // AircraftLoaded event
									fmt.Printf("   ✈️  Aircraft LOADED (Data: %d)\n", parsedEvent.EventData)
								default:
									fmt.Printf("   🎪 Unknown Event ID: %d\n", parsedEvent.EventID)
								}
							} else {
								fmt.Printf("   ⚠️  Could not cast event data to EventData, got type: %T\n", eventData)
							}
						} else {
							fmt.Printf("   ⚠️  No event field found in EVENT message\n")
						}

					case "QUIT":
						fmt.Printf("   👋 SimConnect connection closed\n")

					default:
						// Show what we have access to for other message types
						fmt.Printf("   📋 Available message fields: %v\n", getMapKeys(msgMap))
					}
				}
			}
		case <-timeout:
			fmt.Printf("⏰ Timeout reached. Received %d messages total.\n", messageCount)
			if !periodicStopped {
				fmt.Println("🛑 Cleaning up: Stopping any remaining periodic requests...")
				sdk.StopPeriodicRequest(500)
				sdk.StopPeriodicRequest(600)
			}
			fmt.Println("🛑 Initiating graceful shutdown...")
			return
		}
	}
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
