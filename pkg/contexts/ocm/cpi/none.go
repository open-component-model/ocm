// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

// Type is the access type for no blob.
const (
	NoneType       = "none"
	NoneLegacyType = "None"
)

func IsNoneAccess(kind string) bool {
	return kind == NoneType || kind == NoneLegacyType
}
