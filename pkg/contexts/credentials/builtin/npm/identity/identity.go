package identity

import (
	"fmt"
	"net/http"
	"path"

	. "net/url"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/npm"
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
	url, err := Parse(rawURL)
	if err != nil {
		return nil
	}

	url.Path = path.Join(url.Path, pkgName)
	return hostpath.GetConsumerIdentity(CONSUMER_TYPE, url.String())
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

// BearerToken retrieves the bearer token for the given repository URL and package name.
// Either it's setup in the credentials or it will login to the registry and retrieve it.
func BearerToken(ctx cpi.ContextProvider, repoUrl string, pkgName string) (string, error) {
	// get credentials and TODO cache it
	cred := GetCredentials(ctx, repoUrl, pkgName)
	if cred == nil {
		return "", fmt.Errorf("no credentials found for %s. Couldn't upload '%s'", repoUrl, pkgName)
	}
	log := logging.Context().Logger(REALM)
	log.Debug("found credentials")

	// check if token exists, if not login and retrieve token
	token := cred[ATTR_TOKEN]
	if token != "" {
		log.Debug("token found, skipping login")
		return token, nil
	}

	// use user+pass+mail from credentials to login and retrieve bearer token
	username := cred[ATTR_USERNAME]
	password := cred[ATTR_PASSWORD]
	email := cred[ATTR_EMAIL]
	if username == "" || password == "" || email == "" {
		return "", fmt.Errorf("credentials for %s are invalid. Username, password or email missing! Couldn't upload '%s'", repoUrl, pkgName)
	}
	log = log.WithValues("user", username, "repo", repoUrl)
	log.Debug("login")

	// TODO: check different kinds of .npmrc content
	return npm.Login(repoUrl, username, password, email)
}

// Authorize the given request with the bearer token for the given repository URL and package name.
// If the token is empty (login failed or credentials not found), it will not be set.
func Authorize(req *http.Request, ctx cpi.ContextProvider, repoUrl string, pkgName string) {
	token, err := BearerToken(ctx, repoUrl, pkgName)
	if err != nil {
		log := logging.Context().Logger(REALM)
		log.Debug("Couldn't authorize", "error", err.Error(), "repo", repoUrl, "package", pkgName)
	} else if token != "" {
		req.Header.Set("authorization", "Bearer "+token)
	}
}
