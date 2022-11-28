// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package genericocireg_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci/ocirepo"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg/componentmapping"
	"github.com/open-component-model/ocm/pkg/mime"
)

var DefaultContext = ocm.New()

const COMPONENT = "github.com/mandelsoft/ocm"
const TESTBASE = "testbase.de"

var _ = Describe("component repository mapping", func() {
	var tempfs vfs.FileSystem

	var ocispec oci.RepositorySpec
	var spec *genericocireg.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		ocispec, err = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
		Expect(err).To(Succeed())
		spec = genericocireg.NewRepositorySpec(ocispec, nil)

	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("creates a dummy component", func() {
		repo, err := DefaultContext.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp, err := repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err := comp.NewVersion("v1")
		Expect(err).To(Succeed())

		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		Expect(vers.Close()).To(Succeed())
		Expect(comp.Close()).To(Succeed())
		Expect(repo.Close()).To(Succeed())

		// access it again
		repo, err = DefaultContext.RepositoryForSpec(spec)
		Expect(err).To(Succeed())

		ok, err := repo.ExistsComponentVersion(COMPONENT, "v1")
		Expect(err).To(Succeed())
		Expect(ok).To(BeTrue())

		comp, err = repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err = comp.LookupVersion("v1")
		Expect(err).To(Succeed())
		Expect(vers.GetDescriptor()).To(Equal(compdesc.New(COMPONENT, "v1")))

		Expect(vers.Close()).To(Succeed())
		Expect(comp.Close()).To(Succeed())
		Expect(repo.Close()).To(Succeed())
	})

	It("imports blobs", func() {

		base := func(ctx *storagecontext.StorageContext) string {
			return TESTBASE
		}
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().Register(ocirepo.NewBlobHandler(base))).New()

		blob := accessio.BlobAccessForString(mime.MIME_OCTET, "anydata")

		// create repository
		repo, err := ctx.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp, err := repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())

		vers, err := comp.NewVersion("v1")
		Expect(err).To(Succeed())

		acc, err := vers.AddBlob(blob, "", "", nil)
		Expect(err).To(Succeed())

		// check provided actual access to be local blob
		Expect(acc.GetKind()).To(Equal(localblob.Type))
		l, ok := acc.(*localblob.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(l.LocalReference).To(Equal(blob.Digest().String()))
		Expect(l.GlobalAccess).NotTo(BeNil())

		// check provided global access to be oci blob
		g, err := l.GlobalAccess.Evaluate(DefaultContext)
		Expect(err).To(Succeed())
		o, ok := g.(*ociblob.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(o.Digest).To(Equal(blob.Digest()))
		Expect(o.Reference).To(Equal(TESTBASE + "/" + componentmapping.ComponentDescriptorNamespace + "/" + COMPONENT))
		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())
	})

	It("imports artifact", func() {
		mime := artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest) + "+tar+gzip"
		base := func(ctx *storagecontext.StorageContext) string {
			return TESTBASE
		}
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().Register(ocirepo.NewArtifactHandler(base), cpi.ForMimeType(mime))).New()

		// create artifactset
		opts, err := accessio.AccessOptions(nil, accessio.PathFileSystem(tempfs))
		Expect(err).To(Succeed())
		r, err := artifactset.FormatTGZ.Create("test.tgz", opts, 0700)
		Expect(err).To(Succeed())
		testhelper.DefaultManifestFill(r)
		r.Annotate(artifactset.MAINARTIFACT_ANNOTATION, "sha256:"+testhelper.DIGEST_MANIFEST)
		Expect(r.Close()).To(Succeed())

		// create repository
		repo, err := ctx.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		defer repo.Close()

		ocirepo := repo.(*genericocireg.Repository).GetOCIRepository()

		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp, err := repo.LookupComponent(COMPONENT)
		Expect(err).To(Succeed())
		defer comp.Close()
		vers, err := comp.NewVersion("v1")
		Expect(err).To(Succeed())
		defer vers.Close()
		blob := accessio.BlobAccessForFile(mime, "test.tgz", tempfs)

		acc, err := vers.AddBlob(blob, "", "artifact1", nil)
		Expect(err).To(Succeed())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		o := acc.(*ociartifact.AccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artifact1@sha256:" + testhelper.DIGEST_MANIFEST))
		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		acc, err = vers.AddBlob(blob, "", "artifact2:v1", nil)
		Expect(err).To(Succeed())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		o = acc.(*ociartifact.AccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artifact2:v1"))
		err = comp.AddVersion(vers)
		Expect(err).To(Succeed())

		Expect(vers.Close()).To(Succeed())
		Expect(comp.Close()).To(Succeed())

		ns, err := ocirepo.LookupNamespace("artifact2")
		Expect(err).To(Succeed())
		defer ns.Close()
		art, err := ns.GetArtifact("v1")
		Expect(err).To(Succeed())
		defer art.Close()
		testhelper.CheckArtifact(art)
		Expect(art.Close()).To(Succeed())
		Expect(ns.Close()).To(Succeed())
		Expect(repo.(*genericocireg.Repository).Close()).To(Succeed())
	})

})
