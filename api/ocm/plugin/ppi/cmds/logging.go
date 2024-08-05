package cmds

import (
	"encoding/json"
)

type LoggingHandler interface {
	HandleConfig(data []byte) error
}

var handler LoggingHandler

// RegisterLoggingConfigHandler is used to register a configuration handler
// for logging configration passed by the OCM library.
// If standard mandelsoft logging is used, it can be adapted
// by adding the ananymous import  of the ppi/logging package.
func RegisterLoggingConfigHandler(h LoggingHandler) {
	handler = h
}

// LoggingConfiguration describes logging configuration for a slave executables like
// plugins.
// If mandelsoft logging is used please use ocm.software/ocm/api/utils/cobrautils/logging.LoggingConfiguration,
// instead.
type LoggingConfiguration struct {
	LogFileName string          `json:"logFileName"`
	LogConfig   json.RawMessage `json:"logConfig"`
	Json        bool            `json:"json,omitempty"`
}
