package panics

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/utils/logging"
)

var ReallyCrash = true

type PanicHandler func(interface{}) error

var PanicHandlers = []PanicHandler{logHandler}

// PropagatePanicAsError is intended to be called via defer to
// map a recovered panic to an error and propagate it to the
// error return value of the calling function.
// Use: defer panics.PropagatePanicAsError(&err, false)
// It incorporates the originally returned error by the surrounding function
// and all errors provided by the panic handlers plus the mapped recovered panic.
func PropagatePanicAsError(errp *error, callStandardHandlers bool, additionalHandlers ...PanicHandler) {
	if r := recover(); r != nil {
		list := errors.ErrList().Add(mapRecovered(r))

		if callStandardHandlers {
			// gather errors from standard handler
			for _, fn := range PanicHandlers {
				list.Add(fn(r))
			}
		}

		// add errors from explicit handlers
		for _, fn := range additionalHandlers {
			list.Add(fn(r))
		}

		*errp = list.Result()
	}
}

func HandlePanic(additionalHandlers ...PanicHandler) {
	if r := recover(); r != nil {
		for _, fn := range PanicHandlers {
			_ = fn(r)
		}

		for _, fn := range additionalHandlers {
			_ = fn(r)
		}

		if ReallyCrash {
			panic(r)
		}
	}
}

// RegisterPanicHandler adds handlers to the panic handler.
func RegisterPanicHandler(handler PanicHandler) {
	PanicHandlers = append(PanicHandlers, handler)
}

func mapRecovered(r interface{}) error {
	if err, ok := r.(error); ok {
		return err
	} else {
		// Same as stdlib http server code. Manually allocate stack trace buffer size
		// to prevent excessively large logs
		const size = 64 << 10
		stacktrace := make([]byte, size)
		stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]

		stack := string(stacktrace)
		lines := strings.Split(stack, "\n")
		offset := 1
		for offset < len(lines) && !strings.HasPrefix(lines[offset], "panic(") {
			offset++
		}
		if offset < len(lines) {
			stack = strings.Join(append(lines[:1], lines[offset:]...), "\n")
		}
		if _, ok := r.(string); ok {
			return fmt.Errorf("Observed a panic: %#v\n%s", r, stack)
		}
		return fmt.Errorf("Observed a panic: %#v (%v)\n%s", r, r, stack)
	}
}

func logHandler(r interface{}) error {
	logging.Logger().Error(mapRecovered(r).Error())
	return nil
}
