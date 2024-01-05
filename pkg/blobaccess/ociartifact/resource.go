// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ociartifact

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

func BlobAccessForOCIArtifact(refname string, opts ...Option) (blobaccess.BlobAccess, string, error) {
	eff := optionutils.EvalOptions(opts...)

	eff.Printf("image %s\n", refname)
	ref, err := oci.ParseRef(refname)
	if err != nil {
		return nil, "", err
	}

	spec, err := eff.OCIContext().MapUniformRepositorySpec(&ref.UniformRepositorySpec)
	if err != nil {
		return nil, "", err
	}

	repo, err := eff.OCIContext().RepositoryForSpec(spec)
	if err != nil {
		return nil, "", err
	}
	ns, err := repo.LookupNamespace(ref.Repository)
	if err != nil {
		return nil, "", err
	}

	version := ref.Version()
	if version == "" || version == "latest" {
		version = eff.Version
	}
	if version == "" {
		return nil, "", fmt.Errorf("no version specified")
	}
	blob, err := artifactset.SynthesizeArtifactBlobWithFilter(ns, version, eff.Filter)
	if err != nil {
		return nil, "", err
	}
	return blob, version, nil
}

func BlobAccessProviderForOCIArtifact(name string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, _, err := BlobAccessForOCIArtifact(name, opts...)
		return b, err
	})
}
