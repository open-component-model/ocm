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

package ctf_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/digester/digesters/artefact"
	"github.com/open-component-model/ocm/pkg/mime"
)

type DummyMethod struct {
	accessio.BlobAccess
}

var _ ocm.AccessMethod = (*DummyMethod)(nil)
var _ accessio.DigestSource = (*DummyMethod)(nil)

func (d *DummyMethod) GetKind() string {
	return localblob.Type
}

func CheckBlob(blob accessio.BlobAccess) oci.NamespaceAccess {
	set, err := artefactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
	Expect(err).To(Succeed())
	defer func() {
		if set != nil {
			set.Close()
		}
	}()

	idx := set.GetIndex()
	Expect(idx.Annotations).To(Equal(map[string]string{
		artefactset.MAINARTEFACT_ANNOTATION: "sha256:" + DIGEST_MANIFEST,
	}))
	Expect(idx.Manifests).To(Equal([]artdesc.Descriptor{
		{
			MediaType: artdesc.MediaTypeImageManifest,
			Digest:    "sha256:" + DIGEST_MANIFEST,
			Size:      362,
			Annotations: map[string]string{
				"cloud.gardener.ocm/tags": "v1",
			},
		},
	}))

	art, err := set.GetArtefact("sha256:" + DIGEST_MANIFEST)
	Expect(err).To(Succeed())
	defer Close(art)
	m, err := art.Manifest()
	Expect(err).To(Succeed())
	Expect(m.Config).To(Equal(artdesc.Descriptor{
		MediaType: mime.MIME_OCTET,
		Digest:    "sha256:" + DIGEST_CONFIG,
		Size:      2,
	}))

	layer, err := art.GetBlob(digest.Digest("sha256:" + DIGEST_LAYER))
	Expect(err).To(Succeed())
	Expect(layer.Get()).To(Equal([]byte("testdata")))

	result := set
	set = nil
	return result
}

var _ = Describe("syntheses", func() {
	var tempfs vfs.FileSystem
	var spec *ctf.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t
		spec = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("synthesize", func() {
		r, err := ctf.FormatDirectory.Create(oci.DefaultContext(), "test", spec.Options, 0700)
		Expect(err).To(Succeed())
		n, err := r.LookupNamespace("mandelsoft/test")
		Expect(err).To(Succeed())
		DefaultManifestFill(n)
		Expect(n.Close()).To(Succeed())
		Expect(r.Close()).To(Succeed())

		r, err = ctf.Open(oci.DefaultContext(), accessobj.ACC_READONLY, "test", 0, spec.Options)
		Expect(err).To(Succeed())
		defer Close(r, "ctf")
		n, err = r.LookupNamespace("mandelsoft/test")
		Expect(err).To(Succeed())
		defer Close(n, "namespace")
		blob, err := artefactset.SynthesizeArtefactBlob(n, TAG)
		Expect(err).To(Succeed())
		defer Close(blob, "blob")
		path := blob.Path()
		Expect(path).To(MatchRegexp(filepath.Join(blob.FileSystem().FSTempDir(), "artefactblob.*\\.tgz")))
		Expect(vfs.Exists(blob.FileSystem(), path)).To(BeTrue())

		set := CheckBlob(blob)
		defer Close(set, "set")

		Expect(blob.Close()).To(Succeed())
		Expect(vfs.Exists(blob.FileSystem(), path)).To(BeFalse())

		// use syntesized blob to extract new blob, useless but should work
		newblob, err := artefactset.SynthesizeArtefactBlob(set, TAG)
		Expect(err).To(Succeed())
		defer Close(newblob, "newblob")

		Expect(CheckBlob(newblob).Close()).To(Succeed())

		meth := &DummyMethod{newblob}
		digest, err := artefact.New(digest.SHA256).DetermineDigest("", meth, nil)
		Expect(err).To(Succeed())
		Expect(digest.Value).To(Equal(DIGEST_MANIFEST))
		Expect(digest.NormalisationAlgorithm).To(Equal(artefact.OciArtifactDigestV1))
		Expect(digest.HashAlgorithm).To(Equal(sha256.Algorithm))

		digests, err := ocm.DefaultContext().BlobDigesters().DetermineDigests("", nil, signing.DefaultRegistry(), meth)
		Expect(err).To(Succeed())
		Expect(digests).To(Equal([]cpi.DigestDescriptor{
			{
				Value:                  DIGEST_MANIFEST,
				HashAlgorithm:          sha256.Algorithm,
				NormalisationAlgorithm: artefact.OciArtifactDigestV1,
			},
		}))

	})
})
