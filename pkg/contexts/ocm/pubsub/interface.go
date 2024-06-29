package pubsub

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/runtime/descriptivetype"
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
// to a concrete implementation of the  pub/sub adapter.
// to forward events to the described system.
type PubSubSpec interface {
	runtime.VersionedTypedObject

	PubSubMethod(repo cpi.Repository) (PubSubMethod, error)
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
// Pub Sub types. A PubSub types is finally able to
// provide an implementation for notifying a dedicated
// Pub Sub instance.
type TypeScheme descriptivetype.TypeScheme[PubSubSpec, PubSubType]

func NewTypeScheme(base ...TypeScheme) TypeScheme {
	return descriptivetype.NewTypeScheme[PubSubSpec, PubSubType, TypeScheme]("PubSub type", nil, &UnknownPubSubSpec{}, false, base...)
}

func NewStrictTypeScheme(base ...TypeScheme) runtime.VersionedTypeRegistry[PubSubSpec, PubSubType] {
	return descriptivetype.NewTypeScheme[PubSubSpec, PubSubType, TypeScheme]("PubSub type", nil, &UnknownPubSubSpec{}, false, base...)
}

// DefaultTypeScheme contains all globally known PubSub serializer.
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
	return fmt.Sprintf("unknown PubSub method type %q", s.GetType())
}

var _ PubSubSpec = &UnknownPubSubSpec{}
