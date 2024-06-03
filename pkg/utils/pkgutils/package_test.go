package pkgutils_test

import (
	"github.com/mandelsoft/goutils/generics"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"
	me "github.com/open-component-model/ocm/pkg/utils/pkgutils"
	"github.com/open-component-model/ocm/pkg/utils/pkgutils/testpackage"
	"reflect"
)

type typ struct{}

var _ = Describe("package tests", func() {
	DescribeTable("determine package type for ", func(typ interface{}) {
		Expect(Must(me.GetPackageName(typ))).To(Equal(reflect.TypeOf(testpackage.MyStruct{}).PkgPath()))
	},
		Entry("struct", &testpackage.MyStruct{}),
		Entry("array", &testpackage.MyArray{}),
		Entry("list", &testpackage.MyList{}),
		Entry("map", &testpackage.MyMap{}),
		Entry("chan", make(testpackage.MyChan)),
		Entry("func", testpackage.MyFunc),
		Entry("func type", generics.TypeOf[testpackage.MyFuncType]()),
		Entry("struct type", generics.TypeOf[testpackage.MyStruct]()),
	)
	It("determine package for caller func", func() {
		Expect(Must(testpackage.MyFunc())).To(Equal(reflect.TypeOf(testpackage.MyStruct{}).PkgPath()))
		Expect(Must(testpackage.MyFunc(1))).To(Equal(reflect.TypeOf(typ{}).PkgPath()))
	})
})
