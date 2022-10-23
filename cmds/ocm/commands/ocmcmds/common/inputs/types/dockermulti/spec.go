// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dockermulti

import (
	"fmt"

	"github.com/opencontainers/go-digest"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociimage"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/docker"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Spec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Repository is the repository hint for the index artefact
	Repository string `json:"repository"`
	// Variants holds the list of repository path and tag of the images in the docker daemon
	// used to compose a multi-arch image.
	Variants []string `json:"variants"`
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(pathtags ...string) *Spec {
	return &Spec{
		ObjectVersionedType: runtime.ObjectVersionedType{
			Type: TYPE,
		},
		Variants: pathtags,
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = ociimage.ValidateRepository(fldPath.Child("repository"), allErrs, s.Repository)
	variantsField := fldPath.Child("variants")
	if len(s.Variants) == 0 {
		allErrs = append(allErrs, field.Required(variantsField, fmt.Sprintf("variants is required for input of type %q and must has at least one entry", s.GetType())))
	}
	for i, variant := range s.Variants {
		variantField := fldPath.Index(i)
		if variant == "" {
			allErrs = append(allErrs, field.Required(variantField, fmt.Sprintf("non-empty image name is required input of type %q", s.GetType())))
		} else {
			_, _, err := docker.ParseGenericRef(variant)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(variantField, variant, err.Error()))
			}
		}
	}
	return allErrs
}

func (s *Spec) getVariant(ctx clictx.Context, finalize *utils.Finalizer, variant string) (oci.ArtefactAccess, error) {
	locator, version, err := docker.ParseGenericRef(variant)
	if err != nil {
		return nil, err
	}
	if version == "" {
		return nil, fmt.Errorf("artefact version required")
	}
	spec := docker.NewRepositorySpec()
	repo, err := ctx.OCIContext().RepositoryForSpec(spec)
	if err != nil {
		return nil, err
	}
	finalize.Close(repo)
	ns, err := repo.LookupNamespace(locator)
	if err != nil {
		return nil, err
	}
	finalize.Close(ns)

	art, err := ns.GetArtefact(version)
	if err != nil {
		return nil, artefactset.GetArtifactError{Original: err, Ref: version}
	}
	finalize.Close(art)
	return art, nil
}

func (s *Spec) GetBlob(ctx inputs.Context, nv common.NameVersion, inputFilePath string) (accessio.TemporaryBlobAccess, string, error) {
	index := artdesc.NewIndexArtefact()
	i := 0

	feedback := func(blob accessio.BlobAccess, art cpi.ArtefactAccess) error {
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
		index.Index().AddManifest(desc)
		return nil
	}

	blob, err := artefactset.SynthesizeArtefactBlobFor(nv.GetVersion(), func() (fac artefactset.ArtefactFactory, main bool, err error) {
		var art cpi.ArtefactAccess
		var blob accessio.BlobAccess

		switch {
		case i > len(s.Variants):
			// end loop
		case i == len(s.Variants):
			// provide index (main) artefact
			ctx.Printf("image %d: INDEX\n", i)
			fac = func(set *artefactset.ArtefactSet) (digest.Digest, string, error) {
				art, err = set.NewArtefact(index)
				if err != nil {
					return "", "", errors.Wrapf(err, "cannot create index artefact")
				}
				defer art.Close()
				blob, err = set.AddArtefact(art)
				if err != nil {
					return "", "", errors.Wrapf(err, "cannot add index artefact")
				}
				defer blob.Close()
				return blob.Digest(), blob.MimeType(), nil
			}
			main = true
		default:
			// provide variant
			ctx.Printf("image %d: %s\n", i, s.Variants[i])
			var finalize utils.Finalizer

			art, err = s.getVariant(ctx, &finalize, s.Variants[i])

			if err == nil {
				blob, err = art.Blob()
				if err == nil {
					finalize.Close(art)
					fac = artefactset.ArtefactTransferCreator(art, &finalize, feedback)
				}
			}
		}
		i++
		return
	})
	if err != nil {
		return nil, "", err
	}
	return blob, ociimage.Hint(nv, nv.GetName(), s.Repository, nv.GetVersion()), nil
}
