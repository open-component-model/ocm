package gardenerconfig_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	local "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig"
	gardenercfg_cpi "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
)

var _ = Describe("gardener config", func() {
	props := common.Properties{
		"username": "abc",
		"password": "123",
	}

	creds := `{
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
	encryptedCredentials := "Uz4mfePXFOUbjUEZnRrnG8zP2T7lRH6bR2rFHYgWDwZUXfW7D5wArwY4dsBACPVFNapF7kcM9z79+LvJXd2kNoIfvUyMOhrSDAyv4LtUqYSKBOoRH/aJMnXjmN9GQBCXSRSJs/Fu21AoDNo8fA9zYvvc7WxTldkYC/vHxLVNJu5j176e1QiaS9hwDjgNhgyUT3XUjHUyQ19PcRgwDglRLfiL4Cs/fYPPxdg4YZQdCnc="

	specdata := `{"type":"GardenerConfig","url":"http://localhost:8080/container_registry","configType":"container_registry","cipher":"PLAINTEXT","key":null,"propagateConsumerIdentity":true}`

	var defaultContext credentials.Context

	BeforeEach(func() {
		defaultContext = credentials.New()
	})

	It("serializes repo spec", func() {
		spec := local.NewRepositorySpec("http://localhost:8080/container_registry", "container_registry", local.Plaintext, nil, true)
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(specdata)))
	})

	It("deserializes repo spec", func() {
		spec, err := defaultContext.RepositorySpecForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*gardenerconfig.RepositorySpec"))

		parsedSpec := spec.(*local.RepositorySpec)
		Expect(parsedSpec.URL).To(Equal("http://localhost:8080/container_registry"))
		Expect(parsedSpec.ConfigType).To(Equal(gardenercfg_cpi.ContainerRegistry))
		Expect(parsedSpec.Cipher).To(Equal(local.Plaintext))
		Expect(parsedSpec.Key).To(BeNil())
	})

	It("resolves repository", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(creds))
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		specdata := fmt.Sprintf(`{"type":"GardenerConfig","url":"%s/container_registry","configType":"container_registry","cipher":"PLAINTEXT","key":null,"propagateConsumerIdentity":true}`, svr.URL)

		repo, err := defaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*gardenerconfig.Repository"))
	})

	It("retrieves credentials from unencrypted server", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(creds))
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		repo, err := local.NewRepository(
			defaultContext,
			svr.URL+"/container_registry",
			gardenercfg_cpi.ContainerRegistry,
			local.Plaintext,
			nil,
			true,
		)
		Expect(err).ToNot(HaveOccurred())

		credentials, err := repo.LookupCredentials("test-credentials")
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

		repo, err := local.NewRepository(
			defaultContext,
			svr.URL+"/container_registry",
			gardenercfg_cpi.ContainerRegistry,
			local.AESECB,
			[]byte(encryptionKey),
			true,
		)
		Expect(err).ToNot(HaveOccurred())

		credentials, err := repo.LookupCredentials("test-credentials")
		Expect(err).ToNot(HaveOccurred())

		Expect(credentials.Properties()).To(Equal(props))
	})

	It("retrieves credentials from file", func() {
		filename := "/container_registry"
		fs := memoryfs.New()
		vfsattr.Set(defaultContext, fs)

		file, err := fs.Create(filename)
		Expect(err).ToNot(HaveOccurred())

		_, err = file.Write([]byte(creds))
		Expect(err).ToNot(HaveOccurred())

		err = file.Close()
		Expect(err).ToNot(HaveOccurred())

		repo, err := local.NewRepository(
			defaultContext,
			"file://"+filename,
			gardenercfg_cpi.ContainerRegistry,
			local.Plaintext,
			nil,
			true,
		)
		Expect(err).ToNot(HaveOccurred())

		credentials, err := repo.LookupCredentials("test-credentials")
		Expect(err).ToNot(HaveOccurred())

		Expect(credentials.Properties()).To(Equal(props))
	})

})
