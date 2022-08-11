// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package errors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/errors"
)

var _ = Describe("errors", func() {
	Context("ErrReadOnly", func() {
		It("identifies kind error", func() {
			uerr := errors.ErrReadOnly("KIND", "obj")

			Expect(errors.IsErrReadOnlyKind(uerr, "KIND")).To(BeTrue())
			Expect(errors.IsErrReadOnlyKind(uerr, "other")).To(BeFalse())

		})
		It("message with elem", func() {
			uerr := errors.ErrReadOnly("KIND", "obj")

			Expect(uerr.Error()).To(Equal("KIND \"obj\" is readonly"))
		})
		It("message without elem", func() {
			uerr := errors.ErrReadOnly()

			Expect(uerr.Error()).To(Equal("readonly"))
		})
	})
	Context("ErrUnkown", func() {
		It("identifies kind error", func() {
			uerr := errors.ErrUnknown("KIND", "obj")

			Expect(errors.IsErrUnknownKind(uerr, "KIND")).To(BeTrue())
			Expect(errors.IsErrUnknownKind(uerr, "other")).To(BeFalse())

		})
		It("find error in history", func() {
			uerr := errors.ErrUnknown("KIND", "obj")
			werr := errors.Wrapf(uerr, "wrapped")

			Expect(errors.IsErrUnknownKind(werr, "KIND")).To(BeTrue())
			Expect(errors.IsErrUnknownKind(werr, "other")).To(BeFalse())
		})
	})

})
