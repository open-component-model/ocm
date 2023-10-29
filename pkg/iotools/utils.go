// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package iotools

type NopCloser struct{}

type _nopCloser = NopCloser

func (NopCloser) Close() error {
	return nil
}
