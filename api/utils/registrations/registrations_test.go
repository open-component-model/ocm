package registrations_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/registrations"
)

type Target interface{}

type Option struct{}

////////////////////////////////////////////////////////////////////////////////
// Test instantiation of generic type

type (
	HandlerRegistrationRegistry = registrations.HandlerRegistrationRegistry[Target, Option]
	HandlerRegistrationHandler  = registrations.HandlerRegistrationHandler[Target, Option]
	RegistrationHandlerInfo     = registrations.RegistrationHandlerInfo[Target, Option]
)

func NewHandlerRegistrationRegistry(base ...HandlerRegistrationRegistry) HandlerRegistrationRegistry {
	return registrations.NewHandlerRegistrationRegistry[Target, Option](base...)
}

func NewRegistrationHandlerInfo(path string, handler HandlerRegistrationHandler) *RegistrationHandlerInfo {
	return registrations.NewRegistrationHandlerInfo[Target, Option](path, handler)
}

////////////////////////////////////////////////////////////////////////////////

type Handler struct{}

func (b Handler) Do() {
}

type TestRegistrationHandler struct {
	name       string
	registered map[string]interface{}
}

func NewTestRegistrationHandler(name string) *TestRegistrationHandler {
	return &TestRegistrationHandler{
		name:       name,
		registered: map[string]interface{}{},
	}
}

func (t *TestRegistrationHandler) RegisterByName(handler string, target Target, config registrations.HandlerConfig, opts ...Option) (bool, error) {
	path := registrations.NewNamePath(handler)
	if len(path) < 1 || path[0] != "match" {
		return false, nil
	}
	t.registered[handler] = nil
	return true, nil
}

func (t *TestRegistrationHandler) GetHandlers(target Target) registrations.HandlerInfos {
	infos := registrations.HandlerInfos{}

	for _, n := range utils.StringMapKeys(t.registered) {
		infos = append(infos, registrations.NewLeafHandlerInfo(n, "")...)
	}
	return infos
}

var _ = Describe("handler registry test", func() {
	Context("registration registry", func() {
		var reg HandlerRegistrationRegistry

		var ha *TestRegistrationHandler
		var hab *TestRegistrationHandler
		var habc *TestRegistrationHandler
		var habd *TestRegistrationHandler
		var habe *TestRegistrationHandler
		var hb *TestRegistrationHandler

		BeforeEach(func() {
			reg = NewHandlerRegistrationRegistry()
			ha = NewTestRegistrationHandler("a")
			hab = NewTestRegistrationHandler("a/b")
			habc = NewTestRegistrationHandler("a/b/c")
			habd = NewTestRegistrationHandler("a/b/d")
			habe = NewTestRegistrationHandler("a/b/e")
			hb = NewTestRegistrationHandler("b")
		})

		It("registers ordered prefixes", func() {
			reg.RegisterRegistrationHandler("a", ha)
			reg.RegisterRegistrationHandler("a/b/d", habd)
			reg.RegisterRegistrationHandler("a/b/c", habc)
			reg.RegisterRegistrationHandler("a/b/e", habe)
			reg.RegisterRegistrationHandler("a/b", hab)
			reg.RegisterRegistrationHandler("b", hb)

			Expect(reg.GetRegistrationHandlers("a/b/c")).To(Equal([]*RegistrationHandlerInfo{
				NewRegistrationHandlerInfo("a/b/c", habc),
				NewRegistrationHandlerInfo("a/b", hab),
				NewRegistrationHandlerInfo("a", ha),
			}))
		})

		It("registers ordered prefixes", func() {
			reg.RegisterRegistrationHandler("a", ha)
			reg.RegisterRegistrationHandler("a/b/d", habd)
			reg.RegisterRegistrationHandler("a/b/c", habc)
			reg.RegisterRegistrationHandler("a/b/e", habe)
			reg.RegisterRegistrationHandler("a/b", hab)
			reg.RegisterRegistrationHandler("b", hb)

			_, err := reg.RegisterByName("a/b/c/d", nil, nil)
			MustFailWithMessage(err, "no registration handler found for a/b/c/d")
			Expect(Must(reg.RegisterByName("a/b/c/match/d", nil, nil))).To(BeTrue())
			Expect(Must(reg.RegisterByName("a/b/c/match", nil, nil))).To(BeTrue())

			Expect(ha.registered).To(Equal(map[string]interface{}{}))
			Expect(hb.registered).To(Equal(map[string]interface{}{}))
			Expect(hab.registered).To(Equal(map[string]interface{}{}))
			Expect(habd.registered).To(Equal(map[string]interface{}{}))
			Expect(habe.registered).To(Equal(map[string]interface{}{}))
			Expect(habc.registered).To(Equal(map[string]interface{}{"match/d": nil, "match": nil}))
		})

		It("registers ordered prefixes", func() {
			reg.RegisterRegistrationHandler("a/b/d", habd)
			reg.RegisterRegistrationHandler("a/b/c", habc)
			reg.RegisterRegistrationHandler("a/b/e", habe)
			reg.RegisterRegistrationHandler("a", ha)

			derived := NewHandlerRegistrationRegistry(reg)
			derived.RegisterRegistrationHandler("a/b", hab)
			derived.RegisterRegistrationHandler("b", hb)

			list := derived.GetRegistrationHandlers("a/b/e")
			Expect(list).To(Equal([]*RegistrationHandlerInfo{
				NewRegistrationHandlerInfo("a/b/e", habe),
				NewRegistrationHandlerInfo("a/b", hab),
				NewRegistrationHandlerInfo("a", ha),
			}))

			_, err := reg.RegisterByName("a/b/e/d", nil, nil)
			MustFailWithMessage(err, "no registration handler found for a/b/e/d")
			Expect(Must(reg.RegisterByName("a/b/e/match/d", nil, nil))).To(BeTrue())
			Expect(Must(reg.RegisterByName("a/b/e/match", nil, nil))).To(BeTrue())

			Expect(ha.registered).To(Equal(map[string]interface{}{}))
			Expect(hb.registered).To(Equal(map[string]interface{}{}))
			Expect(hab.registered).To(Equal(map[string]interface{}{}))
			Expect(habd.registered).To(Equal(map[string]interface{}{}))
			Expect(habc.registered).To(Equal(map[string]interface{}{}))
			Expect(habe.registered).To(Equal(map[string]interface{}{"match/d": nil, "match": nil}))
		})
	})
})
