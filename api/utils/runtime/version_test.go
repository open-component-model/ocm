package runtime_test

import (
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type1   = "testType1"
	Type1V1 = Type1 + "/v1"
	Type1V2 = Type1 + "/v2"

	Type2   = "testType2"
	Type2V1 = Type2 + "/v1"
)

var versions runtime.Scheme[TestSpecRealm, TestType]

func init() {
	versions = runtime.MustNewDefaultScheme[TestSpecRealm, TestType](nil, false, nil)

	versions.RegisterByDecoder(Type1, runtime.NewVersionedTypedObjectTypeByConverter[TestSpecRealm, *TestSpec1, *Spec1V1](Type1, &converterSpec1V1{}))
	versions.RegisterByDecoder(Type1V1, runtime.NewVersionedTypedObjectTypeByConverter[TestSpecRealm, *TestSpec1, *Spec1V1](Type1V1, &converterSpec1V1{}))

	versions.RegisterByDecoder(Type2, runtime.NewVersionedTypedObjectType[TestSpecRealm, *TestSpec2](Type2))
	versions.RegisterByDecoder(Type2V1, runtime.NewVersionedTypedObjectType[TestSpecRealm, *TestSpec2](Type2V1))
}

type TestType runtime.TypedObjectDecoder[TestSpecRealm]

// TestSpec is the realm.
type TestSpecRealm interface {
	runtime.VersionedTypedObject
	TestFunction()
}

// TestSpec1 is a first implementation of the realm TestSpec.
// It is used as internal version.
type TestSpec1 struct {
	runtime.InternalVersionedTypedObject[TestSpecRealm]
	Field string `json:"field"`
}

func (a TestSpec1) MarshalJSON() ([]byte, error) {
	return runtime.MarshalVersionedTypedObject(&a)
}

func (a *TestSpec2) TestFunction() {}

func NewTestSpec1(field string) *TestSpec1 {
	return &TestSpec1{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[TestSpecRealm](versions, Type1),
		Field:                        field,
	}
}

// Spec1V1 is an old v1 version of a TestSpec1.
type Spec1V1 struct {
	runtime.ObjectVersionedType
	OldField string `json:"oldField"`
}

type converterSpec1V1 struct{}

var _ runtime.Converter[*TestSpec1, *Spec1V1] = (*converterSpec1V1)(nil)

func (_ converterSpec1V1) ConvertFrom(in *TestSpec1) (*Spec1V1, error) {
	return &Spec1V1{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		OldField:            in.Field,
	}, nil
}

func (_ converterSpec1V1) ConvertTo(in *Spec1V1) (*TestSpec1, error) {
	return &TestSpec1{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[TestSpecRealm](versions, in.Type),
		Field:                        in.OldField,
	}, nil
}

// Spec1V2 is an old v1 version of a TestSpec1.
type Spec1V2 struct {
	runtime.ObjectVersionedType
	Field string `json:"field"`
}

type converterSpec1V2 struct{}

var _ runtime.Converter[*TestSpec1, *Spec1V2] = (*converterSpec1V2)(nil)

func (_ converterSpec1V2) ConvertFrom(in *TestSpec1) (*Spec1V2, error) {
	return &Spec1V2{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		Field:               in.Field,
	}, nil
}

func (_ converterSpec1V2) ConvertTo(in *Spec1V2) (*TestSpec1, error) {
	return &TestSpec1{
		InternalVersionedTypedObject: runtime.NewInternalVersionedTypedObject[TestSpecRealm](versions, in.Type),
		Field:                        in.Field,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// TestSpec2 is a second implementation of the realm TestSpec.
// It is used a internal version and v2.
type TestSpec2 struct {
	runtime.VersionedObjectType
	Field string `json:"field"`
}

func (a *TestSpec1) TestFunction() {}

func NewTestSpec2(field string) *TestSpec2 {
	return &TestSpec2{
		VersionedObjectType: runtime.VersionedObjectType{Type2},
		Field:               field,
	}
}

////////////////////////////////////////////////////////////////////////////////

type encoder interface {
	getEncoder() int
}

type object struct{}

func (_ *object) getEncoder() int {
	return 1
}

var _ = Describe("versioned types", func() {
	var versions runtime.Scheme[TestSpecRealm, TestType]

	versions = runtime.MustNewDefaultScheme[TestSpecRealm, TestType](nil, false, nil)

	versions.RegisterByDecoder(Type1, runtime.NewVersionedTypedObjectTypeByConverter[TestSpecRealm, *TestSpec1, *Spec1V1](Type1, &converterSpec1V1{}))
	versions.RegisterByDecoder(Type1V1, runtime.NewVersionedTypedObjectTypeByConverter[TestSpecRealm, *TestSpec1, *Spec1V1](Type1V1, &converterSpec1V1{}))

	versions.RegisterByDecoder(Type2, runtime.NewVersionedTypedObjectType[TestSpecRealm, *TestSpec2](Type2))
	versions.RegisterByDecoder(Type2V1, runtime.NewVersionedTypedObjectType[TestSpecRealm, *TestSpec2](Type2V1))

	It("marshals version for TestSpec1", func() {
		s1 := NewTestSpec1("value")

		data := Must(json.Marshal(s1))
		Expect(string(data)).To(StringEqualWithContext(`{"type":"testType1","oldField":"value"}`))

		spec := Must(versions.Decode(data, runtime.DefaultJSONEncoding))
		Expect(spec).To(Equal(s1))
	})

	It("unmarshal version for TestSpec2", func() {
		s1 := NewTestSpec2("value")

		data := Must(json.Marshal(s1))
		Expect(string(data)).To(StringEqualWithContext(`{"type":"testType2","field":"value"}`))

		spec := Must(versions.Decode(data, runtime.DefaultJSONEncoding))
		Expect(spec).To(Equal(s1))
	})
})
