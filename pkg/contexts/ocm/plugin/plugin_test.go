// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
)

var _ = Describe("setup plugin cache", func() {
	var ctx ocm.Context
	var registry cache.Cache

	BeforeEach(func() {
		ctx = ocm.New()
		plugindirattr.Set(ctx, "testdata")
		registry = plugincacheattr.Get(ctx)
	})

	It("finds plugin", func() {
		p := registry.GetPlugin("test")
		Expect(p).NotTo(BeNil())
		Expect(p.GetDescriptor().Short).To(Equal("a test plugin"))
	})

	It("registers access methods", func() {
		p := registry.GetPlugin("test")
		Expect(p).NotTo(BeNil())
		Expect(len(p.GetDescriptor().AccessMethods)).To(Equal(2))
		Expect(registry.RegisterExtensions(nil)).To(Succeed())
		t := ctx.AccessMethods().GetAccessType("test")
		Expect(t).NotTo(BeNil())
		raw := `
type: test
someattr: value
`
		s, err := ctx.AccessSpecForConfig([]byte(raw), nil)
		Expect(err).To(Succeed())
		spec := s.(*access.AccessSpec)
		h := spec.Handler()
		info, err := h.Info(spec)
		Expect(err).To(Succeed())
		Expect(info).To(Equal(&plugin.AccessSpecInfo{
			Short:     "a test",
			MediaType: "plain/text",
			Hint:      "testfile",
			ConsumerId: credentials.ConsumerIdentity{
				"type":     "test",
				"hostname": "localhost",
			},
		}))
	})
})
