package resolvers_test

import (
	"fmt"
	"strings"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/ocm/resolvers"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const (
	ARCH      = "ctf"
	COMPONENT = "ocm.software/test"
	VERSION   = "1.0.0"
)

var _ = Describe("resolver", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lookup cv per standard resolver", func() {
		// ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, accessio.ALLOC_REALM))

		ctx := ocm.New()

		spec := Must(ctf.NewRepositorySpec(accessobj.ACC_READONLY, ARCH, env))
		ctx.AddResolverRule("ocm.software", spec, 10)

		cv := Must(ctx.GetResolver().LookupComponentVersion(COMPONENT, VERSION))

		/*
			err := cv.Repository().Close()
			if err != nil {
				defer cv.Close()
				Expect(err).To(Succeed())
			}
		*/
		Close(cv)
		Expect(ctx.Finalize()).To(Succeed())
	})

	It("orders resolver rules", func() {
		m := resolvers.NewMatchingResolver(ocm.DefaultContext()).(*internal.MatchingResolver)

		spec1 := ocireg.NewRepositorySpec("host1.org", nil)
		spec2 := ocireg.NewRepositorySpec("host2.org", nil)
		spec3 := ocireg.NewRepositorySpec("host3.org", nil)

		m.AddRule("github.com/open-component-model", spec1, 1)
		m.AddRule("", spec1, 1)
		m.AddRule("github.com", spec1, 1)
		m.AddRule("github.com", spec2, 2)
		m.AddRule("", spec2, 2)
		m.AddRule("github.com/open-component-model", spec2, 2)
		m.AddRule("github.com/open-component-model", spec3, 3)
		m.AddRule("", spec3, 3)
		m.AddRule("github.com", spec3, 3)

		rules := m.GetRules()
		Expect(len(rules)).To(Equal(9))
		Print(rules)
		Check(rules[0], "github.com/open-component-model", spec3, 3)
		Check(rules[1], "github.com", spec3, 3)
		Check(rules[2], "", spec3, 3)

		Check(rules[3], "github.com/open-component-model", spec2, 2)
		Check(rules[4], "github.com", spec2, 2)
		Check(rules[5], "", spec2, 2)

		Check(rules[6], "github.com/open-component-model", spec1, 1)
		Check(rules[7], "github.com", spec1, 1)
		Check(rules[8], "", spec1, 1)
	})
})

func Check(r resolvers.ResolverRule, prefix string, spec ocm.RepositorySpec, prio int) {
	ExpectWithOffset(1, r.GetPrefix()).To(Equal(prefix))
	ExpectWithOffset(1, r.GetPriority()).To(Equal(prio))
	ExpectWithOffset(1, r.GetSpecification()).To(BeIdenticalTo(spec))
}

func Print(rules []resolvers.ResolverRule) {
	list := []string{}
	for _, r := range rules {
		list = append(list, fmt.Sprintf("[%d]%s", r.GetPriority(), r.GetPrefix()))
	}

	fmt.Printf("order: %s\n", strings.Join(list, ", "))
}
