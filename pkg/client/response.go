package client

import (
	"fmt"
	"unsafe"

	"github.com/mycrew-online/sdk/pkg/types"
)

// This is just temp for better API design. This needs to be more precise
func parseSimConnectData(ppData uintptr, pcbData uint32) {
	if ppData == 0 || pcbData == 0 {
		fmt.Println("No data received")
		return
	}

	// Cast the pointer to the base SIMCONNECT_RECV structure
	recv := (*types.SIMCONNECT_RECV)(unsafe.Pointer(ppData))
	fmt.Printf("Received message - Size: %d, Version: %d, ID: %d\n",
		recv.DwSize, recv.DwVersion, recv.DwID)
	// Check what type of message we received based on the ID
	switch recv.DwID {
	case types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA:
		fmt.Println("📊 SIMOBJECT_DATA received")
		//parseSimObjectData(ppData, pcbData)
	case types.SIMCONNECT_RECV_ID_OPEN:
		fmt.Println("🔓 OPEN confirmation received")
	case types.SIMCONNECT_RECV_ID_EXCEPTION:
		fmt.Println("❌ EXCEPTION received")
	case types.SIMCONNECT_RECV_ID_SYSTEM_STATE:
		fmt.Println("🔧 SYSTEM_STATE received")
		//parseSystemStateData(ppData, pcbData)
	case types.SIMCONNECT_RECV_ID_EVENT:
		fmt.Println("📡 EVENT received")
		//parseEventData(ppData, pcbData)
	case types.SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENTS:
		fmt.Println("🎮 ENUMERATE_INPUT_EVENTS received")
		//parseEnumerateInputEventsData(ppData, pcbData)
	case types.SIMCONNECT_RECV_ID_SUBSCRIBE_INPUT_EVENT:
		fmt.Println("🔗 SUBSCRIBE_INPUT_EVENT received")
		//parseSubscribeInputEventData(ppData, pcbData)
	case types.SIMCONNECT_RECV_ID_QUIT:
		fmt.Println("👋 QUIT received")
	default:
		fmt.Printf("❓ Unknown message type: %d\n", recv.DwID)
	}
}

// parseSimConnectToChannelMessage converts SimConnect data to a channel message
func parseSimConnectToChannelMessage(ppData uintptr, pcbData uint32) any {
	if ppData == 0 || pcbData == 0 {
		return nil
	}

	// Cast the pointer to the base SIMCONNECT_RECV structure
	recv := (*types.SIMCONNECT_RECV)(unsafe.Pointer(ppData))

	// Create a simple message structure for the channel
	msg := map[string]any{
		"size":       recv.DwSize,
		"version":    recv.DwVersion,
		"type":       getMessageTypeName(recv.DwID),
		"id":         recv.DwID,
		"data":       ppData,
		"size_bytes": pcbData,
	}

	return msg
}

// getMessageTypeName converts message ID to readable string
func getMessageTypeName(id types.SimConnectRecvID) string {
	switch id {
	case types.SIMCONNECT_RECV_ID_NULL:
		return "NULL"
	case types.SIMCONNECT_RECV_ID_EXCEPTION:
		return "EXCEPTION"
	case types.SIMCONNECT_RECV_ID_OPEN:
		return "OPEN"
	case types.SIMCONNECT_RECV_ID_QUIT:
		return "QUIT"
	case types.SIMCONNECT_RECV_ID_EVENT:
		return "EVENT"
	case types.SIMCONNECT_RECV_ID_SIMOBJECT_DATA:
		return "SIMOBJECT_DATA"
	case types.SIMCONNECT_RECV_ID_SYSTEM_STATE:
		return "SYSTEM_STATE"
	case types.SIMCONNECT_RECV_ID_ENUMERATE_INPUT_EVENTS:
		return "ENUMERATE_INPUT_EVENTS"
	case types.SIMCONNECT_RECV_ID_SUBSCRIBE_INPUT_EVENT:
		return "SUBSCRIBE_INPUT_EVENT"
	default:
		return "UNKNOWN"
	}
}
