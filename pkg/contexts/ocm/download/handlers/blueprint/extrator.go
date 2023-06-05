package blueprint

import (
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const BlUEPRINT_MIMETYPE = "application/vnd.gardener.landscaper.blueprint.layer.v1.tar"
const BLUEPRINT_MIMETYPE_LEGACY = "application/vnd.gardener.landscaper.blueprint.v1+tar+gzip"

func ExtractArchive(access accessio.DataAccess, path string, fs vfs.FileSystem) (rerr error) {
	rawReader, err := access.Reader()
	if err != nil {
		return err
	}
	defer errors.PropagateError(&rerr, rawReader.Close)
	reader, _, err := compression.AutoDecompress(rawReader)
	if err != nil {
		return err
	}
	defer errors.PropagateError(&rerr, reader.Close)
	err = fs.MkdirAll(path, 0o700)
	if err != nil {
		return err
	}

	pfs, err := projectionfs.New(fs, path)
	if err != nil {
		return err
	}
	err = tarutils.ExtractTarToFs(pfs, reader)
	if err != nil {
		return err
	}
	return nil
}

func ExtractArtifact(access accessio.DataAccess, path string, fs vfs.FileSystem) (rerr error) {
	rd, err := access.Reader()
	if err != nil {
		return err
	}
	defer errors.PropagateError(&rerr, rd.Close)

	set, err := artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(rd))
	if err != nil {
		return err
	}
	defer errors.PropagateError(&rerr, set.Close)

	art, err := set.GetArtifact(set.GetMain().String())
	if err != nil {
		return err
	}
	defer errors.PropagateError(&rerr, art.Close)

	desc := art.ManifestAccess().GetDescriptor().Layers[0]
	if desc.MediaType != BlUEPRINT_MIMETYPE && desc.MediaType != BLUEPRINT_MIMETYPE_LEGACY {
		return errors.Newf("MIME type is not %v or %v", BlUEPRINT_MIMETYPE, BLUEPRINT_MIMETYPE_LEGACY)
	}

	blob, err := art.GetBlob(desc.Digest)
	if err != nil {
		return err
	}
	err = ExtractArchive(blob, path, fs)
	if err != nil {
		return err
	}
	return nil
}
