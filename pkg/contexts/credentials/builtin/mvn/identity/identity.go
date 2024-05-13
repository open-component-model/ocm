package identity

import (
	"net/http"
	"path"

	. "net/url"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/logging"
)

const (
	// CONSUMER_TYPE is the mvn repository type.
	CONSUMER_TYPE = "Repository.maven.apache.org"

	// ATTR_USERNAME is the username attribute. Required for login at any mvn registry.
	ATTR_USERNAME = cpi.ATTR_USERNAME
	// ATTR_PASSWORD is the password attribute. Required for login at any mvn registry.
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
)

// Logging Realm.
var REALM = logging.DefineSubRealm("Maven repository", "mvn")

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "the basic auth user name",
		ATTR_PASSWORD, "the basic auth password",
	})

	cpi.RegisterStandardIdentity(CONSUMER_TYPE, hostpath.IdentityMatcher(CONSUMER_TYPE), `MVN repository

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

func GetConsumerId(rawURL, groupId string) cpi.ConsumerIdentity {
	url, err := Parse(rawURL)
	if err != nil {
		debug("GetConsumerId", "error", err.Error(), "url", rawURL)
		return nil
	}

	url.Path = path.Join(url.Path, groupId)
	return hostpath.GetConsumerIdentity(CONSUMER_TYPE, url.String())
}

func GetCredentials(ctx cpi.ContextProvider, repoUrl, groupId string) common.Properties {
	id := GetConsumerId(repoUrl, groupId)
	if id == nil {
		return nil
	}
	credentials, err := cpi.CredentialsForConsumer(ctx.CredentialsContext(), id)
	if err != nil {
		debug("GetCredentials", "error", err.Error())
		return nil
	}
	if credentials == nil {
		debug("no credentials found")
		return nil
	}
	return credentials.Properties()
}

func BasicAuth(req *http.Request, ctx accspeccpi.Context, repoUrl, groupId string) {
	credentials := GetCredentials(ctx, repoUrl, groupId)
	if credentials == nil {
		return
	}
	username := credentials[ATTR_USERNAME]
	password := credentials[ATTR_PASSWORD]
	if username == "" || password == "" {
		return
	}
	req.SetBasicAuth(username, password)
}

// debug uses a dynamic logger to log a debug message.
func debug(msg string, keypairs ...interface{}) {
	logging.DynamicLogger(REALM).Debug(msg, keypairs...)
}
