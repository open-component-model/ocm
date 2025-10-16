package vault_test

import (
	"encoding/json"
	"fmt"
	"reflect"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	me "ocm.software/ocm/api/credentials/extensions/repositories/vault"
	"ocm.software/ocm/api/credentials/extensions/repositories/vault/identity"
	common "ocm.software/ocm/api/utils/misc"
)

const (
	VAULT_ADDRESS    = "127.0.0.1:8200"
	VAULT_HTTP_URL   = "http://" + VAULT_ADDRESS
	VAULT_NAMESPACE  = "test-namespace"
	VAULT_MOUNT_PATH = "secret"
	VAULT_PATH_REPO1 = "mysecrets/repo1"
	VAULT_PATH_REPO2 = "mysecrets/repo2"
)

var _ = Describe("", func() {
	Context("serialization and deserialization", func() {
		DefaultContext := credentials.New()

		specdata := fmt.Sprintf("{\"type\": %q, \"serverURL\": %q, \"namespace\": %q, \"mountPath\": %q, \"path\": %q, \"secrets\": [\"secret1\", \"secret2\", \"secret3\"], \"propagateConsumerIdentity\": true }", me.Type, "http://"+VAULT_ADDRESS, VAULT_NAMESPACE, VAULT_MOUNT_PATH, VAULT_PATH_REPO1)
		spec := me.NewRepositorySpec("http://"+VAULT_ADDRESS, me.WithNamespace(VAULT_NAMESPACE), me.WithMountPath(VAULT_MOUNT_PATH), me.WithPath(VAULT_PATH_REPO1), me.WithSecrets("secret1", "secret2", "secret3"), me.WithPropagation())

		specdata2 := fmt.Sprintf("{\"type\": %q, \"serverURL\": %q }", me.Type, "http://"+VAULT_ADDRESS)
		spec2 := me.NewRepositorySpec("http://" + VAULT_ADDRESS)

		It("serializes repo spec", func() {
			data := Must(json.Marshal(spec))
			Expect(data).To(YAMLEqual([]byte(specdata)))

			data = Must(json.Marshal(spec2))
			Expect(data).To(YAMLEqual([]byte(specdata2)))
		})

		It("deserializes repo spec", func() {
			localspec := Must(DefaultContext.RepositorySpecForConfig([]byte(specdata), nil))
			Expect(reflect.TypeOf(localspec).String()).To(Equal("*vault.RepositorySpec"))
			Expect(localspec).To(Equal(spec))

			localspec = Must(DefaultContext.RepositorySpecForConfig([]byte(specdata2), nil))
			Expect(reflect.TypeOf(localspec).String()).To(Equal("*vault.RepositorySpec"))
			Expect(localspec).To(Equal(spec2))
		})

		It("resolves repository", func() {
			// Since vault always requires credentials to be accessed, RepositoryForConfig checks whether credentials
			// for a corresponding consumer exist. Thus, creating such credentials is required to test the method even
			// though they are not used
			consumerId := Must(identity.GetConsumerId(VAULT_HTTP_URL, VAULT_NAMESPACE, VAULT_MOUNT_PATH, VAULT_PATH_REPO1))
			creds := credentials.NewCredentials(common.Properties{
				identity.ATTR_AUTHMETH: identity.AUTH_TOKEN,
				identity.ATTR_TOKEN:    "token",
			})
			DefaultContext.SetCredentialsForConsumer(consumerId, creds)

			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))
			Expect(repo).ToNot(BeNil())
		})
	})
})
