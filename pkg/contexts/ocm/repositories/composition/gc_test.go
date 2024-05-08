package composition_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/refmgmt"
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
