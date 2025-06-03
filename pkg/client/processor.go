package client

import (
	"context"
	"time"
	"unsafe"
)

func (e *Engine) Listen() <-chan any {
	// Thread-safe check for connection status
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return nil // No messages to process if not connected
	}

	// Thread-safe check if already listening
	e.mu.RLock()
	alreadyListening := e.isListening
	e.mu.RUnlock()

	if alreadyListening {
		return e.stream // Return existing stream if already listening
	}

	// Use sync.Once to ensure context and goroutine are initialized only once
	e.contextOnce.Do(func() {
		e.ctx, e.cancel = context.WithCancel(context.Background())
		e.done = make(chan struct{})
	})

	// Use sync.Once to ensure only one dispatch goroutine is started
	e.startOnce.Do(func() {
		e.mu.Lock()
		e.isListening = true
		e.mu.Unlock()

		// Start a goroutine to dispatch messages and not block the main thread
		go func() {
			defer close(e.done)
			if err := e.dispatch(); err != nil {
				// Handle error (e.g., log it)
				// fmt.Println("Error in dispatch:", err)
			}
			// Mark as no longer listening when dispatch exits
			e.mu.Lock()
			e.isListening = false
			e.mu.Unlock()
		}()
	})

	return e.stream
}

func (e *Engine) dispatch() error {
	// Thread-safe check for connection status
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
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

			// Thread-safe access to handle
			e.mu.RLock()
			handle := e.handle
			e.mu.RUnlock()

			// Call SimConnect_GetNextDispatch
			responseDispatch, _, _ := SimConnect_GetNextDispatch.Call(
				uintptr(handle),                   // hSimConnect
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
	msg := e.parseSimConnectToChannelMessage(ppData, pcbData)
	// Handle QUIT messages for natural shutdown
	if msg != nil && e.isQuitMessage(msg) {
		// Thread-safe update of connection status
		e.system.mu.Lock()
		e.system.IsConnected = false
		e.system.mu.Unlock()

		// Thread-safe access to cancel function
		e.mu.RLock()
		cancel := e.cancel
		e.mu.RUnlock()

		if cancel != nil {
			cancel() // Signal shutdown
		}
		return
	}

	// Send to stream if channel is available (non-blocking with buffered channel)
	if msg != nil {
		select {
		case e.stream <- msg:
			// Message sent successfully
		default:
			// Channel full, drop message to prevent blocking
			// Consider logging this event for debugging
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
