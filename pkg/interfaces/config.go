package interfaces

import "github.com/muleiwu/gsr"

// ConfigInterface aliases gsr.Provider so go-web's Viper-backed config provider
// satisfies the dwz interface used throughout app/* code.
type ConfigInterface = gsr.Provider
