package ocirepo

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"
	metav1 "ocm.software/ocm/api/ocm/refhints"
	oci2 "ocm.software/ocm/api/tech/oci"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

func init() {
	for _, mime := range artdesc.ArchiveBlobTypes() {
		cpi.RegisterBlobHandler(NewArtifactHandler(), cpi.ForMimeType(mime), cpi.WithPrio(10))
	}
}

////////////////////////////////////////////////////////////////////////////////

// artifactHandler stores artifact blobs as OCIArtifacts regardless of the
// intended OCM target repository.
// It acts on the OCI upload attribute to determine the target OCI repository.
// If none is configured, it does nothing.
type artifactHandler struct {
	spec *ociuploadattr.Attribute
}

func NewArtifactHandler(repospec ...*ociuploadattr.Attribute) cpi.BlobHandler {
	return &artifactHandler{utils.Optional(repospec...)}
}

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, artType string, hints metav1.ReferenceHints, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	attr := b.spec
	if attr == nil {
		attr = ociuploadattr.Get(ctx.GetContext())
	}
	if attr == nil {
		return nil, nil
	}

	mediaType := blob.MimeType()
	if !artdesc.IsOCIMediaType(mediaType) || (!strings.HasSuffix(mediaType, "+tar") && !strings.HasSuffix(mediaType, "+tar+gzip")) {
		return nil, nil
	}

	repo, base, prefix, err := attr.GetInfo(ctx.GetContext())
	if err != nil {
		return nil, err
	}

	// this section is for logging, only
	target, err := json.Marshal(repo.GetSpecification())
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal target specification")
	}

	var hint string
	if h := hints.GetReferenceHint(oci2.ReferenceHintType, ""); h != nil {
		hint = h.GetReference()
	}

	values := []interface{}{
		"arttype", artType,
		"mediatype", mediaType,
		"hint", hint,
		"target", string(target),
	}
	if m, ok := blob.(blobaccess.AnnotatedBlobAccess[cpi.AccessMethod]); ok {
		// prepare for optimized point to point implementation
		cpi.BlobHandlerLogger(ctx.GetContext()).Debug("oci generic artifact handler with ocm access source",
			sliceutils.CopyAppend[any](values, "sourcetype", m.Source().AccessSpec().GetType())...,
		)
	} else {
		cpi.BlobHandlerLogger(ctx.GetContext()).Debug("oci generic artifact handler", values...)
	}

	var namespace oci.NamespaceAccess
	var version string
	var name string
	var tag string

	if hint == "" {
		name = path.Join(prefix, ctx.TargetComponentName())
	} else {
		i := strings.LastIndex(hint, "@")
		if i > 0 {
			version = hint[i:]
			name = hint[:i]
		} else {
			name = hint
		}
		i = strings.LastIndex(name, ":")
		if i > 0 {
			tag = name[i+1:]
			name = name[:i]
		}
		name = path.Join(prefix, name)
	}

	namespace, err = repo.LookupNamespace(name)
	if err != nil {
		return nil, errors.Wrapf(err, "lookup namespace %s in target repository %s", name, attr.Ref)
	}
	defer namespace.Close()

	set, err := artifactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
	if err != nil {
		return nil, err
	}
	defer set.Close()
	digest := set.GetMain()
	if version == "" {
		version = "@" + digest.String()
	} else {
		if version != "@"+digest.String() {
			return nil, fmt.Errorf("corrupted digest: hint requests %q, but found %q", version[1:], digest.String())
		}
	}
	art, err := set.GetArtifact(digest.String())
	if err != nil {
		return nil, err
	}
	defer art.Close()

	err = artifactset.TransferArtifact(art, namespace, oci.AsTags(tag)...)
	if err != nil {
		return nil, err
	}
	if tag != "" {
		tag = ":" + tag
	}
	ref := base.ComposeRef(namespace.GetNamespace() + tag + version)
	return ociartifact.New(ref), nil
}
