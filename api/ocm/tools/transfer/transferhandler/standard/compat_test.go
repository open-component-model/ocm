package standard_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	COMPAT_ARCH = "/testdata/v0.18.0"
	COMPAT_COMP = "github.com/mandelsoft/test1"
	COMPAT_VERS = "1.0.0"
)

var _ = Describe("Transfer Test Environment", func() {
	Context("extraid compatibility transfer", func() {
		var env *TestEnv

		BeforeEach(func() {
			env = NewTestEnv(TestData())
		})

		It("", func() {
			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, COMPAT_ARCH, 0, env))
			defer Close(src, "source")
			cv := Must(src.LookupComponentVersion(COMPAT_COMP, COMPAT_VERS))
			defer Close(cv, "source cv")
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer Close(tgt, "target")

			p, buf := common.NewBufferedPrinter()
			MustBeSuccessful(transfer.Transfer(cv, tgt, transfer.WithPrinter(p)))
			Expect(env.DirExists(OUT)).To(BeTrue())

			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "github.com/mandelsoft/test1:1.0.0"...
...resource 0 multi-implicit[plainText]...
...resource 1 multi-implicit[plainText]...
...resource 2 multi-explicit[plainText]...
...resource 3 multi-explicit[plainText]...
...source 0 multi-implicit[plainText]...
...source 1 multi-implicit[plainText]...
...source 2 multi-explicit[plainText]...
...source 3 multi-explicit[plainText]...
...adding component version...
`))

			tcv := Must(tgt.LookupComponentVersion(COMPAT_COMP, COMPAT_VERS))
			defer Close(tcv, "target cv")
			Expect(tcv.GetDescriptor()).To(YAMLEqual(cv.GetDescriptor()))

			Expect(tcv.GetDescriptor().Resources[0].ExtraIdentity).To(BeNil())
			Expect(tcv.GetDescriptor().Resources[1].ExtraIdentity).To(BeNil())
			Expect(tcv.GetDescriptor().Sources[0].ExtraIdentity).To(BeNil())
			Expect(tcv.GetDescriptor().Sources[1].ExtraIdentity).To(BeNil())
		})
	})
})
