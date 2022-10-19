// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package elemhdlr

type Option interface {
	Apply(handler *TypeHandler)
}

type forceEmpty struct {
	flag bool
}

func (o forceEmpty) Apply(handler *TypeHandler) {
	handler.forceEmpty = o.flag
}

func ForceEmpty(b bool) Option {
	return forceEmpty{b}
}
