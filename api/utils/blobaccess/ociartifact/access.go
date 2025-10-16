package ociartifact

import (
	"fmt"

	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

func BlobAccess(refname string, opts ...Option) (bpi.BlobAccess, string, error) {
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

func Provider(name string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, _, err := BlobAccess(name, opts...)
		return b, err
	})
}
