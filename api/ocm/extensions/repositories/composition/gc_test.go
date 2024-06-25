package composition_test

import (
	"runtime"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/api/datacontext"
	"github.com/open-component-model/ocm/api/ocm"
	me "github.com/open-component-model/ocm/api/ocm/extensions/repositories/composition"
	"github.com/open-component-model/ocm/api/utils/refmgmt"
)

var _ = Describe("repository", func() {
	It("finalizes with context", func() {
		ctx := ocm.New(datacontext.MODE_EXTENDED)

		repo := me.NewRepository(ctx, "test")
		MustBeSuccessful(repo.Close())

		ctx = nil
		runtime.GC()
		time.Sleep(time.Second)
		runtime.GC()
		time.Sleep(time.Second)
		Expect(refmgmt.ReferenceCount(repo)).To(Equal(0))
	})
})
