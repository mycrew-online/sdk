package main

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	S_OK         = uint32(0x00000000)
	E_FAIL       = uint32(0x80004005)
	E_INVALIDARG = uint32(0x80070057)

	// SimConnect data type constants
	SIMCONNECT_DATATYPE_FLOAT64 = 0
	SIMCONNECT_DATATYPE_FLOAT32 = 1
	SIMCONNECT_DATATYPE_INT32   = 2
	SIMCONNECT_DATATYPE_INT64   = 3

	// SimConnect period constants
	SIMCONNECT_PERIOD_NEVER        = 0
	SIMCONNECT_PERIOD_ONCE         = 1
	SIMCONNECT_PERIOD_VISUAL_FRAME = 2
	SIMCONNECT_PERIOD_SIM_FRAME    = 3
	SIMCONNECT_PERIOD_SECOND       = 4

	// SimConnect object ID constants
	SIMCONNECT_OBJECT_ID_USER                 = 0 // SimConnect receive ID constants
	SIMCONNECT_RECV_ID_NULL                   = 0
	SIMCONNECT_RECV_ID_EXCEPTION              = 1
	SIMCONNECT_RECV_ID_OPEN                   = 2
	SIMCONNECT_RECV_ID_QUIT                   = 3
	SIMCONNECT_RECV_ID_EVENT                  = 4
	SIMCONNECT_RECV_ID_SIMOBJECT_DATA         = 8
	SIMCONNECT_RECV_ID_SYSTEM_STATE           = 15
	SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENTS = 34
	SIMCONNECT_RECV_ID_SUBSCRIBE_INPUT_EVENT  = 36

	// Client event constants for electrical events
	EVENT_ID_TOGGLE_EXTERNAL_POWER = 10011511
	EVENT_ID_TOGGLE_MASTER_BATTERY = 10025115
	EVENT_ID_TOGGLE_MASTER_ALT     = 10031515
	EVENT_ID_TOGGLE_BEACON_LIGHTS  = 10041515
	EVENT_ID_TOGGLE_NAV_LIGHTS     = 10051514
)

// Data definition and request IDs
const (
	DATA_DEFINITION_1 = 1
	DATA_REQUEST_1    = 1

	// Electrical system data definitions
	DATA_DEFINITION_ELECTRICAL = 2000
	DATA_REQUEST_ELECTRICAL    = 2000
)

// Base SimConnect receive structure
type SIMCONNECT_RECV struct {
	DwSize    uint32 // Total size of the returned structure in bytes
	DwVersion uint32 // Version number of the SimConnect server
	DwID      uint32 // ID of the returned structure (SIMCONNECT_RECV_ID)
}

// SimConnect SimObject data receive structure
type SIMCONNECT_RECV_SIMOBJECT_DATA struct {
	SIMCONNECT_RECV        // Inherits from base structure
	DwRequestID     uint32 // ID of the client defined request
	DwObjectID      uint32 // ID of the client defined object
	DwDefineID      uint32 // ID of the client defined data definition
	DwFlags         uint32 // Flags that were set for this data request
	DwEntryNumber   uint32 // Index number of this object (1-based)
	DwOutOf         uint32 // Total number of objects being returned
	DwDefineCount   uint32 // Number of 8-byte elements in the data array
	DwData          uint32 // Start of data array (actual data follows)
}

// SimConnect System State receive structure
// MAX_PATH is typically 260 characters in Windows
const MAX_PATH = 260

type SIMCONNECT_RECV_SYSTEM_STATE struct {
	SIMCONNECT_RECV                // Inherits from base structure
	DwRequestID     uint32         // ID of the client defined request
	DwInteger       uint32         // Integer or boolean value
	FFloat          float32        // Float value
	SzString        [MAX_PATH]byte // Null-terminated string array
}

// SimConnect Event receive structure
type SIMCONNECT_RECV_EVENT struct {
	SIMCONNECT_RECV        // Inherits from base structure
	UGroupID        uint32 // ID of the group that the event belongs to
	UEventID        uint32 // ID of the event
	DwData          uint32 // Data associated with the event
}

// SimConnect List Template (base for list-based responses)
type SIMCONNECT_RECV_LIST_TEMPLATE struct {
	SIMCONNECT_RECV        // Inherits from base structure
	DwRequestID     uint32 // ID of the client defined request
	DwArraySize     uint32 // Number of elements in the list within this packet
	DwEntryNumber   uint32 // Index number of this list packet (0-based)
	DwOutOf         uint32 // Total number of packets used to transmit the list
}

// SimConnect Input Event Descriptor
type SIMCONNECT_INPUT_EVENT_DESCRIPTOR struct {
	Name      [64]byte   // Name of the Input Event (SIMCONNECT_STRING(Name, 64))
	Hash      uint32     // Hash ID for the event
	Type      uint32     // Expected datatype (from SIMCONNECT_DATATYPE enum)
	NodeNames [1024]byte // List of node names separated by ';' (SIMCONNECT_STRING(NodeNames, 1024))
}

// SimConnect Enumerate Input Events receive structure
type SIMCONNECT_RECV_ENUMERATE_INPUT_EVENTS struct {
	SIMCONNECT_RECV_LIST_TEMPLATE // Inherits from list template
	// rgData follows immediately after this structure in memory
	// We'll access it using pointer arithmetic in the parsing function
}

// SimConnect Subscribe Input Event receive structure
type SIMCONNECT_RECV_SUBSCRIBE_INPUT_EVENT struct {
	SIMCONNECT_RECV        // Inherits from base structure
	DwDefineID      uint32 // ID of the client defined input event definition
}

// Camera state data structure that matches our data definition
type CameraState struct {
	State int32
}

// Electrical system data structure
type ElectricalSystemState struct {
	ExternalPowerAvailable int32   // EXTERNAL POWER AVAILABLE (bool)
	ExternalPowerOn        int32   // EXTERNAL POWER ON (bool)
	MasterBattery          int32   // ELECTRICAL MASTER BATTERY (bool)
	MasterAlternator       int32   // GENERAL ENG MASTER ALTERNATOR (bool)
	BeaconLightSwitch      int32   // LIGHT BEACON (bool)
	NavLightSwitch         int32   // LIGHT NAV (bool)
	BatteryVoltage         float32 // ELECTRICAL MAIN BUS VOLTAGE (volts)
	BatteryLoad            float32 // ELECTRICAL BATTERY LOAD (amperes)
}

// As we are mapping processes to the SimConnect processes, we need to define a entity that will hold the process information.
var dll = syscall.NewLazyDLL("C:/MSFS 2024 SDK/SimConnect SDK/lib/SimConnect.dll")

var handle syscall.Handle // Handle to the SimConnect server

// SimConnectProcesses is a map that holds the SimConnect.dll procedures.
var (
	SimConnect_Open                              *syscall.LazyProc // SimConnect_Open procedure
	SimConnect_Close                             *syscall.LazyProc // SimConnect_Close procedure
	SimConnect_GetNextDispatch                   *syscall.LazyProc // SimConnect_GetNextDispatch procedure
	SimConnect_AddToDataDefinition               *syscall.LazyProc // SimConnect_AddToDataDefinition procedure
	SimConnect_RequestDataOnSimObject            *syscall.LazyProc // SimConnect_RequestDataOnSimObject procedure
	SimConnect_ClearDataDefinition               *syscall.LazyProc // SimConnect_ClearDataDefinition procedure
	SimConnect_RequestSystemState                *syscall.LazyProc // SimConnect_RequestSystemState procedure
	SimConnect_SetDataOnSimObject                *syscall.LazyProc // SimConnect_SetDataOnSimObject procedure
	SimConnect_SubscribeToSystemEvent            *syscall.LazyProc // SimConnect_SubscribeToSystemEvent procedure
	SimConnect_SetSystemEventState               *syscall.LazyProc // SimConnect_SetSystemEventState procedure
	SimConnect_EnumerateInputEvents              *syscall.LazyProc // SimConnect_EnumerateInputEvents procedure
	SimConnect_SubscribeInputEvent               *syscall.LazyProc // SimConnect_SubscribeInputEvents procedure
	SimConnect_MapClientEventToSimEvent          *syscall.LazyProc // SimConnect_MapClientEventToSimEvent procedure
	SimConnect_TransmitClientEvent               *syscall.LazyProc // SimConnect_TransmitClientEvent procedure
	SimConnect_AddClientEventToNotificationGroup *syscall.LazyProc // SimConnect_AddClientEventToNotificationGroup procedure
	SimConnect_SetNotificationGroupPriority      *syscall.LazyProc // SimConnect_SetNotificationGroupPriority procedure
)

func initProcedures() error {
	// SimConnect_Open procedure
	SimConnect_Open = dll.NewProc("SimConnect_Open")
	// SimConnect_Close procedure
	SimConnect_Close = dll.NewProc("SimConnect_Close")
	// SimConnect_GetNextDispatch procedure
	SimConnect_GetNextDispatch = dll.NewProc("SimConnect_GetNextDispatch")
	// SimConnect_AddToDataDefinition procedure
	SimConnect_AddToDataDefinition = dll.NewProc("SimConnect_AddToDataDefinition")
	// SimConnect_RequestDataOnSimObject procedure
	SimConnect_RequestDataOnSimObject = dll.NewProc("SimConnect_RequestDataOnSimObject")
	// SimConnect_ClearDataDefinition procedure
	SimConnect_ClearDataDefinition = dll.NewProc("SimConnect_ClearDataDefinition")
	// SimConnect_RequestSystemState procedure
	SimConnect_RequestSystemState = dll.NewProc("SimConnect_RequestSystemState")
	// SimConnect_SetDataOnSimObject procedure
	SimConnect_SetDataOnSimObject = dll.NewProc("SimConnect_SetDataOnSimObject")
	// SimConnect_SubscribeToSystemEvent procedure
	SimConnect_SubscribeToSystemEvent = dll.NewProc("SimConnect_SubscribeToSystemEvent")
	// SimConnect_SetSystemEventState procedure
	SimConnect_SetSystemEventState = dll.NewProc("SimConnect_SetSystemEventState")
	// SimConnect_EnumerateInputEventParams
	SimConnect_EnumerateInputEvents = dll.NewProc("SimConnect_EnumerateInputEvents")
	// SimConnect_SubscribeInputEvent procedure
	SimConnect_SubscribeInputEvent = dll.NewProc("SimConnect_SubscribeInputEvent")
	// SimConnect_MapClientEventToSimEvent procedure
	SimConnect_MapClientEventToSimEvent = dll.NewProc("SimConnect_MapClientEventToSimEvent")
	// SimConnect_TransmitClientEvent procedure
	SimConnect_TransmitClientEvent = dll.NewProc("SimConnect_TransmitClientEvent")
	// SimConnect_AddClientEventToNotificationGroup procedure
	SimConnect_AddClientEventToNotificationGroup = dll.NewProc("SimConnect_AddClientEventToNotificationGroup")
	// SimConnect_SetNotificationGroupPriority procedure
	SimConnect_SetNotificationGroupPriority = dll.NewProc("SimConnect_SetNotificationGroupPriority")
	// Here we would implement the logic to initialize the processes.
	// This might involve loading process information from the SimConnect server, setting up any necessary event handlers, etc.
	// For now, we will just return nil to indicate success.
	return nil
}

func IsHRESULTSuccess(hresult uint32) bool {
	return hresult == S_OK
}
func IsHRESULTFailure(hresult uint32) bool {
	return hresult != S_OK
}

// Simple function to parse SimConnect response data
func parseSimConnectData(ppData uintptr, pcbData uint32) {
	if ppData == 0 || pcbData == 0 {
		fmt.Println("No data received")
		return
	}

	// Cast the pointer to the base SIMCONNECT_RECV structure
	recv := (*SIMCONNECT_RECV)(unsafe.Pointer(ppData))
	fmt.Printf("Received message - Size: %d, Version: %d, ID: %d\n",
		recv.DwSize, recv.DwVersion, recv.DwID)
	// Check what type of message we received based on the ID
	switch recv.DwID {
	case SIMCONNECT_RECV_ID_SIMOBJECT_DATA:
		fmt.Println("📊 SIMOBJECT_DATA received")
		parseSimObjectData(ppData, pcbData)
	case SIMCONNECT_RECV_ID_OPEN:
		fmt.Println("🔓 OPEN confirmation received")
	case SIMCONNECT_RECV_ID_EXCEPTION:
		fmt.Println("❌ EXCEPTION received")
	case SIMCONNECT_RECV_ID_SYSTEM_STATE:
		fmt.Println("🔧 SYSTEM_STATE received")
		parseSystemStateData(ppData, pcbData)
	case SIMCONNECT_RECV_ID_EVENT:
		fmt.Println("📡 EVENT received")
		parseEventData(ppData, pcbData)
	case SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENTS:
		fmt.Println("🎮 ENUMERATE_INPUT_EVENTS received")
		parseEnumerateInputEventsData(ppData, pcbData)
	case SIMCONNECT_RECV_ID_SUBSCRIBE_INPUT_EVENT:
		fmt.Println("🔗 SUBSCRIBE_INPUT_EVENT received")
		parseSubscribeInputEventData(ppData, pcbData)
	case SIMCONNECT_RECV_ID_QUIT:
		fmt.Println("👋 QUIT received")
	default:
		fmt.Printf("❓ Unknown message type: %d\n", recv.DwID)
	}
}

// Parse system state data messages
func parseSystemStateData(ppData uintptr, pcbData uint32) {
	// Cast to system state structure
	systemState := (*SIMCONNECT_RECV_SYSTEM_STATE)(unsafe.Pointer(ppData))

	fmt.Printf("  Request ID: %d\n", systemState.DwRequestID)
	fmt.Printf("  Integer Value: %d\n", systemState.DwInteger)
	fmt.Printf("  Float Value: %f\n", systemState.FFloat)

	// Find the null terminator in the string array
	stringLength := 0
	for i := 0; i < MAX_PATH && systemState.SzString[i] != 0; i++ {
		stringLength++
	}

	if stringLength > 0 {
		stringValue := string(systemState.SzString[:stringLength])
		fmt.Printf("  📄 String Value: '%s'\n", stringValue)
	} else {
		fmt.Println("  📄 String Value: (empty)")
	}
}

// Parse event data messages (like system events)
func parseEventData(ppData uintptr, pcbData uint32) {
	// Cast to event structure
	eventData := (*SIMCONNECT_RECV_EVENT)(unsafe.Pointer(ppData))

	fmt.Printf("  Group ID: %d\n", eventData.UGroupID)
	fmt.Printf("  Event ID: %d\n", eventData.UEventID)
	fmt.Printf("  Data: %d\n", eventData.DwData)

	// Decode specific events based on Event ID
	switch eventData.UEventID {
	case 1010: // This matches the UserEventID we used for "Paused" subscription
		if eventData.DwData == 1 {
			fmt.Println("  🎮 Simulation is PAUSED")
		} else {
			fmt.Println("  ▶️ Simulation is RUNNING")
		}
	default:
		fmt.Printf("  📡 Unknown event ID: %d with data: %d\n", eventData.UEventID, eventData.DwData)
	}
}

// Parse SimObject data messages (like CAMERA STATE)
func parseSimObjectData(ppData uintptr, pcbData uint32) {
	// Cast to SimObject data structure
	simObjData := (*SIMCONNECT_RECV_SIMOBJECT_DATA)(unsafe.Pointer(ppData))
	fmt.Printf("  Request ID: %d, Object ID: %d, Define ID: %d\n",
		simObjData.DwRequestID, simObjData.DwObjectID, simObjData.DwDefineID)
	fmt.Printf("  Define Count: %d, Entry: %d of %d\n",
		simObjData.DwDefineCount, simObjData.DwEntryNumber, simObjData.DwOutOf)

	// Check if this is our camera state data (DATA_REQUEST_1)
	if simObjData.DwRequestID == DATA_REQUEST_1 && simObjData.DwDefineCount > 0 {
		// For single data elements, the data is actually stored in the DwData field itself
		// When DwDefineCount is 1, DwData contains the actual value, not a pointer
		cameraState := int32(simObjData.DwData)

		// Decode camera state value to human-readable format
		var cameraDesc string
		switch cameraState {
		case 0:
			cameraDesc = "Cockpit view"
		case 1:
			cameraDesc = "External view"
		case 2:
			cameraDesc = "Spot view"
		case 3:
			cameraDesc = "Tower view"
		default:
			cameraDesc = "Unknown view"
		}
		fmt.Printf("  📹 Camera State: %d (%s)\n", cameraState, cameraDesc)
	} else if simObjData.DwRequestID == DATA_REQUEST_ELECTRICAL && simObjData.DwDefineCount > 0 {
		// Parse electrical system data
		parseElectricalSystemData(ppData, pcbData)
	}
}

// Parse electrical system data
func parseElectricalSystemData(ppData uintptr, pcbData uint32) {
	fmt.Printf("  🔌 ELECTRICAL SYSTEM DATA:\n")

	// The electrical data starts after the SIMCONNECT_RECV_SIMOBJECT_DATA header
	// Calculate offset to the actual data
	headerSize := unsafe.Sizeof(SIMCONNECT_RECV_SIMOBJECT_DATA{})
	dataPtr := ppData + headerSize

	// Cast to our electrical data structure
	if headerSize+unsafe.Sizeof(ElectricalSystemState{}) <= uintptr(pcbData) {
		electricalData := (*ElectricalSystemState)(unsafe.Pointer(dataPtr))

		fmt.Printf("    ⚡ External Power Available: %s\n", boolToYesNo(electricalData.ExternalPowerAvailable))
		fmt.Printf("    🔌 External Power On: %s\n", boolToYesNo(electricalData.ExternalPowerOn))
		fmt.Printf("    🔋 Master Battery: %s\n", boolToYesNo(electricalData.MasterBattery))
		fmt.Printf("    🌟 Master Alternator: %s\n", boolToYesNo(electricalData.MasterAlternator))
		fmt.Printf("    🚨 Beacon Light: %s\n", boolToYesNo(electricalData.BeaconLightSwitch))
		fmt.Printf("    🧭 Nav Lights: %s\n", boolToYesNo(electricalData.NavLightSwitch))
		fmt.Printf("    🔋 Battery Voltage: %.1f V\n", electricalData.BatteryVoltage)
		fmt.Printf("    ⚡ Battery Load: %.1f A\n", electricalData.BatteryLoad)
	} else {
		fmt.Printf("    ⚠️  Insufficient data for electrical system parsing\n")
	}
}

// Helper function to convert integer bool to yes/no string
func boolToYesNo(value int32) string {
	if value != 0 {
		return "YES"
	}
	return "NO"
}

// Setup electrical system data definitions and requests
func setupElectricalSystemData() {
	fmt.Println("🔌 Setting up electrical system data monitoring...")

	// Define electrical system variables
	electricalVars := []struct {
		name     string
		units    string
		dataType uint32
	}{
		{"EXTERNAL POWER AVAILABLE", "Bool", SIMCONNECT_DATATYPE_INT32},
		{"EXTERNAL POWER ON", "Bool", SIMCONNECT_DATATYPE_INT32},
		{"ELECTRICAL MASTER BATTERY", "Bool", SIMCONNECT_DATATYPE_INT32},
		{"GENERAL ENG MASTER ALTERNATOR", "Bool", SIMCONNECT_DATATYPE_INT32},
		{"LIGHT BEACON", "Bool", SIMCONNECT_DATATYPE_INT32},
		{"LIGHT NAV", "Bool", SIMCONNECT_DATATYPE_INT32},
		{"ELECTRICAL MAIN BUS VOLTAGE", "Volts", SIMCONNECT_DATATYPE_FLOAT32},
		{"ELECTRICAL BATTERY LOAD", "Amperes", SIMCONNECT_DATATYPE_FLOAT32},
	}

	// Add each variable to the electrical data definition
	for i, v := range electricalVars {
		varNameBytes, err := syscall.BytePtrFromString(v.name)
		if err != nil {
			fmt.Printf("  ❌ Failed to convert %s to bytes: %v\n", v.name, err)
			continue
		}

		unitsBytes, err := syscall.BytePtrFromString(v.units)
		if err != nil {
			fmt.Printf("  ❌ Failed to convert units %s to bytes: %v\n", v.units, err)
			continue
		}

		val := 2000 + i

		response, _, _ := SimConnect_AddToDataDefinition.Call(
			uintptr(handle),                       // hSimConnect
			uintptr(val),                          // DefineID
			uintptr(unsafe.Pointer(varNameBytes)), // DatumName
			uintptr(unsafe.Pointer(unitsBytes)),   // UnitsName
			uintptr(v.dataType),                   // DatumType
			uintptr(0),                            // fEpsilon
		)

		hresult := uint32(response)
		if IsHRESULTSuccess(hresult) {
			fmt.Printf("  ✅ Added %s to electrical data definition\n", v.name)
		} else {
			fmt.Printf("  ❌ Failed to add %s to electrical data definition: %d\n", v.name, hresult)
		}
	}

	// Request electrical data on the SimObject
	response, _, _ := SimConnect_RequestDataOnSimObject.Call(
		uintptr(handle),                     // hSimConnect
		uintptr(DATA_REQUEST_ELECTRICAL),    // RequestID
		uintptr(DATA_DEFINITION_ELECTRICAL), // DefineID
		uintptr(SIMCONNECT_OBJECT_ID_USER),  // ObjectID - user aircraft
		uintptr(SIMCONNECT_PERIOD_SECOND),   // Period - request data every second
		uintptr(0),                          // Flags
		uintptr(0),                          // origin
		uintptr(0),                          // interval
		uintptr(0),                          // limit
	)

	hresult := uint32(response)
	if IsHRESULTSuccess(hresult) {
		fmt.Println("  ✅ Successfully requested electrical system data")
	} else {
		fmt.Printf("  ❌ Failed to request electrical system data: %d\n", hresult)
	}
}

// Setup electrical events for interaction
func setupElectricalEvents() {
	fmt.Println("⚡ Setting up electrical event mappings...")

	// Map electrical events to client event IDs
	electricalEvents := []struct {
		eventName string
		eventID   uint32
	}{
		{"TOGGLE_EXTERNAL_POWER", EVENT_ID_TOGGLE_EXTERNAL_POWER},
		{"TOGGLE_MASTER_BATTERY", EVENT_ID_TOGGLE_MASTER_BATTERY},
		{"TOGGLE_MASTER_ALTERNATOR", EVENT_ID_TOGGLE_MASTER_ALT},
		{"TOGGLE_BEACON_LIGHTS", EVENT_ID_TOGGLE_BEACON_LIGHTS},
		{"TOGGLE_NAV_LIGHTS", EVENT_ID_TOGGLE_NAV_LIGHTS},
	}

	for _, event := range electricalEvents {
		eventNameBytes, err := syscall.BytePtrFromString(event.eventName)
		if err != nil {
			fmt.Printf("  ❌ Failed to convert event name %s to bytes: %v\n", event.eventName, err)
			continue
		}

		response, _, _ := SimConnect_MapClientEventToSimEvent.Call(
			uintptr(handle),                         // hSimConnect
			uintptr(event.eventID),                  // EventID
			uintptr(unsafe.Pointer(eventNameBytes)), // EventName
		)

		hresult := uint32(response)
		if IsHRESULTSuccess(hresult) {
			fmt.Printf("  ✅ Mapped event %s to ID %d\n", event.eventName, event.eventID)
		} else {
			fmt.Printf("  ❌ Failed to map event %s: %d\n", event.eventName, hresult)
		}

		// Add event to notification group (optional - for receiving event notifications)
		response2, _, _ := SimConnect_AddClientEventToNotificationGroup.Call(
			uintptr(handle),        // hSimConnect
			uintptr(0),             // GroupID (use 0 for default group)
			uintptr(event.eventID), // EventID
		)

		hresult2 := uint32(response2)
		if IsHRESULTSuccess(hresult2) {
			fmt.Printf("  📡 Added event %s to notification group\n", event.eventName)
		}
	}

	// Set notification group priority
	response, _, _ := SimConnect_SetNotificationGroupPriority.Call(
		uintptr(handle), // hSimConnect
		uintptr(0),      // GroupID
		uintptr(1000),   // uPriority
	)

	hresult := uint32(response)
	if IsHRESULTSuccess(hresult) {
		fmt.Println("  📡 Set notification group priority")
	}
}

// Function to trigger an electrical event (example usage)
func triggerElectricalEvent(eventID uint32, data uint32) {
	response, _, _ := SimConnect_TransmitClientEvent.Call(
		uintptr(handle),                    // hSimConnect
		uintptr(SIMCONNECT_OBJECT_ID_USER), // ObjectID
		uintptr(eventID),                   // EventID
		uintptr(data),                      // dwData
		uintptr(5),                         // Priority (SIMCONNECT_GROUP_PRIORITY_HIGHEST)
		uintptr(16),                        // Flags (SIMCONNECT_EVENT_FLAG_GROUPID_IS_PRIORITY)
	)

	hresult := uint32(response)
	if IsHRESULTSuccess(hresult) {
		fmt.Printf("  ⚡ Successfully triggered event ID %d with data %d\n", eventID, data)
	} else {
		fmt.Printf("  ❌ Failed to trigger event ID %d: %d\n", eventID, hresult)
	}
}

// Parse subscribe input event data messages
func parseSubscribeInputEventData(ppData uintptr, pcbData uint32) {
	// Cast to subscribe input event structure
	subscribeData := (*SIMCONNECT_RECV_SUBSCRIBE_INPUT_EVENT)(unsafe.Pointer(ppData))

	fmt.Printf("  ✅ Successfully subscribed to input event\n")
	fmt.Printf("  Define ID: %d\n", subscribeData.DwDefineID)

	// The define ID should match what we passed to SimConnect_SubscribeInputEvent
	if subscribeData.DwDefineID == 0 {
		fmt.Printf("  🔗 Subscription confirmed for DefineID 0 (all input events)\n")
	} else {
		fmt.Printf("  🔗 Subscription confirmed for DefineID %d\n", subscribeData.DwDefineID)
	}
}

// Parse enumerate input events data messages
func parseEnumerateInputEventsData(ppData uintptr, pcbData uint32) {
	foundToggleExternalPower := false

	// Cast to enumerate input events structure
	enumInputEvents := (*SIMCONNECT_RECV_ENUMERATE_INPUT_EVENTS)(unsafe.Pointer(ppData))

	fmt.Printf("  Request ID: %d\n", enumInputEvents.DwRequestID)
	fmt.Printf("  Array Size: %d\n", enumInputEvents.DwArraySize)
	fmt.Printf("  Entry Number: %d of %d\n", enumInputEvents.DwEntryNumber+1, enumInputEvents.DwOutOf)

	// Safety check: ensure we have a reasonable array size
	if enumInputEvents.DwArraySize == 0 || enumInputEvents.DwArraySize > 1000 {
		fmt.Printf("  ⚠️  Invalid array size: %d, skipping parsing\n", enumInputEvents.DwArraySize)
		return
	}

	// The input event descriptors start immediately after the header structure
	// Calculate the offset to the data array
	headerSize := unsafe.Sizeof(SIMCONNECT_RECV_ENUMERATE_INPUT_EVENTS{})
	dataOffset := ppData + headerSize

	// Calculate the size of each descriptor
	descriptorSize := unsafe.Sizeof(SIMCONNECT_INPUT_EVENT_DESCRIPTOR{})

	// Safety check: ensure we don't read beyond the message boundaries
	totalDataSize := uintptr(enumInputEvents.DwArraySize) * descriptorSize
	if headerSize+totalDataSize > uintptr(pcbData) {
		fmt.Printf("  ⚠️  Data size exceeds message bounds, limiting to safe range\n")
		// Calculate how many complete descriptors we can safely read
		availableDataSize := uintptr(pcbData) - headerSize
		safeArraySize := uint32(availableDataSize / descriptorSize)
		if safeArraySize < enumInputEvents.DwArraySize {
			enumInputEvents.DwArraySize = safeArraySize
		}
	}

	// Parse each input event descriptor in this packet
	for i := uint32(0); i < enumInputEvents.DwArraySize && i < 100; i++ { // Limit to 100 for safety
		// Calculate pointer to the current descriptor
		descriptorPtr := dataOffset + uintptr(i)*descriptorSize

		// Safety check: ensure the descriptor pointer is within bounds
		if descriptorPtr+descriptorSize > ppData+uintptr(pcbData) {
			fmt.Printf("  ⚠️  Descriptor %d would exceed message bounds, stopping\n", i+1)
			break
		}

		descriptor := (*SIMCONNECT_INPUT_EVENT_DESCRIPTOR)(unsafe.Pointer(descriptorPtr))
		// Extract the null-terminated string from the Name array with bounds checking
		nameLength := 0
		for j := 0; j < 64 && j < len(descriptor.Name) && descriptor.Name[j] != 0; j++ {
			// Only count printable ASCII characters to avoid garbage
			if descriptor.Name[j] >= 32 && descriptor.Name[j] <= 126 {
				nameLength++
			} else {
				break // Stop at first non-printable character
			}
		}

		var name string
		if nameLength > 2 && nameLength <= 64 { // Require at least 3 characters for valid name
			name = string(descriptor.Name[:nameLength])
		} else {
			name = "(invalid name)"
		}
		// Extract the null-terminated string from the NodeNames array with bounds checking
		nodeNamesLength := 0
		for j := 0; j < 1024 && j < len(descriptor.NodeNames) && descriptor.NodeNames[j] != 0; j++ {
			if descriptor.NodeNames[j] >= 32 && descriptor.NodeNames[j] <= 126 {
				nodeNamesLength++
			} else {
				break
			}
		}

		var nodeNames string
		if nodeNamesLength > 0 && nodeNamesLength <= 1024 {
			nodeNames = string(descriptor.NodeNames[:nodeNamesLength])
		} else {
			nodeNames = "(none)"
		}
		// Decode the data type
		var dataTypeStr string
		switch descriptor.Type {
		case SIMCONNECT_DATATYPE_FLOAT32:
			dataTypeStr = "FLOAT32"
		case SIMCONNECT_DATATYPE_FLOAT64:
			dataTypeStr = "FLOAT64"
		case SIMCONNECT_DATATYPE_INT32:
			dataTypeStr = "INT32"
		case SIMCONNECT_DATATYPE_INT64:
			dataTypeStr = "INT64"
		default:
			dataTypeStr = fmt.Sprintf("TYPE_%d", descriptor.Type)
		}

		// Skip obviously invalid entries early
		if name == "(invalid name)" || len(name) < 3 {
			continue
		}

		// Check if this is the TOGGLE_EXTERNAL_POWER event we're looking for
		if name == "TOGGLE_EXTERNAL_POWER" {
			foundToggleExternalPower = true
			fmt.Printf("  🔋 *** FOUND TOGGLE_EXTERNAL_POWER! *** 🔋\n")
			fmt.Printf("    📛 Name: '%s'\n", name)
			fmt.Printf("    🔗 Hash: %d\n", descriptor.Hash)
			fmt.Printf("    📊 Type: %s (%d)\n", dataTypeStr, descriptor.Type)
			fmt.Printf("    🔗 Node Names: '%s'\n", nodeNames)
			fmt.Printf("  🔋 *** END TOGGLE_EXTERNAL_POWER *** 🔋\n")
		} else if strings.Contains(strings.ToUpper(name), "POWER") ||
			strings.Contains(strings.ToUpper(name), "EXTERNAL") ||
			strings.Contains(strings.ToUpper(name), "TOGGLE") {
			// Show power-related events prominently
			fmt.Printf("  ⚡ Power-Related Event:\n")
			fmt.Printf("    📛 Name: '%s'\n", name)
			fmt.Printf("    🔗 Hash: %d\n", descriptor.Hash)
			fmt.Printf("    📊 Type: %s (%d)\n", dataTypeStr, descriptor.Type)
			fmt.Printf("    🔗 Node Names: '%s'\n", nodeNames)
		} else if i < 10 || i%25 == 0 {
			// Only show first 10 events, then every 25th to reduce noise
			fmt.Printf("  🎮 Input Event %d:\n", i+1)
			fmt.Printf("    📛 Name: '%s'\n", name)
			fmt.Printf("    🔗 Hash: %d\n", descriptor.Hash)
			fmt.Printf("    📊 Type: %s (%d)\n", dataTypeStr, descriptor.Type)
			fmt.Printf("    🔗 Node Names: '%s'\n", nodeNames)
		}

		// Stop if we've processed enough events for this demonstration
		if i >= 50 {
			fmt.Printf("  ℹ️  Processed %d events, stopping for brevity\n", i+1)
			break
		}
	}

	// Summary
	if foundToggleExternalPower {
		fmt.Printf("  ✅ SUCCESS: TOGGLE_EXTERNAL_POWER event was found in this enumeration!\n")
	} else {
		fmt.Printf("  ❌ TOGGLE_EXTERNAL_POWER event not found in this enumeration packet\n")
	}
}

func main() {
	// Initialize the SimConnect procedures
	initProcedures()

	nameBytes, err := syscall.BytePtrFromString("Matyho SimConnect Example")
	if err != nil {
		fmt.Printf("failed to convert name to bytes: %v", err)
		return
	}
	// Call SimConnect_Open
	// HRESULT SimConnect_Open(HANDLE* phSimConnect, LPCSTR szName, HWND hWnd,
	//                         DWORD UserEventWin32, HANDLE hEventHandle, DWORD ConfigIndex)
	response, _, _ := SimConnect_Open.Call(
		uintptr(unsafe.Pointer(&handle)),   // phSimConnect
		uintptr(unsafe.Pointer(nameBytes)), // szName
		0,                                  // hWnd (NULL)
		0,                                  // UserEventWin32
		0,                                  // hEventHandle
		uintptr(0),                         // ConfigIndex
	)

	hresult := uint32(response)
	// Verify handle was set
	if handle == 0 {
		fmt.Println("SimConnect_Open succeeded but handle is null")
		return
	}

	fmt.Println("SimConnect_Open response:", IsHRESULTSuccess(hresult), IsHRESULTFailure(hresult))

	nameBytes2, _ := syscall.BytePtrFromString("AircraftLoaded")

	responseSystemState, _, _ := SimConnect_RequestSystemState.Call(
		uintptr(handle),                     // phSimConnect
		uintptr(0),                          // DefineID (0 for default)
		uintptr(unsafe.Pointer(nameBytes2)), // DataName
	)
	hresultSystemState := uint32(responseSystemState)
	fmt.Println("SimConnect_RequestSystemState response:", IsHRESULTSuccess(hresultSystemState), IsHRESULTFailure(hresultSystemState))

	// Define CAMERA STATE simulation variable (this is a valid SimConnect system state)
	cameraStateName, err := syscall.BytePtrFromString("CAMERA STATE")
	if err != nil {
		fmt.Printf("failed to convert Camera State to bytes: %v", err)
		return
	}
	enumUnits, err := syscall.BytePtrFromString("Enum")
	if err != nil {
		fmt.Printf("failed to convert Enum units to bytes: %v", err)
		return
	}

	// Add camera state to data definition
	response2, _, _ := SimConnect_AddToDataDefinition.Call(
		uintptr(handle),                          // hSimConnect
		uintptr(DATA_DEFINITION_1),               // DefineID - use unique ID
		uintptr(unsafe.Pointer(cameraStateName)), // DatumName - CAMERA STATE is valid
		uintptr(unsafe.Pointer(enumUnits)),       // UnitsName - Enum for camera state
		uintptr(SIMCONNECT_DATATYPE_INT32),       // DatumType - Camera state is typically an integer enum
		uintptr(0),                               // fEpsilon - default 0
	)

	hresult2 := uint32(response2)
	fmt.Println("SimConnect_AddToDataDefinition response:", IsHRESULTSuccess(hresult2), IsHRESULTFailure(hresult2))

	// Request data on the SimObject with correct parameter order
	response4, _, _ := SimConnect_RequestDataOnSimObject.Call(
		uintptr(handle),                    // hSimConnect
		uintptr(DATA_REQUEST_1),            // RequestID - unique request ID
		uintptr(DATA_DEFINITION_1),         // DefineID - matches our data definition
		uintptr(SIMCONNECT_OBJECT_ID_USER), // ObjectID - user aircraft
		uintptr(SIMCONNECT_PERIOD_SECOND),  // Period - request data every second
		uintptr(0),                         // Flags - default (no special flags)
		uintptr(0),                         // origin - start immediately
		uintptr(0),                         // interval - every period
		uintptr(0),                         // limit - unlimited
	)
	hresult4 := uint32(response4)
	fmt.Println("SimConnect_RequestDataOnSimObject response:", IsHRESULTSuccess(hresult4), IsHRESULTFailure(hresult4))

	// Set up electrical system data definition and request
	setupElectricalSystemData()

	// Set up electrical events mapping
	setupElectricalEvents()

	responseSystemState2, _, _ := SimConnect_RequestSystemState.Call(
		uintptr(handle),                     // phSimConnect
		uintptr(0),                          // DefineID (0 for default)
		uintptr(unsafe.Pointer(nameBytes2)), // DataName
	)

	hresultSystemState2 := uint32(responseSystemState2)
	fmt.Println("SimConnect_RequestSystemState response:", IsHRESULTSuccess(hresultSystemState2), IsHRESULTFailure(hresultSystemState2))

	name3Bytes, _ := syscall.BytePtrFromString("Pause")

	resp4, _, _ := SimConnect_SubscribeToSystemEvent.Call(
		uintptr(handle),                     // hSimConnect
		uintptr(1010),                       // UserEventID (0 for default)
		uintptr(unsafe.Pointer(name3Bytes)), // szEventName

	)

	hresult4 = uint32(resp4)
	fmt.Println("SimConnect_SubscribeToSystemEvent response:", IsHRESULTSuccess(hresult4), IsHRESULTFailure(hresult4))

	resp5, _, _ := SimConnect_EnumerateInputEvents.Call(
		uintptr(handle), // hSimConnect
		uintptr(90010),  // DefineID (0 for default)
	)

	hresult5 := uint32(resp5)
	fmt.Println("SimConnect_EnumerateInputEvents response:", IsHRESULTSuccess(hresult5), IsHRESULTFailure(hresult5))

	resp6, _, _ := SimConnect_SubscribeInputEvent.Call(
		uintptr(handle), // hSimConnect
		uintptr(56940),  // DefineID (0 for default)
	)
	hresult6 := uint32(resp6)

	fmt.Println("SimConnect_SubscribeInputEvents response:", IsHRESULTSuccess(hresult6), IsHRESULTFailure(hresult6))

	// Example: Trigger an electrical event after a few seconds
	fmt.Println("🔌 Waiting 5 seconds before triggering electrical events...")
	time.Sleep(5 * time.Second)

	// Example: Toggle external power
	fmt.Println("⚡ Triggering TOGGLE_EXTERNAL_POWER...")
	triggerElectricalEvent(EVENT_ID_TOGGLE_EXTERNAL_POWER, 0)

	time.Sleep(2 * time.Second)

	// Example: Toggle master battery
	fmt.Println("🔋 Triggering TOGGLE_MASTER_BATTERY...")
	triggerElectricalEvent(EVENT_ID_TOGGLE_MASTER_BATTERY, 0)

	// Hold for a while and dispatch messages
	for i := 0; i < 100000; i++ { // Changed to finite loop for demonstration
		var ppData uintptr
		var pcbData uint32

		// Call SimConnect_GetNextDispatch
		responseDispatch, _, _ := SimConnect_GetNextDispatch.Call(
			uintptr(handle),                   // hSimConnect
			uintptr(unsafe.Pointer(&ppData)),  // ppData
			uintptr(unsafe.Pointer(&pcbData)), // pcbData
		)

		hresultDispatch := uint32(responseDispatch)

		//fmt.Println("SimConnect_GetNextDispatch response:", IsHRESULTSuccess(hresultDispatch), IsHRESULTFailure(hresultDispatch))
		if IsHRESULTSuccess(hresultDispatch) {
			// Parse the received SimConnect data
			parseSimConnectData(ppData, pcbData)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Cleanup - close SimConnect connection
	SimConnect_Close.Call(uintptr(handle))
	time.Sleep(100 * time.Millisecond) // Give some time for the close operation to complete
}
