package ociuploadattr_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm"
	me "ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("attribute", func() {
	var ctx ocm.Context
	var cfgctx config.Context

	attr := &me.Attribute{Ref: "ref"}

	BeforeEach(func() {
		cfgctx = config.WithSharedAttributes(datacontext.New(nil)).New()
		credctx := credentials.WithConfigs(cfgctx).New()
		ocictx := oci.WithCredentials(credctx).New()
		ctx = ocm.WithOCIRepositories(ocictx).New()
	})
	It("local setting", func() {
		Expect(me.Get(ctx)).To(BeNil())
		Expect(me.Set(ctx, attr)).To(Succeed())
		Expect(me.Get(ctx)).To(BeIdenticalTo(attr))
	})

	It("global setting", func() {
		Expect(me.Get(cfgctx)).To(BeNil())
		Expect(me.Set(ctx, attr)).To(Succeed())
		Expect(me.Get(ctx)).To(BeIdenticalTo(attr))
	})

	It("parses string", func() {
		Expect(me.AttributeType{}.Decode([]byte("ref"), runtime.DefaultJSONEncoding)).To(Equal(&me.Attribute{Ref: "ref"}))
	})

	It("parses spec", func() {
		spec, err := oci.ToGenericRepositorySpec(ocireg.NewRepositorySpec("ghcr.io"))
		Expect(err).To(Succeed())
		attr := &me.Attribute{
			Repository:      spec,
			NamespacePrefix: "ref",
		}
		data, err := json.Marshal(attr)
		Expect(err).To(Succeed())
		Expect(me.AttributeType{}.Decode(data, runtime.DefaultJSONEncoding)).To(Equal(attr))
	})
})
