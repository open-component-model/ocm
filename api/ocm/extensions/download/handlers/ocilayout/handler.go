package ocilayout

import (
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/logging"
	common "ocm.software/ocm/api/utils/misc"
)

const PRIORITY = 200

type Handler struct{}

func New() download.Handler {
	return &Handler{}
}

func (h *Handler) Download(p common.Printer, racc cpi.ResourceAccess, path string, fs vfs.FileSystem) (ok bool, _ string, err error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	// Step 1: Get access method to read resource content
	m, err := racc.AccessMethod()
	if err != nil {
		return false, "", err
	}
	finalize.Close(m)

	// Step 2: Check MIME type - only handle OCI artifacts (tar/tar+gzip)
	if !isOCIArtifact(m.MimeType()) {
		logging.Logger().Debug("skipping non-OCI artifact", "mime", m.MimeType())
		return false, "", nil
	}

	if path == "" {
		path = racc.Meta().GetName()
	}

	// Step 3: Open resource blob as artifact set (contains OCI image)
	src, err := artifactset.OpenFromDataAccess(accessobj.ACC_READONLY, m.MimeType(), m)
	if err != nil {
		return true, "", errors.Wrapf(err, "open artifact set")
	}
	finalize.Close(src)

	// Step 4: Get the main artifact from the set
	art, err := src.GetArtifact(src.GetMain().String())
	if err != nil {
		return true, "", errors.Wrapf(err, "get artifact")
	}
	finalize.Close(art)

	// Step 5: Create target directory with OCI format (index.json + oci-layout)
	target, err := artifactset.Create(accessobj.ACC_CREATE, path, 0o755,
		accessio.PathFileSystem(fs),
		accessobj.FormatDirectory,
		artifactset.StructureFormat(artifactset.FORMAT_OCI),
	)
	if err != nil {
		return true, "", errors.Wrapf(err, "create OCI layout")
	}

	// Step 6: Transfer all manifests and blobs to target with hybrid tagging:
	// - Original tags from source (e.g., "latest", "linux")
	// - Resource version (e.g., "1.0.0")
	tags := collectTags(src, racc.Meta().GetVersion())
	if err := artifactset.TransferArtifact(art, target, tags...); err != nil {
		err = errors.Join(err, target.Close())
		return true, "", errors.Wrapf(err, "transfer artifact")
	}

	if err := target.Close(); err != nil {
		return true, "", errors.Wrapf(err, "close target")
	}

	p.Printf("%s: downloaded to OCI layout\n", path)
	return true, path, nil
}

func isOCIArtifact(mime string) bool {
	return artdesc.IsOCIMediaType(mime) &&
		(strings.HasSuffix(mime, "+tar") || strings.HasSuffix(mime, "+tar+gzip"))
}

// collectTags returns a deduplicated list of tags combining:
// - Resource version FIRST (becomes org.opencontainers.image.ref.name for ORAS resolution)
// - Original tags from the source artifact set (preserves mutable refs like "latest")
func collectTags(src *artifactset.ArtifactSet, version string) []string {
	seen := make(map[string]struct{})
	var tags []string

	// Add resource version first - it becomes the primary tag (org.opencontainers.image.ref.name)
	if version != "" {
		seen[version] = struct{}{}
		tags = append(tags, version)
	}

	// Add original tags from source index annotations
	mainDigest := src.GetMain()
	for _, m := range src.GetIndex().Manifests {
		if m.Digest == mainDigest && m.Annotations != nil {
			if tagStr := artifactset.RetrieveTags(m.Annotations); tagStr != "" {
				for _, t := range strings.Split(tagStr, ",") {
					t = strings.TrimSpace(t)
					if _, ok := seen[t]; !ok && t != "" {
						seen[t] = struct{}{}
						tags = append(tags, t)
					}
				}
			}
		}
	}

	return tags
}
