//go:build unix

package plugin_test

import (
	"crypto"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/plugin"
	"ocm.software/ocm/api/tech/signing/handlers/plugin/testdata/plugin/signinghandlers"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils/mime"

	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/plugin/plugins"
)

const PLUGIN = "signing"

const KEY = "some key"

const DIGEST = "mydigest"
const DN = "O=acme.org,CN=donald"

var _ = Describe("Signing Command Test Environment", func() {
	Context("plugin execution", func() {
		var env *TestEnv
		var plugindir TempPluginDir
		var registry plugins.Set

		BeforeEach(func() {
			env = NewTestEnv(TestData())
			plugindir = Must(ConfigureTestPlugins(env, "testdata/plugins"))
			registry = plugincacheattr.Get(env)
		})

		AfterEach(func() {
			plugindir.Cleanup()
			env.Cleanup()
		})

		It("loads plugin", func() {
			//	Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get(PLUGIN)
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
		})

		It("calls signer", func() {
			p := registry.Get(PLUGIN)

			sctx := &signing.DefaultSigningContext{
				Hash:       crypto.SHA256,
				PrivateKey: KEY,
				PublicKey:  nil,
				Issuer:     Must(signutils.ParseDN(DN)),
			}

			sig := Must(p.Sign(signinghandlers.NAME, DIGEST, nil, sctx))
			Expect(sig).NotTo(BeNil())

			Expect(sig.Value).To(Equal(DIGEST + ":" + KEY))
			Expect(sig.MediaType).To(Equal(mime.MIME_TEXT))
			Expect(sig.Algorithm).To(Equal(signinghandlers.NAME))

			dn := Must(signutils.ParseDN(DN))
			Expect(sig.Issuer).To(Equal(signutils.DNAsString(*dn)))
		})

		It("calls verifier", func() {
			p := registry.Get(PLUGIN)

			sctx := &signing.DefaultSigningContext{
				Hash:       crypto.SHA256,
				PrivateKey: nil,
				PublicKey:  KEY,
				Issuer:     Must(signutils.ParseDN(DN)),
			}

			sig := &signing.Signature{
				Value:     DIGEST + ":" + KEY,
				MediaType: mime.MIME_TEXT,
				Algorithm: signinghandlers.NAME,
				Issuer:    DN,
			}
			MustBeSuccessful(p.Verify(signinghandlers.NAME, DIGEST, sig, sctx))
		})

		Context("extension with creds", func() {
			It("calls signer", func() {
				p := registry.Get(PLUGIN)

				signer := plugin.NewSigner(p, signinghandlers.NAME)

				env.Context.CredentialsContext().SetCredentialsForConsumer(
					credentials.NewConsumerIdentity(signinghandlers.CID_TYPE),
					credentials.CredentialsFromList(signinghandlers.SUFFIX_KEY, KEY),
				)
				sctx := &signing.DefaultSigningContext{
					Hash:       crypto.SHA256,
					PrivateKey: signinghandlers.SUFFIX,
					Issuer:     Must(signutils.ParseDN(DN)),
				}

				sig := Must(signer.Sign(env.Context.CredentialsContext(), DIGEST, sctx))
				Expect(sig).NotTo(BeNil())

				Expect(sig.Value).To(Equal(DIGEST + ":" + signinghandlers.SUFFIX + ":" + KEY))
				Expect(sig.MediaType).To(Equal(mime.MIME_TEXT))
				Expect(sig.Algorithm).To(Equal(signinghandlers.NAME))

				dn := Must(signutils.ParseDN(DN))
				Expect(sig.Issuer).To(Equal(signutils.DNAsString(*dn)))
			})

			It("calls verifier", func() {
				p := registry.Get(PLUGIN)

				verifier := plugin.NewVerifier(p, signinghandlers.NAME)

				sctx := &signing.DefaultSigningContext{
					Hash:      crypto.SHA256,
					PublicKey: signinghandlers.SUFFIX_KEY + ":" + KEY,
					Issuer:    Must(signutils.ParseDN(DN)),
				}
				sig := &signing.Signature{
					Value:     DIGEST + ":" + signinghandlers.SUFFIX_KEY + ":" + KEY,
					MediaType: mime.MIME_TEXT,
					Algorithm: signinghandlers.NAME,
					Issuer:    DN,
				}

				MustBeSuccessful(verifier.Verify(DIGEST, sig, sctx))
			})
		})

		Context("extension without creds", func() {
			It("calls signer", func() {
				p := registry.Get(PLUGIN)

				signer := plugin.NewSigner(p, signinghandlers.NAME)

				sctx := &signing.DefaultSigningContext{
					Hash:       crypto.SHA256,
					PrivateKey: KEY,
					Issuer:     Must(signutils.ParseDN(DN)),
				}

				signer.Sign(nil, DIGEST, sctx)

				sig := Must(signer.Sign(nil, DIGEST, sctx))
				Expect(sig).NotTo(BeNil())

				Expect(sig.Value).To(Equal(DIGEST + ":" + KEY))
				Expect(sig.MediaType).To(Equal(mime.MIME_TEXT))
				Expect(sig.Algorithm).To(Equal(signinghandlers.NAME))

				dn := Must(signutils.ParseDN(DN))
				Expect(sig.Issuer).To(Equal(signutils.DNAsString(*dn)))
			})

			It("calls verifier", func() {
				p := registry.Get(PLUGIN)

				verifier := plugin.NewVerifier(p, signinghandlers.NAME)

				sctx := &signing.DefaultSigningContext{
					Hash:      crypto.SHA256,
					PublicKey: KEY,
					Issuer:    Must(signutils.ParseDN(DN)),
				}
				sig := &signing.Signature{
					Value:     DIGEST + ":" + KEY,
					MediaType: mime.MIME_TEXT,
					Algorithm: signinghandlers.NAME,
					Issuer:    DN,
				}

				MustBeSuccessful(verifier.Verify(DIGEST, sig, sctx))
			})
		})
	})
})
