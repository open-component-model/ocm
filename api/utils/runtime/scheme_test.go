package runtime_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

type TType runtime.TypedObjectDecoder[T]

type T interface {
	runtime.TypedObject
	TFunc()
}

type T1 struct {
	runtime.ObjectTypedObject
	T1 string
}

type T2 struct {
	runtime.ObjectTypedObject
	T2 string
}

func (t *T1) TFunc() {}
func (t *T2) TFunc() {}

var (
	T1Decoder = runtime.MustNewDirectDecoder[T](&T1{})
	T2Decoder = runtime.MustNewDirectDecoder[T](&T2{})
)

var (
	t1data = []byte(`{"type":"t1","t1":"v1"}`)
	t2data = []byte(`{"type":"t2","t2":"v2"}`)
)

var (
	t1 = &T1{runtime.ObjectTypedObject{"t1"}, "v1"}
	t2 = &T2{runtime.ObjectTypedObject{"t2"}, "v2"}
)

var _ = Describe("scheme", func() {
	var scheme runtime.Scheme[T, TType]

	BeforeEach(func() {
		scheme = Must(runtime.NewDefaultScheme[T, TType](&runtime.UnstructuredTypedObject{}, false, nil))
		MustBeSuccessful(scheme.RegisterByDecoder("t1", T1Decoder))
	})

	It("decodes object", func() {
		Expect(Must(scheme.Decode(t1data, nil))).To(Equal(t1))
		Expect(scheme.KnownTypeNames()).To(Equal([]string{"t1"}))
		Expect(utils.StringMapKeys(scheme.KnownTypes())).To(Equal([]string{"t1"}))
	})

	It("handles derived scheme", func() {
		derived := Must(runtime.NewDefaultScheme[T, TType](&runtime.UnstructuredTypedObject{}, false, nil, scheme))
		Expect(Must(derived.Decode(t1data, nil))).To(Equal(t1))
	})

	It("extends derived scheme", func() {
		derived := Must(runtime.NewDefaultScheme[T, TType](&runtime.UnstructuredTypedObject{}, false, nil, scheme))
		MustBeSuccessful(derived.RegisterByDecoder("t2", T2Decoder))
		Expect(Must(derived.Decode(t2data, nil))).To(Equal(t2))

		Expect(scheme.KnownTypeNames()).To(Equal([]string{"t1"}))
		Expect(derived.KnownTypeNames()).To(Equal([]string{"t1", "t2"}))

		Expect(utils.StringMapKeys(scheme.KnownTypes())).To(Equal([]string{"t1"}))
		Expect(utils.StringMapKeys(derived.KnownTypes())).To(Equal([]string{"t1", "t2"}))
	})
})
