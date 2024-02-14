// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/listformat"
)

// CONSUMER_TYPE is the wget access method type.
const CONSUMER_TYPE = "wget"

// used identity properties.
const (
	ID_TYPE       = hostpath.ID_TYPE
	ID_HOSTNAME   = hostpath.ID_HOSTNAME
	ID_PORT       = hostpath.ID_PORT
	ID_PATHPREFIX = hostpath.ID_PATHPREFIX
	ID_SCHEME     = hostpath.ID_SCHEME
)

// used credential properties.
const (
	ATTR_USERNAME              = cpi.ATTR_USERNAME
	ATTR_PASSWORD              = cpi.ATTR_PASSWORD
	ATTR_IDENTITY_TOKEN        = cpi.ATTR_IDENTITY_TOKEN
	ATTR_CERTIFICATE_AUTHORITY = cpi.ATTR_CERTIFICATE_AUTHORITY
	ATTR_CERTIFICATE           = cpi.ATTR_CERTIFICATE
	ATTR_PRIVATE_KEY           = cpi.ATTR_PRIVATE_KEY
)

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "the basic auth user name",
		ATTR_PASSWORD, "the basic auth password",
		ATTR_IDENTITY_TOKEN, "the bearer token used for non-basic auth authorization",
		ATTR_CERTIFICATE_AUTHORITY, "the certificate authority certificate used to verify certificates presented by the server",
		ATTR_CERTIFICATE, "the certificate used to present to the server",
		ATTR_PRIVATE_KEY, "the private key corresponding to the certificate",
	})

	cpi.RegisterStandardIdentity(CONSUMER_TYPE, IdentityMatcher, `wget credential matcher

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

var identityMatcher = hostpath.IdentityMatcher(CONSUMER_TYPE)

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
	return identityMatcher(pattern, cur, id)
}

func GetConsumerId(url string) cpi.ConsumerIdentity {
	return hostpath.GetConsumerIdentity(CONSUMER_TYPE, url)
}
