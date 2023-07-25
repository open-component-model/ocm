// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package toi

import (
	logging2 "github.com/open-component-model/ocm/v2/pkg/logging"
)

var REALM = logging2.DefineSubRealm("TOI logging", "toi")

var Log = logging2.DynamicLogger(REALM)
