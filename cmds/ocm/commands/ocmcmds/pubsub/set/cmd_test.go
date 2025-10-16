package set_test

import (
	"bytes"
	"encoding/json"
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

const ARCH = "ctf"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory)

		attr := pubsub.For(env)
		attr.ProviderRegistry.Register(ctf.Type, &ocireg.Provider{})
		attr.TypeScheme.Register(pubsub.NewPubSubType[*Spec](Type))
		attr.TypeScheme.Register(pubsub.NewPubSubType[*Spec](TypeV1))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("sets pubsub", func() {
		var buf bytes.Buffer

		spec := Must(json.Marshal(NewSpec("testtarget")))

		MustBeSuccessful(env.CatchOutput(&buf).Execute("set", "pubsub", ARCH, string(spec)))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
set pubsub spec "test" for repository "ctf"
`))

		repo := Must(ctf.Open(env, ctf.ACC_WRITABLE, ARCH, 0o600, env))
		defer Close(repo)
		raw := Must(pubsub.SpecForRepo(repo))
		Expect(raw).To(YAMLEqual(spec))
	})

	It("removes pubsub for non-existing", func() {
		var buf bytes.Buffer

		MustBeSuccessful(env.CatchOutput(&buf).Execute("set", "pubsub", "-d", ARCH))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
no pubsub spec configured for repository "ctf"
`))
	})

	It("removes pubsub", func() {
		var buf bytes.Buffer

		repo := Must(ctf.Open(env, ctf.ACC_WRITABLE, ARCH, 0o600, env))
		err := pubsub.SetForRepo(repo, NewSpec("testtarget"))
		MustBeSuccessful(repo.Close())
		MustBeSuccessful(err)

		MustBeSuccessful(env.CatchOutput(&buf).Execute("set", "pubsub", "-d", ARCH))
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
removed pubsub spec "test" for repository "ctf"
`))

		repo = Must(ctf.Open(env, ctf.ACC_WRITABLE, ARCH, 0o600, env))
		defer Close(repo)
		Expect(Must(pubsub.SpecForRepo(repo))).To(BeNil())
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
