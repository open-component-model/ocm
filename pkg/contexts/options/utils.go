// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"reflect"

	"github.com/open-component-model/ocm/pkg/generics"
)

func FilterOptions[T any, O any](opts []O) []T {
	var found []T

	t := generics.TypeOf[T]()
	for _, o := range opts {
		if reflect.TypeOf(o).AssignableTo(t) {
			found = append(found, generics.As[T](o))
		}
	}
	return found
}
