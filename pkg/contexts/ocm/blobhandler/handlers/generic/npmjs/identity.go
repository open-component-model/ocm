package npmjs

import (
	"path"

	. "net/url"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/listformat"
)

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "the basic auth user name",
		ATTR_PASSWORD, "the basic auth password",
		ATTR_EMAIL, "NPM registry, require an email address",
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
