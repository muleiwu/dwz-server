package interfaces

import "github.com/muleiwu/gsr"

// LoggerInterface aliases gsr.Logger so that any implementation satisfying gsr
// (notably go-web's logger) is directly usable everywhere the dwz codebase
// references the legacy interface name.
type LoggerInterface = gsr.Logger

// LoggerFieldInterface aliases gsr.LoggerField for the same reason.
type LoggerFieldInterface = gsr.LoggerField
