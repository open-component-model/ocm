package builder_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/api/helper/builder"
	. "github.com/open-component-model/ocm/api/ocm/testhelper"

	"github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/localblob"
	"github.com/open-component-model/ocm/api/ocm/extensions/attrs/compositionmodeattr"
	"github.com/open-component-model/ocm/api/ocm/extensions/repositories/ctf"
	ocmutils "github.com/open-component-model/ocm/api/ocm/ocmutils"
	"github.com/open-component-model/ocm/api/utils/accessio"
	"github.com/open-component-model/ocm/api/utils/accessobj"
)

const (
	ARCH      = "/tmp/ctf"
	ARCH2     = "/tmp/ctf2"
	PROVIDER  = "open-component-model"
	VERSION   = "v1"
	COMPONENT = "github.com/open-component-model/test"
	OUT       = "/tmp/res"
)

var _ = Describe("Transfer handler", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()
		compositionmodeattr.Set(env.OCMContext(), true)
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					TestDataResource(env)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("add ocm resource", func() {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))

		Expect(len(cv.GetDescriptor().Resources)).To(Equal(1))

		r := Must(cv.GetResourceByIndex(0))
		a := Must(r.Access())
		Expect(a.GetType()).To(Equal(localblob.Type))

		data := Must(ocmutils.GetResourceData(r))
		Expect(string(data)).To(Equal(S_TESTDATA))
	})
})
