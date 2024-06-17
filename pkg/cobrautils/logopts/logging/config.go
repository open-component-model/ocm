package logging

import (
	"runtime"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	logcfg "github.com/mandelsoft/logging/config"
	"github.com/mandelsoft/logging/logrusl"
	"github.com/mandelsoft/logging/logrusl/adapter"
	"github.com/mandelsoft/logging/logrusr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/sirupsen/logrus"
)

// LoggingConfiguration describes logging configuration for a slave executables like
// plugins.
type LoggingConfiguration struct {
	LogFileName string        `json:"logFileName"`
	LogConfig   logcfg.Config `json:"logConfig"`
	Json        bool          `json:"json,omitempty"`
}

func (c *LoggingConfiguration) Apply() error {
	if GlobalLogFileOverride == c.LogFileName {
		return nil
	}
	logctx := logging.DefaultContext()
	if c.LogFileName != "" {
		logfile, err := LogFileFor(c.LogFileName, osfs.OsFs)
		if err != nil {
			return errors.Wrapf(err, "cannot open log file %q", c.LogFileName)
		}
		ConfigureLogrusFor(logctx, false, logfile)
		GlobalLogFile = logfile
		GlobalLogFileOverride = c.LogFileName
	} else {
		// overwrite current log formatter in case of a logrus logger is
		// used as logging backend.
		var f logrus.Formatter = adapter.NewJSONFormatter()
		if !c.Json {
			f = adapter.NewTextFmtFormatter()
		}
		logrusr.SetFormatter(logging.UnwrapLogSink(logctx.GetSink()), f)
	}
	return nil
}

func ConfigureLogrusFor(logctx logging.Context, human bool, logfile *LogFile) {
	settings := logrusl.Adapter().WithWriter(logfile.File())
	if human {
		settings = settings.Human()
	} else {
		settings = settings.WithFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	}

	log := settings.NewLogrus()
	logctx.SetBaseLogger(logrusr.New(log))
	runtime.SetFinalizer(log, func(_ *logrus.Logger) { logfile.Close() })
}
