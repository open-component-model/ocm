package hostpath

import (
	"net/url"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
)

// IDENTITY_TYPE is the identity of this matcher.
const IDENTITY_TYPE = "hostpath"

// ID_TYPE is the type of the consumer.
const ID_TYPE = cpi.ID_TYPE

// ID_HOSTNAME is a hostname.
const ID_HOSTNAME = "hostname"

// ID_PORT is a port.
const ID_PORT = "port"

// ID_PATHPREFIX is the path prefix below the host.
const ID_PATHPREFIX = "pathprefix"

// ID_SCHEME is the scheme prefix.
const ID_SCHEME = "scheme"

func init() {
	cpi.RegisterStandardIdentityMatcher(IDENTITY_TYPE, Matcher, `Host and path based credential matcher

This matcher works on the following properties:

- *<code>`+ID_TYPE+`</code>* (required if set in pattern): the identity type 
- *<code>`+ID_HOSTNAME+`</code>* (required if set in pattern): the hostname of a server
- *<code>`+ID_SCHEME+`</code>* (optional): the URL scheme of a server
- *<code>`+ID_PORT+`</code>* (optional): the port of a server
- *<code>`+ID_PATHPREFIX+`</code>* (optional): a path prefix to match. The 
  element with the most matching path components is selected (separator is <code>/</code>).
`)
}

var Matcher = IdentityMatcher("")

func IdentityMatcher(identityType string) cpi.IdentityMatcher {
	return func(request, cur, id cpi.ConsumerIdentity) bool {
		if request[ID_TYPE] != "" && request[ID_TYPE] != id[ID_TYPE] {
			return false
		}

		if identityType != "" && request[ID_TYPE] != "" && identityType != request[ID_TYPE] {
			return false
		}

		if request[ID_HOSTNAME] != "" && id[ID_HOSTNAME] != "" && request[ID_HOSTNAME] != id[ID_HOSTNAME] {
			return false
		}

		if request[ID_PORT] != "" {
			if id[ID_PORT] != "" && id[ID_PORT] != request[ID_PORT] {
				return false
			}
		}

		if request[ID_SCHEME] != "" {
			if id[ID_SCHEME] != "" && id[ID_SCHEME] != request[ID_SCHEME] {
				return false
			}
		}

		if request[ID_PATHPREFIX] != "" {
			if id[ID_PATHPREFIX] != "" {
				if len(id[ID_PATHPREFIX]) > len(request[ID_PATHPREFIX]) {
					return false
				}
				pcomps := strings.Split(request[ID_PATHPREFIX], "/")
				icomps := strings.Split(id[ID_PATHPREFIX], "/")
				if len(icomps) > len(pcomps) {
					return false
				}
				for i := range icomps {
					if pcomps[i] != icomps[i] {
						return false
					}
				}
			}
		} else {
			if id[ID_PATHPREFIX] != "" {
				return false
			}
		}

		// ok now it basically matches, check against current match
		if len(cur) == 0 {
			return true
		}

		if cur[ID_HOSTNAME] == "" && id[ID_HOSTNAME] != "" {
			return true
		}
		if cur[ID_PORT] == "" && (id[ID_PORT] != "" && request[ID_PORT] != "") {
			return true
		}
		if cur[ID_SCHEME] == "" && (id[ID_SCHEME] != "" && request[ID_SCHEME] != "") {
			return true
		}

		if len(cur[ID_PATHPREFIX]) < len(id[ID_PATHPREFIX]) {
			return true
		}
		return false
	}
}

func GetConsumerIdentity(typ, _url string) cpi.ConsumerIdentity {
	u, err := url.Parse(_url)
	if err != nil {
		return nil
	}

	id := cpi.NewConsumerIdentity(typ)
	if u.Host != "" {
		parts := strings.Split(u.Host, ":")
		if len(parts) > 1 {
			id[ID_PORT] = parts[1]
		} else {
			switch u.Scheme {
			case "https":
				id[ID_PORT] = "443"
			case "http":
				id[ID_PORT] = "80"
			}
		}
		id[ID_HOSTNAME] = parts[0]
	}
	if u.Scheme != "" {
		id[ID_SCHEME] = u.Scheme
	}
	path := strings.Trim(u.Path, "/")
	if path != "" {
		id[ID_PATHPREFIX] = path
	}
	return id
}
