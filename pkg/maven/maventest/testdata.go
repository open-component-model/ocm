// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maventest

import (
	"github.com/open-component-model/ocm/pkg/env"
)

func TestData(dest ...string) env.Option {
	return env.ProjectTestDataForCaller(dest...)
}
