package identity

import (
	. "net/url"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/logging"
)

const (
	// CONSUMER_TYPE is the npm repository type.
	CONSUMER_TYPE = "NpmRegistry"

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
var REALM = logging.DefineSubRealm("NPM registry", "npm")

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "the basic auth user name",
		ATTR_PASSWORD, "the basic auth password",
		ATTR_EMAIL, "NPM registry, require an email address",
		ATTR_TOKEN, "the token attribute. May exist after login at any npm registry. Check your .npmrc file!",
	})

	cpi.RegisterStandardIdentity(CONSUMER_TYPE, hostpath.IdentityMatcher(CONSUMER_TYPE), `NPM registry

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

var identityMatcher = hostpath.IdentityMatcher(CONSUMER_TYPE)

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
	return identityMatcher(pattern, cur, id)
}

func GetConsumerId(rawURL, groupId string) (cpi.ConsumerIdentity, error) {
	url, err := JoinPath(rawURL, groupId)
	if err != nil {
		return nil, err
	}
	return hostpath.GetConsumerIdentity(CONSUMER_TYPE, url), nil
}

func GetCredentials(ctx cpi.ContextProvider, repoUrl string, pkgName string) (cpi.Credentials, error) {
	id, err := GetConsumerId(repoUrl, pkgName)
	if err != nil {
		return nil, err
	}
	if id == nil {
		logging.DynamicLogger(REALM).Debug("No consumer identity found.", "url", repoUrl, "groupId", pkgName)
		return nil, nil
	}
	return cpi.CredentialsForConsumer(ctx.CredentialsContext(), id)
}
