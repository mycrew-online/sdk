package client

import (
	"syscall"

	"github.com/mycrew-online/sdk/pkg/types"
)

const (
	DLL_DEFAULT_PATH = "C:/MSFS 2024 SDK/SimConnect SDK/lib/SimConnect.dll"
	// Default buffer size for the message stream channel
	DEFAULT_STREAM_BUFFER_SIZE = 100
)

func New(name string) Connection {
	return NewWithCustomDLL(
		name,
		DLL_DEFAULT_PATH,
	)
}

func NewWithCustomDLL(name string, path string) Connection {
	state := &SystemState{
		IsConnected: false,
	}
	client := &Engine{
		dll:              dll(path),
		name:             name,
		system:           state,
		stream:           make(chan any, DEFAULT_STREAM_BUFFER_SIZE), // Buffered channel for message processing
		dataTypeRegistry: make(map[uint32]types.SimConnectDataType),  // Initialize data type tracking
	}

	// TODO Error handling for DLL loading???
	client.bootstrap()

	return client
}

func dll(path string) *syscall.LazyDLL {
	// This method seems useless, but we can extend it later if needed.
	return syscall.NewLazyDLL(path)
}
