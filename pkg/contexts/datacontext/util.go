// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package datacontext

import (
	"context"
)

// ForContextByKey retrieves the context for a given key to be used for a context.Context.
// If not defined, it returns the given default context and false.
func ForContextByKey(ctx context.Context, key interface{}, def Context) (Context, bool) {
	c := ctx.Value(key)
	if c == nil {
		return def, false
	}
	return c.(Context), true
}
