// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/core"
)

const (
	KIND_OCIARTEFACT = core.KIND_OCIARTEFACT
	KIND_MEDIATYPE   = accessio.KIND_MEDIATYPE
	KIND_BLOB        = accessio.KIND_BLOB
)

const CONTEXT_TYPE = core.CONTEXT_TYPE

const CommonTransportFormat = core.CommonTransportFormat

type (
	Context                          = core.Context
	Repository                       = core.Repository
	RepositorySpecHandlers           = core.RepositorySpecHandlers
	RepositorySpecHandler            = core.RepositorySpecHandler
	UniformRepositorySpec            = core.UniformRepositorySpec
	RepositoryTypeScheme             = core.RepositoryTypeScheme
	RepositorySpec                   = core.RepositorySpec
	IntermediateRepositorySpecAspect = core.IntermediateRepositorySpecAspect
	GenericRepositorySpec            = core.GenericRepositorySpec
	ArtefactAccess                   = core.ArtefactAccess
	NamespaceLister                  = core.NamespaceLister
	NamespaceAccess                  = core.NamespaceAccess
	ManifestAccess                   = core.ManifestAccess
	IndexAccess                      = core.IndexAccess
	BlobAccess                       = core.BlobAccess
	DataAccess                       = core.DataAccess
)

func DefaultContext() core.Context {
	return core.DefaultContext
}

func ForContext(ctx context.Context) Context {
	return core.ForContext(ctx)
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	return core.DefinedForContext(ctx)
}

func IsErrBlobNotFound(err error) bool {
	return accessio.IsErrBlobNotFound(err)
}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	return core.ToGenericRepositorySpec(spec)
}
