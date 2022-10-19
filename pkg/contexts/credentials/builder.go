// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

func WithContext(ctx context.Context) core.Builder {
	return core.Builder{}.WithContext(ctx)
}

func WithConfigs(ctx config.Context) core.Builder {
	return core.Builder{}.WithConfig(ctx)
}

func WithRepositoyTypeScheme(scheme RepositoryTypeScheme) core.Builder {
	return core.Builder{}.WithRepositoyTypeScheme(scheme)
}

func WithStandardConumerMatchers(matchers core.IdentityMatcherRegistry) core.Builder {
	return core.Builder{}.WithStandardConumerMatchers(matchers)
}

func New(mode ...datacontext.BuilderMode) Context {
	return core.Builder{}.New(mode...)
}
