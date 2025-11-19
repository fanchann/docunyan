package constants

import "time"

// File permissions
const (
	DefaultFilePermission = 0644
)

// Timing constants
const (
	FileWatcherDebounce = 100 * time.Millisecond
)

// WebSocket constants
const (
	WebSocketReadBufferSize  = 1024
	WebSocketWriteBufferSize = 1024
)
