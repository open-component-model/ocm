package ocirepo_test

import (
	"strings"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	ctfoci "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/oci/grammar"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/extensions/download/handlers/ocirepo"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
)

const (
	COMP = "github.com/compa"
	VERS = "1.0.0"
	CTF  = "ctf"
)

const (
	HINT   = "ocm.software/test"
	UPLOAD = "ocm.software/upload"
)

const (
	TARGETHOST = "target"
	TARGETPATH = "/tmp/target"
)

const (
	OCIHOST      = "source"
	OCIPATH      = "/tmp/source"
	OCINAMESPACE = "ocm/value"
	OCIVERSION   = "v2.0"
)

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
			defer Close(cv, "version")

			racc := Must(cv.GetResourceByIndex(0))

			ok, path := Must2(download.For(env).Download(nil, racc, TARGETHOST+".alias"+grammar.RepositorySeparator+UPLOAD, env))
			Expect(ok).To(BeTrue())
			Expect(path).To(Equal("target.alias/ocm.software/upload:1.0.0"))

			MustBeSuccessful(env.OCMContext().Finalize())

			target := Must(ctfoci.Open(env.OCIContext(), accessobj.ACC_READONLY, TARGETPATH, 0, env))
			defer Close(target, "download target")
			Expect(target.ExistsArtifact(path[strings.Index(path, grammar.RepositorySeparator)+1:strings.Index(path, ":")], VERS)).To(BeTrue())
		})
	})
})
