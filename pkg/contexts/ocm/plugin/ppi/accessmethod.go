// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"github.com/open-component-model/ocm/pkg/runtime"
)

type AccessSpec runtime.VersionedTypedObject

type AccessSpecProvider func() AccessSpec
