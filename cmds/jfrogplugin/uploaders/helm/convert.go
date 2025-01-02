package helm

import (
	"errors"
	"fmt"
	"io"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

func ConvertArtifactSetHelmChartToPlainTGZChart(reader io.Reader) (_ io.ReadCloser, _ string, err error) {
	set, err := artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(io.NopCloser(reader)))
	if err != nil {
		return nil, "", fmt.Errorf("failed to open helm OCI artifact as artifact set: %w", err)
	}
	defer func() {
		err = errors.Join(err, set.Close())
	}()

	art, err := set.GetArtifact(set.GetMain().String())
	if err != nil {
		return nil, "", fmt.Errorf("failed to get artifact from set: %w", err)
	}
	defer func() {
		err = errors.Join(err, art.Close())
	}()

	chartTgz, provenance, err := accessSingleLayerOCIHelmChart(art)
	if err != nil {
		return nil, "", fmt.Errorf("failed to access OCI artifact as a single layer helm OCI image: %w", err)
	}
	defer func() {
		err = errors.Join(err, chartTgz.Close())
		if provenance != nil {
			err = errors.Join(err, provenance.Close())
		}
	}()

	chartReader, err := chartTgz.Reader()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get reader for chart tgz: %w", err)
	}

	digest := chartTgz.Digest().String()

	return chartReader, digest, nil
}

func accessSingleLayerOCIHelmChart(art oci.ArtifactAccess) (chart oci.BlobAccess, prov oci.BlobAccess, err error) {
	m := art.ManifestAccess()
	if m == nil {
		return nil, nil, errors.New("artifact is no image manifest")
	}
	if len(m.GetDescriptor().Layers) < 1 {
		return nil, nil, errors.New("no layers found")
	}

	if chart, err = m.GetBlob(m.GetDescriptor().Layers[0].Digest); err != nil {
		return nil, nil, err
	}

	if len(m.GetDescriptor().Layers) > 1 {
		if prov, err = m.GetBlob(m.GetDescriptor().Layers[1].Digest); err != nil {
			return nil, nil, err
		}
	}

	return chart, prov, nil
}
