// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

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
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
)

var _ = Describe("setup plugin cache", func() {
	var ctx ocm.Context
	var registry plugins.Set

	BeforeEach(func() {
		cache.DirectoryCache.Reset()
		ctx = ocm.New()
		plugindirattr.Set(ctx, "testdata")
		registry = plugincacheattr.Get(ctx)
	})

	It("finds plugin", func() {
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
		Expect(p.GetDescriptor().Short).To(Equal("a test plugin"))
	})

	It("scans only once", func() {
		ctx = ocm.New()
		plugindirattr.Set(ctx, "testdata")
		registry = plugincacheattr.Get(ctx)

		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
		Expect(p.GetDescriptor().Short).To(Equal("a test plugin"))

		Expect(cache.DirectoryCache.Count()).To(Equal(1))
		Expect(cache.DirectoryCache.Requests()).To(Equal(2))
	})

	It("registers access methods", func() {
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
		Expect(len(p.GetDescriptor().AccessMethods)).To(Equal(2))
		Expect(registration.RegisterExtensions(registry.GetContext())).To(Succeed())
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
