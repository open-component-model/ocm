// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/internal"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/rootcertsattr"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
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

func GetRootCAs(ctx ContextProvider, creds Credentials) *x509.CertPool {
	var rootCAs *x509.CertPool
	if creds != nil {
		c := creds.GetProperty(internal.ATTR_CERTIFICATE_AUTHORITY)
		if c != "" {
			rootCAs = x509.NewCertPool()
			rootCAs.AppendCertsFromPEM([]byte(c))
		}
	}
	if rootCAs == nil {
		rootCAs = rootcertsattr.Get(ctx.CredentialsContext()).GetRootCertPool(true)
	}
	return rootCAs
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
