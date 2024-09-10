package wget

import (
	gocontext "context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"mime"
	"net/http"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/tech/wget/identity"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/blobaccess/file"
	ocmmime "ocm.software/ocm/api/utils/mime"
)

const (
	CACHE_CONTENT_THRESHOLD = 4096
)

func DataAccess(url string, opts ...Option) (bpi.DataAccess, error) {
	blobAccess, err := BlobAccess(url, opts...)
	if err != nil {
		return nil, err
	}
	return blobAccess, nil
}

func BlobAccess(url string, opts ...Option) (_ bpi.BlobAccess, rerr error) {
	eff := optionutils.EvalOptions(opts...)
	log := eff.Logger("URL", url)

	creds, err := eff.GetCredentials(url)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		log.Debug("no credentials found for url {{url}}", "url", url)
	}

	// configure http client
	rootCAs, err := credentials.GetRootCAs(eff.CredentialContext, creds)
	if rerr != nil {
		return nil, err
	}
	clientCerts, err := credentials.GetClientCerts(eff.CredentialContext, creds)
	if err != nil {
		return nil, errors.New("client certificate and private key provided in credentials could not be loaded " +
			"as tls certificate")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:   tls.VersionTLS13,
			RootCAs:      rootCAs,
			Certificates: clientCerts,
		},
	}

	var redirectFunc func(req *http.Request, via []*http.Request) error = nil
	if eff.NoRedirect != nil && *eff.NoRedirect {
		redirectFunc = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	client := &http.Client{
		CheckRedirect: redirectFunc,
		Transport:     transport,
	}

	if eff.Verb == "" {
		eff.Verb = http.MethodGet
	}

	// configure http request
	request, err := http.NewRequestWithContext(gocontext.Background(), eff.Verb, url, eff.Body)
	if err != nil {
		return nil, err
	}

	if eff.Header != nil {
		for key, arr := range eff.Header {
			for _, el := range arr {
				request.Header.Add(key, el)
			}
		}
	}

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

	// make http request
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer errors.PropagateError(&rerr, resp.Body.Close)
	log.Debug("http status code {{code}}", "code", resp.StatusCode)

	// determine effective mime type
	if eff.MimeType == "" {
		log.Debug("no mime type provided as option, trying to extract mime type from content type " +
			"response header")

		contentType := resp.Header.Get("Content-Type")
		eff.MimeType, _, err = mime.ParseMediaType(contentType)
		if err != nil {
			log.Debug("failed to get mime type from content type response header with error {{err}}",
				"err", err)
		}
	}
	if eff.MimeType == "" {
		log.Debug("no mime type was provided as content type header of the http response, trying to" +
			"extract mime type from url")
		ext, err := utils.GetFileExtensionFromUrl(url)
		if err == nil && ext != "" {
			eff.MimeType = mime.TypeByExtension(ext)
		} else if err != nil {
			log.Debug(err.Error())
		}
	}
	if eff.MimeType == "" {
		eff.MimeType = ocmmime.MIME_OCTET
		log.Debug("no mime type could be extract from the url, defaulting to {{default}}", "default",
			eff.MimeType)
	}

	// download content
	var blob cpi.BlobAccess
	if resp.ContentLength < 0 || resp.ContentLength > CACHE_CONTENT_THRESHOLD {
		log.Debug("download to file because content length is unknown or greater than {{threshold}}", "threshold", CACHE_CONTENT_THRESHOLD)
		f, err := file.NewTempFile("", "wget")
		if err != nil {
			return nil, err
		}
		defer errors.PropagateError(&rerr, f.Close)

		n, err := io.Copy(f.Writer(), resp.Body)
		if err != nil {
			return nil, err
		}
		log.Debug("downloaded size {{size}} to {{filepath}}", "size", n, "filepath", f.Name())

		blob = f.AsBlob(eff.MimeType)
	} else {
		log.Debug("download to memory because content length is less than {{threshold}}", "threshold", CACHE_CONTENT_THRESHOLD)
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		blob = blobaccess.ForData(eff.MimeType, buf)
	}

	return blob, nil
}

func Provider(url string, opts ...Option) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		b, err := BlobAccess(url, opts...)
		return b, err
	})
}
