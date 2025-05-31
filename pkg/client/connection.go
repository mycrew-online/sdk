package client

import (
	"fmt"
	"syscall"
	"unsafe"
)

type Connection interface {
	Open() error
	Close() error
	Listen() <-chan any
}

func (e *Engine) Open() error {
	if e.system.IsConnected {
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

	e.system.IsConnected = true

	return nil
}

func (e *Engine) Close() error {
	if !e.system.IsConnected {
		return nil // No need to close if not connected
		//return fmt.Errorf("client and server have not opened connection, skipping")
	}

	// Signal graceful shutdown to dispatch goroutine
	if e.cancel != nil {
		e.cancel()
		// Wait for dispatch to finish if done channel exists
		if e.done != nil {
			<-e.done
		}
	}

	// Call SimConnect_Close
	// HRESULT SimConnect_Close(HANDLE hSimConnect)
	hresult, _, _ := SimConnect_Close.Call(e.handle)

	response := uint32(hresult)

	if !IsHRESULTSuccess(response) {
		return fmt.Errorf("SimConnect_Close failed with HRESULT: 0x%08X", response)
	}

	e.system.IsConnected = false
	e.handle = 0

	return nil
}
