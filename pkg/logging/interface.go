package logging

import (
	"bytes"

	"github.com/mandelsoft/logging"
	"github.com/tonglil/buflogr"
)

type LogProvider interface {
	logging.ContextProvider
	Logger(messageContext ...logging.MessageContext) logging.Logger
}

func NewDefaultContext() logging.Context {
	return logging.NewWithBase(logging.DefaultContext())
}

func NewBufferedContext() (logging.Context, *bytes.Buffer) {
	var buf bytes.Buffer
	return logging.New(buflogr.NewWithBuffer(&buf)), &buf
}

var liblogcontext = logging.NewWithBase(logging.DefaultContext())

// DefaultContext is the local log context used all over the ocm library.
func DefaultContext() logging.Context {
	return liblogcontext
}

// Logger provides a logger for the logging context of the ocm library.
func Logger(mctx ...logging.MessageContext) logging.Logger {
	return liblogcontext.Logger(mctx...)
}
