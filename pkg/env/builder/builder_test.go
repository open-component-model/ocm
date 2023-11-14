// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/exception"
)

var _ = Describe("Builder", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("catches builder error", func() {
		err := env.Build(func(e *Builder) {
			e.ExtraIdentity("a", "b")
		})
		Expect(err).To(MatchError("builder.(*Builder).ExtraIdentity(29): element with metadata required"))
	})

	It("catches explicit error", func() {
		err := env.Build(func(e *Builder) {
			exception.Throw(fmt.Errorf("dedicated"))
		})
		Expect(err).To(MatchError("dedicated"))
	})

	It("catches explicit env error", func() {
		err := env.Build(func(e *Builder) {
			env.Fail("dedicated")
		})
		Expect(err).To(MatchError("env.(*Environment).Fail(43): dedicated"))
	})

	It("catches explicit env error", func() {
		err := env.Build(func(e *Builder) {
			env.FailOnErr(fmt.Errorf("dedicated"), "context")
		})
		Expect(err).To(MatchError("env.(*Environment).FailOnErr(50): context: dedicated"))
	})

	It("catches outer error", func() {
		Expect(Build(func(e *Builder) {
			e.ExtraIdentity("a", "b")
		})).To(MatchError("builder.(*Builder).ExtraIdentity(57): element with metadata required"))
	})

	/*
		It("catches outer error", func() {
			NewBuilder().ExtraIdentity("a", "b")
		})
	*/
})

func Build(funcs ...func(e *Builder)) (err error) {
	env := New()
	defer env.Cleanup()
	defer env.PropagateError(&err)
	for _, f := range funcs {
		f(env)
	}
	return nil
}
