package credentials

import (
	"crypto/tls"
	"crypto/x509"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/texttheater/golang-levenshtein/levenshtein"

	"ocm.software/ocm/api/credentials/internal"
	"ocm.software/ocm/api/datacontext/attrs/rootcertsattr"
	"ocm.software/ocm/api/utils"
)

func GetProvidedConsumerId(obj interface{}, uctx ...UsageContext) ConsumerIdentity {
	return utils.UnwrappingCall(obj, func(provider ConsumerIdentityProvider) ConsumerIdentity {
		return provider.GetConsumerId(uctx...)
	})
}

func GetProvidedIdentityMatcher(obj interface{}) string {
	return utils.UnwrappingCall(obj, func(provider ConsumerIdentityProvider) string {
		return provider.GetIdentityMatcher()
	})
}

func CredentialsFor(ctx ContextProvider, obj interface{}, uctx ...UsageContext) (Credentials, error) {
	id := GetProvidedConsumerId(obj, uctx...)
	if id == nil {
		return nil, errors.ErrNotSupported(KIND_CONSUMER)
	}
	return CredentialsForConsumer(ctx, id)
}

func GetRootCAs(ctx ContextProvider, creds Credentials) (*x509.CertPool, error) {
	var rootCAs *x509.CertPool
	var err error

	if creds != nil {
		c := creds.GetProperty(internal.ATTR_CERTIFICATE_AUTHORITY)
		if c != "" {
			rootCAs = x509.NewCertPool()
			rootCAs.AppendCertsFromPEM([]byte(c))
		}
	}
	if rootCAs == nil {
		if ctx != nil {
			rootCAs = rootcertsattr.Get(ctx.CredentialsContext()).GetRootCertPool(true)
		} else {
			rootCAs, err = x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
		}
	}
	return rootCAs, nil
}

func GetClientCerts(ctx ContextProvider, creds Credentials) ([]tls.Certificate, error) {
	if creds != nil {
		cert := creds.GetProperty(internal.ATTR_CERTIFICATE)
		priv := creds.GetProperty(internal.ATTR_PRIVATE_KEY)
		if cert == "" && priv == "" {
			return nil, nil
		}
		if cert == "" || priv == "" {
			return nil, errors.New("both, private key and certificate are required for tls client authentication")
		}
		if cert != "" && priv != "" {
			tlsCert, err := tls.X509KeyPair([]byte(cert), []byte(priv))
			if err != nil {
				return nil, err
			}
			return []tls.Certificate{tlsCert}, nil
		}
	}
	return nil, nil
}

func GuessConsumerType(ctxp ContextProvider, spec string) string {
	matchers := ctxp.CredentialsContext().ConsumerIdentityMatchers()
	lspec := strings.ToLower(spec)

	if matchers.Get(spec) == nil {
		fix := ""
		for _, i := range matchers.List() {
			idx := strings.Index(i.Type, ".")
			if idx > 0 && i.Type[:idx] == spec {
				fix = i.Type
				break
			}
		}
		if fix == "" {
			for _, i := range matchers.List() {
				if strings.ToLower(i.Type) == lspec {
					fix = i.Type
					break
				}
			}
		}
		if fix == "" {
			for _, i := range matchers.List() {
				idx := strings.Index(i.Type, ".")
				if idx > 0 && strings.ToLower(i.Type[:idx]) == lspec {
					fix = i.Type
					break
				}
			}
		}
		if fix == "" {
			minVal := -1
			for _, i := range matchers.List() {
				idx := strings.Index(i.Type, ".")
				if idx > 0 {
					d := levenshtein.DistanceForStrings([]rune(lspec), []rune(strings.ToLower(i.Type[:idx])), levenshtein.DefaultOptions)
					if d < 5 && fix == "" || minVal > d {
						fix = i.Type
						minVal = d
					}
				}
			}
		}
		if fix == "" {
			minVal := -1
			for _, i := range matchers.List() {
				d := levenshtein.DistanceForStrings([]rune(lspec), []rune(strings.ToLower(i.Type)), levenshtein.DefaultOptions)
				if d < 5 && fix == "" || minVal > d {
					fix = i.Type
					minVal = d
				}
			}
		}
		if fix != "" {
			return fix
		}
	}
	return spec
}
