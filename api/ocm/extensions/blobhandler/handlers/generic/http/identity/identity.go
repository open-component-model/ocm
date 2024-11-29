package identity

import (
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/utils/listformat"
	common "ocm.software/ocm/api/utils/misc"
)

// CONSUMER_TYPE is the Helm chart repository type.
const CONSUMER_TYPE = "HTTPUploader"

// ID_TYPE is the type field of a consumer identity.
const ID_TYPE = cpi.ID_TYPE

// ID_SCHEME is the scheme of the repository.
const ID_SCHEME = hostpath.ID_SCHEME

// ID_HOSTNAME is the hostname of a repository.
const ID_HOSTNAME = hostpath.ID_HOSTNAME

// ID_PORT is the port number of a repository.
const ID_PORT = hostpath.ID_PORT

// ID_PATHPREFIX is the path of a repository.
const ID_PATHPREFIX = hostpath.ID_PATHPREFIX

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "the basic auth user name",
		ATTR_PASSWORD, "the basic auth password",
		ATTR_CERTIFICATE, "TLS client certificate",
		ATTR_PRIVATE_KEY, "TLS private key",
		ATTR_CERTIFICATE_AUTHORITY, "TLS certificate authority",
	})

	cpi.RegisterStandardIdentity(CONSUMER_TYPE, IdentityMatcher, `HTTPUploader

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

var identityMatcher = hostpath.IdentityMatcher("")

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
	return identityMatcher(pattern, cur, id)
}

// used credential attributes

const (
	ATTR_USERNAME              = cpi.ATTR_USERNAME
	ATTR_PASSWORD              = cpi.ATTR_PASSWORD
	ATTR_CERTIFICATE_AUTHORITY = cpi.ATTR_CERTIFICATE_AUTHORITY
	ATTR_CERTIFICATE           = cpi.ATTR_CERTIFICATE
	ATTR_PRIVATE_KEY           = cpi.ATTR_PRIVATE_KEY
	ATTR_TOKEN                 = cpi.ATTR_TOKEN
)

func GetCredentials(ctx cpi.ContextProvider, consumerType string, url string) common.Properties {
	if consumerType == "" {
		consumerType = CONSUMER_TYPE
	}
	id := hostpath.GetConsumerIdentity(consumerType, url)
	if id == nil {
		return nil
	}
	creds, err := cpi.CredentialsForConsumer(ctx.CredentialsContext(), id)
	if creds == nil || err != nil {
		return nil
	}
	return creds.Properties()
}
