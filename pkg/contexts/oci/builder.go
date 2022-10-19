// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci/core"
)

func WithContext(ctx context.Context) core.Builder {
	return core.Builder{}.WithContext(ctx)
}

func WithCredentials(ctx credentials.Context) core.Builder {
	return core.Builder{}.WithCredentials(ctx)
}

func WithRepositoyTypeScheme(scheme RepositoryTypeScheme) core.Builder {
	return core.Builder{}.WithRepositoyTypeScheme(scheme)
}

func WithRepositorySpecHandlers(reg RepositorySpecHandlers) core.Builder {
	return core.Builder{}.WithRepositorySpecHandlers(reg)
}

func New(mode ...datacontext.BuilderMode) Context {
	return core.Builder{}.New(mode...)
}
