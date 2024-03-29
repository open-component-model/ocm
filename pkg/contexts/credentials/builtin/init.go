// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package builtin

import (
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/github"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/helm/identity"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/wget/identity"
)
