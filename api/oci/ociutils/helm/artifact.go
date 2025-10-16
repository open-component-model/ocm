package helm

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/registry"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/tech/helm/loader"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/file"
)

func SynthesizeArtifactBlob(loader loader.Loader) (artifactset.ArtifactBlob, error) {
	return artifactset.SythesizeArtifactSet(func(set *artifactset.ArtifactSet) (string, error) {
		chart, blob, err := TransferAsArtifact(loader, set)
		if err != nil {
			return "", fmt.Errorf("unable to transfer as artifact: %w", err)
		}

		if chart.Metadata.Version != "" {
			err = set.AddTags(blob.Digest, chart.Metadata.Version)
			if err != nil {
				return "", fmt.Errorf("unable to add tag: %w", err)
			}
		}

		set.Annotate(artifactset.MAINARTIFACT_ANNOTATION, blob.Digest.String())

		return artdesc.MediaTypeImageManifest, nil
	})
}

func TransferAsArtifact(loader loader.Loader, ns oci.NamespaceAccess) (*chart.Chart, *artdesc.Descriptor, error) {
	chart, err := loader.Chart()
	if err != nil {
		return nil, nil, err
	}
	err = chart.Validate()
	if err != nil {
		return nil, nil, errors.ErrInvalidWrap(err, "helm chart")
	}

	provData, err := loader.Provenance()
	if err != nil {
		return nil, nil, err
	}

	var blob blobaccess.BlobAccess
	blob, err = loader.ChartArchive()
	if err != nil {
		return nil, nil, err
	}
	if blob == nil {
		dir, err := os.MkdirTemp("", "helmchart-")
		if err != nil {
			return chart, nil, errors.Wrapf(err, "cannot create temporary directory for helm chart")
		}
		defer os.RemoveAll(dir)
		path, err := chartutil.Save(chart, dir)
		if err != nil {
			return chart, nil, err
		}
		blob = file.BlobAccess(registry.ChartLayerMediaType, path, osfs.New())
	} else {
		defer blob.Close()
	}
	meta := chart.Metadata

	configData, err := json.Marshal(meta)
	if err != nil {
		return chart, nil, err
	}

	art, err := ns.NewArtifact()
	if err != nil {
		return chart, nil, err
	}
	defer art.Close()
	m := art.ManifestAccess()

	err = m.SetConfigBlob(blobaccess.ForData(registry.ConfigMediaType, configData), nil)
	if err != nil {
		return chart, nil, err
	}
	_, err = m.AddLayer(blob, nil)
	if err != nil {
		return chart, nil, err
	}
	if provData != nil {
		_, err = m.AddLayer(blobaccess.ForData(registry.ProvLayerMediaType, provData), nil)
		if err != nil {
			return chart, nil, err
		}
	}
	blob, err = ns.AddArtifact(art)
	if err != nil {
		return chart, nil, err
	}
	return chart, artdesc.DefaultBlobDescriptor(blob), err
}
