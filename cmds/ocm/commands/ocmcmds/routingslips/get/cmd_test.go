package get_test

import (
	"bytes"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/goutils/finalizer"

	"ocm.software/ocm/api/ocm/extensions/labels/routingslip"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/types/comment"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const (
	ARCH     = "/tmp/ca"
	VERSION  = "v1"
	COMP     = "test.de/x"
	PROVIDER = "acme.org"
	OTHER    = "a.company.com"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var e1a *routingslip.HistoryEntry

	BeforeEach(func() {
		env = NewTestEnv()
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
				})
			})
		})
		env.RSAKeyPair(PROVIDER, OTHER)

		repo := Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		defer Close(cv)

		e1a = Must(routingslip.AddEntry(cv, PROVIDER, rsa.Algorithm, comment.New("first entry"), nil))
		MustBeSuccessful(cv.Update())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("detects manipulation", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		repo := Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
		finalize.Close(repo)
		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		finalize.Close(cv)
		slip := Must(routingslip.GetSlip(cv, PROVIDER))
		slip.Get(0).Signature.Value = slip.Get(0).Signature.Value[1:] + "0"
		MustBeSuccessful(routingslip.SetSlip(cv, slip))
		MustBeSuccessful(cv.Update())
		MustBeSuccessful(finalize.Finalize())

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "routingslip", "-v", "--fail-on-error", ARCH)).To(MatchError("validation failed: for details see output"))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
TYPE    TIMESTAMP            DESCRIPTION
                             Error: cannot verify entry ` + e1a.Digest.String() + `: signature verification failed, crypto/rsa: verification error
comment ` + e1a.Timestamp.String() + ` Comment: first entry

`))
	})

	It("gets single entry", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "routingslip", "-v", ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT-VERSION NAME     TYPE    TIMESTAMP            DESCRIPTION
test.de/x:v1      acme.org comment ` + e1a.Timestamp.String() + ` Comment: first entry

`))
	})

	Context("multiple slips", func() {
		var e2a *routingslip.HistoryEntry
		var e2b *routingslip.HistoryEntry
		var e2c *routingslip.HistoryEntry

		BeforeEach(func() {
			repo := Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMP, VERSION))
			defer Close(cv)

			e2a = Must(routingslip.AddEntry(cv, OTHER, rsa.Algorithm, comment.New("first other entry"), nil))
			e2b = Must(routingslip.AddEntry(cv, OTHER, rsa.Algorithm, comment.New("second other entry"), nil))

			entry := Must(routingslip.NewGenericEntryWith("acme.org/test",
				"name", "unit-tests",
				"status", "passed",
			))
			e2c = Must(routingslip.AddEntry(cv, OTHER, rsa.Algorithm, entry, nil))

			MustBeSuccessful(cv.Update())
		})

		It("gets different slips", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "routingslip", ARCH)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NAME          TYPE          TIMESTAMP            DESCRIPTION
a.company.com comment       ` + e2a.Timestamp.String() + ` Comment: first other entry
a.company.com comment       ` + e2b.Timestamp.String() + ` Comment: second other entry
a.company.com acme.org/test ` + e2c.Timestamp.String() + ` name: unit-tests, status: passed
acme.org      comment       ` + e1a.Timestamp.String() + ` Comment: first entry
`))
		})

		It("gets dedicated slip", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, "a.company.com")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
TYPE          TIMESTAMP            DESCRIPTION
comment       ` + e2a.Timestamp.String() + ` Comment: first other entry
comment       ` + e2b.Timestamp.String() + ` Comment: second other entry
acme.org/test ` + e2c.Timestamp.String() + ` name: unit-tests, status: passed
`))
		})

		It("gets dedicated wide slip", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, "a.company.com", "-owide")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
TYPE          DIGEST   PARENT   TIMESTAMP            LINKS DESCRIPTION
comment       ` + digests(e2a, nil) + `       Comment: first other entry
comment       ` + digests(e2b, e2a) + `       Comment: second other entry
acme.org/test ` + digests(e2c, e2b) + `       name: unit-tests, status: passed
`))
		})

		It("gets dedicated yaml slip", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, "a.company.com", "-ojson")).To(Succeed())
			/*
			   {
			     "items": [
			       {
			         "component": "test.de/x",
			         "version": "v1",
			         "routingSlip": "a.company.com",
			         "entry": {
			           "payload": {
			             "comment": "first other entry",
			             "type": "comment"
			           },
			           "timestamp": "2023-12-19T09:21:10Z",
			           "digest": "sha256:bc27d55d56b83c53cdceb97294cdade29d99354ac1d8a5f7efd0e5cb1a238065",
			         }
			       },
			       {
			         "component": "test.de/x",
			         "version": "v1",
			         "routingSlip": "a.company.com",
			         "entry": {
			           "payload": {
			             "comment": "second other entry",
			             "type": "comment"
			           },
			           "timestamp": "2023-12-19T09:21:10Z",
			           "parent": "sha256:bc27d55d56b83c53cdceb97294cdade29d99354ac1d8a5f7efd0e5cb1a238065",
			           "digest": "sha256:61971501378144a22e03d12b5f61436026ce11ed7b237dd1dc08b680e9f31c8a",
			         }
			       },
			       {
			         "component": "test.de/x",
			         "version": "v1",
			         "routingSlip": "a.company.com",
			         "entry": {
			           "payload": {
			             "name": "unit-tests",
			             "status": "passed",
			             "type": "acme.org/test"
			           },
			           "timestamp": "2023-12-19T09:21:10Z",
			           "parent": "sha256:61971501378144a22e03d12b5f61436026ce11ed7b237dd1dc08b680e9f31c8a",
			           "digest": "sha256:bc61fef066810bb73f51a3811918780736cb8246b1ea5574f5c1c4e12221e7f2",
			           "signature": {
			             "algorithm": "RSASSA-PKCS1-V1_5",
			             "value": "89ef043b6cb61af11f22ca31940473540577d66dc75d6e585160d0e847d0ab086a3623246350a1028df42b2563adaee1407c68ff4c1df873db2a8bab3fcc2efb5bfa46de17ef1b3602b359f28e912ad80e419a5deb12dc6b1573e79613795e1f0baee20b91aa0f32fa0b0150fbf7a4fd4268ea8dba2048ddec0f0b0788ecd7ea6562f2e113643c82702e2221b887078f99153a222038d5042a5e9de92008b9de14f102bc5366abca6e5b9fa1c2bfcbb608e4fda3bc080c658d94852783b8b2e0aa3ef6c428e9e400961d1ff4b5444bc428c2c6f6ad0bf3b39da73f7dc72194c91e32bbe02befb0ed6c27849de26e6c3eca793d8b8f9f6e7a5fa6527f99bce1bb",
			             "mediaType": "application/vnd.ocm.signature.rsa"
			           }
			         }
			       }
			     ]
			   }
			*/
			fmt.Printf("\n%s\n", buf.String())
			Expect(len(buf.String())).To(Equal(2015))
		})

		Context("with links", func() {
			var e2d *routingslip.HistoryEntry

			BeforeEach(func() {
				repo := Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
				defer Close(repo)
				cv := Must(repo.LookupComponentVersion(COMP, VERSION))
				defer Close(cv)

				e2d = Must(routingslip.AddEntry(cv, OTHER, rsa.Algorithm, comment.New("linked entry"), []routingslip.Link{{
					Name:   PROVIDER,
					Digest: e1a.Digest,
				}}))
				MustBeSuccessful(cv.Update())
			})
			It("gets dedicated wide slip with link", func() {
				buf := bytes.NewBuffer(nil)
				Expect(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, "a.company.com", "-owide")).To(Succeed())
				Expect(buf.String()).To(StringEqualTrimmedWithContext(
					`
TYPE          DIGEST   PARENT   TIMESTAMP            LINKS             DESCRIPTION
comment       ` + digests(e2a, nil) + `                   Comment: first other entry
comment       ` + digests(e2b, e2a) + `                   Comment: second other entry
acme.org/test ` + digests(e2c, e2b) + `                   name: unit-tests, status: passed
comment       ` + digests(e2d, e2c) + ` acme.org@` + e1a.Digest.Encoded()[:8] + ` Comment: linked entry
`))
			})
		})
	})
})

func digests(e1, e2 *routingslip.HistoryEntry) string {
	d := ""
	if e2 != nil {
		d = e2.Digest.Encoded()[:8]
	}
	return fmt.Sprintf("%8s %8s %s", e1.Digest.Encoded()[:8], d, e1.Timestamp)
}
