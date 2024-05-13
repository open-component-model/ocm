package identity

import (
	. "net/url"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/logging"
)

const (
	// CONSUMER_TYPE is the npm repository type.
	CONSUMER_TYPE = "Registry.npmjs.com"

	// ATTR_USERNAME is the username attribute. Required for login at any npm registry.
	ATTR_USERNAME = cpi.ATTR_USERNAME
	// ATTR_PASSWORD is the password attribute. Required for login at any npm registry.
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
	// ATTR_EMAIL is the email attribute. Required for login at any npm registry.
	ATTR_EMAIL = cpi.ATTR_EMAIL
	// ATTR_TOKEN is the token attribute. May exist after login at any npm registry.
	ATTR_TOKEN = cpi.ATTR_TOKEN
)

// Logging Realm.
var REALM = logging.DefineSubRealm("NPM registry", "NPM")

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "the basic auth user name",
		ATTR_PASSWORD, "the basic auth password",
		ATTR_EMAIL, "NPM registry, require an email address",
		ATTR_TOKEN, "the token attribute. May exist after login at any npm registry. Check your .npmrc file!",
	})

	cpi.RegisterStandardIdentity(CONSUMER_TYPE, hostpath.IdentityMatcher(CONSUMER_TYPE), `NPM repository

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

func GetConsumerId(rawURL string, pkgName string) cpi.ConsumerIdentity {
	url, err := JoinPath(rawURL, pkgName)
	if err != nil {
		debug("GetConsumerId", "error", err.Error(), "url", rawURL)
		return nil
	}

	return hostpath.GetConsumerIdentity(CONSUMER_TYPE, url)
}

func GetCredentials(ctx cpi.ContextProvider, repoUrl string, pkgName string) common.Properties {
	id := GetConsumerId(repoUrl, pkgName)
	if id == nil {
		return nil
	}
	credentials, err := cpi.CredentialsForConsumer(ctx.CredentialsContext(), id)
	if credentials == nil || err != nil {
		return nil
	}
	return credentials.Properties()
}

// debug uses a dynamic logger to log a debug message.
func debug(msg string, keypairs ...interface{}) {
	logging.DynamicLogger(REALM).Debug(msg, keypairs...)
}
