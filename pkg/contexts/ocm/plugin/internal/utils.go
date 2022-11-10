// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/registry"
)

type Named interface {
	GetName() string
}

type Element[K registry.Key[K]] interface {
	Named
	GetConstraints() []K
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
