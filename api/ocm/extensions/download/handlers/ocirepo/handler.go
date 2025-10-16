package ocirepo

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/oci/grammar"
	"ocm.software/ocm/api/oci/tools/transfer"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
)

////////////////////////////////////////////////////////////////////////////////

type handler struct {
	spec *ociuploadattr.Attribute
}

func New(repospec ...*ociuploadattr.Attribute) download.Handler {
	return &handler{spec: general.Optional(repospec...)}
}

func (h *handler) Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (accepted bool, target string, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagationf(&err, "upload to OCI registry")

	ctx := racc.GetOCMContext()
	m, err := racc.AccessMethod()
	if err != nil {
		return false, "", err
	}
	finalize.Close(m, "access method for download")

	mediaType := m.MimeType()

	if !artdesc.IsOCIMediaType(mediaType) || (!strings.HasSuffix(mediaType, "+tar") && !strings.HasSuffix(mediaType, "+tar+gzip")) {
		return false, "", nil
	}

	log := download.Logger(ctx).WithName("ocireg")

	var repo oci.Repository

	var tag string

	aspec := m.AccessSpec()
	namespace := racc.ReferenceHint()
	if l, ok := aspec.(*localblob.AccessSpec); namespace == "" && ok {
		namespace = l.ReferenceName
	}

	// get rid of digest
	i := strings.LastIndex(namespace, "@")
	if i >= 0 {
		namespace = namespace[:i] // remove digest
	}

	i = strings.LastIndex(namespace, ":")
	if i > 0 {
		tag = namespace[i:]
		tag = tag[1:] // remove colon
		namespace = namespace[:i]
	}

	ocictx := ctx.OCIContext()

	var artspec *oci.ArtSpec
	var prefix string
	var result oci.RefSpec

	if h.spec == nil {
		log.Debug("no config set")
		if path == "" {
			return false, "", fmt.Errorf("path required as target repo specification")
		}
		ref, err := oci.ParseRef(path)
		if err != nil {
			return true, "", err
		}
		result.UniformRepositorySpec = ref.UniformRepositorySpec
		repospec, err := ocictx.MapUniformRepositorySpec(&ref.UniformRepositorySpec)
		if err != nil {
			return true, "", err
		}
		repo, err = ocictx.RepositoryForSpec(repospec)
		if err != nil {
			return true, "", err
		}
		finalize.Close(repo, "repository for downloading OCI artifact")
		artspec = &ref.ArtSpec
	} else {
		log.Debug("evaluating config")
		if path != "" {
			artspec, err = oci.ParseArt(path)
			if err != nil {
				return true, "", err
			}
		}
		var us *oci.UniformRepositorySpec
		repo, us, prefix, err = h.spec.GetInfo(ctx)
		if err != nil {
			return true, "", err
		}
		result.UniformRepositorySpec = *us
	}

	if artspec != nil {
		log.Debug("using artifact spec", "spec", artspec.String())
		if artspec.IsDigested() {
			return true, "", fmt.Errorf("digest not possible for target")
		}

		if artspec.Repository != "" {
			namespace = artspec.Repository
		}
		if artspec.IsTagged() {
			tag = *artspec.Tag
		}
	}

	if prefix != "" && namespace != "" {
		namespace = prefix + grammar.RepositorySeparator + namespace
	}
	if tag == "" || tag == "latest" {
		tag = racc.Meta().GetVersion()
	}
	log.Debug("using final target", "namespace", namespace, "tag", tag)
	if namespace == "" {
		return true, "", fmt.Errorf("no OCI namespace")
	}

	var art oci.ArtifactAccess

	cand := m
	if local, ok := aspec.(*localblob.AccessSpec); ok {
		if local.GlobalAccess != nil {
			s, err := ctx.AccessSpecForSpec(local.GlobalAccess)
			if err == nil {
				_ = s
				// c, err := s.AccessMethod()  // TODO: try global access for direct artifact access
				// set cand to oci access method
			}
		}
	}
	if ocimeth, ok := accspeccpi.GetAccessMethodImplementation(cand).(ociartifact.AccessMethodImpl); ok {
		// prepare for optimized point to point implementation
		art, _, err = ocimeth.GetArtifact()
		if err != nil {
			return true, "", errors.Wrapf(err, "cannot access source artifact")
		}
		finalize.Close(art)
	}

	ns, err := repo.LookupNamespace(namespace)
	if err != nil {
		return true, "", err
	}
	finalize.Close(ns)

	if art == nil {
		log.Debug("using artifact set transfer mode")
		set, err := artifactset.OpenFromDataAccess(accessobj.ACC_READONLY, m.MimeType(), m)
		if err != nil {
			return true, "", errors.Wrapf(err, "opening resource blob as artifact set")
		}
		finalize.Close(set)
		art, err = set.GetArtifact(set.GetMain().String())
		if err != nil {
			return true, "", errors.Wrapf(err, "get artifact from blob")
		}
		finalize.Close(art)
	} else {
		log.Debug("using direct transfer mode")
	}

	p.Printf("uploading resource %s to %s[%s:%s]...\n", racc.Meta().GetName(), repo.GetSpecification().UniformRepositorySpec(), namespace, tag)
	err = transfer.TransferArtifact(art, ns, oci.AsTags(tag)...)
	if err != nil {
		return true, "", errors.Wrapf(err, "transfer artifact")
	}

	result.Repository = namespace
	result.Tag = &tag
	return true, result.String(), nil
}
