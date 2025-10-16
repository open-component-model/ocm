package config_test

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/extensions/repositories/memory"
	. "ocm.software/ocm/api/helper/env"
	common "ocm.software/ocm/api/utils/misc"
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
