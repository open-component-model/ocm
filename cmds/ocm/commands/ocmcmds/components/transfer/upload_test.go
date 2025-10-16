package transfer_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/oci"
	ctfoci "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/oci/grammar"
	. "ocm.software/ocm/api/oci/testhelper"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	COMP   = "github.com/compa"
	VERS   = "1.0.0"
	CA     = "ca"
	CTF    = "ctf"
	COPY   = "./ctf.copy"
	TARGET = "/tmp/target"
)

const OCISRC = "/tmp/source"

var _ = Describe("upload", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()

		// fake OCI registry
		spec := Must(ctfoci.NewRepositorySpec(accessobj.ACC_READONLY, OCISRC, accessio.PathFileSystem(env.FileSystem())))
		env.OCIContext().SetAlias(OCIHOST, spec)

		env.OCICommonTransport(OCISRC, accessio.FormatDirectory, func() {
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

		env.OCICommonTransport(TARGET, accessio.FormatDirectory)

		env.ComponentArchive(CA, accessio.FormatDirectory, COMP, VERS, func() {
			env.Provider("mandelsoft")
			env.Resource("value", "", resourcetypes.OCI_IMAGE, v1.LocalRelation, func() {
				env.Access(
					ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
				)
			})
		})

		ca := Must(comparch.Open(env.OCMContext(), accessobj.ACC_READONLY, CA, 0, env))
		oca := accessio.OnceCloser(ca)
		defer Close(oca)

		ctf := Must(ctfocm.Create(env.OCMContext(), accessobj.ACC_CREATE, CTF, 0o700, env))
		octf := accessio.OnceCloser(ctf)
		defer Close(octf)

		handler := Must(standard.New(standard.ResourcesByValue()))

		MustBeSuccessful(transfer.TransferVersion(nil, nil, ca, ctf, handler))

		// now we have a transport archive with local blob for the image
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers oci artifact with named handler", func() {
		ctx := env.OCMContext()
		config := Must(json.Marshal(ociuploadattr.New(TARGET + grammar.RepositorySeparator + grammar.RepositorySeparator + "copy")))

		buf := bytes.NewBuffer(nil)
		err := env.CatchOutput(buf).Execute("transfer", "components", "--uploader", "ocm/ociArtifacts="+string(config), "--copy-resources", CTF, COPY)
		if err != nil {
			fmt.Printf("%s\n", buf.String())
			Expect(err).To(Succeed())
		}
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/compa:1.0.0"...
...resource 0 value[ociImage](ocm/value:v2.0)...
...adding component version...
1 versions transferred

`))

		copy := Must(ctfocm.Open(ctx, accessobj.ACC_READONLY, COPY, 0o700, env))
		defer Close(copy)

		// check type
		cv2 := Must(copy.LookupComponentVersion(COMP, VERS))
		defer Close(cv2)
		ra := Must(cv2.GetResourceByIndex(0))
		acc := Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		val := Must(ctx.AccessSpecForSpec(acc))
		// TODO: the result is invalid for ctf: better handling for ctf refs
		Expect(val.(*ociartifact.AccessSpec).ImageReference).To(Equal("/tmp/target//copy/ocm/value:v2.0@sha256:" + D_OCIMANIFEST1))

		target, err := ctfoci.Open(ctx.OCIContext(), accessobj.ACC_READONLY, TARGET, 0, env)
		Expect(err).To(Succeed())
		defer Close(target)
		Expect(target.ExistsArtifact("copy/ocm/value", "v2.0")).To(BeTrue())
	})
})
