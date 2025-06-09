package client

import (
	"context"
	"sync"
	"syscall"

	"github.com/mycrew-online/sdk/pkg/types"
)

type Engine struct {
	dll    *syscall.LazyDLL
	handle uintptr
	name   string
	system *SystemState
	stream chan any

	// Shutdown coordination with async safety
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	// Async safety controls
	mu          sync.RWMutex // Protects shared state
	startOnce   sync.Once    // Ensures Listen() is called only once
	contextOnce sync.Once    // Ensures context initialization happens only once
	closeOnce   sync.Once    // Ensures Close() is called only once
	isListening bool         // Protected by mu, tracks if listening is active
	// Data type tracking for sim variables
	dataTypeRegistry map[uint32]types.SimConnectDataType // DefineID → DataType mapping

	// Unhandled message tracking for monitoring and debugging
	unhandledMessageStats map[types.SimConnectRecvID]int64 // MessageType → Count
	lastUnhandledCheck    int64                            // Timestamp of last stats check
}

type SystemState struct {
	mu          sync.RWMutex // Protects IsConnected
	IsConnected bool
}
