// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ociutils

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/utils"
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
		s += utils.IndentLines(h.Description(m, config), "  ")
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
		s += utils.IndentLines(PrintLayer(blob), "  ")
	}
	return s
}

func PrintLayer(blob accessio.BlobAccess) string {
	reader, err := blob.Reader()
	if err != nil {
		return "cannot read blob: " + err.Error()
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
			if errors.Is(err, io.EOF) {
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
			s += "  resolved artefact:\n"
			s += utils.IndentLines(PrintArtefact(a), "    ")
		}
	}
	return s
}
