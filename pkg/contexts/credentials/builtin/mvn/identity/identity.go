package identity

import (
	"path"

	. "net/url"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/logging"
)

const (
	// CONSUMER_TYPE is the mvn repository type.
	CONSUMER_TYPE = "Repository.maven.apache.org"

	// FIXME use correct settings for maven repo authentication

	// ATTR_USERNAME is the username attribute. Required for login at any mvn registry.
	ATTR_USERNAME = cpi.ATTR_USERNAME
	// ATTR_PASSWORD is the password attribute. Required for login at any mvn registry.
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
)

// Logging Realm.
var REALM = logging.DefineSubRealm("MVN repository", "MVN")

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
	if credentials == nil || err != nil {
		return nil
	}
	return credentials.Properties()
}
