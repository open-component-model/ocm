package command

import (
	"context"
)

type CommandConfigHandler interface {
	HandleConfig(ctx context.Context, data []byte) (context.Context, error)
}

var handler CommandConfigHandler

// RegisterCommandConfigHandler is used to register a configuration handler
// for OCM configuration passed by the OCM library.
// If the OCM config framework is , it can be adapted
// by adding the ananymous import  of the ppi/config package.
func RegisterCommandConfigHandler(h CommandConfigHandler) {
	handler = h
}
