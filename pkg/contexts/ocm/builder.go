// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
)

func WithContext(ctx context.Context) core.Builder {
	return core.Builder{}.WithContext(ctx)
}

func WithCredentials(ctx credentials.Context) core.Builder {
	return core.Builder{}.WithCredentials(ctx)
}

func WithOCIRepositories(ctx oci.Context) core.Builder {
	return core.Builder{}.WithOCIRepositories(ctx)
}

func WithRepositoyTypeScheme(scheme RepositoryTypeScheme) core.Builder {
	return core.Builder{}.WithRepositoyTypeScheme(scheme)
}

func WithAccessypeScheme(scheme AccessTypeScheme) core.Builder {
	return core.Builder{}.WithAccessTypeScheme(scheme)
}

func WithRepositorySpecHandlers(reg RepositorySpecHandlers) core.Builder {
	return core.Builder{}.WithRepositorySpecHandlers(reg)
}

func WithBlobHandlers(reg BlobHandlerRegistry) core.Builder {
	return core.Builder{}.WithBlobHandlers(reg)
}

func WithBlobDigesters(reg BlobDigesterRegistry) core.Builder {
	return core.Builder{}.WithBlobDigesters(reg)
}

func New(mode ...datacontext.BuilderMode) Context {
	return core.Builder{}.New(mode...)
}
