package logopts

import (
	"runtime"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	loggingopt "ocm.software/ocm/api/utils/cobrautils/logopts/logging"
)

var _ = Describe("log file", func() {
	var fs vfs.FileSystem

	BeforeEach(func() {
		fs = Must(osfs.NewTempFileSystem())
	})

	AfterEach(func() {
		vfs.Cleanup(fs)
	})

	It("closes log file", func() {
		ctx := ocm.New(datacontext.MODE_INITIAL)
		lctx := logging.NewDefault()

		vfsattr.Set(ctx, fs)

		opts := &Options{
			ConfigFragment: ConfigFragment{
				LogLevel:    "debug",
				LogFileName: "debug.log",
			},
		}

		MustBeSuccessful(opts.Configure(ctx, lctx))

		Expect(loggingopt.GetLogFileFor(opts.LogFileName, fs)).NotTo(BeNil())
		lctx = nil
		for i := 1; i < 100; i++ {
			time.Sleep(1 * time.Millisecond)
			runtime.GC()
		}
		Expect(loggingopt.GetLogFileFor(opts.LogFileName, fs)).To(BeNil())
	})
})
