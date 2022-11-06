// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

type Named interface {
	GetName() string
}

type List[T Named] []T

func (l List[T]) Get(name string) *T {
	for _, m := range l {
		if m.GetName() == name {
			return &m
		}
	}
	return nil
}
