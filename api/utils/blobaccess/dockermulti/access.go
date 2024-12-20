package dockermulti

import (
	"fmt"

	. "github.com/mandelsoft/goutils/finalizer"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/annotations"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/oci/extensions/repositories/docker"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
)

func (o *Options) OCIContext() oci.Context {
	if o.Context == nil {
		return oci.DefaultContext()
	}
	return o.Context
}

func (s *Options) getVariant(ctx oci.Context, finalize *Finalizer, variant string) (oci.ArtifactAccess, error) {
	locator, version, err := docker.ParseGenericRef(variant)
	if err != nil {
		return nil, err
	}
	if version == "" {
		return nil, fmt.Errorf("artifact version required")
	}
	spec := docker.NewRepositorySpec()
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return nil, err
	}
	finalize.Close(repo)
	ns, err := repo.LookupNamespace(locator)
	if err != nil {
		return nil, err
	}
	finalize.Close(ns)

	art, err := ns.GetArtifact(version)
	if err != nil {
		return nil, artifactset.GetArtifactError{Original: err, Ref: locator + ":" + version}
	}
	finalize.Close(art)
	return art, nil
}

func BlobAccess(opts ...Option) (bpi.BlobAccess, error) {
	eff := optionutils.EvalOptions(opts...)
	ctx := eff.OCIContext()

	index := artdesc.NewIndex()
	i := 0

	version := eff.Version
	if eff.Origin != nil {
		if version == "" {
			version = eff.Origin.GetVersion()
		}
		index.SetAnnotation(annotations.COMPVERS_ANNOTATION, eff.Origin.String())
	}
	if version == "" {
		return nil, fmt.Errorf("no version specified")
	}

	feedback := func(blob bpi.BlobAccess, art cpi.ArtifactAccess) error {
		desc := artdesc.DefaultBlobDescriptor(blob)
		if art.IsManifest() {
			cfgBlob, err := art.ManifestAccess().GetConfigBlob()
			if err != nil {
				return errors.Wrapf(err, "cannot get config blob")
			}
			cfg, err := artdesc.ParseImageConfig(cfgBlob)
			if err != nil {
				return errors.Wrapf(err, "cannot parse config blob")
			}
			if cfg.Architecture != "" {
				desc.Platform = &artdesc.Platform{
					Architecture: cfg.Architecture,
					OS:           cfg.OS,
					Variant:      cfg.Variant,
				}
			}
		}
		index.AddManifest(desc)
		return nil
	}

	blob, err := artifactset.SynthesizeArtifactBlobFor(version, func() (fac artifactset.ArtifactFactory, main bool, err error) {
		var art cpi.ArtifactAccess
		var blob bpi.BlobAccess

		switch {
		case i > len(eff.Variants):
			// end loop
		case i == len(eff.Variants):
			// provide index (main) artifact
			if eff.Printer != nil {
				eff.Printer.Printf("image %d: INDEX\n", i)
			}
			fac = func(set *artifactset.ArtifactSet) (digest.Digest, string, error) {
				art, err = set.NewArtifact(index)
				if err != nil {
					return "", "", errors.Wrapf(err, "cannot create index artifact")
				}
				defer art.Close()
				blob, err = set.AddArtifact(art)
				if err != nil {
					return "", "", errors.Wrapf(err, "cannot add index artifact")
				}
				defer blob.Close()
				return blob.Digest(), blob.MimeType(), nil
			}
			main = true
		default:
			// provide variant
			if eff.Printer != nil {
				eff.Printer.Printf("image %d: %s\n", i, eff.Variants[i])
			}
			var finalize Finalizer

			art, err = eff.getVariant(ctx, &finalize, eff.Variants[i])

			if err == nil {
				if eff.Origin != nil {
					art.Artifact().SetAnnotation(annotations.COMPVERS_ANNOTATION, eff.Origin.String())
				}
				blob, err = art.Blob()
				if err == nil {
					finalize.Close(art)
					fac = artifactset.ArtifactTransferCreator(art, &finalize, feedback)
				}
			}
		}
		i++
		return
	})
	if err != nil {
		return nil, err
	}
	return blob, nil
}

func Provider(opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		return BlobAccess(opts...)
	})
}
