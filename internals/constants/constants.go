package constants

import "time"

// File permissions
const (
	DefaultFilePermission = 0644
	DefaultDirPermission  = 0755
)

// Timing constants
const (
	FileWatcherDebounce = 300 * time.Millisecond
	SpinnerCharSetIndex = 14
	SpinnerInterval     = 100 * time.Millisecond
	SpinnerColor        = "cyan"
)

// Spinner timing stages for user feedback
const (
	SpinnerScanningDelay   = 300 * time.Millisecond
	SpinnerParsingDelay    = 400 * time.Millisecond
	SpinnerBuildingDelay   = 350 * time.Millisecond
	SpinnerGeneratingDelay = 250 * time.Millisecond
)

// WebSocket constants
const (
	WebSocketReadBufferSize  = 1024
	WebSocketWriteBufferSize = 1024
)

// File patterns and defaults
const (
	DefaultConfigFileName = "docunyan.yml"
	DefaultDTODirectory   = "dto"
	DefaultOutputFile     = "generated/generated.json"
	DefaultOutputPattern  = "docunyan_gen_20060102_150405.json"
)

// HTTP server constants
const (
	MinPortNumber = 3000
	MaxPortNumber = 9000
	ServerReadTimeout  = 30 * time.Second
	ServerWriteTimeout = 30 * time.Second
)

// Validation limits
const (
	MaxPathLength      = 260
	MaxTitleLength     = 100
	MaxVersionLength   = 20
	MaxDescriptionLength = 1000
)
