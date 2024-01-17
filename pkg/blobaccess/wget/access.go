package wget

import (
	"crypto/tls"
	"fmt"
	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/wget/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/optionutils"
	"io"
	"net/http"
)

const (
	CACHE_CONTENT_THRESHOLD = 4096
)

func DataAccessForWget(url string, opts ...Option) (blobaccess.DataAccess, error) {
	blobAccess, err := BlobAccessForWget(url, opts...)
	if err != nil {
		return nil, err
	}
	return blobAccess, nil
}

func BlobAccessForWget(url string, opts ...Option) (_ blobaccess.BlobAccess, rerr error) {
	eff := optionutils.EvalOptions(opts...)
	log := eff.Logger("URL", fmt.Sprintf("%s", url))

	creds, err := eff.GetCredentials(url)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		log.Debug("no credentials found for", "url", url)
	}

	// configure http client
	rootCAs, err := credentials.GetRootCAs(eff.CredentialContext, creds)
	clientCerts, err := credentials.GetClientCerts(eff.CredentialContext, creds)
	if err != nil {
		return nil, errors.New("client certificate and private key provided in credentials could not be loaded " +
			"as tls certificate")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      rootCAs,
			Certificates: clientCerts,
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	// configure http request
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if creds != nil {
		user := creds.GetProperty(identity.ATTR_USERNAME)
		password := creds.GetProperty(identity.ATTR_PASSWORD)
		token := creds.GetProperty(identity.ATTR_IDENTITY_TOKEN)

		if user != "" && password != "" {
			request.SetBasicAuth(user, password)
		} else if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer errors.PropagateError(&rerr, resp.Body.Close)
	log.Debug("http status code", "", resp.StatusCode)

	var blob cpi.BlobAccess
	if resp.ContentLength < 0 || resp.ContentLength > CACHE_CONTENT_THRESHOLD {
		log.Debug("download to file because content length is", "unkown or greater than", CACHE_CONTENT_THRESHOLD)
		f, err := blobaccess.NewTempFile("", "wget")
		if err != nil {
			return nil, err
		}
		defer errors.PropagateError(&rerr, f.Close)

		n, err := io.Copy(f.Writer(), resp.Body)
		if err != nil {
			return nil, err
		}
		log.Debug("downloaded", "size", n, "to", f.Name())

		if eff.MimeType == "" {
			eff.MimeType = mime.MIME_OCTET
			log.Debug("provided mimeType is empty and is therefore defaulted to", "mimeType", mime.MIME_OCTET)
		}
		blob = f.AsBlob(eff.MimeType)
	} else {
		log.Debug("download to memory because content length is", "less than", CACHE_CONTENT_THRESHOLD)
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		blob = blobaccess.ForData(eff.MimeType, buf)
	}

	return blob, nil
}

func BlobAccessProviderForWget(url string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, err := BlobAccessForWget(url, opts...)
		return b, err
	})
}
