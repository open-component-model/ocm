package npmjs

import (
	"path"

	. "net/url"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/listformat"
)

// CONSUMER_TYPE is the npmjs repository type.
const CONSUMER_TYPE = "Registry.npmjs.com"

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "the basic auth user name",
		ATTR_PASSWORD, "the basic auth password",
		ATTR_EMAIL, "npmjs registries, require an email address",
	})

	cpi.RegisterStandardIdentity(CONSUMER_TYPE, hostpath.IdentityMatcher(CONSUMER_TYPE), `Npmjs repository

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

const (
	ATTR_USERNAME = cpi.ATTR_USERNAME
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
	ATTR_EMAIL    = cpi.ATTR_EMAIL
)

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
