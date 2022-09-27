package logger

import (
	"github.com/bombsimon/logrusr/v3"
	"github.com/mandelsoft/logging"
	"github.com/sirupsen/logrus"
)

func NewDefaultLoggerContext() logging.Context {
	logrusLog := logrus.New()
	log := logrusr.New(logrusLog)

	log = log.WithName("ocm-logger")
	return logging.New(log)
}

type LoggingContextProvider interface {
	Logger(messageContext ...logging.MessageContext) logging.Logger
}
