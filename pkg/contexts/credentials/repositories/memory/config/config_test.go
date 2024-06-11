package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory"
)

var _ = Describe("configure credentials", func() {
	var env *Environment
	var ctx credentials.Context
	var cfg config.Context

	BeforeEach(func() {
		env = NewEnvironment(TestData())
		cfg = config.New()
		ctx = credentials.WithConfigs(cfg).New()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("reads config with ref", func() {
		data, err := vfs.ReadFile(env, "/testdata/creds.yaml")
		Expect(err).To(Succeed())
		_, err = cfg.ApplyData(data, nil, "creds.yaml")
		Expect(err).To(Succeed())

		spec := memory.NewRepositorySpec("default")
		repo, err := ctx.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		mem := repo.(*memory.Repository)
		Expect(mem.ExistsCredentials("ref")).To(BeTrue())
		creds, err := mem.LookupCredentials("ref")
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(common.Properties{"username": "mandelsoft", "password": "specialsecret"}))
	})

	It("reads config with direct", func() {
		data, err := vfs.ReadFile(env, "/testdata/creds.yaml")
		Expect(err).To(Succeed())
		_, err = cfg.ApplyData(data, nil, "creds.yaml")
		Expect(err).To(Succeed())

		spec := memory.NewRepositorySpec("default")
		repo, err := ctx.RepositoryForSpec(spec)
		Expect(err).To(Succeed())
		mem := repo.(*memory.Repository)
		Expect(mem.ExistsCredentials("direct")).To(BeTrue())
		creds, err := mem.LookupCredentials("direct")
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(common.Properties{"username": "mandelsoft2", "password": "specialsecret2"}))
	})
})
