package client

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/mycrew-online/sdk/pkg/types"
)

type Connection interface {
	Open() error
	Close() error
	Listen() <-chan any
	AddSimVar(defID uint32, varName string, units string, dataType types.SimConnectDataType) error
	RequestSimVarData(defID uint32, requestID uint32) error
	SetSimVar(defID uint32, value interface{}) error
}

func (e *Engine) Open() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Thread-safe check for connection status
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if isConnected {
		return fmt.Errorf("client, server connection is already open, skipping")
	}

	// Convert name to null-terminated byte array
	nameBytes, err := syscall.BytePtrFromString(e.name)
	if err != nil {
		return fmt.Errorf("failed to convert name to bytes: %v", err)
	}
	// Call SimConnect_Open
	// HRESULT SimConnect_Open(HANDLE* phSimConnect, LPCSTR szName, HWND hWnd,
	//                         DWORD UserEventWin32, HANDLE hEventHandle, DWORD ConfigIndex)
	hresult, _, _ := SimConnect_Open.Call(
		uintptr(unsafe.Pointer(&e.handle)), // phSimConnect
		uintptr(unsafe.Pointer(nameBytes)), // szName
		0,                                  // hWnd (NULL)
		0,                                  // UserEventWin32
		0,                                  // hEventHandle
		uintptr(0),                         // ConfigIndex
	)

	response := uint32(hresult)

	if !IsHRESULTSuccess(response) {
		return fmt.Errorf("SimConnect_Open failed with HRESULT: 0x%08X", response)
	}

	// Verify handle was set or return an error
	if e.handle == 0 {
		return fmt.Errorf("SimConnect_Open succeeded but handle is null")
	}

	// Thread-safe update of connection status
	e.system.mu.Lock()
	e.system.IsConnected = true
	e.system.mu.Unlock()

	return nil
}

func (e *Engine) Close() error {
	var closeErr error

	// Use sync.Once to ensure close operations happen only once
	e.closeOnce.Do(func() {
		// Thread-safe check for connection status first
		e.system.mu.RLock()
		isConnected := e.system.IsConnected
		e.system.mu.RUnlock()

		if !isConnected {
			closeErr = nil // No need to close if not connected
			return
		}

		// Get cancel function and done channel while holding lock briefly
		e.mu.Lock()
		cancel := e.cancel
		done := e.done
		e.mu.Unlock()

		// Signal graceful shutdown to dispatch goroutine (without holding the main mutex)
		if cancel != nil {
			cancel()
			// Wait for dispatch to finish if done channel exists
			if done != nil {
				<-done
			}
		}
		// Now acquire the lock for the actual close operations
		e.mu.Lock()
		defer e.mu.Unlock()

		// Call SimConnect_Close
		// HRESULT SimConnect_Close(HANDLE hSimConnect)
		hresult, _, _ := SimConnect_Close.Call(e.handle)

		response := uint32(hresult)

		if !IsHRESULTSuccess(response) {
			closeErr = fmt.Errorf("SimConnect_Close failed with HRESULT: 0x%08X", response)
			return
		}

		// Thread-safe update of connection status and handle
		e.system.mu.Lock()
		e.system.IsConnected = false
		e.system.mu.Unlock()

		e.handle = 0
		e.isListening = false

		closeErr = nil
	})

	return closeErr
}
