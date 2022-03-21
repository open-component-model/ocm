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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/compression"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
)

func PrintArtefact(art cpi.ArtefactAccess) string {
	if art.IsManifest() {
		return fmt.Sprintf("type: %s\n", artdesc.MediaTypeImageManifest) + PrintManifest(art.ManifestAccess())
	}
	if art.IsIndex() {
		return fmt.Sprintf("type: %s\n", artdesc.MediaTypeImageIndex+PrintIndex(art.IndexAccess()))
	}
	return "unspecific"
}

func IndentLines(orig string, gap string) string {
	s := ""
	for _, l := range strings.Split(orig, "\n") {
		s += gap + l + "\n"
	}
	return s
}

func PrintManifest(m cpi.ManifestAccess) string {
	s := ""
	data, err := accessio.BlobData(m.Blob())
	if err != nil {
		s += fmt.Sprintf("descriptor: invalid: %s\n", err)
	} else {
		s += fmt.Sprintf("descriptor: %s\n", string(data))
	}
	man := m.GetDescriptor()
	s += "config:\n"
	s += fmt.Sprintf("  type:        %s\n", man.Config.MediaType)
	s += fmt.Sprintf("  digest:      %s\n", man.Config.Digest)
	s += fmt.Sprintf("  size:        %d\n", man.Config.Size)

	config, err := accessio.BlobData(m.GetBlob(man.Config.Digest))
	if err != nil {
		s += "  error getting config blob: " + err.Error() + "\n"
	} else {
		s += fmt.Sprintf("  config json: %s\n", string(config))
	}
	h := getHandler(man.Config.MediaType)

	if h != nil {
		s += IndentLines(h.Description(m, config), "  ")
	}
	s += "layers:\n"
	for _, l := range man.Layers {
		s += fmt.Sprintf("- type:   %s\n", l.MediaType)
		s += fmt.Sprintf("  digest: %s\n", l.Digest)
		s += fmt.Sprintf("  size:   %d\n", l.Size)
		blob, err := m.GetBlob(l.Digest)
		if err != nil {
			s += "  error getting blob: " + err.Error() + "\n"
		}
		s += IndentLines(PrintLayer(blob), "  ")
	}
	return s
}

func PrintLayer(blob accessio.BlobAccess) string {
	reader, err := blob.Reader()
	if err != nil {
		return "cannot read blob: " + err.Error()
	}
	file, err := os.Create("/tmp/blob")
	if err == nil {
		io.Copy(file, reader)
		file.Close()
		reader.Close()
		reader, err = blob.Reader()
		if err != nil {
			return "cannot read blob: " + err.Error()
		}
	}
	defer reader.Close()
	reader, _, err = compression.AutoDecompress(reader)
	if err != nil {
		return "cannot decompress blob: " + err.Error()
	}
	tr := tar.NewReader(reader)
	s := ""
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return s
			}
			if s == "" {
				return "no tar"
			}
			return s + fmt.Sprintf("tar error: %s", err)
		}
		if s == "" {
			s = "tar filesystem:\n"
		}

		switch header.Typeflag {
		case tar.TypeDir:
			s += fmt.Sprintf("  dir:  %s\n", header.Name)
		case tar.TypeReg:
			s += fmt.Sprintf("  file: %s\n", header.Name)
		}
	}
	return s
}

func PrintIndex(i cpi.IndexAccess) string {
	s := "manifests:\n"
	for _, l := range i.GetDescriptor().Manifests {
		s += fmt.Sprintf("- type:   %s\n", l.MediaType)
		s += fmt.Sprintf("  digest: %s\n", l.Digest)
		a, err := i.GetArtefact(l.Digest)
		if err != nil {
			s += fmt.Sprintf("  error: %s\n", err)
		} else {
			s += fmt.Sprintf("  resolved artefact:\n")
			s += IndentLines(PrintArtefact(a), "    ")
		}
	}
	return s
}
