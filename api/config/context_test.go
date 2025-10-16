package config_test

import (
	"encoding/json"
	"reflect"

	"github.com/mandelsoft/goutils/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/datacontext"
)

var _ = Describe("config handling", func() {
	var scheme config.ConfigTypeScheme
	var cfgctx config.Context

	BeforeEach(func() {
		scheme = config.NewConfigTypeScheme()
		cfgctx = config.WithConfigTypeScheme(scheme).New()
		Expect(cfgctx.AttributesContext().GetId()).NotTo(BeIdenticalTo(datacontext.DefaultContext.GetId()))
	})

	It("can deserialize unknown", func() {
		cfg := NewConfig("a", "b")
		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		result, err := cfgctx.GetConfigForData(data, nil)
		Expect(err).To(Succeed())
		Expect(config.IsGeneric(result)).To(BeTrue())
	})

	It("can deserialize known", func() {
		RegisterAt(scheme)

		cfg := NewConfig("a", "b")
		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		result, err := cfgctx.GetConfigForData(data, nil)
		Expect(err).To(Succeed())
		Expect(config.IsGeneric(result)).To(BeFalse())
		Expect(reflect.TypeOf(result).String()).To(Equal("*config_test.Config"))
	})

	It("it applies to existing context", func() {
		RegisterAt(scheme)

		d := newDummy(cfgctx)

		cfg := NewConfig("a", "b")

		err := cfgctx.ApplyConfig(cfg, "test")

		Expect(err).To(Succeed())

		Expect(d.getApplied()).To(Equal([]*Config{cfg}))
	})

	It("it applies to new context", func() {
		RegisterAt(scheme)

		cfg := NewConfig("a", "b")

		err := cfgctx.ApplyConfig(cfg, "test")
		Expect(err).To(Succeed())

		d := newDummy(cfgctx)
		Expect(d.applied).To(Equal([]*Config{cfg}))
	})

	It("it applies generic to new context", func() {
		cfg := NewConfig("a", "b")
		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		gen, err := cfgctx.ApplyData(data, nil, "test")
		Expect(err).To(HaveOccurred())
		Expect(errors.IsErrUnknownKind(err, config.KIND_CONFIGTYPE)).To(BeTrue())
		Expect(config.IsGeneric(gen)).To(BeTrue())

		RegisterAt(scheme)
		d := newDummy(cfgctx)
		Expect(d.getApplied()).To(Equal([]*Config{cfg}))
	})
})
