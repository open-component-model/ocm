// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/mandelsoft/logging"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

var TAG = logging.NewTag("transfer")

type ContextProvider interface {
	GetContext() ocm.Context
}

func LogInfo(c ContextProvider, msg string, keyValuePairs ...interface{}) {
	c.GetContext().Logger(TAG).Info(msg, keyValuePairs...)
}

func Logger(c ContextProvider, keyValuePairs ...interface{}) logging.Logger {
	return c.GetContext().Logger(TAG).WithValues(keyValuePairs...)
}
