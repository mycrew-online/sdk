package client

import "syscall"

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

func (e *Engine) bootstrap() error {
	// Load the procedures from the SimConnect DLL to make them available for use.
	e.loadProcedures()
	// Here we would implement the logic to initialize the processes.
	// This might involve loading process information from the SimConnect server, setting up any necessary event handlers, etc.
	// For now, we will just return nil to indicate success.
	return nil
}

func (e *Engine) loadProcedures() error {
	// SimConnect_Open procedure
	SimConnect_Open = e.dll.NewProc("SimConnect_Open")
	// SimConnect_Close procedure
	SimConnect_Close = e.dll.NewProc("SimConnect_Close")
	// SimConnect_GetNextDispatch procedure
	SimConnect_GetNextDispatch = e.dll.NewProc("SimConnect_GetNextDispatch")
	// SimConnect_AddToDataDefinition procedure
	SimConnect_AddToDataDefinition = e.dll.NewProc("SimConnect_AddToDataDefinition")
	// SimConnect_RequestDataOnSimObject procedure
	SimConnect_RequestDataOnSimObject = e.dll.NewProc("SimConnect_RequestDataOnSimObject")
	// SimConnect_ClearDataDefinition procedure
	SimConnect_ClearDataDefinition = e.dll.NewProc("SimConnect_ClearDataDefinition")
	// SimConnect_RequestSystemState procedure
	SimConnect_RequestSystemState = e.dll.NewProc("SimConnect_RequestSystemState")
	// SimConnect_SetDataOnSimObject procedure
	SimConnect_SetDataOnSimObject = e.dll.NewProc("SimConnect_SetDataOnSimObject")
	// SimConnect_SubscribeToSystemEvent procedure
	SimConnect_SubscribeToSystemEvent = e.dll.NewProc("SimConnect_SubscribeToSystemEvent")
	// SimConnect_SetSystemEventState procedure
	SimConnect_SetSystemEventState = e.dll.NewProc("SimConnect_SetSystemEventState")
	// SimConnect_EnumerateInputEventParams
	SimConnect_EnumerateInputEvents = e.dll.NewProc("SimConnect_EnumerateInputEvents")
	// SimConnect_SubscribeInputEvent procedure
	SimConnect_SubscribeInputEvent = e.dll.NewProc("SimConnect_SubscribeInputEvent")
	// SimConnect_MapClientEventToSimEvent procedure
	SimConnect_MapClientEventToSimEvent = e.dll.NewProc("SimConnect_MapClientEventToSimEvent")
	// SimConnect_TransmitClientEvent procedure
	SimConnect_TransmitClientEvent = e.dll.NewProc("SimConnect_TransmitClientEvent")
	// SimConnect_AddClientEventToNotificationGroup procedure
	SimConnect_AddClientEventToNotificationGroup = e.dll.NewProc("SimConnect_AddClientEventToNotificationGroup")
	// SimConnect_SetNotificationGroupPriority procedure
	SimConnect_SetNotificationGroupPriority = e.dll.NewProc("SimConnect_SetNotificationGroupPriority")
	// Return nil to indicate that the procedures were loaded successfully, as there is no error handling on syscall.NewLazyProc.
	return nil
}
