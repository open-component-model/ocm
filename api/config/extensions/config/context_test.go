package config_test

import (
	"os"
	"reflect"
	"runtime"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/general"
	"sigs.k8s.io/yaml"

	"ocm.software/ocm/api/config"
	local "ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/datacontext"
)

func CheckRefs(ctx config.Context, n int) {
	runtime.GC()
	time.Sleep(time.Second)
	Expect(datacontext.GetContextRefCount(ctx)).To(Equal(n)) // all temp refs have been finalized
}

var _ = Describe("generic config handling", func() {
	var scheme config.ConfigTypeScheme
	var cfgctx config.Context

	testdataconfig, _ := os.ReadFile("testdata/config.yaml")
	testdatajson, _ := yaml.YAMLToJSON(testdataconfig)

	nesteddataconfig, _ := os.ReadFile("testdata/nested.yaml")

	_ = testdatajson

	BeforeEach(func() {
		scheme = config.NewConfigTypeScheme()
		scheme.AddKnownTypes(config.DefaultContext().ConfigTypes())
		cfgctx = config.WithConfigTypeScheme(scheme).New()
	})

	It("can deserialize config", func() {
		result, err := cfgctx.GetConfigForData(testdataconfig, nil)
		Expect(err).To(Succeed())
		Expect(config.IsGeneric(result)).To(BeFalse())
		Expect(reflect.TypeOf(result).String()).To(Equal("*config.Config"))

		CheckRefs(cfgctx, 1)
	})

	It("it applies to existing context", func() {
		RegisterAt(scheme)
		d := newDummy(cfgctx)

		cfg, err := cfgctx.GetConfigForData(testdataconfig, nil)
		Expect(err).To(Succeed())

		err = cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(Succeed())
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))

		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", ""), NewConfig("", "bob")}))

		CheckRefs(cfgctx, 1)
	})

	It("it applies nested to existing context", func() {
		RegisterAt(scheme)
		d := newDummy(cfgctx)

		cfg, err := cfgctx.GetConfigForData(nesteddataconfig, nil)
		Expect(err).To(Succeed())

		err = cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(Succeed())
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(4)))
		Expect(len(cfgs)).To(Equal(4))

		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", ""), NewConfig("", "bob")}))

		CheckRefs(cfgctx, 1)
	})

	It("it applies unknown type to existing context", func() {
		cfg, err := cfgctx.GetConfigForData(testdataconfig, nil)
		Expect(err).To(Succeed())

		err = cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(StringEqualWithContext("testconfig: config apply errors: {config entry 0--testconfig: config type \"Dummy\" is unknown, config entry 1--testconfig: config type \"Dummy\" is unknown}"))
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))

		RegisterAt(scheme)
		d := newDummy(cfgctx)
		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", ""), NewConfig("", "bob")}))

		CheckRefs(cfgctx, 1)
	})

	It("it applies composed config to existing context", func() {
		RegisterAt(scheme)
		d := newDummy(cfgctx)

		cfg := local.New()

		nested := NewConfig("alice", "")
		cfg.AddConfig(nested)
		nested = NewConfig("", "bob")
		cfg.AddConfig(nested)

		err := cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(Succeed())
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))

		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", ""), NewConfig("", "bob")}))

		CheckRefs(cfgctx, 1)
	})

	It("it applies composed config set to existing context", func() {
		RegisterAt(scheme)
		d := newDummy(cfgctx)

		cfg := local.New()

		nested := NewConfig("alice", "")
		cfg.AddConfigToSet("test", nested)
		nested = NewConfig("", "bob")
		cfg.AddConfig(nested)

		err := cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(Succeed())
		gen, cfgs := cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(2)))
		Expect(len(cfgs)).To(Equal(2))

		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("", "bob")}))

		err = cfgctx.ApplyConfigSet("test")
		Expect(err).To(Succeed())

		gen, cfgs = cfgctx.GetConfig(config.AllGenerations, nil)
		Expect(gen).To(Equal(int64(3)))
		Expect(len(cfgs)).To(Equal(3))
		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("", "bob"), NewConfig("alice", "")}))

		CheckRefs(cfgctx, 1)
	})

	It("it applies compig to storing target", func() {
		RegisterAt(scheme)
		d := newDummy(cfgctx)

		cfg := NewConfig("alice", "")

		err := cfgctx.ApplyConfig(cfg, "testconfig")
		Expect(err).To(Succeed())

		Expect(d.getApplied()).To(Equal([]*Config{NewConfig("alice", "")}))

		target := dummyTarget{}
		MustBeSuccessful(cfgctx.ApplyTo(0, &target))
		Expect(target.used).NotTo(BeNil())
		Expect(target.used.GetId()).To(Equal(cfgctx.GetId()))

		CheckRefs(cfgctx, general.Conditional(datacontext.MULTI_REF, 2, 1)) // config context stored in target with separate ref
		target.used.GetId()
	})
})
