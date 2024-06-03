package accessio_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/goutils/errors"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

var Retry = accessio.Retry
var RetriableError = accessio.RetriableError

var _ = Describe("retry", func() {
	It("retries to success", func() {
		cnt := 0

		MustBeSuccessful(Retry(10, time.Second, func() error {
			cnt++
			if cnt <= 5 {
				return RetriableError(fmt.Errorf("retriable problem detected"))
			}
			return nil
		}))
		Expect(cnt).To(Equal(6))
	})

	It("retries to failure", func() {
		cnt := 0

		Expect(Retry(10, time.Second, func() error {
			cnt++
			return RetriableError(fmt.Errorf("retriable problem detected"))
		})).To(MatchError("retriable problem detected"))
		Expect(cnt).To(Equal(11))
	})

	It("retries to non-retriable failure", func() {
		cnt := 0

		Expect(Retry(10, time.Second, func() error {
			cnt++
			if cnt <= 5 {
				return errors.Wrapf(RetriableError(fmt.Errorf("retriable problem detected")), "wrapped error")
			}
			return fmt.Errorf("non-problem detected")
		})).To(MatchError("non-problem detected"))
		Expect(cnt).To(Equal(6))
	})
})
