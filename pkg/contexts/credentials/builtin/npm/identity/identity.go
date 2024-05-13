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
	// ConsumerType is the npm repository type.
	ConsumerType = "Registry.npmjs.com"

	// Username is the username attribute. Required for login at any npm registry.
	Username = cpi.ATTR_USERNAME
	// Password is the password attribute. Required for login at any npm registry.
	Password = cpi.ATTR_PASSWORD
	// Email is the email attribute. Required for login at any npm registry.
	Email = cpi.ATTR_EMAIL
	// Token is the token attribute. May exist after login at any npm registry.
	Token = cpi.ATTR_TOKEN
)

// REALM the logging realm / prefix.
var REALM = logging.DefineSubRealm("NPM registry", "NPM")

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		Username, "the basic auth user name",
		Password, "the basic auth password",
		Email, "NPM registry, require an email address",
		Token, "the token attribute. May exist after login at any npm registry. Check your .npmrc file!",
	})

	cpi.RegisterStandardIdentity(ConsumerType, hostpath.IdentityMatcher(ConsumerType), `NPM repository

It matches the <code>`+ConsumerType+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

func GetConsumerId(rawURL string, pkgName string) cpi.ConsumerIdentity {
	url, err := JoinPath(rawURL, pkgName)
	if err != nil {
		debug("GetConsumerId", "error", err.Error(), "url", rawURL)
		return nil
	}

	return hostpath.GetConsumerIdentity(ConsumerType, url)
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
