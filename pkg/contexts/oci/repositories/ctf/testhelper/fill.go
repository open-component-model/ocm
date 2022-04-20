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

package testhelper

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/opencontainers/go-digest"

	. "github.com/onsi/gomega"
)

const TAG = "v1"
const DIGEST_MANIFEST = "3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a"
const DIGEST_LAYER = "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"
const DIGEST_CONFIG = "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"

func DefaultManifestFill(n cpi.NamespaceAccess) {
	art := NewArtefact(n)
	blob, err := n.AddArtefact(art)
	Expect(err).To(Succeed())
	n.AddTags(blob.Digest(), TAG)
}

func NewArtefact(n cpi.NamespaceAccess) cpi.ArtefactAccess {
	art, err := n.NewArtefact()
	Expect(err).To(Succeed())
	Expect(art.AddLayer(accessio.BlobAccessForString(mime.MIME_OCTET, "testdata"), nil)).To(Equal(0))
	desc, err := art.Manifest()
	Expect(err).To(Succeed())
	Expect(desc).NotTo(BeNil())

	Expect(desc.Layers[0].Digest).To(Equal(digest.FromString("testdata")))
	Expect(desc.Layers[0].MediaType).To(Equal(mime.MIME_OCTET))
	Expect(desc.Layers[0].Size).To(Equal(int64(8)))

	config := accessio.BlobAccessForData(mime.MIME_OCTET, []byte("{}"))
	Expect(n.AddBlob(config)).To(Succeed())
	desc.Config = *artdesc.DefaultBlobDescriptor(config)
	return art
}

func CheckArtefact(art oci.ArtefactAccess) {
	Expect(art.IsManifest()).To(BeTrue())
	blob, err := art.GetBlob("sha256:" + DIGEST_LAYER)
	Expect(err).To(Succeed())
	Expect(blob.Get()).To(Equal([]byte("testdata")))
	Expect(blob.MimeType()).To(Equal(mime.MIME_OCTET))
	blob, err = art.GetBlob("sha256:" + DIGEST_CONFIG)
	Expect(err).To(Succeed())
	Expect(blob.Get()).To(Equal([]byte("{}")))
	Expect(blob.MimeType()).To(Equal(mime.MIME_OCTET))
}
