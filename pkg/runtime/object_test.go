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

package runtime_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var _ = Describe("*** basic types", func() {

	Context("type name", func() {
		It("one arg", func() {
			t := runtime.TypeName("test")
			Expect(t).To(Equal("test"))
		})
		It("two arg", func() {
			t := runtime.TypeName("test", "v1")
			Expect(t).To(Equal("test" + runtime.VersionSeparator + "v1"))
		})
		It("two arg empty", func() {
			t := runtime.TypeName("test", "")
			Expect(t).To(Equal("test"))
		})
		It("two arg", func() {
			defer func() {
				e := recover()
				Expect(e).NotTo(BeNil())
			}()
			runtime.TypeName("test", "v1", "v3")
			Fail("no panic")
		})
	})
	Context("object type", func() {
		It("gets the type", func() {
			t := runtime.NewObjectType("test")
			Expect(t.GetType()).To(Equal("test"))
		})
		It("sets the type", func() {
			t := runtime.NewObjectType("test")
			t.SetType("other")
			Expect(t.GetType()).To(Equal("other"))
		})
	})

	Context("versioned object type", func() {
		It("get type and version of unversioned type", func() {
			t := runtime.NewVersionedObjectType("test", "")
			Expect(t.GetType()).To(Equal("test"))
			Expect(t.GetKind()).To(Equal("test"))
			Expect(t.GetVersion()).To(Equal("v1"))
		})
		It("get type and version of versioned type", func() {
			t := runtime.NewVersionedObjectType("test", "v2")
			Expect(t.GetType()).To(Equal(runtime.TypeName("test", "v2")))
			Expect(t.GetKind()).To(Equal("test"))
			Expect(t.GetVersion()).To(Equal("v2"))
		})

		It("set type", func() {
			t := runtime.NewVersionedObjectType("test", "v2")
			t.SetType(runtime.TypeName("other", "v3"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("other", "v3")))
			Expect(t.GetKind()).To(Equal("other"))
			Expect(t.GetVersion()).To(Equal("v3"))
		})

		It("set kind on unversioned", func() {
			t := runtime.NewVersionedObjectType("test")
			t.SetKind(runtime.TypeName("other"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("other")))
			Expect(t.GetKind()).To(Equal("other"))
			Expect(t.GetVersion()).To(Equal("v1"))
		})
		It("set version on unversioned", func() {
			t := runtime.NewVersionedObjectType("test")
			t.SetVersion(runtime.TypeName("v3"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("test", "v3")))
			Expect(t.GetKind()).To(Equal("test"))
			Expect(t.GetVersion()).To(Equal("v3"))
		})

		It("set kind on versioned", func() {
			t := runtime.NewVersionedObjectType("test", "v2")
			t.SetKind(runtime.TypeName("other"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("other", "v2")))
			Expect(t.GetKind()).To(Equal("other"))
			Expect(t.GetVersion()).To(Equal("v2"))
		})
		It("set version on unversioned", func() {
			t := runtime.NewVersionedObjectType("test", "v2")
			t.SetVersion(runtime.TypeName("v3"))
			Expect(t.GetType()).To(Equal(runtime.TypeName("test", "v3")))
			Expect(t.GetKind()).To(Equal("test"))
			Expect(t.GetVersion()).To(Equal("v3"))
		})
	})
})
