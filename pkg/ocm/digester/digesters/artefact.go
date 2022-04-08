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

package digesters

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	artefactset2 "github.com/gardener/ocm/pkg/oci/repositories/artefactset"
	"github.com/gardener/ocm/pkg/ocm/digester/core"
	"github.com/opencontainers/go-digest"
)

func init() {
	d := &ArtefactDigester{}
	core.RegisterDigester(d, artdesc.ArchiveBlobTypes()...)
}

var ARTEFACT_DIGESTER = core.DigesterType{
	Kind:    "artefact",
	Version: "v1",
}

type ArtefactDigester struct{}

var _ core.BlobDigester = (*ArtefactDigester)(nil)

func (d ArtefactDigester) GetType() core.DigesterType {
	return ARTEFACT_DIGESTER
}

func (d ArtefactDigester) DetermineDigest(blob accessio.BlobAccess) (*core.DigestDescriptor, error) {
	mime := blob.MimeType()
	r, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var reader io.Reader = r
	if strings.HasSuffix(mime, "+gzip") {
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
	}
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil, errors.ErrInvalid("artefact archive")
			}
			return nil, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
		case tar.TypeReg:
			if header.Name == artefactset2.ArtefactSetDescriptorFileName {
				data, err := io.ReadAll(tr)
				if err != nil {
					return nil, fmt.Errorf("unable to read descriptor from archive: %w", err)
				}
				index, err := artdesc.DecodeIndex(data)
				if err != nil {
					return nil, err
				}
				if index == nil {
					return nil, fmt.Errorf("no main artefact found")
				}
				main := index.Annotations[artefactset2.MAINARTEFACT_ANNOTATION]
				if main == "" {
					return nil, fmt.Errorf("no main artefact found")
				}
				return core.NewDigestDescriptor(digest.Digest(main), d.GetType()), nil
			}
		}
	}
	return nil, fmt.Errorf("unable to read descriptor from archive: %w", err)
}
