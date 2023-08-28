// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/entrytypes/comment"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const PROVIDER = "acme.org"
const OTHER = "a.company.com"

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

		e1a = Must(routingslip.AddEntry(cv, PROVIDER, rsa.Algorithm, comment.New("first entry")))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("gets single entry", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		repo := Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
		finalize.Close(repo)
		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		finalize.Close(cv)
		slip := Must(routingslip.GetSlip(cv, PROVIDER))
		slip.Entries[0].Signature.Value = slip.Entries[0].Signature.Value[1:] + "0"
		MustBeSuccessful(routingslip.SetSlip(cv, slip))
		MustBeSuccessful(finalize.Finalize())

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "routingslip", "-v", "--fail-on-error", ARCH)).To(MatchError("validation failed: for details see output"))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
TYPE    TIMESTAMP                 DESCRIPTION
                                  Error: cannot verify entry ` + e1a.Digest.String() + `: signature verification failed, crypto/rsa: verification error
comment ` + e1a.Timestamp.String() + ` Comment: first entry

`))
	})

	It("detects manipulation", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "routingslip", "-v", ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT-VERSION NAME     TYPE    TIMESTAMP                 DESCRIPTION
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

			e2a = Must(routingslip.AddEntry(cv, OTHER, rsa.Algorithm, comment.New("first other entry")))
			e2b = Must(routingslip.AddEntry(cv, OTHER, rsa.Algorithm, comment.New("second other entry")))

			te := Must(routingslip.NewGenericEntryWith("acme.org/test",
				"name", "unit-tests",
				"status", "passed",
			))
			e2c = Must(routingslip.AddEntry(cv, OTHER, rsa.Algorithm, te))
		})

		It("gets different slips", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "routingslip", ARCH)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NAME          TYPE          TIMESTAMP                 DESCRIPTION
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
TYPE          TIMESTAMP                 DESCRIPTION
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
TYPE          DIGEST   PARENT   TIMESTAMP                 DESCRIPTION
comment       ` + digests(e2a, nil) + ` Comment: first other entry
comment       ` + digests(e2b, e2a) + ` Comment: second other entry
acme.org/test ` + digests(e2c, e2b) + ` name: unit-tests, status: passed
`))
		})

		It("gets dedicated yaml slip", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, "a.company.com", "-ojson")).To(Succeed())
			Expect(len(buf.String())).To(Equal(3485))
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
