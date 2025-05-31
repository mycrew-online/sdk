package client

import (
	"context"
	"time"
	"unsafe"
)

func (e *Engine) Listen() <-chan any {
	if !e.system.IsConnected {
		return nil // No messages to process if not connected
	}

	// Initialize context for graceful shutdown if not already done
	if e.ctx == nil {
		e.ctx, e.cancel = context.WithCancel(context.Background())
		e.done = make(chan struct{})
	}

	// Start a goroutine to dispatch messages and not block the main thread
	go func() {
		defer close(e.done)
		if err := e.dispatch(); err != nil {
			// Handle error (e.g., log it)
			// fmt.Println("Error in dispatch:", err)
		}
	}()

	return e.stream
}

func (e *Engine) dispatch() error {
	if !e.system.IsConnected {
		return nil // No messages to process if not connected
	}

	// We should also request some internal check to ensure sim state

	// Process messages from the SimConnect server with graceful shutdown
	for {
		select {
		case <-e.ctx.Done():
			return e.ctx.Err() // Graceful shutdown requested
		default:
			var ppData uintptr
			var pcbData uint32

			// Call SimConnect_GetNextDispatch
			responseDispatch, _, _ := SimConnect_GetNextDispatch.Call(
				uintptr(e.handle),                 // hSimConnect
				uintptr(unsafe.Pointer(&ppData)),  // ppData
				uintptr(unsafe.Pointer(&pcbData)), // pcbData
			)

			hresultDispatch := uint32(responseDispatch)

			if IsHRESULTSuccess(hresultDispatch) {
				// Parse and send message to channel (non-blocking)
				e.handleMessage(ppData, pcbData)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// handleMessage processes messages and sends them to the stream channel
func (e *Engine) handleMessage(ppData uintptr, pcbData uint32) {
	// Parse the message
	msg := parseSimConnectToChannelMessage(ppData, pcbData)

	// Handle QUIT messages for natural shutdown
	if msg != nil && e.isQuitMessage(msg) {
		e.system.IsConnected = false
		if e.cancel != nil {
			e.cancel() // Signal shutdown
		}
		return
	}

	// Send to stream if channel is available (non-blocking)
	if msg != nil {
		select {
		case e.stream <- msg:
		default:
			// Channel full, drop message to prevent blocking
		}
	}
}

// isQuitMessage checks if the message is a QUIT signal
func (e *Engine) isQuitMessage(msg any) bool {
	// Simple type assertion to check for quit message
	if msgMap, ok := msg.(map[string]any); ok {
		if msgType, exists := msgMap["type"]; exists {
			return msgType == "QUIT"
		}
	}
	return false
}
