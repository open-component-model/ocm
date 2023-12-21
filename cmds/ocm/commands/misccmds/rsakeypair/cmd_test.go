// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rsakeypair_test

import (
	"bytes"
	"crypto/x509/pkix"
	"encoding/pem"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/encrypt"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/signutils"
)

var ISSUER = &pkix.Name{CommonName: "mandelsoft"}

const KEYNAME = "test"

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var defaultContext credentials.Context

	BeforeEach(func() {
		env = NewTestEnv()
		defaultContext = credentials.New()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("create key pair", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "key.priv")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair key.priv[key.pub]
`))
		priv, err := env.ReadFile("key.priv")
		Expect(err).To(Succeed())
		pub, err := env.ReadFile("key.pub")
		Expect(err).To(Succeed())

		sctx := &signing.DefaultSigningContext{
			Hash:       0,
			PrivateKey: priv,
			PublicKey:  nil,
			RootCerts:  nil,
			Issuer:     ISSUER,
		}
		d := digest.FromBytes([]byte("digest"))
		sig, err := rsa.Handler{}.Sign(defaultContext, d.Hex(), sctx)
		Expect(err).To(Succeed())
		Expect(sig.Algorithm).To(Equal(rsa.Algorithm))
		Expect(sig.MediaType).To(Equal(rsa.MediaType))

		err = rsa.Handler{}.Verify(d.Hex(), sig, &signing.DefaultSigningContext{PublicKey: pub})
		Expect(err).To(Succeed())
	})

	It("create self-signed key pair", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "key.priv", "CN=mandelsoft")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair key.priv[key.cert]
`))
		priv, err := env.ReadFile("key.priv")
		Expect(err).To(Succeed())
		pub, err := env.ReadFile("key.cert")
		Expect(err).To(Succeed())

		sctx := &signing.DefaultSigningContext{
			Hash:       0,
			PrivateKey: priv,
			PublicKey:  nil,
			RootCerts:  nil,
			Issuer:     ISSUER,
		}
		d := digest.FromBytes([]byte("digest"))
		sig, err := rsa.Handler{}.Sign(defaultContext, d.Hex(), sctx)
		Expect(err).To(Succeed())
		Expect(sig.Algorithm).To(Equal(rsa.Algorithm))
		Expect(sig.MediaType).To(Equal(rsa.MediaType))

		err = rsa.Handler{}.Verify(d.Hex(), sig, &signing.DefaultSigningContext{PublicKey: pub})
		Expect(err).To(Succeed())
	})

	Context("encryption", func() {
		It("creates encrypted key with new encryption key", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "-E", "key.priv")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created encrypted rsa key pair key.priv[key.pub][key.priv.ekey]
`))
			pub := Must(env.ReadFile("key.pub"))
			Expect(pub).NotTo(BeNil())

			priv := Must(env.ReadFile("key.priv"))
			Expect(priv).NotTo(BeNil())
			block, rest := pem.Decode(priv)
			Expect(len(rest)).To(Equal(0))
			Expect(block).NotTo(BeNil())
			Expect(block.Type).To(Equal(encrypt.PEM_ENCRYPTED_DATA))

			ekey := Must(env.ReadFile("key.priv.ekey"))
			block, rest = pem.Decode(ekey)
			Expect(len(rest)).To(Equal(0))
			Expect(block).NotTo(BeNil())
			Expect(block.Type).To(Equal(encrypt.PEM_ENCRYPTION_KEY))

			reg := signingattr.Get(env)
			reg.RegisterPrivateKey(KEYNAME, priv)
			reg.RegisterPrivateKey(signing.DecryptionKeyName(KEYNAME), ekey)

			key := Must(signing.ResolvePrivateKey(reg, KEYNAME))
			Expect(key).NotTo(BeNil())

			sctx := &signing.DefaultSigningContext{
				Hash:       0,
				PrivateKey: key,
				PublicKey:  nil,
				RootCerts:  nil,
				Issuer:     ISSUER,
			}
			d := digest.FromBytes([]byte("digest"))
			Must(rsa.Handler{}.Sign(defaultContext, d.Hex(), sctx))

			buf.Reset()
			Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "-e", KEYNAME, "other.priv")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created encrypted rsa key pair other.priv[other.pub]
`))
			pub = Must(env.ReadFile("other.pub"))
			Expect(pub).NotTo(BeNil())

			priv = Must(env.ReadFile("other.priv"))
			Expect(priv).NotTo(BeNil())
			block, rest = pem.Decode(priv)
			Expect(len(rest)).To(Equal(0))
			Expect(block).NotTo(BeNil())
			Expect(block.Type).To(Equal(encrypt.PEM_ENCRYPTED_DATA))
		})
	})

	Context("certificate handling", func() {
		It("creates chain", func() {
			buf := bytes.NewBuffer(nil)

			// create Root CA
			Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "--ca", "CN=cerificate-authority", "root.priv")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair root.priv[root.cert]
`))
			Expect(env.FileExists("root.priv")).To(BeTrue())
			Expect(env.FileExists("root.cert")).To(BeTrue())

			// create CA used to create signing certificates
			buf.Reset()
			Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "--ca", "CN=acme.org", "--ca-key", "root.priv", "--ca-cert", "root.cert", "ca.priv")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair ca.priv[ca.cert]
`))
			Expect(env.FileExists("ca.priv")).To(BeTrue())
			Expect(env.FileExists("ca.cert")).To(BeTrue())

			// create signing vcertificate from CA
			buf.Reset()
			Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "--ca", "CN=mandelsoft", "C=DE", "--ca-key", "ca.priv", "--ca-cert", "ca.cert", "--root-certs", "root.cert", "key.priv")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
created rsa key pair key.priv[key.cert]
`))
			Expect(env.FileExists("key.priv")).To(BeTrue())
			Expect(env.FileExists("key.cert")).To(BeTrue())

			root := Must(env.ReadFile("root.cert"))
			certs := Must(env.ReadFile("key.cert"))

			chain := Must(signutils.GetCertificateChain(certs, false))
			Expect(len(chain)).To(Equal(3))
			MustBeSuccessful(signing.VerifyCertDN(chain[1:], root, &pkix.Name{CommonName: "mandelsoft", Country: []string{"DE"}}, chain[0]))
			ExpectError(signing.VerifyCertDN(chain[1:], root, &pkix.Name{CommonName: "mandelsoft", Country: []string{"US"}}, chain[0])).To(MatchError(`country "US" not found`))
		})
	})
})
