package genericocireg_test

import (
	"reflect"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi/repocpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/config"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

var _ = Describe("component repository mapping", func() {
	var tempfs vfs.FileSystem

	var ocispec oci.RepositorySpec
	var spec *genericocireg.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		// ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, accessio.ALLOC_REALM))

		ocispec, err = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
		Expect(err).To(Succeed())
		spec = genericocireg.NewRepositorySpec(ocispec, nil)
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("creates a dummy component with configured chunks", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		ctx := ocm.New(datacontext.MODE_EXTENDED)

		cfg := config.New()
		cfg.AddLimit("@test", 5)
		ctx.ConfigContext().ApplyConfig(cfg, "direct")

		repo := finalizer.ClosingWith(&finalize, Must(ctx.RepositoryForSpec(spec)))
		impl := Must(repocpi.GetRepositoryImplementation(repo))
		Expect(reflect.TypeOf(impl).String()).To(Equal("*genericocireg.RepositoryImpl"))

		Expect(impl.(*genericocireg.RepositoryImpl).GetBlobLimit()).To(Equal(int64(5)))
	})
})
