package cpi

import (
	"github.com/mandelsoft/logging"
)

type OCMContextProvider interface {
	GetContext() Context
}

func Logger(c OCMContextProvider, keyValuePairs ...interface{}) logging.Logger {
	return c.GetContext().Logger().WithValues(keyValuePairs...)
}
