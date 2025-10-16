package identity

import (
	"net"
	"net/url"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/utils/listformat"
)

const CONSUMER_TYPE = "HashiCorpVault"

// identity properties.
const (
	ID_HOSTNAME   = hostpath.ID_HOSTNAME
	ID_SCHEMA     = hostpath.ID_SCHEME
	ID_PORT       = hostpath.ID_PORT
	ID_PATHPREFIX = hostpath.ID_PATHPREFIX
	ID_MOUNTPATH  = "mountPath"
	ID_NAMESPACE  = "namespace"
)

// credential properties.
const (
	ATTR_AUTHMETH = "authmeth"
	ATTR_TOKEN    = cpi.ATTR_TOKEN
	ATTR_ROLEID   = "roleid"
	ATTR_SECRETID = "secretid"
)

const (
	AUTH_APPROLE = "approle"
	AUTH_TOKEN   = "token"
)

var identityMatcher = hostpath.IdentityMatcher(CONSUMER_TYPE)

func IdentityMatcher(request, cur, id cpi.ConsumerIdentity) bool {
	if id[ID_NAMESPACE] != request[ID_NAMESPACE] {
		return false
	}
	if id[ID_MOUNTPATH] != "" && id[ID_MOUNTPATH] != request[ID_MOUNTPATH] {
		return false
	}
	return identityMatcher(request, cur, id)
}

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_AUTHMETH, "auth method",
		ATTR_TOKEN, "vault token",
		ATTR_ROLEID, "app-role role id",
		ATTR_SECRETID, "app-role secret id",
	})
	ids := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ID_HOSTNAME, "vault server host",
		ID_SCHEMA, "(optional) URL scheme",
		ID_PORT, "(optional) server port",
		ID_NAMESPACE, "vault namespace",
		ID_MOUNTPATH, "mount path",
		ID_PATHPREFIX, "path prefix for secret",
	})
	cpi.RegisterStandardIdentity(CONSUMER_TYPE, identityMatcher,
		`HashiCorp Vault credential matcher

This matcher matches credentials for a HashiCorp vault instance.
It uses the following identity attributes:
`+ids,
		attrs+`
The only supported auth methods, so far, are <code>token</code> and <code>approle</code>.
`)
}

func GetConsumerId(serverurl string, namespace string, mountpath string, secretpath string) (cpi.ConsumerIdentity, error) {
	if serverurl == "" {
		return nil, errors.Newf("server address must be given")
	}
	u, err := url.Parse(serverurl)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "server url", serverurl)
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		if strings.LastIndex(host, ":") >= 0 {
			return nil, errors.ErrInvalidWrap(err, "server url", serverurl)
		}
		host = u.Host
	}

	id := cpi.ConsumerIdentity{
		cpi.ID_TYPE: CONSUMER_TYPE,
		ID_HOSTNAME: host,
	}
	if u.Scheme != "" {
		id[ID_SCHEMA] = u.Scheme
	}
	if port != "" {
		id[ID_PORT] = port
	}
	if namespace != "" {
		id[ID_NAMESPACE] = namespace
	}
	if mountpath != "" {
		id[ID_MOUNTPATH] = mountpath
	}

	if secretpath != "" {
		id[ID_PATHPREFIX] = secretpath
	}
	return id, nil
}

func GetCredentials(ctx cpi.ContextProvider, serverurl, namespace string, mountpath, secretpath string) (cpi.Credentials, error) {
	id, err := GetConsumerId(serverurl, namespace, mountpath, secretpath)
	if err != nil {
		return nil, err
	}
	return cpi.CredentialsForConsumer(ctx.CredentialsContext(), id, IdentityMatcher)
}
