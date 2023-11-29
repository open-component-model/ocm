// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
)

func obfuscate(creds credentials.Credentials) string {
	if creds == nil {
		return "no credentials"
	}
	props := creds.Properties()
	if pw, ok := props[credentials.ATTR_PASSWORD]; ok {
		if len(pw) > 5 {
			pw = pw[:5] + "***"
		} else {
			pw = "***"
		}
		props = props.Copy()
		props[credentials.ATTR_PASSWORD] = pw
	}
	return props.String()
}
