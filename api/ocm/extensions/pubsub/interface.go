package pubsub

import (
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/modern-go/reflect2"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/errkind"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/runtime/descriptivetype"
)

const KIND_PUBSUBTYPE = "pub/sub"

type Option = descriptivetype.Option

func WithFormatSpec(fmt string) Option {
	return descriptivetype.WithFormatSpec(fmt)
}

func WithDesciption(desc string) Option {
	return descriptivetype.WithDescription(desc)
}

////////////////////////////////////////////////////////////////////////////////

type PubSubType descriptivetype.TypedObjectType[PubSubSpec]

// PubSubSpec is the interface publish/subscribe specifications
// must fulfill. The main task is to map the specification
// to a concrete implementation of the  pub/sub adapter
// which forwards events to the described system.
type PubSubSpec interface {
	runtime.VersionedTypedObject

	PubSubMethod(repo cpi.Repository) (PubSubMethod, error)
	Describe(ctx cpi.Context) string
}

type (
	PubSubSpecDecoder  = runtime.TypedObjectDecoder[PubSubSpec]
	PubSubTypeProvider = runtime.KnownTypesProvider[PubSubSpec, PubSubType]
)

// PubSubMethod is the handler able to publish
// an OCM component version event.
type PubSubMethod interface {
	NotifyComponentVersion(version common.NameVersion) error
}

// TypeScheme is the registry for specification types for
// PubSub types. A PubSub type is finally able to
// provide an implementation for notifying a dedicated
// PubSub instance.
type TypeScheme descriptivetype.TypeScheme[PubSubSpec, PubSubType]

func NewTypeScheme(base ...TypeScheme) TypeScheme {
	return descriptivetype.NewTypeScheme[PubSubSpec, PubSubType, TypeScheme]("PubSub type", nil, &UnknownPubSubSpec{}, false, base...)
}

func NewStrictTypeScheme(base ...TypeScheme) runtime.VersionedTypeRegistry[PubSubSpec, PubSubType] {
	return descriptivetype.NewTypeScheme[PubSubSpec, PubSubType, TypeScheme]("PubSub type", nil, &UnknownPubSubSpec{}, false, base...)
}

// DefaultTypeScheme contains all globally known PubSub serializers.
var DefaultTypeScheme = NewTypeScheme()

func RegisterType(atype PubSubType) {
	DefaultTypeScheme.Register(atype)
}

func CreatePubSubSpec(t runtime.TypedObject) (PubSubSpec, error) {
	return DefaultTypeScheme.Convert(t)
}

func NewPubSubType[I PubSubSpec](name string, opts ...Option) PubSubType {
	t := descriptivetype.NewTypedObjectTypeObject[PubSubSpec](runtime.NewVersionedTypedObjectType[PubSubSpec, I](name))
	ta := descriptivetype.NewTypeObjectTarget[PubSubSpec](t)
	optionutils.ApplyOptions[descriptivetype.OptionTarget](ta, opts...)
	return t
}

////////////////////////////////////////////////////////////////////////////////

type UnknownPubSubSpec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var (
	_ runtime.TypedObject = &UnknownPubSubSpec{}
	_ runtime.Unknown     = &UnknownPubSubSpec{}
)

func (_ *UnknownPubSubSpec) IsUnknown() bool {
	return true
}

func (s *UnknownPubSubSpec) PubSubMethod(repository cpi.Repository) (PubSubMethod, error) {
	return nil, errors.ErrUnknown(KIND_PUBSUBTYPE, s.GetType())
}

func (s *UnknownPubSubSpec) Describe(ctx cpi.Context) string {
	return fmt.Sprintf("unknown PubSub specification type %q", s.GetType())
}

var _ PubSubSpec = &UnknownPubSubSpec{}

////////////////////////////////////////////////////////////////////////////////

type Unwrapable interface {
	Unwrap(ctx cpi.Context) []PubSubSpec
}

type Evaluatable interface {
	Evaluate(ctx cpi.Context) (PubSubSpec, error)
}

////////////////////////////////////////////////////////////////////////////////

type GenericPubSubSpec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`

	lock       sync.Mutex
	cached     PubSubSpec
	cachedData []byte
}

var (
	_ PubSubSpec  = &GenericPubSubSpec{}
	_ Unwrapable  = &GenericPubSubSpec{}
	_ Evaluatable = &GenericPubSubSpec{}
)

func ToGenericPubSubSpec(spec PubSubSpec) (*GenericPubSubSpec, error) {
	if reflect2.IsNil(spec) {
		return nil, nil
	}
	if g, ok := spec.(*GenericPubSubSpec); ok {
		return g, nil
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return newGenericPubSubSpec(data, runtime.DefaultJSONEncoding)
}

func NewGenericPubSubSpec(data []byte, unmarshaler ...runtime.Unmarshaler) (PubSubSpec, error) {
	return generics.CastPointerR[PubSubSpec](newGenericPubSubSpec(data, general.Optional(unmarshaler...)))
}

func newGenericPubSubSpec(data []byte, unmarshaler runtime.Unmarshaler) (*GenericPubSubSpec, error) {
	unstr := &runtime.UnstructuredVersionedTypedObject{}
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	err := unmarshaler.Unmarshal(data, unstr)
	if err != nil {
		return nil, err
	}
	return &GenericPubSubSpec{UnstructuredVersionedTypedObject: *unstr}, nil
}

func (s *GenericPubSubSpec) Unwrap(ctx cpi.Context) []PubSubSpec {
	eff, err := s.Evaluate(ctx)
	if err != nil {
		return nil
	}
	if u, ok := eff.(Unwrapable); ok {
		return u.Unwrap(ctx)
	}
	return nil
}

func (s *GenericPubSubSpec) Describe(ctx cpi.Context) string {
	eff, err := s.Evaluate(ctx)
	if err != nil {
		return fmt.Sprintf("invalid access specification: %s", err.Error())
	}
	return eff.Describe(ctx)
}

func (s *GenericPubSubSpec) Evaluate(ctx cpi.Context) (PubSubSpec, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.cached != nil && s.cachedData != nil {
		if d, err := s.GetRaw(); err == nil {
			if slices.Equal(d, s.cachedData) {
				return s.cached, nil
			}
		}
		s.cached = nil
		s.cachedData = nil
	}
	raw, err := s.GetRaw()
	if err != nil {
		return nil, err
	}
	s.cached, err = For(ctx).TypeScheme.Decode(raw, runtime.DefaultJSONEncoding)
	if err == nil {
		s.cachedData = raw
	}
	return s.cached, err
}

func (s *GenericPubSubSpec) PubSubMethod(repository cpi.Repository) (PubSubMethod, error) {
	spec, err := s.Evaluate(repository.GetContext())
	if err != nil {
		return nil, err
	}
	if _, ok := spec.(*GenericPubSubSpec); ok {
		return nil, errors.ErrUnknown(errkind.KIND_ACCESSMETHOD, s.GetType())
	}
	return spec.PubSubMethod(repository)
}
