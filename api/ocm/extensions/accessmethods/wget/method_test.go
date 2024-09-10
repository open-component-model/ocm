package wget_test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/ocm/extensions/accessmethods/wget"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/tech/wget/identity"
	"ocm.software/ocm/api/utils/mime"
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

	httpsServerClientAuth http.Server
	httpsServer           http.Server
	httpServer            http.Server
)

const (
	HTTP_PORT                   = ":18080"
	HTTPS_PORT                  = ":1443"
	HTTPS_PORT_WITH_CLIENT_AUTH = ":2443"

	HTTP_HOST                   = "http://localhost" + HTTP_PORT
	HTTPS_HOST                  = "https://localhost" + HTTPS_PORT
	HTTPS_HOST_WITH_CLIENT_AUTH = "https://localhost" + HTTPS_PORT_WITH_CLIENT_AUTH

	TO_MEMORY    = "/tomemory"
	TO_FILE      = "/tofile"
	BASIC_LOGIN  = "/basic-login"
	BEARER_LOGIN = "/bearer-login"
	ECHO_HEADERS = "/headers"
	ECHO_BODY    = "/body"
	ECHO_METHOD  = "/method"
	CONTENT_TYPE = " /content-type"
	DOT_EXT      = "/somefile.tar"
	REDIRECT     = "/redirect"

	USERNAME = "user"
	PASSWORD = "password"
	TOKEN    = "token"

	CONTENT            = "hello world"
	NOREDIRECT_CONTENT = "noredirect"
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
	mux.HandleFunc(BASIC_LOGIN, func(writer http.ResponseWriter, request *http.Request) {
		username, password, ok := request.BasicAuth()
		if !ok {
			n, err := writer.Write([]byte(`failure`))
			_, _ = n, err
		}
		if username != "" && password != "" {
			res := fmt.Sprintf("%s:%s", username, password)
			n, err := writer.Write([]byte(res))
			_, _ = n, err
		} else {
			n, err := writer.Write([]byte(`failure`))
			_, _ = n, err
		}
	})
	mux.HandleFunc(BEARER_LOGIN, func(writer http.ResponseWriter, request *http.Request) {
		auth := request.Header.Get("Authorization")
		if auth == "" {
			n, err := writer.Write([]byte(`failure`))
			_, _ = n, err
		} else {
			bearer, ok := strings.CutPrefix(auth, "Bearer ")
			if !ok {
				n, err := writer.Write([]byte(`failure`))
				_, _ = n, err
			} else {
				n, err := writer.Write([]byte(bearer))
				_, _ = n, err
			}
		}
	})
	mux.HandleFunc(ECHO_HEADERS, func(writer http.ResponseWriter, request *http.Request) {
		err := request.Header.Write(writer)
		_ = err
	})
	mux.HandleFunc(ECHO_BODY, func(writer http.ResponseWriter, request *http.Request) {
		b, err := io.ReadAll(request.Body)
		_, err = writer.Write(b)
		_ = err
	})
	mux.HandleFunc(ECHO_METHOD, func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(request.Method))
		_ = err
	})
	mux.HandleFunc(CONTENT_TYPE, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", mime.MIME_TEXT)
	})
	mux.HandleFunc(DOT_EXT, func(writer http.ResponseWriter, request *http.Request) {})
	mux.HandleFunc(REDIRECT, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Location", TO_MEMORY)
		writer.WriteHeader(307)
		writer.Write([]byte(NOREDIRECT_CONTENT))
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

var _ = AfterSuite(func() {
	MustBeSuccessful(httpsServerClientAuth.Close())
	MustBeSuccessful(httpsServer.Close())
	MustBeSuccessful(httpServer.Close())
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

	It("check that username and password are passed correctly", func() {
		url := HTTP_HOST + BASIC_LOGIN
		spec := New(url)

		ctx := ocm.DefaultContext()
		ctx.CredentialsContext().SetCredentialsForConsumer(identity.GetConsumerId(url), credentials.DirectCredentials{
			identity.ATTR_USERNAME: USERNAME,
			identity.ATTR_PASSWORD: PASSWORD,
		})
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		b := Must(m.Get())
		Expect(string(b)).To(Equal(USERNAME + ":" + PASSWORD))
	})

	It("check that bearer token is passed correctly", func() {
		url := HTTP_HOST + BEARER_LOGIN
		spec := New(url)

		ctx := ocm.DefaultContext()
		ctx.CredentialsContext().SetCredentialsForConsumer(identity.GetConsumerId(url), credentials.DirectCredentials{
			identity.ATTR_IDENTITY_TOKEN: TOKEN,
		})
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		b := Must(m.Get())
		Expect(string(b)).To(Equal(TOKEN))
	})

	It("check that basic auth is merged correctly with other provided headers", func() {
		url := HTTP_HOST + ECHO_HEADERS
		headers := map[string][]string{"Content-Type": {"text/plain"}}
		spec := New(url, WithHeader(headers))

		ctx := ocm.DefaultContext()
		ctx.CredentialsContext().SetCredentialsForConsumer(identity.GetConsumerId(url), credentials.DirectCredentials{
			identity.ATTR_USERNAME: USERNAME,
			identity.ATTR_PASSWORD: PASSWORD,
		})
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		b := Must(m.Get())

		Expect(strings.Contains(string(b), "Content-Type: text/plain")).To(BeTrue())
		Expect(strings.Contains(string(b), "Authorization: Basic")).To(BeTrue())
	})

	It("check detect mime type based on content-type response header", func() {
		url := HTTP_HOST + ECHO_HEADERS
		spec := New(url)

		ctx := ocm.DefaultContext()
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		Expect(m.MimeType()).To(Equal(mime.MIME_TEXT))
	})

	It("check deduction of mime type based on url", func() {
		url := HTTP_HOST + DOT_EXT
		spec := New(url)

		ctx := ocm.DefaultContext()
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		Expect(m.MimeType()).To(Equal("application/x-tar"))
	})

	It("check passing an http body", func() {
		url := HTTP_HOST + ECHO_BODY

		content := `hello world`
		spec := New(url, WithBody(bytes.NewReader([]byte(content))))

		ctx := ocm.DefaultContext()
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		b := Must(m.Get())

		Expect(string(b)).To(Equal(content))
	})

	It("check passing an http verb", func() {
		url := HTTP_HOST + ECHO_METHOD

		method := http.MethodPost
		spec := New(url, WithVerb(method))

		ctx := ocm.DefaultContext()
		m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(m, "method")

		b := Must(m.Get())

		Expect(string(b)).To(Equal(method))
	})

	It("check noredirect behavior", func() {
		url := HTTP_HOST + REDIRECT

		redirectSpec := New(url, WithNoRedirect(false))
		noredirectSpec := New(url, WithNoRedirect(true))

		ctx := ocm.DefaultContext()
		redirectMethod := Must(redirectSpec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(redirectMethod, "redirectmethod")

		noredirectMethod := Must(noredirectSpec.AccessMethod(&cpi.DummyComponentVersionAccess{ctx}))
		defer Close(noredirectMethod, "noredirectmethod")

		redirectContent := Must(redirectMethod.Get())
		Expect(string(redirectContent)).To(Equal(CONTENT))

		noredirectContent := Must(noredirectMethod.Get())
		Expect(string(noredirectContent)).To(Equal(NOREDIRECT_CONTENT))
	})
})
