package compound

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	Type   = "compound"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	pubsub.RegisterType(pubsub.NewPubSubType[*Spec](Type))
	pubsub.RegisterType(pubsub.NewPubSubType[*Spec](TypeV1))
}

// Spec provides a pub sub adapter registering events at its provider.
type Spec struct {
	runtime.ObjectVersionedType
	Specifications []*pubsub.GenericPubSubSpec `json:"specifications,omitempty"`
}

var (
	_ pubsub.PubSubSpec = (*Spec)(nil)
	_ pubsub.Unwrapable = (*Spec)(nil)
)

func New(specs ...pubsub.PubSubSpec) (*Spec, error) {
	var gen []*pubsub.GenericPubSubSpec

	for _, s := range specs {
		g, err := pubsub.ToGenericPubSubSpec(s)
		if err != nil {
			return nil, err
		}
		gen = append(gen, g)
	}
	return &Spec{runtime.NewVersionedObjectType(Type), gen}, nil
}

func (s *Spec) PubSubMethod(repo cpi.Repository) (pubsub.PubSubMethod, error) {
	var meths []pubsub.PubSubMethod

	for _, e := range s.Specifications {
		m, err := e.PubSubMethod(repo)
		if err != nil {
			return nil, err
		}
		meths = append(meths, m)
	}
	return &Method{meths}, nil
}

func (s *Spec) Unwrap(ctx cpi.Context) []pubsub.PubSubSpec {
	return sliceutils.Convert[pubsub.PubSubSpec](s.Specifications)
}

func (s *Spec) Describe(_ cpi.Context) string {
	return fmt.Sprintf("compound pub/sub specification with %d emtries", len(s.Specifications))
}

func (s *Spec) Effective() pubsub.PubSubSpec {
	switch len(s.Specifications) {
	case 0:
		return nil
	case 1:
		return s.Specifications[0]
	default:
		return s
	}
}

// Method finally registers events at contained methods.
type Method struct {
	meths []pubsub.PubSubMethod
}

var _ pubsub.PubSubMethod = (*Method)(nil)

func (m *Method) NotifyComponentVersion(version common.NameVersion) error {
	list := errors.ErrList()
	for _, m := range m.meths {
		list.Add(m.NotifyComponentVersion(version))
	}
	return list.Result()
}
