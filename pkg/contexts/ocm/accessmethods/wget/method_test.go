package wget

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/wget/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/signutils"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"net/http"
	"time"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

var (
	caCert *x509.Certificate
	caPriv *rsa.PrivateKey
	caPub  *rsa.PublicKey
	caPEM  []byte

	serverCert *x509.Certificate
	serverPriv *rsa.PrivateKey
	serverPub  *rsa.PublicKey
	serverPEM  []byte
)

const (
	HTTP_PORT                   = ":18080"
	HTTPS_PORT                  = ":1443"
	HTTPS_PORT_WITH_CLIENT_AUTH = ":2443"

	HTTP_HOST                   = "http://localhost" + HTTP_PORT
	HTTPS_HOST                  = "https://localhost" + HTTPS_PORT
	HTTPS_HOST_WITH_CLIENT_AUTH = "https://localhost" + HTTPS_PORT_WITH_CLIENT_AUTH

	TO_MEMORY = "/tomemory"
	TO_FILE   = "/tofile"

	CONTENT = "hello world"
)

var _ = BeforeSuite(func() {
	// setup certificate authority
	_capriv, _capub := Must2(rsa.Handler{}.CreateKeyPair())
	caPriv = _capriv.(*rsa.PrivateKey)
	caPub = _capub.(*rsa.PublicKey)

	caSpec := &signutils.Specification{
		Subject:      *signutils.CommonName("caCert-authority"),
		Validity:     10 * time.Minute,
		CAPrivateKey: caPriv,
		IsCA:         true,
		Usages:       []interface{}{x509.KeyUsageDigitalSignature},
	}

	caCert, caPEM = Must2(signutils.CreateCertificate(caSpec))

	// use certificate authority to create httpsServer certificate
	_serverPriv, _serverPub := Must2(rsa.Handler{}.CreateKeyPair())
	serverPriv = _serverPriv.(*rsa.PrivateKey)
	serverPub = _serverPub.(*rsa.PublicKey)

	serverSpec := &signutils.Specification{
		IsCA:         false,
		Subject:      pkix.Name{CommonName: "localhost"},
		Validity:     10 * time.Minute,
		RootCAs:      caCert,
		CAChain:      caCert,
		CAPrivateKey: caPriv,
		PublicKey:    serverPub,
		Usages:       []interface{}{x509.ExtKeyUsageServerAuth},
		Hosts:        []string{"localhost", "127.0.0.1"},
	}

	serverCert, serverPEM = Must2(signutils.CreateCertificate(serverSpec))

	// setup tls configuration for the httpsServer for https with the corresponding certs and keys
	serverPrivPEM := Must(rsa.KeyData(serverPriv))
	serverTlsCert := Must(tls.X509KeyPair(serverPEM, serverPrivPEM))

	// ca's used by the server to validate client certificates
	clientCaCertPool := x509.NewCertPool()
	clientCaCertPool.AddCert(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTlsCert},
	}

	tlsConfigClientAuth := &tls.Config{
		Certificates: []tls.Certificate{serverTlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCaCertPool,
	}

	// configure test routes
	mux := http.NewServeMux()
	mux.HandleFunc(TO_MEMORY, func(writer http.ResponseWriter, request *http.Request) {
		n, err := writer.Write([]byte(CONTENT))
		_, _ = n, err
	})

	// setup an https and an http httpsServer
	httpsServerClientAuth := &http.Server{
		Addr:      HTTPS_PORT_WITH_CLIENT_AUTH,
		TLSConfig: tlsConfigClientAuth,
		Handler:   mux,
	}

	httpsServer := &http.Server{
		Addr:      HTTPS_PORT,
		TLSConfig: tlsConfig,
		Handler:   mux,
	}

	httpServer := &http.Server{
		Addr:    HTTP_PORT,
		Handler: mux,
	}

	go func() {
		MustBeSuccessful(httpsServerClientAuth.ListenAndServeTLS("", ""))
	}()

	go func() {
		MustBeSuccessful(httpsServer.ListenAndServeTLS("", ""))
	}()

	go func() {
		MustBeSuccessful(httpServer.ListenAndServe())
	}()
})

var _ = Describe("wget access method", func() {
	It("access content on http server", func() {
		url := HTTP_HOST + TO_MEMORY
		spec := New(url)

		ctx := ocm.DefaultContext()
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		b := Must(m.Get())
		Expect(string(b)).To(Equal(CONTENT))
	})

	It("access content on https server", func() {
		url := HTTPS_HOST + TO_MEMORY
		spec := New(url)

		ctx := ocm.DefaultContext()
		ctx.CredentialsContext().SetCredentialsForConsumer(identity.GetConsumerId(url), credentials.DirectCredentials{
			identity.ATTR_CERTIFICATE_AUTHORITY: string(caPEM),
		})
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		b := Must(m.Get())
		Expect(string(b)).To(Equal(CONTENT))
	})

	It("access content on https server with client authentication", func() {
		// create a client certificate
		_clientPriv, _clientPub := Must2(rsa.Handler{}.CreateKeyPair())
		clientPriv := _clientPriv.(*rsa.PrivateKey)
		clientPrivData := Must(rsa.KeyData(clientPriv))
		clientPub := _clientPub.(*rsa.PublicKey)

		clientSpec := &signutils.Specification{
			IsCA:         false,
			Subject:      pkix.Name{CommonName: "localhost"},
			Validity:     10 * time.Minute,
			RootCAs:      caCert,
			CAChain:      caCert,
			CAPrivateKey: caPriv,
			PublicKey:    clientPub,
			Usages:       []interface{}{x509.ExtKeyUsageClientAuth},
			Hosts:        []string{"localhost", "127.0.0.1"},
		}

		_, clientPEM := Must2(signutils.CreateCertificate(clientSpec))

		// Request
		url := HTTPS_HOST_WITH_CLIENT_AUTH + TO_MEMORY
		spec := New(url)

		ctx := ocm.DefaultContext()
		ctx.CredentialsContext().SetCredentialsForConsumer(identity.GetConsumerId(url), credentials.DirectCredentials{
			identity.ATTR_CERTIFICATE_AUTHORITY: string(caPEM),
			identity.ATTR_CERTIFICATE:           string(clientPEM),
			identity.ATTR_PRIVATE_KEY:           string(clientPrivData),
		})
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		b := Must(m.Get())
		Expect(string(b)).To(Equal(CONTENT))
	})

})
