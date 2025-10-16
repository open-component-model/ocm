package ocireg_test

import (
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	ocictf "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/pubsub"
	"ocm.software/ocm/api/ocm/extensions/pubsub/providers/ocireg"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/componentmapping"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/runtime"
)

const ARCH = "ctf"

var _ = Describe("Test Environment", func() {
	var env *Builder
	var repo ocm.Repository

	BeforeEach(func() {
		env = NewBuilder()
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory)
		attr := pubsub.For(env)
		attr.ProviderRegistry.Register(ctf.Type, &ocireg.Provider{})
		attr.TypeScheme.Register(pubsub.NewPubSubType[*Spec](Type))
		attr.TypeScheme.Register(pubsub.NewPubSubType[*Spec](TypeV1))

		repo = Must(ctf.Open(env, ctf.ACC_WRITABLE, ARCH, 0o600, env))
	})

	AfterEach(func() {
		if repo != nil {
			MustBeSuccessful(repo.Close())
		}
		env.Cleanup()
	})

	It("set provider", func() {
		MustBeSuccessful(pubsub.SetForRepo(repo, NewSpec("testtarget")))

		repo.Close()
		repo = nil

		ocirepo := Must(ocictf.Open(env, ctf.ACC_WRITABLE, ARCH, 0o600, env))
		defer Close(ocirepo, "ocirepo")
		acc := Must(ocirepo.LookupArtifact(componentmapping.ComponentDescriptorNamespace, ocireg.META))
		defer Close(acc, "ociacc")

		m := acc.ManifestAccess()
		Expect(len(m.GetDescriptor().Layers)).To(Equal(1))

		b := Must(m.GetBlob(m.GetDescriptor().Layers[0].Digest))
		defer Close(b, "blob")
		data := Must(b.Get())
		Expect(string(data)).To(Equal(`{"type":"test","target":"testtarget"}`))
	})

	It("set/get provider", func() {
		MustBeSuccessful(pubsub.SetForRepo(repo, NewSpec("testtarget")))

		spec := Must(pubsub.SpecForRepo(repo))
		Expect(spec).To(YAMLEqual(`{"type":"test","target":"testtarget"}`))
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
