package identity

import (
	"net"

	giturls "github.com/chainguard-dev/git-urls"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/utils/listformat"
)

const CONSUMER_TYPE = "Git"

var identityMatcher = hostpath.IdentityMatcher(CONSUMER_TYPE)

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
	return identityMatcher(pattern, cur, id)
}

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "the basic auth user name",
		ATTR_PASSWORD, "the basic auth password",
		ATTR_TOKEN, "HTTP token authentication",
		ATTR_PRIVATE_KEY, "Private Key authentication certificate",
	})
	cpi.RegisterStandardIdentity(CONSUMER_TYPE, identityMatcher,
		`Git credential matcher

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`,
		attrs)
}

const (
	ID_HOSTNAME   = hostpath.ID_HOSTNAME
	ID_PATHPREFIX = hostpath.ID_PATHPREFIX
	ID_PORT       = hostpath.ID_PORT
	ID_SCHEME     = hostpath.ID_SCHEME
)

const (
	ATTR_TOKEN       = cpi.ATTR_TOKEN
	ATTR_USERNAME    = cpi.ATTR_USERNAME
	ATTR_PASSWORD    = cpi.ATTR_PASSWORD
	ATTR_PRIVATE_KEY = cpi.ATTR_PRIVATE_KEY
)

func GetConsumerId(repoURL string) (cpi.ConsumerIdentity, error) {
	host := ""
	port := ""
	defaultPort := ""
	scheme := ""
	path := ""

	if repoURL != "" {
		u, err := giturls.Parse(repoURL)
		if err == nil {
			host = u.Host
		} else {
			return nil, err
		}

		scheme = u.Scheme
		switch scheme {
		case "http":
			defaultPort = "80"
		case "https":
			defaultPort = "443"
		case "git":
			defaultPort = "9418"
		case "ssh":
			defaultPort = "22"
		case "file":
			host = "localhost"
			path = u.Path
		}
	}

	if h, p, err := net.SplitHostPort(host); err == nil {
		host, port = h, p
	}

	id := cpi.ConsumerIdentity{
		cpi.ID_TYPE: CONSUMER_TYPE,
		ID_HOSTNAME: host,
	}

	if port != "" {
		id[ID_PORT] = port
	} else if defaultPort != "" {
		id[ID_PORT] = defaultPort
	}

	if path != "" {
		id[ID_PATHPREFIX] = path
	}

	id[ID_SCHEME] = scheme

	return id, nil
}

func TokenCredentials(token string) cpi.Credentials {
	return cpi.DirectCredentials{
		ATTR_TOKEN: token,
	}
}

func BasicAuthCredentials(username, password string) cpi.Credentials {
	return cpi.DirectCredentials{
		ATTR_USERNAME: username,
		ATTR_PASSWORD: password,
	}
}

func PrivateKeyCredentials(username, privateKey string) cpi.Credentials {
	return cpi.DirectCredentials{
		ATTR_USERNAME:    username,
		ATTR_PRIVATE_KEY: privateKey,
	}
}

func GetCredentials(ctx cpi.ContextProvider, repoURL string) (cpi.Credentials, error) {
	id, err := GetConsumerId(repoURL)
	if err != nil {
		return nil, err
	}
	return cpi.CredentialsForConsumer(ctx.CredentialsContext(), id, IdentityMatcher)
}
