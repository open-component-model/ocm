package identity

import (
	"net/http"

	. "net/url"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/logging"
)

const (
	// ConsumerType is the mvn repository type.
	ConsumerType = "Repository.maven.apache.org"

	// Username is the username attribute. Required for login at any mvn registry.
	Username = cpi.ATTR_USERNAME
	// Password is the password attribute. Required for login at any mvn registry.
	Password = cpi.ATTR_PASSWORD
)

// REALM the logging realm / prefix.
var REALM = logging.DefineSubRealm("Maven repository", "mvn")

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		Username, "the basic auth user name",
		Password, "the basic auth password",
	})

	cpi.RegisterStandardIdentity(ConsumerType, hostpath.IdentityMatcher(ConsumerType), `MVN repository

It matches the <code>`+ConsumerType+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

func GetConsumerId(rawURL, groupId string) cpi.ConsumerIdentity {
	url, err := JoinPath(rawURL, groupId)
	if err != nil {
		debug("GetConsumerId", "error", err.Error(), "url", rawURL)
		return nil
	}

	return hostpath.GetConsumerIdentity(ConsumerType, url)
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
	username := credentials[Username]
	password := credentials[Password]
	if username == "" || password == "" {
		return
	}
	req.SetBasicAuth(username, password)
}

// debug uses a dynamic logger to log a debug message.
func debug(msg string, keypairs ...interface{}) {
	logging.DynamicLogger(REALM).Debug(msg, keypairs...)
}
