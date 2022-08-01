package cc_config_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	local "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/cc_config"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/errors"
)

var _ = Describe("secret server", func() {
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

	specdata := `{"type":"CCConfig","url":"localhost:8080/container_registry","consumerType":"OCIRegistry","cipher":"PLAINTEXT","key":null,"propagate":true}`

	var DefaultContext credentials.Context

	BeforeEach(func() {
		DefaultContext = credentials.New()
	})

	It("serializes repo spec", func() {
		spec := local.NewRepositorySpec("localhost:8080/container_registry", identity.CONSUMER_TYPE, local.Plaintext, nil, true)
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(data).To(Equal([]byte(specdata)))
	})

	It("deserializes repo spec", func() {
		spec, err := DefaultContext.RepositorySpecForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*cc_config.RepositorySpec"))

		parsedSpec := spec.(*local.RepositorySpec)
		Expect(parsedSpec.URL).To(Equal("localhost:8080/container_registry"))
		Expect(parsedSpec.ConsumerType).To(Equal(identity.CONSUMER_TYPE))
		Expect(parsedSpec.Cipher).To(Equal(local.Plaintext))
		Expect(parsedSpec.Key).To(BeNil())
	})

	It("resolves repository", func() {
		repo, err := DefaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).To(Succeed())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*cc_config.Repository"))
	})

	It("retrieves credentials from unencrypted server", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(creds))
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		repo := local.NewRepository(
			DefaultContext,
			svr.URL+"/container_registry",
			identity.CONSUMER_TYPE,
			local.Plaintext,
			nil,
			true,
			nil,
		)

		credentials, err := repo.LookupCredentials("eu.gcr.io/test-project")
		Expect(err).ToNot(HaveOccurred())
		Expect(credentials.Properties()).To(Equal(props))

		credentials, err = repo.LookupCredentials("eu.gcr.io/test-project/my-image:1.0.0")
		Expect(err).ToNot(HaveOccurred())
		Expect(credentials.Properties()).To(Equal(props))

		credentials, err = repo.LookupCredentials("eu.gcr.io")
		Expect(err).To(HaveOccurred())
		Expect(errors.IsErrUnknown(err)).To(BeTrue())
		Expect(credentials).To(BeNil())
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

		repo := local.NewRepository(
			DefaultContext,
			svr.URL+"/container_registry",
			identity.CONSUMER_TYPE,
			local.AESECB,
			[]byte(encryptionKey),
			true,
			nil,
		)

		credentials, err := repo.LookupCredentials("eu.gcr.io/test-project")
		Expect(err).ToNot(HaveOccurred())

		Expect(credentials.Properties()).To(Equal(props))
	})

	It("retrieves credentials from file", func() {
		filename := "/container_registry"
		fs := memoryfs.New()
		file, err := fs.Create(filename)
		Expect(err).ToNot(HaveOccurred())

		_, err = file.Write([]byte(creds))
		Expect(err).ToNot(HaveOccurred())

		err = file.Close()
		Expect(err).ToNot(HaveOccurred())

		repo := local.NewRepository(
			DefaultContext,
			"file://"+filename,
			identity.CONSUMER_TYPE,
			local.Plaintext,
			nil,
			true,
			fs,
		)

		credentials, err := repo.LookupCredentials("eu.gcr.io/test-project")
		Expect(err).ToNot(HaveOccurred())

		Expect(credentials.Properties()).To(Equal(props))
	})

})
