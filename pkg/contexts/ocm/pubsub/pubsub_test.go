package pubsub_test

import (
	"sync"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"

	"github.com/mandelsoft/goutils/sliceutils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const COMP = "acme.org/component"
const VERS = "v1"

// Provider always provides out test pub sub specification
type Provider struct {
	lock      sync.Mutex
	published sliceutils.Slice[common.NameVersion]
}

var _ pubsub.Provider = (*Provider)(nil)

func (p *Provider) For(repo ocm.Repository) (pubsub.PubSubSpec, error) {
	return &Spec{runtime.NewVersionedObjectType(TYPE), p}, nil
}

const TYPE = "test"

// Spec provides a puc sub adapter registering events at its provider.
type Spec struct {
	runtime.ObjectVersionedType
	Provider *Provider
}

var _ pubsub.PubSubSpec = (*Spec)(nil)

func (s *Spec) PubSubMethod(repo ocm.Repository) (pubsub.PubSubMethod, error) {
	return &Method{s.Provider}, nil
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
		prov = &Provider{}
		// we register our test provider for repository type composition
		pubsub.For(ctx).ProviderRegistry.Register(composition.Type, prov)
		// now, we register ouŕ test pub sub type.
		pubsub.For(ctx).TypeScheme.Register(pubsub.NewPubSubType[*Spec](TYPE))
	})

	Context("", func() {
		It("", func() {
			repo := composition.NewRepository(ctx, "testrepo")
			defer Close(repo)
			cv := composition.NewComponentVersion(ctx, COMP, VERS)
			defer Close(cv)

			Expect(repo.GetSpecification().GetKind()).To(Equal(composition.Type))

			Expect(prov.published).To(BeNil())
			MustBeSuccessful(repo.AddComponentVersion(cv))
			Expect(prov.published).To(ConsistOf(common.VersionedElementKey(cv)))
		})
	})
})
