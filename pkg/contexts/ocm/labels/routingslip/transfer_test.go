package routingslip_test

import (
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/finalizer"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compositionmodeattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/types/comment"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const ARCH = "/tmp/ctf"
const TARGET = "/tmp/target"
const COMPONENT = "acme.org/routingslip"
const VERSION = "1.0.0"
const LOCAL = "local.org"

var _ = Describe("management", func() {
	var env *builder.Builder

	BeforeEach(func() {
		env = builder.NewBuilder()
		env.RSAKeyPair(ORG, LOCAL)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	DescribeTable("transfers and updates", func(mode bool) {
		var finalize finalizer.Finalizer

		defer Defer(finalize.Finalize, "finalizer")

		compositionmodeattr.Set(env.OCMContext(), mode)
		e1 := comment.New("start of routing slip")
		e2 := comment.New("additional entry")

		repo := Must(ctf.Open(env, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, ARCH, 0o700, env))
		finalize.Close(repo, "repo")

		c := Must(repo.LookupComponent(COMPONENT))
		finalize.Close(c, "comp")
		cv := Must(c.NewVersion(VERSION))
		finalize.Close(cv, "vers")
		cv.GetDescriptor().Provider.Name = ORG
		MustBeSuccessful(routingslip.AddEntry(cv, ORG, rsa.Algorithm, e1, nil))
		MustBeSuccessful(c.AddVersion(cv))

		target := Must(ctf.Open(env, accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, TARGET, 0o700, env))
		finalize.Close(target, "target")
		pr, buf := common.NewBufferedPrinter()

		MustBeSuccessful(transfer.TransferVersion(pr, nil, cv, target, Must(standard.New())))

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "acme.org/routingslip:1.0.0"...
...adding component version...
`))
		nested := finalize.Nested()
		tc := Must(target.LookupComponent(COMPONENT))
		nested.Close(tc, "target comp")
		tcv := Must(tc.LookupVersion(VERSION))
		nested.Close(tcv)

		slip := Must(routingslip.GetSlip(tcv, ORG))
		MustBeSuccessful(routingslip.AddEntry(tcv, LOCAL, rsa.Algorithm, e1, nil))
		Expect(slip.Len()).To(Equal(1))

		MustBeSuccessful(tc.AddVersion(tcv))
		MustBeSuccessful(nested.Finalize())

		buf.Reset()
		MustBeSuccessful(routingslip.AddEntry(cv, ORG, rsa.Algorithm, e2, nil))
		MustBeSuccessful(transfer.TransferVersion(pr, nil, cv, target, Must(standard.New())))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
transferring version "acme.org/routingslip:1.0.0"...
  updating volatile properties of "acme.org/routingslip:1.0.0"
...adding component version...
`))

		tcv = Must(target.LookupComponentVersion(COMPONENT, VERSION))
		finalize.Close(tcv, "target")
		label := Must(routingslip.Get(tcv))
		Expect(len(label)).To(Equal(2))
		Expect(len(label[ORG])).To(Equal(2))
		Expect(len(label[LOCAL])).To(Equal(1))
		fmt.Printf("*** routing slips:\n%s\n", Must(runtime.DefaultYAMLEncoding.Marshal(label)))
	},
		Entry("with direct mode", false),
		Entry("with composition mode", true),
	)
})
