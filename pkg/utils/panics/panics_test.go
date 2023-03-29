// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package panics_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/utils/panics"
)

func caller(topanic interface{}, outerr error, handlers ...panics.PanicHandler) (err error) {
	defer panics.PropagatePanicAsError(&err, false, handlers...)

	err = outerr
	callee(topanic)
	return
}

func callee(topanic interface{}) {
	if topanic != nil {
		panic(topanic)
	}
}

var _ = Describe("catch panics", func() {
	It("propagates catched panic", func() {
		defer func() {
			Expect(recover()).To(BeNil())
		}()
		err := caller("exception", nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp(`(?s)Observed a panic: "exception"
goroutine [0-9]* \[running\]:
panic.*$`))
	})

	It("propagates catched panic with handlers", func() {
		defer func() {
			Expect(recover()).To(BeNil())
		}()
		err := caller("exception", nil, func(any) error { return fmt.Errorf("handler") })
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp(`(?s){Observed a panic: "exception"
goroutine [0-9]* \[running\]:
panic.*
, handler}$`))
	})

	It("regular error", func() {
		defer func() {
			Expect(recover()).To(BeNil())
		}()
		err := caller(nil, fmt.Errorf("exception"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("exception"))
	})

})
