// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blob

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/finalizer"
)

func Download(p common.Printer, ctx oci.Context, ref string, path string, fs vfs.FileSystem) (err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagationf(&err, "downloading helm chart %q", ref)

	r, err := oci.ParseRef(ref)
	if err != nil {
		return err
	}

	spec, err := ctx.MapUniformRepositorySpec(&r.UniformRepositorySpec)
	if err != nil {
		return err
	}

	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return err
	}
	finalize.Close(repo)

	art, err := repo.LookupArtifact(r.Repository, r.Version())
	if err != nil {
		return err
	}
	finalize.Close(art)

	return download(p, art, path, fs)
}
