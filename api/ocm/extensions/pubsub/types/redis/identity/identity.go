package identity

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/utils/listformat"
)

const CONSUMER_TYPE = "Github"

// identity properties.
const (
	ID_HOSTNAME   = hostpath.ID_HOSTNAME
	ID_PORT       = hostpath.ID_PORT
	ID_PATHPREFIX = hostpath.ID_PATHPREFIX
	ID_CHANNEL    = "channel"
	ID_DATABASE   = "database"
)

// credential properties.
const (
	ATTR_USERNAME = cpi.ATTR_USERNAME
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
)

func IdentityMatcher(request, cur, id cpi.ConsumerIdentity) bool {
	match, better := hostpath.Match(CONSUMER_TYPE, request, cur, id)
	if !match {
		return false
	}

	if request[ID_CHANNEL] != "" {
		if id[ID_CHANNEL] != "" && id[ID_CHANNEL] != request[ID_CHANNEL] {
			return false
		}
	}
	if request[ID_DATABASE] != "" {
		if id[ID_DATABASE] != "" && id[ID_DATABASE] != request[ID_DATABASE] {
			return false
		}
	}

	// ok now it basically matches, check against current match

	if cur[ID_CHANNEL] == "" && request[ID_CHANNEL] != "" {
		return true
	}
	if cur[ID_DATABASE] == "" && request[ID_DATABASE] != "" {
		return true
	}
	return better
}

func init() {
	attrs := listformat.FormatListElements("", listformat.StringElementDescriptionList{
		ATTR_USERNAME, "Redis username",
		ATTR_PASSWORD, "Redis password",
	})
	cpi.RegisterStandardIdentity(CONSUMER_TYPE, IdentityMatcher,
		`Redis PubSub credential matcher

This matcher is a hostpath matcher with additional attributes:

- *<code>`+ID_CHANNEL+`</code>* (required if set in pattern): the channel name 
- *<code>`+ID_DATABASE+`</code>* the database number
`,
		attrs)
}

func PATCredentials(user, pass string) cpi.Credentials {
	return cpi.DirectCredentials{
		ATTR_USERNAME: user,
		ATTR_PASSWORD: pass,
	}
}

func GetConsumerId(serveraddr string, channel string, db int) cpi.ConsumerIdentity {
	p := ""
	host, port, err := ParseAddress(serveraddr)
	if err == nil {
		host = serveraddr
	}

	id := cpi.ConsumerIdentity{
		cpi.ID_TYPE: CONSUMER_TYPE,
		ID_HOSTNAME: host,
		ID_CHANNEL:  channel,
		ID_DATABASE: fmt.Sprintf("%d", db),
	}
	if port != 0 {
		id[ID_PORT] = fmt.Sprintf("%d", port)
	}
	if p != "" {
		id[ID_PATHPREFIX] = p
	}
	return id
}

func GetCredentials(ctx cpi.ContextProvider, serverurl string, channel string, db int) (cpi.Credentials, error) {
	id := GetConsumerId(serverurl, channel, db)
	return cpi.CredentialsForConsumer(ctx.CredentialsContext(), id, IdentityMatcher)
}

func ParseAddress(addr string) (string, int, error) {
	idx := strings.Index(addr, ":")
	if idx < 0 {
		return addr, 6379, nil
	}
	p, err := strconv.ParseInt(addr[idx+1:], 10, 32)
	if err != nil {
		return "", 0, errors.Wrapf(err, "invalid port in redis address")
	}
	return addr[:idx], int(p), nil
}
