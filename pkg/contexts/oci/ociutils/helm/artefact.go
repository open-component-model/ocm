// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/errors"
)

func SynthesizeArtefactBlob(path string, fss ...vfs.FileSystem) (artefactset.ArtefactBlob, error) {
	return artefactset.SythesizeArtefactSet(artdesc.MediaTypeImageManifest, func(set *artefactset.ArtefactSet) error {
		chart, blob, err := TransferAsArtefact(path, set, fss...)
		if err != nil {
			return err
		}
		if chart.Metadata.Version != "" {
			err = set.AddTags(blob.Digest, chart.Metadata.Version)
			if err != nil {
				return err
			}
		}
		set.Annotate(artefactset.MAINARTEFACT_ANNOTATION, blob.Digest.String())
		return err
	})
}
func TransferAsArtefact(path string, ns oci.NamespaceAccess, fss ...vfs.FileSystem) (*chart.Chart, *artdesc.Descriptor, error) {
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

	art, err := ns.NewArtefact()
	if err != nil {
		return chart, nil, err
	}
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
	blob, err := ns.AddArtefact(art)
	if err != nil {
		return chart, nil, err
	}
	return chart, artdesc.DefaultBlobDescriptor(blob), err
}
