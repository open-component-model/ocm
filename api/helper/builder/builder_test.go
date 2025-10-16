package builder

import (
	"fmt"

	"github.com/mandelsoft/goutils/exception"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		Expect(err).To(MatchError("builder.(*Builder).ExtraIdentity(25): element with metadata required"))
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
		Expect(err).To(MatchError("env.(*Environment).Fail(39): dedicated"))
	})

	It("catches explicit env error", func() {
		err := env.Build(func(e *Builder) {
			env.FailOnErr(fmt.Errorf("dedicated"), "context")
		})
		Expect(err).To(MatchError("env.(*Environment).FailOnErr(46): context: dedicated"))
	})

	It("catches outer error", func() {
		Expect(Build(func(e *Builder) {
			e.ExtraIdentity("a", "b")
		})).To(MatchError("builder.(*Builder).ExtraIdentity(53): element with metadata required"))
	})
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
