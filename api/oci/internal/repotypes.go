package internal

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/modern-go/reflect2"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

type RepositoryType interface {
	runtime.VersionedTypedObjectType[RepositorySpec]
}

type IntermediateRepositorySpecAspect interface {
	IsIntermediate() bool
}

type RepositorySpec interface {
	runtime.VersionedTypedObject

	Name() string
	UniformRepositorySpec() *UniformRepositorySpec
	Repository(Context, credentials.Credentials) (Repository, error)

	Validate(Context, credentials.Credentials, ...credentials.UsageContext) error
}

type (
	RepositorySpecDecoder  = runtime.TypedObjectDecoder[RepositorySpec]
	RepositoryTypeProvider = runtime.KnownTypesProvider[RepositorySpec, RepositoryType]
)

type RepositoryTypeScheme interface {
	runtime.TypeScheme[RepositorySpec, RepositoryType]
}

type _Scheme = runtime.TypeScheme[RepositorySpec, RepositoryType]

type repositoryTypeScheme struct {
	_Scheme
}

func NewRepositoryTypeScheme(defaultDecoder RepositorySpecDecoder, base ...RepositoryTypeScheme) RepositoryTypeScheme {
	scheme := runtime.MustNewDefaultTypeScheme[RepositorySpec, RepositoryType](&UnknownRepositorySpec{}, true, defaultDecoder, utils.Optional(base...))
	return &repositoryTypeScheme{scheme}
}

func NewStrictRepositoryTypeScheme(base ...RepositoryTypeScheme) runtime.VersionedTypeRegistry[RepositorySpec, RepositoryType] {
	scheme := runtime.MustNewDefaultTypeScheme[RepositorySpec, RepositoryType](nil, false, nil, utils.Optional(base...))
	return &repositoryTypeScheme{scheme}
}

func (t *repositoryTypeScheme) KnownTypes() runtime.KnownTypes[RepositorySpec, RepositoryType] {
	return t._Scheme.KnownTypes()
}

// DefaultRepositoryTypeScheme contains all globally known access serializer.
var DefaultRepositoryTypeScheme = NewRepositoryTypeScheme(nil)

func RegisterRepositoryType(atype RepositoryType) {
	DefaultRepositoryTypeScheme.Register(atype)
}

func CreateRepositorySpec(t runtime.TypedObject) (RepositorySpec, error) {
	return DefaultRepositoryTypeScheme.Convert(t)
}

////////////////////////////////////////////////////////////////////////////////

type UnknownRepositorySpec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var (
	_ RepositorySpec  = &UnknownRepositorySpec{}
	_ runtime.Unknown = &UnknownRepositorySpec{}
)

func (r *UnknownRepositorySpec) IsUnknown() bool {
	return true
}

func (r *UnknownRepositorySpec) Name() string {
	return "unknown-" + r.GetKind()
}

func (r *UnknownRepositorySpec) UniformRepositorySpec() *UniformRepositorySpec {
	return UniformRepositorySpecForUnstructured(&r.UnstructuredVersionedTypedObject)
}

func (r *UnknownRepositorySpec) Repository(Context, credentials.Credentials) (Repository, error) {
	return nil, errors.ErrUnknown("repository type", r.GetType())
}

func (r *UnknownRepositorySpec) Validate(Context, credentials.Credentials, ...credentials.UsageContext) error {
	return errors.ErrUnknown("repository type", r.GetType())
}

////////////////////////////////////////////////////////////////////////////////

type GenericRepositorySpec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var _ RepositorySpec = &GenericRepositorySpec{}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	if reflect2.IsNil(spec) {
		return nil, nil
	}
	if g, ok := spec.(*GenericRepositorySpec); ok {
		return g, nil
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return newGenericRepositorySpec(data, runtime.DefaultJSONEncoding)
}

func NewGenericRepositorySpec(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	return generics.CastPointerR[RepositorySpec](newGenericRepositorySpec(data, unmarshaler))
}

func newGenericRepositorySpec(data []byte, unmarshaler runtime.Unmarshaler) (*GenericRepositorySpec, error) {
	unstr := &runtime.UnstructuredVersionedTypedObject{}
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	err := unmarshaler.Unmarshal(data, unstr)
	if err != nil {
		return nil, err
	}
	return &GenericRepositorySpec{*unstr}, nil
}

func (s *GenericRepositorySpec) Name() string {
	return "generic-" + s.GetKind()
}

func (s *GenericRepositorySpec) UniformRepositorySpec() *UniformRepositorySpec {
	return UniformRepositorySpecForUnstructured(&s.UnstructuredVersionedTypedObject)
}

func (s *GenericRepositorySpec) Evaluate(ctx Context) (RepositorySpec, error) {
	raw, err := s.GetRaw()
	if err != nil {
		return nil, err
	}
	return ctx.RepositoryTypes().Decode(raw, runtime.DefaultJSONEncoding)
}

func (s *GenericRepositorySpec) Repository(ctx Context, creds credentials.Credentials) (Repository, error) {
	spec, err := s.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return spec.Repository(ctx, creds)
}

func (s *GenericRepositorySpec) Validate(ctx Context, creds credentials.Credentials, context ...credentials.UsageContext) error {
	spec, err := s.Evaluate(ctx)
	if err != nil {
		return err
	}
	return spec.Validate(ctx, creds, context...)
}

////////////////////////////////////////////////////////////////////////////////
