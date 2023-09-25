// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocirepo_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/grammar"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	ctfoci "github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/ociuploadattr"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download/handlers/ocirepo"
	ctfocm "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
)

const COMP = "github.com/compa"
const VERS = "1.0.0"
const CTF = "ctf"

const HINT = "ocm.software/test"
const UPLOAD = "ocm.software/upload"

const TARGETHOST = "target"
const TARGETPATH = "/tmp/target"

const OCIHOST = "source"
const OCIPATH = "/tmp/source"
const OCINAMESPACE = "ocm/value"
const OCIVERSION = "v2.0"

const ARTIFACTSET = "/tmp/set.tgz"

var _ = Describe("upload", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()

		// fake OCI registry
		spec := Must(ctfoci.NewRepositorySpec(accessobj.ACC_WRITABLE, TARGETPATH, env))
		env.OCIContext().SetAlias(TARGETHOST, spec)

		env.OCICommonTransport(TARGETPATH, accessio.FormatDirectory)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("local blob", func() {
		BeforeEach(func() {
			env.ArtifactSet(ARTIFACTSET, accessio.FormatTGZ, func() {
				env.Manifest(OCIVERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "manifestlayer")
					})
				})
				env.Annotation(artifactset.MAINARTIFACT_ANNOTATION, OCIVERSION)
			})

			env.OCMCommonTransport(CTF, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP, VERS, func() {
					env.Provider("mandelsoft")
					env.Resource("value", "", resourcetypes.OCI_IMAGE, v1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(ociv1.MediaTypeImageManifest), ARTIFACTSET)
						env.Hint(HINT)
					})
				})
			})
		})

		It("uploads local oci artifact blob", func() {
			download.For(env).Register(ocirepo.New(), download.ForArtifactType(resourcetypes.OCI_IMAGE))

			src := Must(ctfocm.Open(env, accessobj.ACC_READONLY, CTF, 0, env))
			defer Close(src, "source ctf")

			cv := Must(src.LookupComponentVersion(COMP, VERS))
			defer Close(cv)

			racc := Must(cv.GetResourceByIndex(0))

			ok, path := Must2(download.For(env).Download(nil, racc, TARGETHOST+".alias"+grammar.RepositorySeparator+UPLOAD, env))
			Expect(ok).To(BeTrue())
			Expect(path).To(Equal("target.alias/ocm.software/upload:1.0.0"))

			env.OCMContext().Finalize()

			target, err := ctfoci.Open(env.OCIContext(), accessobj.ACC_READONLY, TARGETPATH, 0, env)
			Expect(err).To(Succeed())
			defer Close(target)
			Expect(target.ExistsArtifact(path[strings.Index(path, grammar.RepositorySeparator)+1:strings.Index(path, ":")], VERS)).To(BeTrue())
		})

		It("uploads local oci artifact blob using named handler", func() {
			download.RegisterHandlerByName(env, ocirepo.PATH, nil, download.ForArtifactType(resourcetypes.OCI_IMAGE))

			src := Must(ctfocm.Open(env.OCMContext(), accessobj.ACC_READONLY, CTF, 0, accessio.PathFileSystem(env)))
			defer Close(src, "source ctf")

			cv := Must(src.LookupComponentVersion(COMP, VERS))
			defer Close(cv)

			racc := Must(cv.GetResourceByIndex(0))

			ok, path := Must2(download.For(env).Download(nil, racc, TARGETHOST+".alias"+grammar.RepositorySeparator+UPLOAD, env))
			Expect(ok).To(BeTrue())
			Expect(path).To(Equal("target.alias/ocm.software/upload:1.0.0"))

			env.OCMContext().Finalize()

			target, err := ctfoci.Open(env.OCIContext(), accessobj.ACC_READONLY, TARGETPATH, 0, env)
			Expect(err).To(Succeed())
			defer Close(target)
			Expect(target.ExistsArtifact(path[strings.Index(path, grammar.RepositorySeparator)+1:strings.Index(path, ":")], VERS)).To(BeTrue())
		})

		It("uploads local oci artifact blob using named handler and config", func() {
			cfg := ociuploadattr.Attribute{
				Ref: TARGETHOST + ".alias" + grammar.RepositorySeparator + "upload",
			}
			download.RegisterHandlerByName(env, ocirepo.PATH, cfg, download.ForArtifactType(resourcetypes.OCI_IMAGE))

			src := Must(ctfocm.Open(env.OCMContext(), accessobj.ACC_READONLY, CTF, 0, accessio.PathFileSystem(env)))
			defer Close(src, "source ctf")

			cv := Must(src.LookupComponentVersion(COMP, VERS))
			defer Close(cv)

			racc := Must(cv.GetResourceByIndex(0))

			ok, path := Must2(download.For(env).Download(nil, racc, "", env))
			Expect(ok).To(BeTrue())
			// Expect(path).To(Equal("CommonTransportFormat::/tmp/target//upload/ocm.software/test:1.0.0"))
			Expect(path).To(Equal("target.alias/upload/ocm.software/test:1.0.0"))

			env.OCMContext().Finalize()

			target, err := ctfoci.Open(env.OCIContext(), accessobj.ACC_READONLY, TARGETPATH, 0, env)
			Expect(err).To(Succeed())
			defer Close(target)
			// Expect(target.ExistsArtifact(path[strings.Index(path, grammar.RepositorySeparator+grammar.RepositorySeparator)+2:strings.LastIndex(path, ":")], VERS)).To(BeTrue())
			Expect(target.ExistsArtifact(path[strings.Index(path, grammar.RepositorySeparator)+1:strings.Index(path, ":")], VERS)).To(BeTrue())
		})
	})

	Context("oci ref", func() {
		BeforeEach(func() {
			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				env.Namespace(OCINAMESPACE, func() {
					env.Manifest(OCIVERSION, func() {
						env.Config(func() {
							env.BlobStringData(mime.MIME_JSON, "{}")
						})
						env.Layer(func() {
							env.BlobStringData(mime.MIME_TEXT, "manifestlayer")
						})
					})
				})
			})

			// fake OCI registry
			spec := Must(ctfoci.NewRepositorySpec(accessobj.ACC_WRITABLE, OCIPATH, env))
			env.OCIContext().SetAlias(OCIHOST, spec)

			env.OCMCommonTransport(CTF, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP, VERS, func() {
					env.Provider("mandelsoft")
					env.Resource("value", "", resourcetypes.OCI_IMAGE, v1.LocalRelation, func() {
						env.Access(ociartifact.New(OCIHOST + ".alias" + grammar.RepositorySeparator + OCINAMESPACE + grammar.TagSeparator + OCIVERSION))
					})
				})
			})
		})

		It("uploads oci artifact ref", func() {
			download.For(env).Register(ocirepo.New(), download.ForArtifactType(resourcetypes.OCI_IMAGE))

			src := Must(ctfocm.Open(env.OCMContext(), accessobj.ACC_READONLY, CTF, 0, accessio.PathFileSystem(env)))
			defer Close(src, "source ctf")

			cv := Must(src.LookupComponentVersion(COMP, VERS))
			defer Close(cv)

			racc := Must(cv.GetResourceByIndex(0))

			ok, path := Must2(download.For(env).Download(nil, racc, TARGETHOST+".alias"+grammar.RepositorySeparator+UPLOAD, env))
			Expect(ok).To(BeTrue())
			Expect(path).To(Equal("target.alias/ocm.software/upload:1.0.0"))

			env.OCMContext().Finalize()

			target, err := ctfoci.Open(env.OCIContext(), accessobj.ACC_READONLY, TARGETPATH, 0, env)
			Expect(err).To(Succeed())
			defer Close(target)
			Expect(target.ExistsArtifact(path[strings.Index(path, grammar.RepositorySeparator)+1:strings.Index(path, ":")], VERS)).To(BeTrue())
		})
	})
})
