package hash_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/mime"
)

const (
	ARCH     = "/tmp/ca"
	VERSION  = "v1"
	COMP     = "test.de/x"
	COMP2    = "test.de/y"
	PROVIDER = "mandelsoft"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("hash component archive", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION HASH                                                             NORMALIZED FORM
test.de/x v1      37f7f500d87f4b0a8765649f7c047db382e272b73e042805131df57279991b2b [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
`))
	})

	It("normalize component archive v1", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "-O", "-", "-o", "norm")).To(Succeed())
		Expect(buf.String()).To(Equal(`[{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
`))
	})

	It("normalize component archive v2", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "-N", "jsonNormalisation/v2", "-o", "norm")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`{"component":{"componentReferences":[],"name":"test.de/x","provider":{"name":"mandelsoft"},"resources":[],"sources":[],"version":"v1"}}
`))
	})

	It("check hash", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "-o", "yaml")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
---
component: test.de/x
context: []
hash: 37f7f500d87f4b0a8765649f7c047db382e272b73e042805131df57279991b2b
normalized: '[{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]'
version: v1
`))

		h := sha256.Sum256([]byte(`[{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]`))
		Expect(hex.EncodeToString(h[:])).To(Equal("37f7f500d87f4b0a8765649f7c047db382e272b73e042805131df57279991b2b"))
	})

	It("hash component archive with resources", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Resource("test", VERSION, resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
				env.BlobStringData(mime.MIME_TEXT, "testdata")
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT      : test.de/x
VERSION        : v1
HASH           : 9d8fc24cf27d1092f58098286d9f63c6824c2daf739c19789f64c062d1f30cc5
NORMALIZED FORM: [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[[{"digest":[{"hashAlgorithm":"SHA-256"},{"normalisationAlgorithm":"genericBlobDigest/v1"},{"value":"810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"}]},{"extraIdentity":null},{"name":"test"},{"relation":"local"},{"type":"plainText"},{"version":"v1"}]]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
---`))
	})

	It("hash component archive with resources", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Resource("test", VERSION, resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
				env.BlobStringData(mime.MIME_TEXT, "testdata")
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "--actual", "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT      : test.de/x
VERSION        : v1
HASH           : 9d8fc24cf27d1092f58098286d9f63c6824c2daf739c19789f64c062d1f30cc5
NORMALIZED FORM: [{"component":[{"componentReferences":[]},{"name":"test.de/x"},{"provider":"mandelsoft"},{"resources":[[{"digest":[{"hashAlgorithm":"SHA-256"},{"normalisationAlgorithm":"genericBlobDigest/v1"},{"value":"810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"}]},{"extraIdentity":null},{"name":"test"},{"relation":"local"},{"type":"plainText"},{"version":"v1"}]]},{"version":"v1"}]},{"meta":[{"schemaVersion":"v2"}]}]
---`))
	})

	It("hash component archive with v2", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Resource("test", VERSION, resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
				env.BlobStringData(mime.MIME_TEXT, "testdata")
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", ARCH, "-N", compdesc.JsonNormalisationV2, "--actual", "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT      : test.de/x
VERSION        : v1
HASH           : 6e8e9eb0af1c4c0b9dcc4161168b3f0ad913bc85e4234688dd6d4b283fe4b956
NORMALIZED FORM: {"component":{"componentReferences":[],"name":"test.de/x","provider":{"name":"mandelsoft"},"resources":[{"digest":{"hashAlgorithm":"SHA-256","normalisationAlgorithm":"genericBlobDigest/v1","value":"810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"},"name":"test","relation":"local","type":"plainText","version":"v1"}],"sources":[],"version":"v1"}}
---`))
	})

	It("hash component recursively", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Provider(PROVIDER)
			})
			env.ComponentVersion(COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Reference("ref", COMP2, VERSION)
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", "-r", ARCH+"//test.de/x:v1")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
REFERENCEPATH COMPONENT VERSION HASH                                                             IDENTITY
              test.de/x v1      b74cee6c6b8215f470efd0e3c49618bb98610fc80de36a2e121d0550650b9cdc 
test.de/x:v1  test.de/y v1      e60c791a20091abcf8d35742a134b3a99ce811d874fd721870b28ea90ef5ad2a "name"="ref"
`))
	})

	It("hash component recursively", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Provider(PROVIDER)
			})
			env.ComponentVersion(COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Reference("ref", COMP2, VERSION)
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", "-r", "--repo", ARCH, "test.de/x:v1")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
REFERENCEPATH COMPONENT VERSION HASH                                                             IDENTITY
              test.de/x v1      b74cee6c6b8215f470efd0e3c49618bb98610fc80de36a2e121d0550650b9cdc 
test.de/x:v1  test.de/y v1      e60c791a20091abcf8d35742a134b3a99ce811d874fd721870b28ea90ef5ad2a "name"="ref"
`))
	})

	It("hash components recursively", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Provider(PROVIDER)
			})
			env.ComponentVersion(COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Reference("ref", COMP2, VERSION)
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", "-r", ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
REFERENCEPATH COMPONENT VERSION HASH                                                             IDENTITY
              test.de/x v1      b74cee6c6b8215f470efd0e3c49618bb98610fc80de36a2e121d0550650b9cdc 
test.de/x:v1  test.de/y v1      e60c791a20091abcf8d35742a134b3a99ce811d874fd721870b28ea90ef5ad2a "name"="ref"
              test.de/y v1      e60c791a20091abcf8d35742a134b3a99ce811d874fd721870b28ea90ef5ad2a
`))
	})

	It("hash component recursively and updates hashes", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Provider(PROVIDER)
			})
			env.ComponentVersion(COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Reference("ref", COMP2, VERSION)
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("hash", "components", "-r", "--repo", ARCH, "-U", "test.de/x:v1")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
REFERENCEPATH COMPONENT VERSION HASH                                                             IDENTITY
              test.de/x v1      b74cee6c6b8215f470efd0e3c49618bb98610fc80de36a2e121d0550650b9cdc 
test.de/x:v1  test.de/y v1      e60c791a20091abcf8d35742a134b3a99ce811d874fd721870b28ea90ef5ad2a "name"="ref"
`))

		repo := Must(ctf.Open(env, ctf.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")

		cvy := Must(repo.LookupComponentVersion(COMP2, VERSION))
		defer Close(cvy, "cvy")
		data := Must(compdesc.Encode(cvy.GetDescriptor()))
		fmt.Printf("%s:\n%s\n", COMP2, string(data))

		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		defer Close(cv, "cv")

		data = Must(compdesc.Encode(cv.GetDescriptor()))
		fmt.Printf("%s:\n%s\n", COMP, string(data))
		ref := Must(cv.GetReferenceByIndex(0))
		d := ref.GetDigest()
		Expect(d).NotTo(BeNil())
		Expect(d.Value).To(Equal("e60c791a20091abcf8d35742a134b3a99ce811d874fd721870b28ea90ef5ad2a"))
	})
})
