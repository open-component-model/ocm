package gardenerconfig_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/memoryfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	local "ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig"
	gardenercfgcpi "ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig/cpi"
	"ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig/identity"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	ociidentity "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils"
)

var _ = Describe("gardener config", func() {
	containerRegistryCfg := `{
	"container_registry": {
		"test-credentials": {
			"username": "abc",
			"password": "123",
			"image_reference_prefixes": [
				"eu.gcr.io/test-project"
			]
		}
	}
}`
	encryptionKey := "abcdefghijklmnop"
	encryptedContainerRegistryCfg := "Uz4mfePXFOUbjUEZnRrnG8zP2T7lRH6bR2rFHYgWDwZUXfW7D5wArwY4dsBACPVFNapF7kcM9z79+LvJXd2kNoIfvUyMOhrSDAyv4LtUqYSKBOoRH/aJMnXjmN9GQBCXSRSJs/Fu21AoDNo8fA9zYvvc7WxTldkYC/vHxLVNJu5j176e1QiaS9hwDjgNhgyUT3XUjHUyQ19PcRgwDglRLfiL4Cs/fYPPxdg4YZQdCnc="
	expectedCreds := cpi.DirectCredentials{
		cpi.ATTR_USERNAME: "abc",
		cpi.ATTR_PASSWORD: "123",
	}

	repoSpecTemplate := `{"type":"GardenerConfig","url":"%s","configType":"container_registry","cipher":"%s","propagateConsumerIdentity":true}`

	var defaultContext credentials.Context

	BeforeEach(func() {
		defaultContext = credentials.New()
	})

	It("serializes repo spec", func() {
		const (
			url    = "http://localhost:8080/container_registry"
			cipher = local.Plaintext
		)
		expectedSpec := fmt.Sprintf(repoSpecTemplate, url, cipher)

		spec := local.NewRepositorySpec("http://localhost:8080/container_registry", "container_registry", local.Plaintext, true)
		data, err := json.Marshal(spec)
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(Equal([]byte(expectedSpec)))
	})

	It("deserializes repo spec", func() {
		const (
			url    = "http://localhost:8080/container_registry"
			cipher = local.Plaintext
		)
		specdata := fmt.Sprintf(repoSpecTemplate, url, cipher)

		spec, err := defaultContext.RepositorySpecForConfig([]byte(specdata), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*gardenerconfig.RepositorySpec"))

		parsedSpec := spec.(*local.RepositorySpec)
		Expect(parsedSpec.URL).To(Equal(url))
		Expect(parsedSpec.ConfigType).To(Equal(gardenercfgcpi.ContainerRegistry))
		Expect(parsedSpec.Cipher).To(Equal(cipher))
	})

	It("resolves repository", func() {
		const (
			url    = "http://localhost:8080/container_registry"
			cipher = local.Plaintext
		)
		specdata := fmt.Sprintf(repoSpecTemplate, url, cipher)

		repo, err := defaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*gardenerconfig.Repository"))
	})

	It("retrieves credentials from unencrypted server", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(containerRegistryCfg))
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		spec := fmt.Sprintf(repoSpecTemplate, svr.URL, local.Plaintext)

		repo, err := defaultContext.RepositoryForConfig([]byte(spec), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())

		credentialsFromRepo, err := repo.LookupCredentials("test-credentials")
		Expect(err).ToNot(HaveOccurred())
		Expect(credentialsFromRepo).To(Equal(expectedCreds))
	})

	It("propagates credentials with consumer ids in the context", func() {
		expectedConsumerId := cpi.ConsumerIdentity{
			cpi.ID_TYPE:               ociidentity.CONSUMER_TYPE,
			ociidentity.ID_HOSTNAME:   "eu.gcr.io",
			ociidentity.ID_PATHPREFIX: "test-project",
		}

		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(containerRegistryCfg))
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		spec := fmt.Sprintf(repoSpecTemplate, svr.URL, local.Plaintext)

		repo, err := defaultContext.RepositoryForConfig([]byte(spec), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())

		credentialsFromCtx, err := credentials.CredentialsForConsumer(defaultContext, expectedConsumerId)
		Expect(err).ToNot(HaveOccurred())
		Expect(credentialsFromCtx).To(Equal(expectedCreds))
	})

	It("retrieves credentials from encrypted server", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			data, err := base64.StdEncoding.DecodeString(encryptedContainerRegistryCfg)
			Expect(err).ToNot(HaveOccurred())
			_, err = writer.Write(data)
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		parsedURL, err := utils.ParseURL(svr.URL)
		Expect(err).ToNot(HaveOccurred())

		id := cpi.NewConsumerIdentity(identity.CONSUMER_TYPE)
		id.SetNonEmptyValue(identity.ID_HOSTNAME, parsedURL.Host)
		id.SetNonEmptyValue(identity.ID_SCHEME, parsedURL.Scheme)
		id.SetNonEmptyValue(identity.ID_PATHPREFIX, strings.Trim(parsedURL.Path, "/"))
		id.SetNonEmptyValue(identity.ID_PORT, parsedURL.Port())

		creds := credentials.DirectCredentials{
			cpi.ATTR_KEY: encryptionKey,
		}
		defaultContext.SetCredentialsForConsumer(id, creds)

		spec := fmt.Sprintf(repoSpecTemplate, svr.URL, local.AESECB)

		repo, err := defaultContext.RepositoryForConfig([]byte(spec), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())

		credentialsFromRepo, err := repo.LookupCredentials("test-credentials")
		Expect(err).ToNot(HaveOccurred())
		Expect(credentialsFromRepo).To(Equal(expectedCreds))
	})

	It("retrieves credentials from file", func() {
		filename := "/container_registry"
		fs := memoryfs.New()
		vfsattr.Set(defaultContext, fs)

		file, err := fs.Create(filename)
		Expect(err).ToNot(HaveOccurred())

		_, err = file.Write([]byte(containerRegistryCfg))
		Expect(err).ToNot(HaveOccurred())

		err = file.Close()
		Expect(err).ToNot(HaveOccurred())

		spec := fmt.Sprintf(repoSpecTemplate, "file://"+filename, local.Plaintext)

		repo, err := defaultContext.RepositoryForConfig([]byte(spec), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())

		credentialsFromRepo, err := repo.LookupCredentials("test-credentials")
		Expect(err).ToNot(HaveOccurred())
		Expect(credentialsFromRepo).To(Equal(expectedCreds))
	})
})
