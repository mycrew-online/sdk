package client

import (
	"fmt"
	"time"
	"unsafe"
)

func (e *Engine) Listen() error {
	if !e.system.IsConnected {
		return nil // No messages to process if not connected
	}

	// Start a goroutine to dispatch messages and not block the main thread
	go func() {
		if err := e.dispatch(); err != nil {
			// Handle error (e.g., log it)
			// fmt.Println("Error in dispatch:", err)
		}
	}()

	return nil
}

func (e *Engine) dispatch() error {
	if !e.system.IsConnected {
		return nil // No messages to process if not connected
	}

	// Process messages from the SimConnect server
	for {
		var ppData uintptr
		var pcbData uint32

		// Call SimConnect_GetNextDispatch
		responseDispatch, _, _ := SimConnect_GetNextDispatch.Call(
			uintptr(e.handle),                 // hSimConnect
			uintptr(unsafe.Pointer(&ppData)),  // ppData
			uintptr(unsafe.Pointer(&pcbData)), // pcbData
		)

		hresultDispatch := uint32(responseDispatch)

		fmt.Println("SimConnect_GetNextDispatch response:", hresultDispatch)

		//fmt.Println("SimConnect_GetNextDispatch response:", IsHRESULTSuccess(hresultDispatch), IsHRESULTFailure(hresultDispatch))
		if IsHRESULTSuccess(hresultDispatch) {
			// Parse the received SimConnect data
			go parseSimConnectData(ppData, pcbData)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
