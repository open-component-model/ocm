//go:build integration

package vault_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	me "ocm.software/ocm/api/credentials/extensions/repositories/vault"
	"ocm.software/ocm/api/credentials/extensions/repositories/vault/identity"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

type vaultMode string

const (
	HTTP  vaultMode = "dev"
	HTTPS vaultMode = "dev-tls"
)

const (
	VAULT_APP_ROLE       = "ocmrole"
	VAULT_APP_ROLE1      = "ocmrole1"
	VAULT_SECRET         = "mysecret"
	VAULT_CUSTOM_SECRETS = "secret-list"
	VAULT_SECRET_2       = "mysecret2"

	VAULT_POLICY_NAME  = "ocm"
	VAULT_POLICY_NAME1 = "ocm1"

	VAULT_ROOT_TOKEN = "toorl"
	VAULT_TLS_DIR    = "./vault-tls"
)

const (
	VAULT_POLICY_RULE = `
path "secret/*"
{
  capabilities = ["read","list"]
}
`
	VAULT_INSUFFICIENT_POLICY_RULE = `
path "secret/notmysecret"
{
  capabilities = ["read", "list"]
}
`
)

var _ = Describe("vault config", func() {
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
		_ = cmd.Wait()
		Expect(os.RemoveAll(VAULT_TLS_DIR)).To(Succeed())
	})

	Context("authentication to vault and reading secrets", func() {
		spec := me.NewRepositorySpec("http://"+VAULT_ADDRESS, me.WithPath(VAULT_PATH_REPO1), me.WithMountPath("secret"))
		spec1 := me.NewRepositorySpec("http://"+VAULT_ADDRESS, me.WithPath(VAULT_PATH_REPO2), me.WithMountPath("secret"))

		It("authenticate with token and retrieve credentials", func() {
			data := map[string]any{
				"password1": "ocm-password-1",
				"password2": "ocm-password-2",
			}
			_ = Must(vaultClient.Secrets.KvV2Write(ctx,
				VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteRequest{Data: data},
				vault.WithMountPath("secret"),
			))

			consumerId := Must(identity.GetConsumerId(vaultClient.Configuration().Address,
				"", "secret", VAULT_PATH_REPO1))
			creds := credentials.NewCredentials(common.Properties{
				identity.ATTR_AUTHMETH: identity.AUTH_TOKEN,
				identity.ATTR_TOKEN:    VAULT_ROOT_TOKEN,
			})
			DefaultContext.SetCredentialsForConsumer(consumerId, creds)

			repo := Must(DefaultContext.RepositoryForSpec(spec, nil))
			Expect(repo).ToNot(BeNil())

			c, err := repo.LookupCredentials(VAULT_SECRET)
			Expect(c.Properties()).To(YAMLEqual(data))
			Expect(err).To(BeNil())
		})

		It("authenticate with approle and retrieve credentials", func() {
			SetUpVaultAccess(ctx, DefaultContext, vaultClient, VAULT_POLICY_RULE)

			data := map[string]any{
				"password1": "ocm-password-1",
				"password2": "ocm-password-2",
			}
			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteRequest{Data: data},
				vault.WithMountPath("secret"),
			))

			repo := Must(DefaultContext.RepositoryForSpec(spec, nil))
			Expect(repo).ToNot(BeNil())

			c, err := repo.LookupCredentials(VAULT_SECRET)
			Expect(c.Properties()).To(YAMLEqual(data))
			Expect(err).To(BeNil())
		})

		It("authenticate with approle with insufficient authorizations and fail to retrieve credentials", func() {
			SetUpVaultAccess(ctx, DefaultContext, vaultClient, VAULT_INSUFFICIENT_POLICY_RULE)

			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET, schema.KvV2WriteRequest{
				Data: map[string]any{
					"password1": "ocm-password-1",
					"password2": "ocm-password-2",
				},
			},
				vault.WithMountPath("secret"),
			))

			repo := Must(DefaultContext.RepositoryForSpec(spec, nil))
			Expect(repo).ToNot(BeNil())

			c, err := repo.LookupCredentials(VAULT_SECRET)
			Expect(err).To(HaveOccurred())
			Expect(c).To(BeNil())
		})

		It("authenticate with approle and specify a subset of secrets at the specified path in the repository spec", func() {
			SetUpVaultAccess(ctx, DefaultContext, vaultClient, VAULT_POLICY_RULE)

			data := map[string]any{
				"password1": "ocm-password-1",
			}
			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteRequest{Data: data},
				vault.WithMountPath("secret"),
			))

			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET_2, schema.KvV2WriteRequest{
				Data: map[string]any{
					"password2": "ocm-password-2",
				},
			},
				vault.WithMountPath("secret"),
			))

			// This is how we restrict the secrets accessible through the repository
			spec.Secrets = append(spec.Secrets, VAULT_SECRET)
			repo := Must(DefaultContext.RepositoryForSpec(spec, nil))
			Expect(repo).ToNot(BeNil())

			c, err := repo.LookupCredentials(VAULT_SECRET)
			Expect(c).To(YAMLEqual(data))
			Expect(err).ToNot(HaveOccurred())

			c, err = repo.LookupCredentials(VAULT_SECRET_2)
			Expect(err).To(BeNil())
			Expect(c).To(BeNil())
		})

		It("authenticate with approle and specify a subset of secrets at the specified path in a dedicated secret", func() {
			SetUpVaultAccess(ctx, DefaultContext, vaultClient, VAULT_POLICY_RULE)

			data := map[string]any{
				"password1": "ocm-password-1",
			}
			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteRequest{Data: data},
				vault.WithMountPath("secret"),
			))

			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET_2, schema.KvV2WriteRequest{
				Data: map[string]any{
					"password2": "ocm-password-2",
				},
			},
				vault.WithMountPath("secret"),
			))

			// You have to specify a value, but it is essentially a placeholder here
			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_CUSTOM_SECRETS, schema.KvV2WriteRequest{
				Data: map[string]any{
					"description": "specify a list in the metadata",
				},
			},
				vault.WithMountPath("secret"),
			))
			metadata := map[string]any{
				me.CUSTOM_SECRETS: VAULT_SECRET,
			}
			_ = Must(vaultClient.Secrets.KvV2WriteMetadata(ctx, VAULT_PATH_REPO1+"/"+VAULT_CUSTOM_SECRETS,
				schema.KvV2WriteMetadataRequest{CustomMetadata: metadata},
				vault.WithMountPath("secret"),
			))

			// This is how we restrict the secrets accessible through the repository
			spec.Secrets = append(spec.Secrets, VAULT_CUSTOM_SECRETS)
			repo := Must(DefaultContext.RepositoryForSpec(spec, nil))
			Expect(repo).ToNot(BeNil())

			c, err := repo.LookupCredentials(VAULT_SECRET)
			Expect(c).To(YAMLEqual(data))
			Expect(err).ToNot(HaveOccurred())

			c, err = repo.LookupCredentials(VAULT_SECRET_2)
			Expect(err).To(BeNil())
			Expect(c).To(BeNil())
		})

		It("authenticate with approle and consume secrets with a consumer id from the provider", func() {
			SetUpVaultAccess(ctx, DefaultContext, vaultClient, VAULT_POLICY_RULE)

			data := map[string]any{
				"password1": "ocm-password-1",
			}
			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteRequest{Data: data},
				vault.WithMountPath("secret"),
			))
			cid := hostpath.GetConsumerIdentity(hostpath.IDENTITY_TYPE, "https://test-url.com")
			cidData := Must(runtime.DefaultJSONEncoding.Marshal(cid))
			metadata := map[string]any{
				me.CUSTOM_CONSUMERID: string(cidData),
			}
			_ = Must(vaultClient.Secrets.KvV2WriteMetadata(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteMetadataRequest{CustomMetadata: metadata},
				vault.WithMountPath("secret"),
			))

			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET_2, schema.KvV2WriteRequest{
				Data: map[string]any{
					"password2": "ocm-password-2",
				},
			},
				vault.WithMountPath("secret"),
			))

			repo := Must(me.NewRepository(DefaultContext, spec))
			Expect(repo).ToNot(BeNil())
			provider := Must(me.NewConsumerProvider(repo))
			c, ok := provider.Get(cid)
			Expect(ok).To(BeTrue())
			Expect(c).ToNot(BeNil())
		})

		It("authenticate with approle and consume secrets with a consumer id from the credential context", func() {
			SetUpVaultAccess(ctx, DefaultContext, vaultClient, VAULT_POLICY_RULE)

			data := map[string]any{
				"password1": "ocm-password-1",
			}
			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteRequest{Data: data},
				vault.WithMountPath("secret"),
			))
			cid := hostpath.GetConsumerIdentity(hostpath.IDENTITY_TYPE, "https://test-url.com")
			cidData := Must(runtime.DefaultJSONEncoding.Marshal(cid))
			metadata := map[string]any{
				me.CUSTOM_CONSUMERID: string(cidData),
			}
			_ = Must(vaultClient.Secrets.KvV2WriteMetadata(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteMetadataRequest{CustomMetadata: metadata},
				vault.WithMountPath("secret"),
			))

			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET_2, schema.KvV2WriteRequest{
				Data: map[string]any{
					"password2": "ocm-password-2",
				},
			},
				vault.WithMountPath("secret"),
			))

			spec.PropgateConsumerIdentity = true
			repo := Must(DefaultContext.RepositoryForSpec(spec))
			Expect(repo).ToNot(BeNil())

			c := Must(DefaultContext.GetCredentialsForConsumer(cid))
			Expect(c).To(YAMLEqual(data))
		})

		It("recursive authentication", func() {
			SetUpVaultAccess(ctx, DefaultContext, vaultClient, fmt.Sprintf(`
path "secret/data/%s/*"
{
  capabilities = ["read"]
}
path "secret/metadata/%s/*"
{
  capabilities = ["list"]
}
`, VAULT_PATH_REPO1, VAULT_PATH_REPO1))

			_ = Must(vaultClient.System.PoliciesWriteAclPolicy(ctx, VAULT_POLICY_NAME1, schema.PoliciesWriteAclPolicyRequest{Policy: fmt.Sprintf(`
path "secret/data/%s/*"
{
  capabilities = ["read"]
}
path "secret/metadata/%s/*"
{
  capabilities = ["list"]
}
`, VAULT_PATH_REPO2, VAULT_PATH_REPO2)}))
			_ = Must(vaultClient.Auth.AppRoleWriteRole(ctx, VAULT_APP_ROLE1, schema.AppRoleWriteRoleRequest{TokenType: "batch", SecretIdTtl: "10m", TokenTtl: "20m", TokenMaxTtl: "30m", SecretIdNumUses: 40, TokenPolicies: []string{VAULT_POLICY_NAME1}}))

			role := Must(vaultClient.Auth.AppRoleReadRoleId(ctx, VAULT_APP_ROLE1))
			roleid := role.Data.RoleId
			// Unfortunately, this function is currently bugged, therefore we fall back to the generic function
			// secretid := Must(client.Auth.AppRoleWriteSecretId(ctx, VAULT_APP_ROLE, schema.AppRoleWriteSecretIdRequest{}))
			secret := Must(vaultClient.Write(ctx, fmt.Sprintf("/v1/auth/approle/role/%s/secret-id", VAULT_APP_ROLE1), map[string]interface{}{}))
			secretid := secret.Data["secret_id"].(string)

			// Write a secret with the credentials for vault repo 2 (VAULT_PATH_REPO2) into vault repo 1 and write the
			// consumer id of vault repo 2 into the secrets metadata
			consumerId := Must(identity.GetConsumerId(VAULT_HTTP_URL, "", "secret", VAULT_PATH_REPO2))
			data := map[string]any{
				identity.ATTR_AUTHMETH: identity.AUTH_APPROLE,
				identity.ATTR_ROLEID:   roleid,
				identity.ATTR_SECRETID: secretid,
			}
			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteRequest{Data: data},
				vault.WithMountPath("secret"),
			))
			consumerIdData := Must(runtime.DefaultJSONEncoding.Marshal(consumerId))
			metadata := map[string]any{
				me.CUSTOM_CONSUMERID: string(consumerIdData),
			}
			_ = Must(vaultClient.Secrets.KvV2WriteMetadata(ctx, VAULT_PATH_REPO1+"/"+VAULT_SECRET,
				schema.KvV2WriteMetadataRequest{CustomMetadata: metadata},
				vault.WithMountPath("secret"),
			))

			// Write a secret with arbitrary data into vault repo 2
			data = map[string]any{
				"password1": "ocm-password-1",
			}
			_ = Must(vaultClient.Secrets.KvV2Write(ctx, VAULT_PATH_REPO2+"/"+VAULT_SECRET,
				schema.KvV2WriteRequest{Data: data},
				vault.WithMountPath("secret"),
			))
			consumerId = hostpath.GetConsumerIdentity(hostpath.IDENTITY_TYPE, "https://test-url.com")
			consumerIdData = Must(runtime.DefaultJSONEncoding.Marshal(consumerId))
			metadata = map[string]any{
				me.CUSTOM_CONSUMERID: string(consumerIdData),
			}
			_ = Must(vaultClient.Secrets.KvV2WriteMetadata(ctx, VAULT_PATH_REPO2+"/"+VAULT_SECRET,
				schema.KvV2WriteMetadataRequest{CustomMetadata: metadata},
				vault.WithMountPath("secret"),
			))

			spec.PropgateConsumerIdentity = true
			repo := Must(DefaultContext.RepositoryForSpec(spec))
			Expect(repo).ToNot(BeNil())

			fmt.Println("***add second provider:")
			spec1.PropgateConsumerIdentity = true
			repo = Must(DefaultContext.RepositoryForSpec(spec1))
			Expect(repo).ToNot(BeNil())

			fmt.Println("***query credential:")
			c := Must(DefaultContext.GetCredentialsForConsumer(consumerId))
			Expect(c).To(YAMLEqual(data))
		})

		// D(irect):
		//  - has general credentials matching parent path of P2
		//  - has credentials for P1
		// P1:
		//  - has specialized credentials for P2
		// P2:
		//  - has credentials for C
		//
		//
		// query C:
		// - D:   -> nothing
		// - P1: query P1
		//    - D: -> found
		//    - P1:   omit (recursion)
		//    - P2: query P2
		//        - D: -> found
		//        - P1:   omit (recursion)
		//        - P2:   omit (recursion)
		//		  explore, whether an additional attempt with P1 BUT only with credentialless providers / direct creds
		//		  would work as a general solution.
		//        -> select D(P2)     WRONG        (a1)
		// - P2: query P2
		//    - D: -> found
		//    - P1: query P1
		//        - D: found
		//        - P1: omit (recursion)
		//        - P2: omit (recursion)
		//        -> select D(P1)     CORRECT      (b)
		//      -> found
		//    - P2: omit (recursion)
		//    -> select P1(P2)  CORRECT            (a2)
		//  -> found
		// -> select P2(C)
		//
		// The Problem here is, that the case a1 and case b are formally indistinguishable. While a2 and b lead to the
		// correct result, we would fail in a1.
		It("recursive authentication with overlapping credentials", func() {
			// TODO
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
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

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
	if err == nil {
		err = PingTCPServer(address, time.Minute)
	}
	return cmd, vaultClient, cancelFunc, err
}

func SetUpVaultAccess(ctx context.Context, credctx credentials.Context, client *vault.Client, policy string) {
	_ = Must(client.System.AuthEnableMethod(ctx, "approle", schema.AuthEnableMethodRequest{Type: "approle"}))
	_ = Must(client.System.PoliciesWriteAclPolicy(ctx, VAULT_POLICY_NAME, schema.PoliciesWriteAclPolicyRequest{Policy: policy}))
	_ = Must(client.Auth.AppRoleWriteRole(ctx, VAULT_APP_ROLE, schema.AppRoleWriteRoleRequest{TokenType: "batch", SecretIdTtl: "10m", TokenTtl: "20m", TokenMaxTtl: "30m", SecretIdNumUses: 40, TokenPolicies: []string{VAULT_POLICY_NAME}}))

	role := Must(client.Auth.AppRoleReadRoleId(ctx, VAULT_APP_ROLE))
	roleid := role.Data.RoleId
	// Unfortunately, this function is currently bugged, therefore we fall back to the generic function
	// secretid := Must(client.Auth.AppRoleWriteSecretId(ctx, VAULT_APP_ROLE, schema.AppRoleWriteSecretIdRequest{}))
	secret := Must(client.Write(ctx, fmt.Sprintf("/v1/auth/approle/role/%s/secret-id", VAULT_APP_ROLE), map[string]interface{}{}))
	secretid := secret.Data["secret_id"].(string)

	consumerId := Must(identity.GetConsumerId(client.Configuration().Address, "", "secret", VAULT_PATH_REPO1))
	creds := credentials.NewCredentials(common.Properties{
		identity.ATTR_AUTHMETH: identity.AUTH_APPROLE,
		identity.ATTR_ROLEID:   roleid,
		identity.ATTR_SECRETID: secretid,
	})
	credctx.SetCredentialsForConsumer(consumerId, creds)
}
