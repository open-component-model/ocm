package filelock_test

import (
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/filelock"
)

var _ = Describe("lock identity", func() {
	It("identity", func() {
		l1 := Must(filelock.MutexFor("testdata/lock"))
		l2 := Must(filelock.MutexFor("testdata/../testdata/lock"))
		Expect(l1).To(BeIdenticalTo(l2))

		Expect(filepath.Base(l1.Path())).To(Equal("lock"))
	})

	It("try lock", func() {
		l := Must(filelock.MutexFor("testdata/lock"))

		c := Must(l.Lock())
		ExpectError(l.TryLock()).To(BeNil())
		c.Close()
		ExpectError(c.Close()).To(BeIdenticalTo(os.ErrClosed))
		c = Must(l.TryLock())
		Expect(c).NotTo(BeNil())
		c.Close()
	})
})
