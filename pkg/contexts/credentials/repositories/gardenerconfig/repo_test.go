package gardenerconfig_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/utils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	local "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig"
	gardenercfgcpi "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig/cpi"
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

	specTemplate := `{"type":"GardenerConfig","url":"%s","configType":"container_registry","cipher":"%s","propagateConsumerIdentity":true}`

	var defaultContext credentials.Context

	BeforeEach(func() {
		defaultContext = credentials.New()
	})

	It("serializes repo spec", func() {
		const (
			url    = "http://localhost:8080/container_registry"
			cipher = local.Plaintext
		)
		expectedSpec := fmt.Sprintf(specTemplate, url, cipher)

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
		specdata := fmt.Sprintf(specTemplate, url, cipher)

		spec, err := defaultContext.RepositorySpecForConfig([]byte(specdata), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(reflect.TypeOf(spec).String()).To(Equal("*gardenerconfig.RepositorySpec"))

		parsedSpec := spec.(*local.RepositorySpec)
		Expect(parsedSpec.URL).To(Equal(url))
		Expect(parsedSpec.ConfigType).To(Equal(gardenercfgcpi.ContainerRegistry))
		Expect(parsedSpec.Cipher).To(Equal(cipher))
	})

	It("resolves repository", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(creds))
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		specdata := fmt.Sprintf(specTemplate, svr.URL, local.Plaintext)

		repo, err := defaultContext.RepositoryForConfig([]byte(specdata), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())
		Expect(reflect.TypeOf(repo).String()).To(Equal("*gardenerconfig.Repository"))
	})

	It("retrieves credentials from unencrypted server", func() {
		svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			_, err := writer.Write([]byte(creds))
			Expect(err).ToNot(HaveOccurred())
		}))
		defer svr.Close()

		spec := fmt.Sprintf(specTemplate, svr.URL, local.Plaintext)

		repo, err := defaultContext.RepositoryForConfig([]byte(spec), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())

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

		parsedURL, err := utils.ParseURL(svr.URL)
		Expect(err).ToNot(HaveOccurred())

		id := cpi.ConsumerIdentity{
			cpi.CONSUMER_ATTR_TYPE: local.CONSUMER_TYPE,
		}
		id.SetNonEmptyValue(hostpath.ID_HOSTNAME, parsedURL.Host)
		id.SetNonEmptyValue(hostpath.ID_SCHEME, parsedURL.Scheme)
		id.SetNonEmptyValue(hostpath.ID_PATHPREFIX, strings.Trim(parsedURL.Path, "/"))
		id.SetNonEmptyValue(hostpath.ID_PORT, parsedURL.Port())

		creds := credentials.NewCredentials(common.Properties{
			cpi.ATTR_KEY: encryptionKey,
		})

		defaultContext.SetCredentialsForConsumer(id, creds)

		spec := fmt.Sprintf(specTemplate, svr.URL, local.AESECB)

		repo, err := defaultContext.RepositoryForConfig([]byte(spec), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())

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

		spec := fmt.Sprintf(specTemplate, "file://"+filename, local.Plaintext)

		repo, err := defaultContext.RepositoryForConfig([]byte(spec), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(repo).ToNot(BeNil())

		credentials, err := repo.LookupCredentials("test-credentials")
		Expect(err).ToNot(HaveOccurred())

		Expect(credentials.Properties()).To(Equal(props))
	})

})
