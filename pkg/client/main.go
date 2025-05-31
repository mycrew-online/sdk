package client

import "syscall"

const (
	DLL_DEFAULT_PATH = "C:/MSFS 2024 SDK/SimConnect SDK/lib/SimConnect.dll"
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
		dll:    dll(path),
		name:   name,
		system: state,
	}

	// TODO Error handling for DLL loading???
	client.bootstrap()

	return client
}

func dll(path string) *syscall.LazyDLL {
	// This method seems useless, but we can extend it later if needed.
	return syscall.NewLazyDLL(path)
}
