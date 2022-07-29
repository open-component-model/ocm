package secretserver_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	local "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/secretserver"
)

var DefaultContext = credentials.New()

var _ = Describe("secret server", func() {
	props := common.Properties{
		"username": "abc",
		"password": "123",
	}

	credentials := `{
	"container_registry": {
		"eu.gcr.io": {
			"username": "abc",
			"password": "123"
		}
	}
}`
	encryptedCredentials := "Uz4mfePXFOUbjUEZnRrnG8sv6oBFmERtiCktBicMtHjGucnCNLekXdRkO0EuUAAeNbvR5/TWBNA0vzTwEzIUJzoXhpfmd32nMIf+tk9MYuYtWof+fYmZzDG3LkGhUXMx"
	encryptionKey := "abcdefghijklmnop"

	specdata := `{"type":"SecretServer","url":"eu.gcr.io","configName":"container_registry","cipher":"PLAINTEXT","key":null}`

	It("serializes repo spec", func() {
		spec := local.NewRepositorySpec("eu.gcr.io", "container_registry", local.Plaintext, nil)
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(specdata)))
	})

	It("deserializes repo spec", func() {
		spec, err := DefaultContext.RepositorySpecForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*secretserver.RepositorySpec"))

		parsedSpec := spec.(*local.RepositorySpec)
		Expect(parsedSpec.URL).To(Equal("eu.gcr.io"))
		Expect(parsedSpec.ConfigName).To(Equal("container_registry"))
		Expect(parsedSpec.Cipher).To(Equal(local.Plaintext))
		Expect(parsedSpec.Key).To(BeNil())
	})

	It("resolves repository", func() {
		repo, err := DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*secretserver.Repository"))
	})

	It("retrieves credentials from unencrypted server", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(credentials))
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		repo := local.NewRepository(svr.URL, "container_registry", local.Plaintext, nil)

		credentials, err := repo.LookupCredentials("eu.gcr.io")
		Expect(err).ToNot(HaveOccurred())

		Expect(credentials.Properties()).To(Equal(props))
	})

	It("retrieves credentials from encrypted server", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			data, err := base64.StdEncoding.DecodeString(encryptedCredentials)
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(data)
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		repo := local.NewRepository(svr.URL, "container_registry", local.AESECB, []byte(encryptionKey))

		credentials, err := repo.LookupCredentials("eu.gcr.io")
		Expect(err).ToNot(HaveOccurred())

		Expect(credentials.Properties()).To(Equal(props))
	})

})
