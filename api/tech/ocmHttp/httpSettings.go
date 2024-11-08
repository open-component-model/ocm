package ocmHttp

import (
	"crypto/tls"
	"encoding/base64"
	"net/http"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils/logging"
)

func NewHttpClient(ctx credentials.ContextProvider, creds credentials.Credentials) *http.Client {
	rootCAs, err := credentials.GetRootCAs(ctx, creds)
	if err != nil {
		logging.DynamicLogger("http").Error("could not load root CAs for http client")
	}
	clientCerts, err := credentials.GetClientCerts(ctx, creds)
	if err != nil {
		logging.DynamicLogger("http").Error("client certificate and private key provided in credentials could not be loaded " +
			"as tls certificate")
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:   tls.VersionTLS13,
			RootCAs:      rootCAs,
			Certificates: clientCerts,
		},
	}
	client := &http.Client{
		Transport: transport,
	}
	return client
}

func Redirect(client *http.Client, noRedirect *bool) *http.Client {
	if noRedirect != nil && *noRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client
}

func Authorize(request *http.Request, creds credentials.Credentials) *http.Request {
	if creds != nil {
		user := creds.GetProperty(identity.ATTR_USERNAME)
		password := creds.GetProperty(identity.ATTR_PASSWORD)
		token := creds.GetProperty(identity.ATTR_IDENTITY_TOKEN)

		if user != "" && password != "" {
			auth := user + ":" + password
			auth = base64.StdEncoding.EncodeToString([]byte(auth))
			request.Header.Add("Authorization", "Basic "+auth)
		} else if token != "" {
			request.Header.Add("Authorization", "Bearer "+token)
		}
	}
	return request
}
