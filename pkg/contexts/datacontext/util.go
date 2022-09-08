// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
