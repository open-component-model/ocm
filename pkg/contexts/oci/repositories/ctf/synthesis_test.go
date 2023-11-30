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

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/digester/digesters/artifact"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

type dummyMethod struct {
	blobaccess.BlobAccess
}

var _ blobaccess.DigestSource = (*dummyMethod)(nil)

func (d *dummyMethod) GetKind() string {
	return localblob.Type
}

func (d *dummyMethod) IsLocal() bool {
	return true
}

func (d *dummyMethod) AccessSpec() cpi.AccessSpec {
	return nil
}

func NewDummyMethod(blob blobaccess.BlobAccess) ocm.AccessMethod {
	m, _ := accspeccpi.AccessMethodForImplementation(&dummyMethod{blob}, nil)
	return m
}

func CheckBlob(blob blobaccess.BlobAccess) oci.NamespaceAccess {
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
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		nested := finalize.Nested()

		// setup the scene
		r := Must(ctf.FormatDirectory.Create(oci.DefaultContext(), "test", &spec.StandardOptions, 0700))
		nested.Close(r, "create ctf")
		n := Must(r.LookupNamespace("mandelsoft/test"))
		nested.Close(n, "ns")
		DefaultManifestFill(n)
		MustBeSuccessful(nested.Finalize())

		r = Must(ctf.Open(oci.DefaultContext(), accessobj.ACC_READONLY, "test", 0, &spec.StandardOptions))
		finalize.Close(r, "ctf")
		n = Must(r.LookupNamespace("mandelsoft/test"))
		finalize.Close(n, "names.pace")

		nested = finalize.Nested()
		blob := Must(artifactset.SynthesizeArtifactBlob(n, TAG))
		nested.Close(blob, "blob")

		info := blobaccess.Cast[blobaccess.FileLocation](blob)
		path := info.Path()
		Expect(path).To(MatchRegexp(filepath.Join(info.FileSystem().FSTempDir(), "artifactblob.*\\.tgz")))
		Expect(vfs.Exists(info.FileSystem(), path)).To(BeTrue())

		set := CheckBlob(blob)
		finalize.Close(set, "set")

		MustBeSuccessful(nested.Finalize())
		Expect(vfs.Exists(info.FileSystem(), path)).To(BeFalse())

		// use syntesized blob to extract new blob, useless but should work
		newblob := Must(artifactset.SynthesizeArtifactBlob(set, TAG))
		finalize.Close(newblob, "newblob")

		finalize.Close(CheckBlob(newblob), "newset")

		meth := NewDummyMethod(newblob)
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
