package ocilayout

import (
	"path/filepath"
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

	// Step 6: Transfer all manifests and blobs to target with resource version as tag
	version := racc.Meta().GetVersion()
	if err := artifactset.TransferArtifact(art, target, version); err != nil {
		target.Close()
		return true, "", errors.Wrapf(err, "transfer artifact")
	}

	if err := target.Close(); err != nil {
		return true, "", errors.Wrapf(err, "close target")
	}

	// Step 7: Convert blob paths from sha256.DIGEST to sha256/DIGEST
	if err := convertBlobPaths(fs, path); err != nil {
		return true, "", err
	}

	p.Printf("%s: downloaded to OCI layout\n", path)
	return true, path, nil
}

func isOCIArtifact(mime string) bool {
	return artdesc.IsOCIMediaType(mime) &&
		(strings.HasSuffix(mime, "+tar") || strings.HasSuffix(mime, "+tar+gzip"))
}

// convertBlobPaths converts blob paths from artifactset format (sha256.DIGEST)
// to OCI Image Layout format (sha256/DIGEST).
//
// This is needed because artifactset uses DigestToFileName which always produces
// "sha256.DIGEST" format. The FORMAT_OCI option only controls the descriptor file
// (index.json) and oci-layout file creation, not the blob path structure.
//
// Call trace where DigestToFileName is invoked:
//
//	TransferArtifact() -> TransferManifest() -> set.AddBlob()
//	  -> accessobj.FileSystemBlobAccess.AddBlob()
//	    -> path := a.DigestPath(blob.Digest())
//	      -> common.DigestToFileName(digest)  // returns "sha256.DIGEST"
func convertBlobPaths(fs vfs.FileSystem, dir string) error {
	blobsDir := filepath.Join(dir, "blobs")
	entries, err := vfs.ReadDir(fs, blobsDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if algo, dig, ok := strings.Cut(name, "."); ok {
			algoDir := filepath.Join(blobsDir, algo)
			fs.MkdirAll(algoDir, 0o755)
			fs.Rename(filepath.Join(blobsDir, name), filepath.Join(algoDir, dig))
		}
	}
	return nil
}
