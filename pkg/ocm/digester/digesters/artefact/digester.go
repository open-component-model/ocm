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

package artefact

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/oci/artdesc"
	artefactset2 "github.com/open-component-model/ocm/pkg/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/ocm/accessmethods/ociregistry"
	"github.com/open-component-model/ocm/pkg/ocm/cpi"
	"github.com/opencontainers/go-digest"
)

func init() {
	d := &Digester{}
	cpi.DefaultBlobDigesterRegistry().RegisterDigester(d, "")
}

var ARTEFACT_DIGESTER = cpi.DigesterType{
	Kind:    "artefact",
	Version: "v1",
}

type Digester struct{}

var _ cpi.BlobDigester = (*Digester)(nil)

func (d Digester) GetType() cpi.DigesterType {
	return ARTEFACT_DIGESTER
}

func (d Digester) DetermineDigest(reftyp string, acc cpi.AccessMethod) (*cpi.DigestDescriptor, error) {
	if acc.GetKind() == localblob.Type {
		mime := acc.MimeType()
		r, err := acc.Reader()
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
					return cpi.NewDigestDescriptor(digest.Digest(main), d.GetType()), nil
				}
			}
		}
		return nil, fmt.Errorf("unable to read descriptor from archive: %w", err)
	}
	if acc.GetKind() == ociregistry.Type {
		dig := acc.(accessio.DigestSource).Digest()
		if dig != "" {
			return cpi.NewDigestDescriptor(dig, d.GetType()), nil
		}
		return nil, errors.Newf("cannot determine digest")
	}
	return nil, nil
}
