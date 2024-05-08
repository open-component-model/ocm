package config

import (
	"github.com/open-component-model/ocm/pkg/contexts/config/internal"
)

var Realm = internal.Realm

var Logger = internal.Logger

func Debug(c Context, msg string, keypairs ...interface{}) {
	c.LoggingContext().Logger(Realm).Debug(msg, append(keypairs, "id", c.GetId())...)
}

func Info(c Context, msg string, keypairs ...interface{}) {
	c.LoggingContext().Logger(Realm).Info(msg, append(keypairs, "id", c.GetId())...)
}
