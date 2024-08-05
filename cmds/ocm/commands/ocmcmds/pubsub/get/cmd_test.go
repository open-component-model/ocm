package get_test

import (
	"bytes"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/pubsub"
	"ocm.software/ocm/api/ocm/extensions/pubsub/providers/ocireg"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ARCH  = "ctf"
	ARCH2 = "ctf2"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory)
		env.OCMCommonTransport(ARCH2, accessio.FormatDirectory)
		attr := pubsub.For(env)
		attr.ProviderRegistry.Register(ctf.Type, &ocireg.Provider{})
		attr.TypeScheme.Register(pubsub.NewPubSubType[*Spec](Type))
		attr.TypeScheme.Register(pubsub.NewPubSubType[*Spec](TypeV1))

		repo := Must(ctf.Open(env, ctf.ACC_WRITABLE, ARCH, 0o600, env))
		defer repo.Close()
		MustBeSuccessful(pubsub.SetForRepo(repo, NewSpec("testtarget")))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get pubsub", func() {
		var buf bytes.Buffer

		MustBeSuccessful(env.CatchOutput(&buf).Execute("get", "pubsub", ARCH))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
REPOSITORY PUBSUBTYPE ERROR
ctf        test       
`))
	})

	It("get pubsub list", func() {
		var buf bytes.Buffer

		MustBeSuccessful(env.CatchOutput(&buf).Execute("get", "pubsub", ARCH, ARCH2, "ARCH2"))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
REPOSITORY PUBSUBTYPE ERROR
ctf        test       
ctf2       -          
ARCH2                 repository "ARCH2" is unknown
`))
	})

	It("get pubsub yaml", func() {
		var buf bytes.Buffer

		MustBeSuccessful(env.CatchOutput(&buf).Execute("get", "pubsub", ARCH, "-o", "yaml"))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
---
pubsub:
  target: testtarget
  type: test
repository: ctf
`))
	})
})

////////////////////////////////////////////////////////////////////////////////

const (
	Type   = "test"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

type Spec struct {
	runtime.ObjectVersionedType
	Target string `json:"target"`
}

var _ pubsub.PubSubSpec = (*Spec)(nil)

func NewSpec(target string) *Spec {
	return &Spec{runtime.NewVersionedObjectType(Type), target}
}

func (s *Spec) PubSubMethod(repo ocm.Repository) (pubsub.PubSubMethod, error) {
	return nil, nil
}

func (s *Spec) Describe(_ ocm.Context) string {
	return fmt.Sprintf("test pubsub")
}
