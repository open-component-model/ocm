// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package artefact

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
)

const OciArtefactDigestV1 string = "ociArtefactDigest/v1"

func init() {
	cpi.MustRegisterDigester(New(digest.SHA256), "")
	cpi.MustRegisterDigester(New(digest.SHA512), "")
}

func New(algo digest.Algorithm) cpi.BlobDigester {
	return &Digester{
		cpi.DigesterType{
			HashAlgorithm:          algo.String(),
			NormalizationAlgorithm: OciArtefactDigestV1,
		},
	}
}

type Digester struct {
	typ cpi.DigesterType
}

var _ cpi.BlobDigester = (*Digester)(nil)

func (d *Digester) GetType() cpi.DigesterType {
	return d.typ
}

func (d *Digester) DetermineDigest(reftyp string, acc cpi.AccessMethod, preferred signing.Hasher) (*cpi.DigestDescriptor, error) {
	if acc.GetKind() == localblob.Type {
		mime := acc.MimeType()
		if !artdesc.IsOCIMediaType(mime) {
			return nil, nil
		}
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

		var desc *cpi.DigestDescriptor
		oci := false
		layout := false
		for {
			header, err := tr.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					if oci {
						if layout {
							return desc, nil
						} else {
							err = fmt.Errorf("oci-layout not found")
						}
					} else {
						err = fmt.Errorf("descriptor not found in archive")
					}
				}
				return nil, errors.ErrInvalidWrap(err, "artefact archive")
			}

			switch header.Typeflag {
			case tar.TypeDir:
			case tar.TypeReg:
				switch header.Name {
				case artefactset.OCILayouFileName:
					layout = true
				case artefactset.OCIArtefactSetDescriptorFileName:
					oci = true
					fallthrough
				case artefactset.ArtefactSetDescriptorFileName:
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
					main := artefactset.RetrieveMainArtefact(index.Annotations)
					if main == "" {
						return nil, fmt.Errorf("no main artefact found")
					}
					if digest.Digest(main).Algorithm() != digest.Algorithm(d.GetType().HashAlgorithm) {
						return nil, nil
					}
					desc = cpi.NewDigestDescriptor(digest.Digest(main).Hex(), d.GetType())
					if !oci {
						return desc, nil
					}
				}
			}
		}
		// not reached (endless for)
	}
	if acc.GetKind() == ociartefact.Type {
		dig := acc.(accessio.DigestSource).Digest()
		if dig != "" {
			if dig.Algorithm() != digest.Algorithm(d.GetType().HashAlgorithm) {
				return nil, nil
			}
			return cpi.NewDigestDescriptor(dig.Hex(), d.GetType()), nil
		}
		return nil, errors.Newf("cannot determine digest")
	}
	return nil, nil
}
