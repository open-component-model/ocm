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
	return NewContext(logging.DefaultContext())
}

func NewBufferedContext() (logging.Context, *bytes.Buffer) {
	var buf bytes.Buffer
	return logging.New(buflogr.NewWithBuffer(&buf)), &buf
}
