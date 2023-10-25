// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package optionutils

type Option[T any] interface {
	ApplyTo(T)
}

func EvalOptions[O any](opts ...Option[*O]) *O {
	var eff O
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	return &eff
}
