package testhelper

import (
	"github.com/mandelsoft/logging"
)

func LogTest(ctx logging.ContextProvider, prefix ...string) {
	LoggerTest(ctx.LoggingContext().Logger(), prefix...)
}

func LoggerTest(logger logging.Logger, prefix ...string) {
	p := ""
	for _, e := range prefix {
		p += e
	}
	logger.Trace(p + "trace")
	logger.Debug(p + "debug")
	logger.Info(p + "info")
	logger.Warn(p + "warn")
	logger.Error(p + "error")
}
