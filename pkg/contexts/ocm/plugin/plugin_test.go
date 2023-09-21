// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package plugin_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	common2 "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/actions/oci-repository-prepare"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/defaultmerge"
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

	It("executes action", func() {
		p := registry.Get("action")
		Expect(p).NotTo(BeNil())
		Expect(p.GetDescriptor().Short).To(Equal("a test plugin"))

		r := Must(p.Action(oci_repository_prepare.Spec("ghcr.io", "mandelsoft"), nil))
		Expect(r).To(Equal(oci_repository_prepare.Result("all good")))
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
		t := ctx.AccessMethods().GetType("test")
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

	Context("inexpensive identity method", func() {
		It("inexpensive identity method compatibility test", func() {
			cv := &cpi.DummyComponentVersionAccess{Context: ctx}
			p := registry.Get("test")
			Expect(p).NotTo(BeNil())
			Expect(len(p.GetDescriptor().AccessMethods)).To(Equal(2))
			Expect(registration.RegisterExtensions(registry.GetContext())).To(Succeed())
			t := ctx.AccessMethods().GetType("test")
			Expect(t).NotTo(BeNil())

			raw := `type: test`
			s, err := ctx.AccessSpecForConfig([]byte(raw), nil)
			Expect(err).To(Succeed())
			spec := s.(*access.AccessSpec)
			Expect(spec.GetInexpensiveContentVersionIdentity(cv)).To(Equal(""))
		})

		It("check inexpensive identity method", func() {
			cv := &cpi.DummyComponentVersionAccess{Context: ctx}
			p := registry.Get("identity")
			Expect(p).NotTo(BeNil())
			Expect(len(p.GetDescriptor().AccessMethods)).To(Equal(1))
			Expect(registration.RegisterExtensions(registry.GetContext())).To(Succeed())
			t := ctx.AccessMethods().GetType("identity")
			Expect(t).NotTo(BeNil())

			raw := `type: identity`
			s, err := ctx.AccessSpecForConfig([]byte(raw), nil)
			Expect(err).To(Succeed())
			spec := s.(*access.AccessSpec)
			Expect(spec.GetInexpensiveContentVersionIdentity(cv)).To(Equal("testidentity"))
		})
	})

	Context("valuemergehandler", func() {
		It("finds plugin", func() {
			p := registry.Get("merge")
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
			Expect(p.IsValid()).To(BeTrue())
			Expect(p.GetDescriptor().Short).To(Equal("a test plugin"))
			Expect(len(p.GetDescriptor().ValueMergeHandlers)).To(Equal(1))
		})

		It("merges", func() {
			p := registry.Get("merge")
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
			Expect(p.IsValid()).To(BeTrue())
			spec := Must(valuemergehandler.NewSpecification("acme.org/test", defaultmerge.NewConfig("test")))

			var local, inbound valuemergehandler.Value
			local.SetValue("local")
			inbound.SetValue("local")
			ok, r := Must2(p.MergeValue(spec, local, inbound))
			Expect(ok).To(BeTrue())
			Expect(r.RawMessage).To(YAMLEqual(`{"mode":"resolved"}`))
		})

		It("provider merge specs", func() {
			p := registry.Get("merge")
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
			Expect(p.IsValid()).To(BeTrue())
			s := p.GetLabelMergeSpecification("testlabel", "v1")
			Expect(s).NotTo(BeNil())
			Expect(s.GetDescription()).To(Equal("generic testlabel merge spec"))
			Expect(s.Algorithm).To(Equal("default"))
			s = p.GetLabelMergeSpecification("testlabel", "v2")
			Expect(s).NotTo(BeNil())
			Expect(s.GetDescription()).To(Equal("v2 testlabel merge spec"))
			Expect(s.Algorithm).To(Equal("simpleMapMerge"))
		})

		It("described plugin", func() {
			p := registry.Get("merge")
			Expect(p).NotTo(BeNil())
			pr, buf := common2.NewBufferedPrinter()
			common.DescribePluginDescriptor(nil, p.GetDescriptor(), pr)
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Plugin Name:      merge
Plugin Version:   v1
Capabilities:     Value Merge Handlers, Label Merge Specifications
Description: 
      a test plugin with value merge algorithm acme.org/test

Value Merge Handlers:
- Name: acme.org/test
    test merger

Label Merge Specifications:
- Name: testlabel
  Algorithm: default
    generic testlabel merge spec
- Name: testlabel@v2
  Algorithm: simpleMapMerge
    v2 testlabel merge spec
`))
		})
	})
})
