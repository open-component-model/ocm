// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/externalartifacts/generic"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repository string, digest digest.Digest, size int64, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	media := eff.MediaType
	if media == "" {
		media = mime.MIME_OCTET
	}
	spec := access.New(repository, digest, media, size)
	// is global access, must work, otherwise there is an error in the lib.
	return generic.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, repository string, digest digest.Digest, size int64, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repository, digest, size, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, repository string, digest digest.Digest, size int64, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repository, digest, size, opts...)
}
