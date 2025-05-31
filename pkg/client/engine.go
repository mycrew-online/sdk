package client

import "syscall"

type Engine struct {
	dll    *syscall.LazyDLL
	handle uintptr
	name   string
	system *SystemState
}

type SystemState struct {
	IsConnected bool
}
