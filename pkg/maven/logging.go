// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maven

import "github.com/open-component-model/ocm/pkg/logging"

var REALM = logging.DefineSubRealm("Maven repository", "mvn")

var Log = logging.DynamicLogger(REALM)
