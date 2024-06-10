package logopts

import (
	"runtime"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"

	"github.com/mandelsoft/logging"
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

		Expect(GetLogFileFor(opts.LogFileName, fs)).NotTo(BeNil())
		lctx = nil
		for i := 1; i < 100; i++ {
			time.Sleep(1 * time.Millisecond)
			runtime.GC()
		}
		Expect(GetLogFileFor(opts.LogFileName, fs)).To(BeNil())
	})
})
