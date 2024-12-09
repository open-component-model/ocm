package internal

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/modern-go/reflect2"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/runtime/descriptivetype"
)

type RepositoryType interface {
	descriptivetype.VersionedTypedObjectType[RepositorySpec]
}

type RepositorySpec interface {
	runtime.VersionedTypedObject

	Repository(Context, Credentials) (Repository, error)
}

type (
	RepositorySpecDecoder  = runtime.TypedObjectDecoder[RepositorySpec]
	RepositoryTypeProvider = runtime.KnownTypesProvider[RepositorySpec, RepositoryType]
)

type RepositoryTypeScheme interface {
	descriptivetype.VersionedTypeScheme[RepositorySpec, RepositoryType]
}

type _Scheme = descriptivetype.VersionedTypeScheme[RepositorySpec, RepositoryType]

type repositoryTypeScheme struct {
	_Scheme
}

func NewRepositoryTypeScheme(defaultDecoder RepositorySpecDecoder, base ...RepositoryTypeScheme) RepositoryTypeScheme {
	scheme := descriptivetype.MustNewDefaultVersionedTypeScheme[RepositorySpec, RepositoryType, RepositoryTypeScheme]("Credential provider", nil, &UnknownRepositorySpec{}, true, defaultDecoder, utils.Optional(base...))
	return &repositoryTypeScheme{scheme}
}

func NewStrictRepositoryTypeScheme(base ...RepositoryTypeScheme) runtime.VersionedTypeRegistry[RepositorySpec, RepositoryType] {
	scheme := descriptivetype.MustNewDefaultVersionedTypeScheme[RepositorySpec, RepositoryType, RepositoryTypeScheme]("Credential provider", nil, nil, false, nil, utils.Optional(base...))
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

func (r *UnknownRepositorySpec) Repository(Context, Credentials) (Repository, error) {
	return nil, errors.ErrUnknown("repository type", r.GetType())
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

func (s *GenericRepositorySpec) Evaluate(ctx Context) (RepositorySpec, error) {
	raw, err := s.GetRaw()
	if err != nil {
		return nil, err
	}
	return ctx.RepositoryTypes().Decode(raw, runtime.DefaultJSONEncoding)
}

func (s *GenericRepositorySpec) Repository(ctx Context, creds Credentials) (Repository, error) {
	spec, err := s.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return spec.Repository(ctx, creds)
}

////////////////////////////////////////////////////////////////////////////////
