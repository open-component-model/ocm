// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package optionutils

import (
	"github.com/open-component-model/ocm/pkg/utils"
)

func PointerTo[T any](v T) *T {
	temp := v
	return &temp
}

func AsValue[T any](p *T) T {
	var r T
	if p != nil {
		r = *p
	}
	return r
}

func AsBool(b *bool, def ...bool) bool {
	return utils.AsBool(b, def...)
}

func ApplyOption[T any](opt *T, tgt **T) {
	if opt != nil {
		*tgt = opt
	}
}
