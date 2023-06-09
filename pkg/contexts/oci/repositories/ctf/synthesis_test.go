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
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/digester/digesters/artifact"
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

func (d *DummyMethod) AccessSpec() cpi.AccessSpec {
	return nil
}

func CheckBlob(blob accessio.BlobAccess) oci.NamespaceAccess {
	set := Must(artifactset.OpenFromBlob(accessobj.ACC_READONLY, blob))
	defer func() {
		if set != nil {
			set.Close()
		}
	}()

	idx := set.GetIndex()
	Expect(idx.Annotations).To(Equal(map[string]string{
		artifactset.MAINARTIFACT_ANNOTATION: "sha256:" + DIGEST_MANIFEST,
	}))
	annos := map[string]string{
		artifactset.TAGS_ANNOTATION: "v1",
	}
	if artifactset.IsOCIDefaultFormat() {
		annos[artifactset.OCITAG_ANNOTATION] = "v1"
	}
	Expect(idx.Manifests).To(Equal([]artdesc.Descriptor{
		{
			MediaType:   artdesc.MediaTypeImageManifest,
			Digest:      "sha256:" + DIGEST_MANIFEST,
			Size:        362,
			Annotations: annos,
		},
	}))

	art := Must(set.GetArtifact("sha256:" + DIGEST_MANIFEST))
	defer Close(art)
	m := Must(art.Manifest())
	Expect(m.Config).To(Equal(artdesc.Descriptor{
		MediaType: mime.MIME_OCTET,
		Digest:    "sha256:" + DIGEST_CONFIG,
		Size:      2,
	}))

	layer := Must(art.GetBlob(digest.Digest("sha256:" + DIGEST_LAYER)))
	Expect(layer.Get()).To(Equal([]byte("testdata")))

	result := set
	set = nil
	return result
}

var _ = Describe("syntheses", func() {
	var tempfs vfs.FileSystem
	var spec *ctf.RepositorySpec

	BeforeEach(func() {
		t := Must(osfs.NewTempFileSystem())
		tempfs = t
		spec = Must(ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory))
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("synthesize", func() {
		r := Must(ctf.FormatDirectory.Create(oci.DefaultContext(), "test", &spec.StandardOptions, 0700))
		n := Must(r.LookupNamespace("mandelsoft/test"))
		DefaultManifestFill(n)
		Expect(n.Close()).To(Succeed())
		Expect(r.Close()).To(Succeed())

		r = Must(ctf.Open(oci.DefaultContext(), accessobj.ACC_READONLY, "test", 0, &spec.StandardOptions))
		defer Close(r, "ctf")
		n = Must(r.LookupNamespace("mandelsoft/test"))
		defer Close(n, "namespace")
		blob := Must(artifactset.SynthesizeArtifactBlob(n, TAG))
		defer Close(blob, "blob")
		path := blob.Path()
		Expect(path).To(MatchRegexp(filepath.Join(blob.FileSystem().FSTempDir(), "artifactblob.*\\.tgz")))
		Expect(vfs.Exists(blob.FileSystem(), path)).To(BeTrue())

		set := CheckBlob(blob)
		defer Close(set, "set")

		Expect(blob.Close()).To(Succeed())
		Expect(vfs.Exists(blob.FileSystem(), path)).To(BeFalse())

		// use syntesized blob to extract new blob, useless but should work
		newblob := Must(artifactset.SynthesizeArtifactBlob(set, TAG))
		defer Close(newblob, "newblob")

		Expect(CheckBlob(newblob).Close()).To(Succeed())

		meth := &DummyMethod{newblob}
		digest := Must(artifact.New(sha256.Algorithm).DetermineDigest("", meth, nil))
		Expect(digest).NotTo(BeNil())
		Expect(digest.Value).To(Equal(DIGEST_MANIFEST))
		Expect(digest.NormalisationAlgorithm).To(Equal(artifact.OciArtifactDigestV1))
		Expect(digest.HashAlgorithm).To(Equal(sha256.Algorithm))

		digests := Must(ocm.DefaultContext().BlobDigesters().DetermineDigests("", nil, signing.DefaultRegistry(), meth))
		Expect(digests).To(Equal([]cpi.DigestDescriptor{
			{
				Value:                  DIGEST_MANIFEST,
				HashAlgorithm:          sha256.Algorithm,
				NormalisationAlgorithm: artifact.OciArtifactDigestV1,
			},
		}))
	})
})
