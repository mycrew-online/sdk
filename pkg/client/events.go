package client

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/mycrew-online/sdk/pkg/types"
)

// SubscribeToSystemEvent subscribes to a system event from the simulator
func (e *Engine) SubscribeToSystemEvent(eventID uint32, eventName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if connected
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to SimConnect")
	}

	// Convert event name to C string
	eventNamePtr, err := syscall.BytePtrFromString(eventName)
	if err != nil {
		return fmt.Errorf("failed to convert event name to C string: %w", err)
	}

	// Call SimConnect_SubscribeToSystemEvent
	r1, _, err := SimConnect_SubscribeToSystemEvent.Call(
		uintptr(e.handle),
		uintptr(eventID),
		uintptr(unsafe.Pointer(eventNamePtr)),
	)

	if r1 != 0 {
		return fmt.Errorf("SimConnect_SubscribeToSystemEvent failed: %w", err)
	}

	return nil
}

// MapClientEventToSimEvent maps a client event ID to a simulator event name
func (e *Engine) MapClientEventToSimEvent(eventID types.ClientEventID, eventName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if connected
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to SimConnect")
	}

	// Convert event name to C string
	eventNamePtr, err := syscall.BytePtrFromString(eventName)
	if err != nil {
		return fmt.Errorf("failed to convert event name to C string: %w", err)
	}

	// Call SimConnect_MapClientEventToSimEvent
	r1, _, err := SimConnect_MapClientEventToSimEvent.Call(
		uintptr(e.handle),
		uintptr(eventID),
		uintptr(unsafe.Pointer(eventNamePtr)),
	)

	if r1 != 0 {
		return fmt.Errorf("SimConnect_MapClientEventToSimEvent failed: %w", err)
	}

	return nil
}

// AddClientEventToNotificationGroup adds a client event to a notification group
func (e *Engine) AddClientEventToNotificationGroup(groupID types.NotificationGroupID, eventID types.ClientEventID, maskable bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if connected
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to SimConnect")
	}

	// Convert bool to int32 (0 = false, 1 = true)
	maskableInt := 0
	if maskable {
		maskableInt = 1
	}

	// Call SimConnect_AddClientEventToNotificationGroup
	r1, _, err := SimConnect_AddClientEventToNotificationGroup.Call(
		uintptr(e.handle),
		uintptr(groupID),
		uintptr(eventID),
		uintptr(maskableInt),
	)

	if r1 != 0 {
		return fmt.Errorf("SimConnect_AddClientEventToNotificationGroup failed: %w", err)
	}

	return nil
}

// SetNotificationGroupPriority sets the priority of a notification group
func (e *Engine) SetNotificationGroupPriority(groupID types.NotificationGroupID, priority uint32) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if connected
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to SimConnect")
	}

	// Call SimConnect_SetNotificationGroupPriority
	r1, _, err := SimConnect_SetNotificationGroupPriority.Call(
		uintptr(e.handle),
		uintptr(groupID),
		uintptr(priority),
	)

	if r1 != 0 {
		return fmt.Errorf("SimConnect_SetNotificationGroupPriority failed: %w", err)
	}

	return nil
}

// TransmitClientEvent transmits a client event to the simulator
func (e *Engine) TransmitClientEvent(objectID uint32, eventID types.ClientEventID, data uint32, groupID types.NotificationGroupID, flags uint32) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if connected
	e.system.mu.RLock()
	isConnected := e.system.IsConnected
	e.system.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected to SimConnect")
	}

	// Call SimConnect_TransmitClientEvent
	r1, _, err := SimConnect_TransmitClientEvent.Call(
		uintptr(e.handle),
		uintptr(objectID),
		uintptr(eventID),
		uintptr(data),
		uintptr(groupID),
		uintptr(flags),
	)

	if r1 != 0 {
		return fmt.Errorf("SimConnect_TransmitClientEvent failed: %w", err)
	}

	return nil
}
