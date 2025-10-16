//go:build unix

package get_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/compdesc/versions/ocm.software/v3alpha1"
	v2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/api/ocm/tools/signing"
)

const (
	COMPONENTA = "acme.org/compa"
	COMPONENTB = "acme.org/compb"
	VERSION    = "v1"
)

const VERIFIED_FILE = "verified.yaml"

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var store signing.VerifiedStore

	BeforeEach(func() {
		env = NewTestEnv()

		store = Must(signing.NewVerifiedStore(VERIFIED_FILE, env))

		cd1 := compdesc.DefaultComponent(&compdesc.ComponentDescriptor{
			Metadata: compdesc.Metadata{
				ConfiguredVersion: v3alpha1.SchemaVersion,
			},
			ComponentSpec: compdesc.ComponentSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:    COMPONENTA,
					Version: VERSION,
					Provider: metav1.Provider{
						Name: "acme.org",
					},
				},
			},
		})
		cd2 := compdesc.DefaultComponent(&compdesc.ComponentDescriptor{
			Metadata: compdesc.Metadata{
				ConfiguredVersion: v2.SchemaVersion,
			},
			ComponentSpec: compdesc.ComponentSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:    COMPONENTB,
					Version: VERSION,
					Provider: metav1.Provider{
						Name: "acme.org",
					},
				},
			},
		})

		store.Add(cd1, "a")
		store.Add(cd2, "b")
		store.Add(cd2, "c")

		MustBeSuccessful(store.Save())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("show verified components", func() {
		var buf bytes.Buffer

		Expect(env.CatchOutput(&buf).Execute("get", "verified", "--verified", VERIFIED_FILE)).To(Succeed())

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
COMPONENT      VERSION
acme.org/compa v1
acme.org/compb v1
`))
	})

	It("show verified components wide", func() {
		var buf bytes.Buffer

		Expect(env.CatchOutput(&buf).Execute("get", "verified", "--verified", VERIFIED_FILE, "-o", "wide")).To(Succeed())

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
COMPONENT      VERSION SIGNATURES
acme.org/compa v1      a
acme.org/compb v1      b, c
`))
	})

	It("show verified dedicated component manifest", func() {
		var buf bytes.Buffer

		Expect(env.CatchOutput(&buf).Execute("get", "verified", "--verified", VERIFIED_FILE, COMPONENTA, "-o", "yaml")).To(Succeed())

		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
---
descriptor:
  component:
    componentReferences: []
    name: acme.org/compa
    provider:
      name: acme.org
    repositoryContexts: []
    resources: []
    sources: []
    version: v1
  meta:
    configuredSchemaVersion: ocm.software/v3alpha1
signatures:
- a
`))
	})
})
