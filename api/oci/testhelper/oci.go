package testhelper

import (
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/api/helper/builder"
	"github.com/open-component-model/ocm/api/oci/extensions/repositories/ctf"
	"github.com/open-component-model/ocm/api/utils/accessio"
	"github.com/open-component-model/ocm/api/utils/accessobj"
)

func FakeOCIRepo(env *builder.Builder, path string, host string) string {
	spec, err := ctf.NewRepositorySpec(accessobj.ACC_READONLY, path, accessio.PathFileSystem(env.FileSystem()))
	ExpectWithOffset(1, err).To(Succeed())
	env.OCIContext().SetAlias(host, spec)
	return host + ".alias"
}
