package npmjs

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	artifact "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/utils"
)

func init() {
	cpi.RegisterBlobHandler(NewArtifactHandler(), cpi.ForMimeType("npmjs"), cpi.WithPrio(10))
}

type artifactHandler struct {
	spec *Attribute
}

func NewArtifactHandler(repospec ...*Attribute) cpi.BlobHandler {
	return &artifactHandler{utils.Optional(repospec...)}
}

func (b *artifactHandler) StoreBlob(blob cpi.BlobAccess, artType, hint string, global cpi.AccessSpec, ctx cpi.StorageContext) (cpi.AccessSpec, error) {
	attr := b.spec
	if attr == nil {
		return nil, nil
	}

	mediaType := blob.MimeType()
	if !strings.HasSuffix(mediaType, "+tar") && !strings.HasSuffix(mediaType, "+tar+gzip") {
		return nil, nil
	}

	repo, base, prefix, err := attr.GetInfo(ctx.GetContext())
	if err != nil {
		return nil, err
	}

	target, err := json.Marshal(repo.GetSpecification())
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal target specification")
	}
	values := []interface{}{
		"arttype", artType,
		"mediatype", mediaType,
		"hint", hint,
		"target", string(target),
	}
	if m, ok := blob.(blobaccess.AnnotatedBlobAccess[cpi.AccessMethod]); ok {
		// prepare for optimized point to point implementation
		cpi.BlobHandlerLogger(ctx.GetContext()).Debug("npmjs generic artifact handler with ocm access source",
			generics.AppendedSlice[any](values, "sourcetype", m.Source().AccessSpec().GetType())...,
		)
	} else {
		cpi.BlobHandlerLogger(ctx.GetContext()).Debug("npmjs generic artifact handler", values...)
	}

	var namespace oci.NamespaceAccess
	var version string
	var name string
	var tag string

	if hint == "" {
		name = path.Join(prefix, ctx.TargetComponentName())
	} else {
		i := strings.LastIndex(hint, ":")
		if i > 0 {
			version = hint[i:]
			name = path.Join(prefix, hint[:i])
			tag = version[1:] // remove colon
		} else {
			name = hint
		}
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

	ref := base.ComposeRef(namespace.GetNamespace() + version)
	return artifact.New(ref), nil
}
