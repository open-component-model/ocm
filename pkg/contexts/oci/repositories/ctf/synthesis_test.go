// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ctf_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

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
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/digester/digesters/artefact"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
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
		artefactset.MAINARTEFACT_ANNOTATION:        "sha256:" + DIGEST_MANIFEST,
		artefactset.LEGACY_MAINARTEFACT_ANNOTATION: "sha256:" + DIGEST_MANIFEST,
	}))
	annos := map[string]string{
		artefactset.TAGS_ANNOTATION:        "v1",
		artefactset.LEGACY_TAGS_ANNOTATION: "v1",
	}
	if artefactset.IsOCIDefaultFormat() {
		annos[artefactset.OCITAG_ANNOTATION] = "v1"
	}
	Expect(idx.Manifests).To(Equal([]artdesc.Descriptor{
		{
			MediaType:   artdesc.MediaTypeImageManifest,
			Digest:      "sha256:" + DIGEST_MANIFEST,
			Size:        362,
			Annotations: annos,
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
		spec, err = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
		Expect(err).To(Succeed())
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("synthesize", func() {
		r, err := ctf.FormatDirectory.Create(oci.DefaultContext(), "test", &spec.StandardOptions, 0700)
		Expect(err).To(Succeed())
		n, err := r.LookupNamespace("mandelsoft/test")
		Expect(err).To(Succeed())
		DefaultManifestFill(n)
		Expect(n.Close()).To(Succeed())
		Expect(r.Close()).To(Succeed())

		r, err = ctf.Open(oci.DefaultContext(), accessobj.ACC_READONLY, "test", 0, &spec.StandardOptions)
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
