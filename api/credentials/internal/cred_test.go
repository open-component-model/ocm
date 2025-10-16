package internal_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/extensions/repositories/memory"
	"ocm.software/ocm/api/credentials/internal"
	common "ocm.software/ocm/api/utils/misc"
)

var DefaultContext = credentials.New()

var _ = Describe("generic credentials", func() {
	props := common.Properties{
		"user":     "USER",
		"password": "PASSWORD",
	}
	credmemdata := "{\"credentialsName\":\"cred\",\"repoName\":\"test\",\"type\":\"Memory\"}"
	memdata := "{\"repoName\":\"test\",\"type\":\"Memory\"}"

	_ = props

	It("de/serializes credentials spec", func() {
		repospec := memory.NewRepositorySpec("test")
		credspec := credentials.NewCredentialsSpec("cred", repospec)

		data, err := json.Marshal(credspec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(credmemdata)))

		credspec = &internal.DefaultCredentialsSpec{}
		err = json.Unmarshal(data, credspec)
		Expect(err).To(Succeed())
		s := credspec.(*internal.DefaultCredentialsSpec)
		Expect(reflect.TypeOf(s.RepositorySpec).String()).To(Equal("*memory.RepositorySpec"))
		Expect(s.CredentialsName).To(Equal("cred"))
		Expect(s.RepositorySpec.(*memory.RepositorySpec).RepositoryName).To(Equal("test"))
	})

	It("de/serializes generic credentials spec", func() {
		credspec := &internal.GenericCredentialsSpec{}

		err := json.Unmarshal([]byte(credmemdata), credspec)
		Expect(err).To(Succeed())

		data, err := json.Marshal(credspec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(credmemdata)))
	})

	It("de/serializes generic repository spec", func() {
		credspec := &internal.GenericRepositorySpec{}

		err := json.Unmarshal([]byte(memdata), credspec)
		Expect(err).To(Succeed())

		data, err := json.Marshal(credspec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(memdata)))
	})

	It("converts credentials spec to generic ones", func() {
		repospec := memory.NewRepositorySpec("test")
		credspec := credentials.NewCredentialsSpec("cred", repospec)
		data, err := json.Marshal(credspec)
		Expect(err).To(Succeed())

		gen, err := credentials.ToGenericCredentialsSpec(credspec)
		Expect(err).To(Succeed())

		Expect(reflect.TypeOf(gen).String()).To(Equal("*internal.GenericCredentialsSpec"))
		Expect(reflect.TypeOf(gen.RepositorySpec).String()).To(Equal("*internal.GenericRepositorySpec"))

		gen2, err := credentials.ToGenericCredentialsSpec(gen)
		Expect(err).To(Succeed())
		Expect(gen2).To(BeIdenticalTo(gen))

		data3, err := json.Marshal(gen)
		Expect(err).To(Succeed())
		Expect(data3).To(Equal(data))
	})
})
