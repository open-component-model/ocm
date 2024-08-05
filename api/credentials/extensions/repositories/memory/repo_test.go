package memory_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/credentials"
	local "ocm.software/ocm/api/credentials/extensions/repositories/memory"
	common "ocm.software/ocm/api/utils/misc"
)

var DefaultContext = credentials.New()

var _ = Describe("direct credentials", func() {
	props := common.Properties{
		"user":     "USER",
		"password": "PASSWORD",
	}

	props2 := common.Properties{
		"user":     "OTHER",
		"password": "OTHERPASSWORD",
	}

	specdata := "{\"type\":\"Memory\",\"repoName\":\"test\"}"

	_ = props

	It("serializes repo spec", func() {
		spec := local.NewRepositorySpec("test")
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(specdata)))
	})
	It("deserializes repo spec", func() {
		spec, err := DefaultContext.RepositorySpecForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*memory.RepositorySpec"))
		Expect(spec.(*local.RepositorySpec).RepositoryName).To(Equal("test"))
	})

	It("resolves repository", func() {
		repo, err := DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*memory.Repository"))
	})

	It("sets and retrieves credentials", func() {
		repo, err := DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())

		_, err = repo.WriteCredentials("bibo", credentials.NewCredentials(props))
		Expect(err).To(Succeed())

		creds, err := repo.LookupCredentials("bibo")
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(props))

		creds, err = repo.LookupCredentials("other")
		Expect(err).NotTo(Succeed())
		Expect(creds).To(BeNil())
	})

	It("caches repo", func() {
		repo, err := DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())

		_, err = repo.WriteCredentials("bibo", credentials.NewCredentials(props))
		Expect(err).To(Succeed())

		// re-request repo by spec
		repo, err = DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())

		creds, err := repo.LookupCredentials("bibo")
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(props))

		creds, err = repo.LookupCredentials("other")
		Expect(err).NotTo(Succeed())
		Expect(creds).To(BeNil())
	})

	It("caches repo in two contexts", func() {
		ctx1 := DefaultContext
		ctx2 := credentials.New()

		// write to first context
		repo1, err := ctx1.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())

		_, err = repo1.WriteCredentials("bibo", credentials.NewCredentials(props))
		Expect(err).To(Succeed())

		// request repo in second context
		repo2, err := ctx2.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())

		creds, err := repo2.LookupCredentials("bibo")
		Expect(err).NotTo(Succeed())
		Expect(creds).To(BeNil())

		// write to second context
		_, err = repo2.WriteCredentials("bibo", credentials.NewCredentials(props2))
		Expect(err).To(Succeed())

		creds, err = repo2.LookupCredentials("bibo")
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(props2))

		// check first context
		creds, err = repo1.LookupCredentials("bibo")
		Expect(err).To(Succeed())
		Expect(creds.Properties()).To(Equal(props))
	})
})
