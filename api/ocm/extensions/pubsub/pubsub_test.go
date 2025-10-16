package pubsub_test

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mandelsoft/goutils/sliceutils"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/pubsub"
	"ocm.software/ocm/api/ocm/extensions/pubsub/types/compound"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	COMP = "acme.org/component"
	VERS = "v1"
)

// Provider always provides out test pub sub specification
type Provider struct {
	lock      sync.Mutex
	settings  map[string]pubsub.PubSubSpec
	published sliceutils.Slice[common.NameVersion]
}

var _ pubsub.Provider = (*Provider)(nil)

func NewProvider() *Provider {
	return &Provider{settings: map[string]pubsub.PubSubSpec{}}
}

func (p *Provider) GetPubSubSpec(repo ocm.Repository) (pubsub.PubSubSpec, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	s := p.settings[key(repo)]
	if s != nil {
		p.set(repo.GetContext(), s)
	}
	return s, nil
}

func (p *Provider) SetPubSubSpec(repo cpi.Repository, spec pubsub.PubSubSpec) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.settings[key(repo)] = spec
	return nil
}

func (p *Provider) set(ctx cpi.Context, s pubsub.PubSubSpec) {
	if e, ok := s.(pubsub.Evaluatable); ok {
		eff, err := e.Evaluate(ctx)
		if err == nil {
			s = eff
		}
	}

	if m, ok := s.(*Spec); ok {
		m.provider = p
	}
	if u, ok := s.(pubsub.Unwrapable); ok {
		for _, n := range u.Unwrap(ctx) {
			p.set(ctx, n)
		}
	}
}

func key(repo cpi.Repository) string {
	s := repo.GetSpecification()
	d, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(d)
}

const TYPE = "test"

// Spec provides a pub sub adapter registering events at its provider.
type Spec struct {
	runtime.ObjectVersionedType
	provider *Provider
}

var _ pubsub.PubSubSpec = (*Spec)(nil)

func NewSpec() pubsub.PubSubSpec {
	return &Spec{runtime.NewVersionedObjectType(TYPE), nil}
}

func (s *Spec) PubSubMethod(repo ocm.Repository) (pubsub.PubSubMethod, error) {
	return &Method{s.provider}, nil
}

func (s *Spec) Describe(ctx ocm.Context) string {
	return fmt.Sprintf("pub/sub spec of kind %q", s.GetKind())
}

// Method finally registers events at its provider.
type Method struct {
	provider *Provider
}

var _ pubsub.PubSubMethod = (*Method)(nil)

func (m *Method) NotifyComponentVersion(version common.NameVersion) error {
	m.provider.lock.Lock()
	defer m.provider.lock.Unlock()

	m.provider.published.Add(version)
	return nil
}

var _ = Describe("Pub SubTest Environment", func() {
	var ctx ocm.Context
	var prov *Provider

	BeforeEach(func() {
		ctx = ocm.New(datacontext.MODE_CONFIGURED)
		prov = NewProvider()
		// we register our test provider for repository type composition
		pubsub.For(ctx).ProviderRegistry.Register(composition.Type, prov)
		// now, we register ouŕ test pub sub type.
		pubsub.For(ctx).TypeScheme.Register(pubsub.NewPubSubType[*Spec](TYPE))
	})

	It("direct setting", func() {
		repo := composition.NewRepository(ctx, "testrepo")
		defer Close(repo)

		pubsub.SetForRepo(repo, NewSpec())

		cv := composition.NewComponentVersion(ctx, COMP, VERS)
		defer Close(cv)

		Expect(repo.GetSpecification().GetKind()).To(Equal(composition.Type))

		Expect(prov.published).To(BeNil())
		MustBeSuccessful(repo.AddComponentVersion(cv))
		Expect(prov.published).To(ConsistOf(common.VersionedElementKey(cv)))
	})

	It("indirect setting", func() {
		repo := composition.NewRepository(ctx, "testrepo")
		defer Close(repo)

		pubsub.SetForRepo(repo, Must(compound.New(NewSpec())))

		cv := composition.NewComponentVersion(ctx, COMP, VERS)
		defer Close(cv)

		Expect(repo.GetSpecification().GetKind()).To(Equal(composition.Type))

		Expect(prov.published).To(BeNil())
		MustBeSuccessful(repo.AddComponentVersion(cv))
		Expect(prov.published).To(ConsistOf(common.VersionedElementKey(cv)))
	})
})
