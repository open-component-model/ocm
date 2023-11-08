// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package githubaccess

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactaccess/genericaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

const TYPE = resourcetypes.DIRECTORY_TREE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repo string, commit string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(repo, eff.APIHostName, commit)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, repo string, commit string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repo, commit, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, repo string, commit string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repo, commit, opts...)
}
