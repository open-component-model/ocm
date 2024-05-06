// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package vault_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	local "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/dockerconfig"
	me "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/vault"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/vault/identity"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/finalizer"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"time"
)

type vaultMode string

const (
	HTTP  vaultMode = "dev"
	HTTPS vaultMode = "dev-tls"
)

const (
	VAULT_NAMESPACE  = "test-namespace"
	VAULT_KV_ENGINE  = "kv"
	VAULT_PATH_REPO1 = "mysecrets/repo1"
	VAULT_PATH_REPO2 = "mysecrets/repo2"
	VAULT_APP_ROLE   = "ocmrole"
	VAULT_SECRET     = "mysecret"

	VAULT_POLICY_RULE = `
path "secret/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
`
	VAULT_POLICY = "ocm"

	VAULT_ADDRESS = "127.0.0.1:8200"

	VAULT_ROOT_TOKEN = "toorl"
	VAULT_TLS_DIR    = "./vault-tls"
)

var _ = Describe("vault config", func() {

	props := common.Properties{
		"username":      "mandelsoft",
		"password":      "password",
		"serverAddress": "https://index.docker.io/v1/",
	}

	props2 := common.Properties{
		"username":      "mandelsoft",
		"password":      "token",
		"serverAddress": "https://ghcr.io",
	}

	var DefaultContext credentials.Context
	var cancelFunc context.CancelFunc
	var vaultClient *vault.Client
	var cmd *exec.Cmd

	ctx := context.Background()

	BeforeEach(func() {
		cmd, vaultClient, cancelFunc = Must3(StartVaultServer(HTTP, VAULT_ROOT_TOKEN, VAULT_ADDRESS))
		DefaultContext = credentials.New()
	})

	AfterEach(func() {
		cancelFunc()
		cmd.Wait()
		os.RemoveAll(VAULT_TLS_DIR)
	})

	Context("vault", func() {
		specdata := fmt.Sprintf("{\"type\": %q, \"serverURL\": %q, \"namespace\": %q, \"secretsEngine\": %q, \"path\": %q, \"secrets\": [\"secret1\", \"secret2\", \"secret3\"], \"propagateConsumerIdentity\": true }", me.Type, "http://"+VAULT_ADDRESS, VAULT_NAMESPACE, VAULT_KV_ENGINE, VAULT_PATH_REPO1)
		spec := me.NewRepositorySpec("http://"+VAULT_ADDRESS, me.WithNamespace(VAULT_NAMESPACE), me.WithSecretsEngine(VAULT_KV_ENGINE), me.WithPath(VAULT_PATH_REPO1), me.WithSecrets("secret1", "secret2", "secret3"), me.WithPropagation())

		specdata2 := fmt.Sprintf("{\"type\": %q, \"serverURL\": %q }", me.Type, VAULT_ADDRESS)
		spec2 := me.NewRepositorySpec(VAULT_ADDRESS)

		It("serializes repo spec", func() {
			localspec := me.NewRepositorySpec(VAULT_ADDRESS, me.WithNamespace(VAULT_NAMESPACE), me.WithSecretsEngine(VAULT_KV_ENGINE), me.WithPath(VAULT_PATH_REPO1), me.WithSecrets("secret1", "secret2", "secret3"), me.WithPropagation())
			data := Must(json.Marshal(localspec))
			Expect(data).To(YAMLEqual([]byte(specdata)))

			localspec = me.NewRepositorySpec(VAULT_ADDRESS)
			data = Must(json.Marshal(localspec))
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

		FIt("resolves repository", func() {
			//err := vaultClient.Sys().EnableAuthWithOptionsWithContext(ctx, "approle", &vault.EnableAuthOptions{Type: "approle"})
			//err = vaultClient.Sys().PutPolicyWithContext(ctx, VAULT_POLICY, VAULT_POLICY_RULE)
			//_, err = vaultClient.Logical().WriteWithContext(ctx, "auth/approle/role/ocm", map[string]interface{}{
			//	"backend": "approle",
			//	"role_name": VAULT_APP_ROLE,
			//	"token_policies": []string{VAULT_POLICY},
			//	"token_no_default_policy": "true",
			//	"bind_secret_id": "true",
			//	"token_period": "0",
			//})

			res := Must(vaultClient.System.AuthEnableMethod(ctx, "approle", schema.AuthEnableMethodRequest{Type: "approle"}))
			res = Must(vaultClient.System.PoliciesWriteAclPolicy(ctx, VAULT_POLICY, schema.PoliciesWriteAclPolicyRequest{Policy: VAULT_POLICY_RULE}))
			res = Must(vaultClient.Auth.AppRoleWriteRole(ctx, VAULT_APP_ROLE, schema.AppRoleWriteRoleRequest{TokenType: "batch", SecretIdTtl: "10m", TokenTtl: "20m", TokenMaxTtl: "30m", SecretIdNumUses: 40, TokenPolicies: []string{VAULT_POLICY}}))
			_ = res

			role := Must(vaultClient.Auth.AppRoleReadRoleId(ctx, VAULT_APP_ROLE))
			roleid := role.Data.RoleId
			fmt.Println(roleid)
			// Unfortunately, this function is currently bugged, therefore we fall back to the generic function
			//secretid := Must(vaultClient.Auth.AppRoleWriteSecretId(ctx, VAULT_APP_ROLE, schema.AppRoleWriteSecretIdRequest{}))
			secret := Must(vaultClient.Write(ctx, fmt.Sprintf("/v1/auth/approle/role/%s/secret-id", VAULT_APP_ROLE), map[string]interface{}{}))
			secretid := secret.Data["secret_id"].(string)
			fmt.Println(secretid)

			consumerId := Must(identity.GetConsumerId(vaultClient.Configuration().Address, VAULT_NAMESPACE, VAULT_KV_ENGINE, VAULT_PATH_REPO1))
			creds := credentials.NewCredentials(common.Properties{
				identity.ATTR_AUTHMETH: identity.AUTH_TOKEN,
				identity.ATTR_TOKEN:    VAULT_ROOT_TOKEN,
			})
			DefaultContext.SetCredentialsForConsumer(consumerId, creds)

			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))
			_ = repo
			//vaultClient.Secrets.KvV2Write(ctx, VAULT_SECRET, schema.KvV2WriteRequest{
			//	Data: map[string]any{
			//		"password1": "abc123",
			//		"password2": "correct horse battery staple",
			//	}},
			//	vault.WithMountPath("secret"),
			//	vault.WithToken(),
			//)
			//
			//repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil, credentials.DirectCredentials{
			//	identity.ATTR_TOKEN: VAULT_ROOT_TOKEN,
			//	identity.ATTR_SECRETID:
			//}))
			//Expect(reflect.TypeOf(repo).String()).To(Equal("*vault.Repository"))
		})

		It("retrieves credentials", func() {
			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))

			creds := Must(repo.LookupCredentials("index.docker.io"))
			Expect(creds.Properties()).To(Equal(props))

			creds = Must(repo.LookupCredentials("ghcr.io"))
			Expect(creds.Properties()).To(Equal(props2))
		})

		It("propagates credentials to consumer identity", func() {
			Must(DefaultContext.RepositoryForConfig([]byte(specdata2), nil))

			creds := Must(credentials.CredentialsForConsumer(DefaultContext, credentials.ConsumerIdentity{
				cpi.ATTR_TYPE:        identity.CONSUMER_TYPE,
				identity.ID_HOSTNAME: "ghcr.io",
			}))
			Expect(creds.Properties()).To(Equal(props2))
		})
	})

	Context("inline data", func() {
		specdata := "{\"type\":\"DockerConfig\",\"dockerConfig\":{\"auths\":{\"https://index.docker.io/v1/\":{\"auth\":\"bWFuZGVsc29mdDpwYXNzd29yZA==\"},\"https://ghcr.io\":{\"auth\":\"bWFuZGVsc29mdDp0b2tlbg==\"}},\"HttpHeaders\":{\"User-Agent\":\"Docker-Client/18.06.1-ce (linux)\"}},\"propagateConsumerIdentity\":true}"

		It("serializes repo spec", func() {
			configdata := Must(os.ReadFile("testdata/dockerconfig.json"))
			spec := local.NewRepositorySpecForConfig(configdata).WithConsumerPropagation(true)
			data := Must(json.Marshal(spec))

			var (
				datajson map[string]interface{}
				specjson map[string]interface{}
			)
			// Comparing the bytes might be problematic as the order of the JSON objects within the config file might change
			// during Marshaling
			MustBeSuccessful(json.Unmarshal([]byte(specdata), &specjson))
			MustBeSuccessful(json.Unmarshal(data, &datajson))
			Expect(datajson).To(Equal(specjson))
		})

		It("deserializes repo spec", func() {
			spec := Must(DefaultContext.RepositorySpecForConfig([]byte(specdata), nil))
			Expect(reflect.TypeOf(spec).String()).To(Equal("*dockerconfig.RepositorySpec"))
			configdata := Must(os.ReadFile("testdata/dockerconfig.json"))
			var (
				configdatajson   map[string]interface{}
				dockerconfigjson map[string]interface{}
			)
			// Comparing the bytes might be problematic as the order of the JSON objects within the config file might change
			// during Marshaling
			MustBeSuccessful(json.Unmarshal(configdata, &configdatajson))
			MustBeSuccessful(json.Unmarshal(spec.(*local.RepositorySpec).DockerConfig, &dockerconfigjson))
			Expect(dockerconfigjson).To(Equal(configdatajson))
		})

		It("resolves repository", func() {
			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))
			Expect(reflect.TypeOf(repo).String()).To(Equal("*dockerconfig.Repository"))
		})

		It("retrieves credentials", func() {
			repo := Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))

			creds := Must(repo.LookupCredentials("index.docker.io"))
			Expect(creds.Properties()).To(Equal(props))

			creds = Must(repo.LookupCredentials("ghcr.io"))
			Expect(creds.Properties()).To(Equal(props2))
		})

		It("propagates credentials to consumer identity", func() {
			Must(DefaultContext.RepositoryForConfig([]byte(specdata), nil))

			creds := Must(credentials.CredentialsForConsumer(DefaultContext, credentials.ConsumerIdentity{
				cpi.ATTR_TYPE:        identity.CONSUMER_TYPE,
				identity.ID_HOSTNAME: "ghcr.io",
			}))
			Expect(creds.Properties()).To(Equal(props2))
		})
	})

	Context("ref handling", func() {
		specdata := "{\"type\":\"DockerConfig\",\"dockerConfigFile\":\"testdata/dockerconfig.json\",\"propagateConsumerIdentity\":true}"

		It("can access the default context", func() {
			ctx := credentials.New()

			r := finalizer.GetRuntimeFinalizationRecorder(ctx)
			Expect(r).NotTo(BeNil())

			Must(ctx.RepositoryForConfig([]byte(specdata), nil))

			runtime.GC()
			time.Sleep(time.Second)
			ctx.GetType()
			Expect(r.Get()).To(BeNil())

			Expect(datacontext.GetContextRefCount(ctx)).To(Equal(1))
			ctx = nil
			runtime.GC()
			time.Sleep(time.Second)

			Expect(r.Get()).To(ContainElement(ContainSubstring(credentials.CONTEXT_TYPE)))
		})
	})
})

func StartVaultServer(mode vaultMode, rootToken, address string) (*exec.Cmd, *vault.Client, context.CancelFunc, error) {
	cmdctx, cancelFunc := context.WithCancel(context.Background())
	if mode == "" {
		mode = HTTP
	}
	url := address
	switch mode {
	case HTTP:
		url = "http://" + url
	case HTTPS:
		url = "https://" + url
	}

	cmd := exec.CommandContext(cmdctx, "../../../../../bin/vault", "server", "-"+string(mode), fmt.Sprintf("-dev-root-token-id=%s", rootToken), fmt.Sprintf("-dev-listen-address=%s", address))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	vaultClient, err := vault.New(
		vault.WithAddress(url),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, nil, cancelFunc, err
	}

	// authenticate with root token
	err = vaultClient.SetToken(rootToken)
	if err != nil {
		return nil, nil, cancelFunc, err
	}

	err = cmd.Start()
	return cmd, vaultClient, cancelFunc, err
}
