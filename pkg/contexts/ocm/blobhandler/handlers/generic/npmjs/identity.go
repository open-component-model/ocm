package npmjs

import (
	"net/url"
	"path"

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

	cpi.RegisterStandardIdentity(CONSUMER_TYPE, IdentityMatcher, `Npmjs repository

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

var identityMatcher = hostpath.IdentityMatcher("")

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
	return identityMatcher(pattern, cur, id)
}

const (
	ATTR_USERNAME = cpi.ATTR_USERNAME
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
	ATTR_EMAIL    = cpi.ATTR_EMAIL
)

func SimpleCredentials(user, passwd string, email string) cpi.Credentials {
	return cpi.DirectCredentials{
		ATTR_USERNAME: user,
		ATTR_PASSWORD: passwd,
		ATTR_EMAIL:    email,
	}
}

func GetConsumerId(repourl string, pkgname string) cpi.ConsumerIdentity {
	u, err := url.Parse(repourl)
	if err != nil {
		return nil
	}

	u.Path = path.Join(u.Path, pkgname)
	return hostpath.GetConsumerIdentity(CONSUMER_TYPE, u.String())
}

func GetCredentials(ctx cpi.ContextProvider, repourl string, pkgname string) common.Properties {
	id := GetConsumerId(repourl, pkgname)
	if id == nil {
		return nil
	}
	creds, err := cpi.CredentialsForConsumer(ctx.CredentialsContext(), id)
	if creds == nil || err != nil {
		return nil
	}
	return creds.Properties()
}
