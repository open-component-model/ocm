package dockerdaemon

import (
	"fmt"

	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/oci/annotations"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/oci/extensions/repositories/docker"
	cpi "ocm.software/ocm/api/oci/types"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

func (o *Options) OCIContext() cpi.Context {
	if o.Context == nil {
		return cpi.DefaultContext()
	}
	return o.Context
}

func ImageInfoFor(name string, opts ...Option) (locator string, version string, err error) {
	eff := optionutils.EvalOptions(opts...)

	locator, version, err = docker.ParseGenericRef(name)
	if err != nil {
		return "", "", err
	}

	if version == "" || version == "latest" || optionutils.AsValue(eff.OverrideVersion) {
		version = eff.Version
	}
	if version == "" {
		return "", "", fmt.Errorf("no version specified")
	}
	return locator, version, nil
}

func Provider(name string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, _, err := BlobAccess(name, opts...)
		return b, err
	})
}

// BlobAccess returns a BlobAccess for the image with the given name.
func BlobAccess(name string, opts ...Option) (bpi.BlobAccess, string, error) {
	eff := optionutils.EvalOptions(opts...)
	ctx := eff.OCIContext()

	locator, version, err := ImageInfoFor(name, eff)
	if err != nil {
		return nil, "", err
	}
	spec := docker.NewRepositorySpec()
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return nil, "", err
	}
	ns, err := repo.LookupNamespace(locator)
	if err != nil {
		return nil, "", err
	}
	blob, err := artifactset.SynthesizeArtifactBlob(ns, version,
		func(art cpi.ArtifactAccess) error {
			if eff.Origin != nil {
				art.Artifact().SetAnnotation(annotations.COMPVERS_ANNOTATION, eff.Origin.String())
			}
			return nil
		},
	)
	if err != nil {
		return nil, "", err
	}
	return blob, version, nil
}
