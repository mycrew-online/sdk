package client

import (
	"context"
	"syscall"
)

type Engine struct {
	dll    *syscall.LazyDLL
	handle uintptr
	name   string
	system *SystemState
	stream chan any

	// Shutdown coordination - minimal additions
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

type SystemState struct {
	IsConnected bool
}
