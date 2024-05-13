package identity

import (
	"errors"
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

var identityMatcher = hostpath.IdentityMatcher(ConsumerType)

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
	return identityMatcher(pattern, cur, id)
}

func GetConsumerId(rawURL, groupId string) (cpi.ConsumerIdentity, error) {
	url, err := JoinPath(rawURL, groupId)
	if err != nil {
		return nil, err
	}
	return hostpath.GetConsumerIdentity(ConsumerType, url), nil
}

func GetCredentials(ctx cpi.ContextProvider, repoUrl, groupId string) (common.Properties, error) {
	id, err := GetConsumerId(repoUrl, groupId)
	if err != nil {
		return nil, err
	}
	if id == nil {
		return nil, nil
	}
	credentials, err := cpi.CredentialsForConsumer(ctx.CredentialsContext(), id)
	if err != nil {
		return nil, err
	}
	if credentials == nil {
		return nil, nil
	}
	return credentials.Properties(), nil
}

func BasicAuth(req *http.Request, ctx accspeccpi.Context, repoUrl, groupId string) (err error) {
	credentials, err := GetCredentials(ctx, repoUrl, groupId)
	if err != nil {
		return
	}
	if credentials == nil {
		return
	}
	username := credentials[Username]
	password := credentials[Password]
	if username == "" || password == "" {
		return errors.New("missing username or password in credentials")
	}
	req.SetBasicAuth(username, password)
	return
}
