// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package panics

import (
	"fmt"
	"runtime"

	"github.com/open-component-model/ocm/pkg/logging"
)

var ReallyCrash = true

type PanicHandler func(interface{})

var PanicHandlers = []PanicHandler{logHandler}

func HandlePanic(additionalHandlers ...PanicHandler) {
	if r := recover(); r != nil {
		for _, fn := range PanicHandlers {
			fn(r)
		}

		for _, fn := range additionalHandlers {
			fn(r)
		}

		if ReallyCrash {
			panic(r)
		}
	}
}

func RegisterPanicHandler(handler PanicHandler) {
	PanicHandlers = append(PanicHandlers, handler)
}

func logHandler(r interface{}) {
	// Same as stdlib http server code. Manually allocate stack trace buffer size
	// to prevent excessively large logs
	const size = 64 << 10
	stacktrace := make([]byte, size)
	stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
	if _, ok := r.(string); ok {
		logging.Logger().Error(fmt.Sprintf("Observed a panic: %#v\n%s", r, stacktrace))
	} else {
		logging.Logger().Error(fmt.Sprintf("Observed a panic: %#v (%v)\n%s", r, r, stacktrace))
	}
}
