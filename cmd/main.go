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
	fmt.Println("✅ Connected successfully!") // Test 1: Register a simple sim variable definition
	fmt.Println("🧪 Testing RegisterSimVarDefinition...")
	if err := sdk.RegisterSimVarDefinition(1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
		fmt.Printf("❌ RegisterSimVarDefinition failed: %v\n", err)
	} else {
		fmt.Println("✅ RegisterSimVarDefinition succeeded - variable registered!")

		// Test 2: Request data for the registered variable
		fmt.Println("🧪 Testing RequestSimVarData...")
		if err := sdk.RequestSimVarData(1, 100); err != nil {
			fmt.Printf("❌ RequestSimVarData failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData succeeded - data requested!")
		}
	}
	// Test 3: Add a second sim variable (different data type)
	fmt.Println("🧪 Testing RegisterSimVarDefinition for CAMERA STATE...")
	if err := sdk.RegisterSimVarDefinition(2, "CAMERA STATE", "Enum", types.SIMCONNECT_DATATYPE_INT32); err != nil {
		fmt.Printf("❌ RegisterSimVarDefinition (camera) failed: %v\n", err)
	} else {
		fmt.Println("✅ RegisterSimVarDefinition (camera) succeeded!")
		// Request camera data
		fmt.Println("🧪 Testing RequestSimVarData for camera...")
		if err := sdk.RequestSimVarData(2, 200); err != nil {
			fmt.Printf("❌ RequestSimVarData (camera) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData (camera) succeeded!")
		}
		// Test 3B: SetSimVar with CAMERA STATE (Baby Step 3B) - Cycle through different camera states
		fmt.Println("🧪 Testing SetSimVar for CAMERA STATE - cycling through values...")
		cameraStates := []int32{2, 3, 4, 5, 6, 2}

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
	fmt.Println("🧪 Testing RegisterSimVarDefinition for ATC TYPE...")
	if err := sdk.RegisterSimVarDefinition(3, "ATC TYPE", "", types.SIMCONNECT_DATATYPE_STRINGV); err != nil {
		fmt.Printf("❌ RegisterSimVarDefinition (ATC TYPE) failed: %v\n", err)
	} else {
		fmt.Println("✅ RegisterSimVarDefinition (ATC TYPE) succeeded!")

		// Request ATC TYPE data
		fmt.Println("🧪 Testing RequestSimVarData for ATC TYPE...")
		if err := sdk.RequestSimVarData(3, 300); err != nil {
			fmt.Printf("❌ RequestSimVarData (ATC TYPE) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData (ATC TYPE) succeeded!")
		}
	}
	// Test 5: Add TITLE variable with proper empty units
	fmt.Println("🧪 Testing RegisterSimVarDefinition for TITLE...")
	if err := sdk.RegisterSimVarDefinition(4, "TITLE", "", types.SIMCONNECT_DATATYPE_STRINGV); err != nil {
		fmt.Printf("❌ RegisterSimVarDefinition (TITLE) failed: %v\n", err)
	} else {
		fmt.Println("✅ RegisterSimVarDefinition (TITLE) succeeded!")

		// Request TITLE data
		fmt.Println("🧪 Testing RequestSimVarData for TITLE...")
		if err := sdk.RequestSimVarData(4, 400); err != nil {
			fmt.Printf("❌ RequestSimVarData (TITLE) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData (TITLE) succeeded!")
		}
	} // Test exception handling by trying to use an invalid SimVar
	fmt.Println("🧪 Testing exception handling with invalid SimVar...")
	if err := sdk.RegisterSimVarDefinition(999, "INVALID_VAR_NAME", "invalid_unit", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
		fmt.Printf("❌ RegisterSimVarDefinition (invalid) failed as expected: %v\n", err)
	} else {
		fmt.Println("✅ RegisterSimVarDefinition (invalid) succeeded - will request to potentially trigger exception...")
		// Try to request data for the invalid variable - this should cause an exception
		if err := sdk.RequestSimVarData(999, 999); err != nil {
			fmt.Printf("❌ RequestSimVarData (invalid) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarData (invalid) submitted - watch for exceptions...")
		}
	}

	// Test 6: Add electrical system data monitoring (matching the GitHub example)
	fmt.Println("🧪 Testing electrical system data monitoring...")
	// Add electrical system variables (matching the GitHub example structure)
	electricalVars := []struct {
		defineID uint32
		name     string
		units    string
		dataType types.SimConnectDataType
	}{
		{7, "EXTERNAL POWER AVAILABLE", "Bool", types.SIMCONNECT_DATATYPE_INT32},
		{8, "EXTERNAL POWER ON", "Bool", types.SIMCONNECT_DATATYPE_INT32},
		{9, "ELECTRICAL MASTER BATTERY", "Bool", types.SIMCONNECT_DATATYPE_INT32},
		{10, "GENERAL ENG MASTER ALTERNATOR", "Bool", types.SIMCONNECT_DATATYPE_INT32},
		{11, "LIGHT BEACON", "Bool", types.SIMCONNECT_DATATYPE_INT32},
		{12, "LIGHT NAV", "Bool", types.SIMCONNECT_DATATYPE_INT32},
		{13, "ELECTRICAL MAIN BUS VOLTAGE", "Volts", types.SIMCONNECT_DATATYPE_FLOAT32},
		{14, "ELECTRICAL BATTERY LOAD", "Amperes", types.SIMCONNECT_DATATYPE_FLOAT32},
	}

	// Register all electrical variables
	for _, v := range electricalVars {
		fmt.Printf("🧪 Adding %s to electrical monitoring...\n", v.name)
		if err := sdk.RegisterSimVarDefinition(v.defineID, v.name, v.units, v.dataType); err != nil {
			fmt.Printf("❌ RegisterSimVarDefinition (%s) failed: %v\n", v.name, err)
		} else {
			fmt.Printf("✅ RegisterSimVarDefinition (%s) succeeded!\n", v.name)

			// Start periodic monitoring for each electrical variable (every second)
			requestID := 700 + v.defineID
			fmt.Printf("🧪 Starting periodic monitoring for %s (every second)...\n", v.name)
			if err := sdk.RequestSimVarDataPeriodic(v.defineID, requestID, types.SIMCONNECT_PERIOD_SECOND); err != nil {
				fmt.Printf("❌ RequestSimVarDataPeriodic (%s) failed: %v\n", v.name, err)
			} else {
				fmt.Printf("✅ RequestSimVarDataPeriodic (%s) succeeded! Monitoring electrical data...\n", v.name)
			}
		}
	}

	// Test 7: Periodic data requests
	fmt.Println("🧪 Testing periodic data requests...") // Add a variable specifically for periodic testing
	fmt.Println("🧪 Adding AIRSPEED INDICATED for periodic testing...")
	if err := sdk.RegisterSimVarDefinition(5, "AIRSPEED INDICATED", "knots", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
		fmt.Printf("❌ RegisterSimVarDefinition (AIRSPEED) failed: %v\n", err)
	} else {
		fmt.Println("✅ RegisterSimVarDefinition (AIRSPEED) succeeded!")

		// Start periodic request every visual frame
		fmt.Println("🧪 Starting periodic request for AIRSPEED (every visual frame)...")
		if err := sdk.RequestSimVarDataPeriodic(5, 500, types.SIMCONNECT_PERIOD_VISUAL_FRAME); err != nil {
			fmt.Printf("❌ RequestSimVarDataPeriodic (AIRSPEED) failed: %v\n", err)
		} else {
			fmt.Println("✅ RequestSimVarDataPeriodic (AIRSPEED) succeeded! Data will flow continuously...")
		}
	} // Add another variable for periodic testing with different frequency
	fmt.Println("🧪 Adding PLANE LATITUDE for periodic testing...")
	if err := sdk.RegisterSimVarDefinition(6, "PLANE LATITUDE", "radians", types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
		fmt.Printf("❌ RegisterSimVarDefinition (LATITUDE) failed: %v\n", err)
	} else {
		fmt.Println("✅ RegisterSimVarDefinition (LATITUDE) succeeded!")

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
	// ===== ELECTRICAL SYSTEM EVENT TESTS =====
	fmt.Println("\n🔌 === ELECTRICAL SYSTEM EVENT TESTING ===")

	// Electrical event constants (from the GitHub example)
	const (
		EVENT_ID_TOGGLE_EXTERNAL_POWER = 10011511
		EVENT_ID_TOGGLE_MASTER_BATTERY = 10025115
		EVENT_ID_TOGGLE_MASTER_ALT     = 10031515
		EVENT_ID_TOGGLE_BEACON_LIGHTS  = 10041515
		EVENT_ID_TOGGLE_NAV_LIGHTS     = 10051514
	)

	// Test 1: Map electrical events to sim events
	fmt.Println("🧪 Testing MapClientEventToSimEvent for TOGGLE_EXTERNAL_POWER...")
	if err := sdk.MapClientEventToSimEvent(EVENT_ID_TOGGLE_EXTERNAL_POWER, "TOGGLE_EXTERNAL_POWER"); err != nil {
		fmt.Printf("❌ MapClientEventToSimEvent (TOGGLE_EXTERNAL_POWER) failed: %v\n", err)
	} else {
		fmt.Println("✅ MapClientEventToSimEvent (TOGGLE_EXTERNAL_POWER) succeeded!")
	}

	fmt.Println("🧪 Testing MapClientEventToSimEvent for TOGGLE_MASTER_BATTERY...")
	if err := sdk.MapClientEventToSimEvent(EVENT_ID_TOGGLE_MASTER_BATTERY, "TOGGLE_MASTER_BATTERY"); err != nil {
		fmt.Printf("❌ MapClientEventToSimEvent (TOGGLE_MASTER_BATTERY) failed: %v\n", err)
	} else {
		fmt.Println("✅ MapClientEventToSimEvent (TOGGLE_MASTER_BATTERY) succeeded!")
	}

	fmt.Println("🧪 Testing MapClientEventToSimEvent for TOGGLE_BEACON_LIGHTS...")
	if err := sdk.MapClientEventToSimEvent(EVENT_ID_TOGGLE_BEACON_LIGHTS, "TOGGLE_BEACON_LIGHTS"); err != nil {
		fmt.Printf("❌ MapClientEventToSimEvent (TOGGLE_BEACON_LIGHTS) failed: %v\n", err)
	} else {
		fmt.Println("✅ MapClientEventToSimEvent (TOGGLE_BEACON_LIGHTS) succeeded!")
	}

	// Test 2: Create electrical notification group and set priority
	fmt.Println("🧪 Testing SetNotificationGroupPriority for electrical events...")
	if err := sdk.SetNotificationGroupPriority(2000, types.SIMCONNECT_GROUP_PRIORITY_HIGHEST); err != nil {
		fmt.Printf("❌ SetNotificationGroupPriority failed: %v\n", err)
	} else {
		fmt.Println("✅ SetNotificationGroupPriority succeeded!")
	}

	// Test 3: Add electrical events to notification group
	fmt.Println("🧪 Testing AddClientEventToNotificationGroup for electrical events...")
	if err := sdk.AddClientEventToNotificationGroup(2000, EVENT_ID_TOGGLE_EXTERNAL_POWER, false); err != nil {
		fmt.Printf("❌ AddClientEventToNotificationGroup (TOGGLE_EXTERNAL_POWER) failed: %v\n", err)
	} else {
		fmt.Println("✅ AddClientEventToNotificationGroup (TOGGLE_EXTERNAL_POWER) succeeded!")
	}

	if err := sdk.AddClientEventToNotificationGroup(2000, EVENT_ID_TOGGLE_MASTER_BATTERY, false); err != nil {
		fmt.Printf("❌ AddClientEventToNotificationGroup (TOGGLE_MASTER_BATTERY) failed: %v\n", err)
	} else {
		fmt.Println("✅ AddClientEventToNotificationGroup (TOGGLE_MASTER_BATTERY) succeeded!")
	}

	if err := sdk.AddClientEventToNotificationGroup(2000, EVENT_ID_TOGGLE_BEACON_LIGHTS, false); err != nil {
		fmt.Printf("❌ AddClientEventToNotificationGroup (TOGGLE_BEACON_LIGHTS) failed: %v\n", err)
	} else {
		fmt.Println("✅ AddClientEventToNotificationGroup (TOGGLE_BEACON_LIGHTS) succeeded!")
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
				} // Stop the latitude periodic request
				if err := sdk.StopPeriodicRequest(600); err != nil {
					fmt.Printf("❌ Failed to stop LATITUDE periodic request: %v\n", err)
				} else {
					fmt.Println("✅ LATITUDE periodic request stopped")
				}

				periodicStopped = true
				fmt.Println("📊 Continuing to listen for remaining 5 seconds (should see fewer messages now)...") // Test electrical event transmission while listening
				fmt.Println("\n⚡ Testing electrical event transmission...")
				fmt.Println("🔌 Triggering TOGGLE_EXTERNAL_POWER...")
				if err := sdk.TransmitClientEvent(types.SIMCONNECT_OBJECT_ID_USER, EVENT_ID_TOGGLE_EXTERNAL_POWER, 0, 2000, types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY); err != nil {
					fmt.Printf("❌ TransmitClientEvent (TOGGLE_EXTERNAL_POWER) failed: %v\n", err)
				} else {
					fmt.Println("✅ TransmitClientEvent (TOGGLE_EXTERNAL_POWER) succeeded! External power should toggle!")
				}

				// Wait a moment, then test master battery toggle
				time.Sleep(1 * time.Second)
				fmt.Println("🔋 Triggering TOGGLE_MASTER_BATTERY...")
				if err := sdk.TransmitClientEvent(types.SIMCONNECT_OBJECT_ID_USER, EVENT_ID_TOGGLE_MASTER_BATTERY, 0, 2000, types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY); err != nil {
					fmt.Printf("❌ TransmitClientEvent (TOGGLE_MASTER_BATTERY) failed: %v\n", err)
				} else {
					fmt.Println("✅ TransmitClientEvent (TOGGLE_MASTER_BATTERY) succeeded! Master battery should toggle!")
				}

				// Wait a moment, then test beacon lights
				time.Sleep(1 * time.Second)
				fmt.Println("🚨 Triggering TOGGLE_BEACON_LIGHTS...")
				if err := sdk.TransmitClientEvent(types.SIMCONNECT_OBJECT_ID_USER, EVENT_ID_TOGGLE_BEACON_LIGHTS, 0, 2000, types.SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY); err != nil {
					fmt.Printf("❌ TransmitClientEvent (TOGGLE_BEACON_LIGHTS) failed: %v\n", err)
				} else {
					fmt.Println("✅ TransmitClientEvent (TOGGLE_BEACON_LIGHTS) succeeded! Beacon lights should toggle!")
				}
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
