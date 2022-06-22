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

package ociutils

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/opencontainers/go-digest"
	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/mime"
)

type BlobInfo struct {
	Error    string          `json:"error,omitempty"`
	Unparsed string          `json:"unparsed,omitempty"`
	Content  json.RawMessage `json:"content,omitempty"`
	Type     string          `json:"type,omitempty"`
	Digest   digest.Digest   `json:"digest,omitempty"`
	Size     int64           `json:"size,omitempty"`
	Info     interface{}     `json:"info,omitempty"`
}
type ArtefactInfo struct {
	Digest     digest.Digest `json:"digest"`
	Type       string        `json:"type"`
	Descriptor interface{}   `json:"descriptor"`
	Config     *BlobInfo     `json:"config,omitempty"`
	Layers     []*BlobInfo   `json:"layers,omitempty"`
	Manifests  []*BlobInfo   `json:"manifests,omitempty"`
}

func GetArtefactInfo(art cpi.ArtefactAccess, layerFiles bool) *ArtefactInfo {
	if art.IsManifest() {
		return GetManifestInfo(art.ManifestAccess(), layerFiles)
	}
	if art.IsIndex() {
		return GetIndexInfo(art.IndexAccess(), layerFiles)
	}
	return &ArtefactInfo{Type: "unspecific"}
}

func GetManifestInfo(m cpi.ManifestAccess, layerFiles bool) *ArtefactInfo {
	info := &ArtefactInfo{
		Type:       artdesc.MediaTypeImageManifest,
		Descriptor: m.GetDescriptor(),
	}
	b, err := m.Blob()
	if err == nil {
		info.Digest = b.Digest()
	}
	man := m.GetDescriptor()
	cfg := &BlobInfo{
		Content: nil,
		Type:    man.Config.MediaType,
		Digest:  man.Config.Digest,
		Size:    man.Config.Size,
	}
	info.Config = cfg

	config, err := accessio.BlobData(m.GetBlob(man.Config.Digest))
	if err != nil {
		cfg.Error = "error getting config blob: " + err.Error()
	} else {
		cfg.Content = json.RawMessage(config)
	}
	h := getHandler(man.Config.MediaType)

	if h != nil {
		cfg.Info = h.Description(m, config)
	}
	for _, l := range man.Layers {
		blobinfo := &BlobInfo{
			Type:   l.MediaType,
			Digest: l.Digest,
			Size:   l.Size,
		}
		blob, err := m.GetBlob(l.Digest)
		if err != nil {
			blobinfo.Error = "error getting blob: " + err.Error()
		} else {
			blobinfo.Info = GetLayerInfo(blob, layerFiles)
		}
		info.Layers = append(info.Layers, blobinfo)
	}
	return info
}

type LayerInfo struct {
	Description string      `json:"description,omitempty"`
	Error       string      `json:"error,omitempty"`
	Unparsed    string      `json:"unparsed,omitempty"`
	Content     interface{} `json:"content,omitempty"`
}

func GetLayerInfo(blob accessio.BlobAccess, layerFiles bool) *LayerInfo {
	info := &LayerInfo{}

	if mime.IsJSON(blob.MimeType()) {
		info.Description = "json document"
		data, err := blob.Get()
		if err != nil {
			info.Error = "cannot read blob: " + err.Error()
			return info
		}
		var j interface{}
		err = json.Unmarshal(data, &j)
		if err != nil {
			if len(data) < 10000 {
				info.Unparsed = string(data)
			}
			info.Error = "invalid json: " + err.Error()
			return info
		}
		info.Content = j
		return info
	}
	if mime.IsYAML(blob.MimeType()) {
		info.Description = "yaml document"
		data, err := blob.Get()
		if err != nil {
			info.Error = "cannot read blob: " + err.Error()
			return info
		}
		var j interface{}
		err = yaml.Unmarshal(data, &j)
		if err != nil {
			if len(data) < 10000 {
				info.Unparsed = string(data)
			}
			info.Error = "invalid yaml: " + err.Error()
			return info
		}
		info.Content = j
		return info
	}
	if !layerFiles {
		return nil
	}
	reader, err := blob.Reader()
	if err != nil {
		info.Error = "cannot read blob: " + err.Error()
		return info
	}
	file, err := os.Create("/tmp/blob")
	if err == nil {
		io.Copy(file, reader)
		file.Close()
		reader.Close()
		reader, err = blob.Reader()
		if err != nil {
			info.Error = "cannot read blob: " + err.Error()
			return info
		}
	}
	defer reader.Close()
	reader, _, err = compression.AutoDecompress(reader)
	if err != nil {
		info.Error = "cannot decompress blob: " + err.Error()
		return info
	}
	var files []string
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return info
			}
			if len(files) == 0 {
				info.Description = "no tar"
				return info
			}
			info.Error = fmt.Sprintf("tar error: %s", err)
			return info
		}
		if len(files) == 0 {
			info.Description = "tar file"
		}

		switch header.Typeflag {
		case tar.TypeDir:
			files = append(files, fmt.Sprintf("dir:  %s\n", header.Name))
		case tar.TypeReg:
			files = append(files, fmt.Sprintf("file: %s\n", header.Name))
		}
	}
	info.Content = files
	return info
}

func GetIndexInfo(i cpi.IndexAccess, layerFiles bool) *ArtefactInfo {
	info := &ArtefactInfo{
		Type:       artdesc.MediaTypeImageIndex,
		Descriptor: i.GetDescriptor(),
	}
	b, err := i.Blob()
	if err == nil {
		info.Digest = b.Digest()
	}
	for _, l := range i.GetDescriptor().Manifests {
		blobinfo := &BlobInfo{
			Type:   l.MediaType,
			Digest: l.Digest,
			Size:   l.Size,
		}
		a, err := i.GetArtefact(l.Digest)
		if err != nil {
			blobinfo.Error = fmt.Sprintf("cannot get artefact: %s\n", err)
		} else {
			blobinfo.Info = GetArtefactInfo(a, layerFiles)
		}
		info.Layers = append(info.Layers, blobinfo)
	}
	return info
}
