package spiff_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/spiff"
)

var _ = Describe("spiff processing", func() {
	var env *TestEnv
	var ictx inputs.Context
	var info inputs.InputResourceInfo

	nv := common.NewNameVersion("test", "v1")

	BeforeEach(func() {
		info = inputs.InputResourceInfo{
			ComponentVersion: nv,
			ElementName:      "elemname",
			InputFilePath:    "/testdata/dummy",
		}
		env = NewTestEnv(TestData())
		ictx = inputs.NewContext(env.Context, common.NewPrinter(env.Context.StdOut()), nil)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("processes template", func() {
		spec, err := spiff.New("test1.yaml", "", false, nil)
		Expect(err).To(Succeed())
		blob, s, err := spec.GetBlob(ictx, info)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(""))
		data, err := blob.Get()
		Expect(err).To(Succeed())
		Expect("\n" + string(data)).To(Equal(`
alice: 24
bob: 25
`))
	})
	It("processes template with values", func() {
		spec, err := spiff.New("test1.yaml", "", false, map[string]interface{}{"diff": 2})
		Expect(err).To(Succeed())
		blob, s, err := spec.GetBlob(ictx, info)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(""))
		data, err := blob.Get()
		Expect(err).To(Succeed())
		Expect("\n" + string(data)).To(Equal(`
alice: 24
bob: 26
`))
	})
	It("processes template with values with local working directory", func() {
		spec, err := spiff.New("test.yaml", "", false, map[string]interface{}{"diff": 2})
		Expect(err).To(Succeed())
		info.InputFilePath = "/testdata/subdir/dummy"
		blob, s, err := spec.GetBlob(ictx, info)
		Expect(err).To(Succeed())
		Expect(s).To(Equal(""))
		data, err := blob.Get()
		Expect(err).To(Succeed())
		Expect("\n" + string(data)).To(Equal(`
alice: 24
bob: 26
`))
	})
})
