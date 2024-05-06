package directory

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	. "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/testutils"
)

var _ = Describe("Input Type", func() {
	var env *InputTest

	var True = true
	var False = false

	BeforeEach(func() {
		env = NewInputTest(TYPE)
	})

	It("simple decode", func() {
		env.Set(options.PathOption, "mypath")
		env.Set(options.CompressOption, "true")
		env.Set(options.MediaTypeOption, "media")
		env.Set(options.PreserveDirOption, "false")
		env.Set(options.FollowSymlinksOption, "true")
		env.Set(options.IncludeOption, "x")
		env.Set(options.ExcludeOption, "a")
		env.Set(options.ExcludeOption, "b")
		env.Check(&Spec{
			MediaFileSpec: cpi.MediaFileSpec{
				PathSpec: cpi.PathSpec{
					Path: "mypath",
				},
				ProcessSpec: cpi.NewProcessSpec("media", true),
			},
			PreserveDir:    &False,
			IncludeFiles:   []string{"x"},
			ExcludeFiles:   []string{"a", "b"},
			FollowSymlinks: &True,
		})
	})

})
