// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/registry"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils/helm/loader"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/errors"
)

func SynthesizeArtifactBlob(path string, fss ...vfs.FileSystem) (artifactset.ArtifactBlob, error) {
	return artifactset.SythesizeArtifactSet(func(set *artifactset.ArtifactSet) (string, error) {
		chart, blob, err := TransferAsArtifact(path, set, fss...)
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

func TransferAsArtifact(path string, ns oci.NamespaceAccess, fss ...vfs.FileSystem) (*chart.Chart, *artdesc.Descriptor, error) {
	fs := accessio.FileSystem(fss...)

	fi, err := fs.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	chart, err := loader.Load(path, fs)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "cannot load helm chart from %q", path)
	}
	err = chart.Validate()
	if err != nil {
		return nil, nil, errors.ErrInvalidWrap(err, "helm chart", path)
	}

	var provData []byte
	provRef := fmt.Sprintf("%s.prov", path)
	if _, err := fs.Stat(provRef); err == nil {
		provData, err = vfs.ReadFile(fs, provRef)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot read provider metadata")
		}
	}

	if fi.IsDir() {
		dir, err := os.MkdirTemp("", "helmchart-")
		if err != nil {
			return chart, nil, errors.Wrapf(err, "cannot create temporary directory for helm chart")
		}
		defer os.RemoveAll(dir)
		path, err = chartutil.Save(chart, dir)
		if err != nil {
			return chart, nil, err
		}
		fs = osfs.New()
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

	err = m.SetConfigBlob(accessio.BlobAccessForData(registry.ConfigMediaType, configData), nil)
	if err != nil {
		return chart, nil, err
	}
	_, err = m.AddLayer(accessio.BlobAccessForFile(registry.ChartLayerMediaType, path, fs), nil)
	if err != nil {
		return chart, nil, err
	}
	if provData != nil {
		_, err = m.AddLayer(accessio.BlobAccessForData(registry.ProvLayerMediaType, provData), nil)
		if err != nil {
			return chart, nil, err
		}
	}
	blob, err := ns.AddArtifact(art)
	if err != nil {
		return chart, nil, err
	}
	return chart, artdesc.DefaultBlobDescriptor(blob), err
}
