// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package spi

type baseAccess interface {
	base() BlobAccessBase
}

func Cast[I interface{}](acc BlobAccess) I {
	var _nil I

	var b BlobAccessBase = acc

	for b != nil {
		if i, ok := b.(I); ok {
			return i
		}
		if i, ok := b.(baseAccess); ok {
			b = i.base()
		} else {
			b = nil
		}
	}
	return _nil
}
