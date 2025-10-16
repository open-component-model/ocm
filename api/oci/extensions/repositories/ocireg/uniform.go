package ocireg

import (
	regex "github.com/mandelsoft/goutils/regexutils"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/grammar"
)

func init() {
	cpi.RegisterRepositorySpecHandler(&repospechandler{}, Type, "")
}

type repospechandler struct{}

func (h *repospechandler) MapReference(ctx cpi.Context, u *cpi.UniformRepositorySpec) (cpi.RepositorySpec, error) {
	scheme := u.Scheme
	host := u.Host
	if u.Host == "" && u.Scheme == "" && u.Info != "" {
		host = u.Info
		match := grammar.AnchoredSchemedRegexp.FindStringSubmatch(host)
		if match != nil {
			scheme = match[1]
			host = match[2]
		}
		if !(regex.Anchored(grammar.HostPortRegexp).MatchString(host) || regex.Anchored(grammar.DomainPortRegexp).MatchString(host)) {
			return nil, nil
		}
	} else if u.Info != "" || u.Host == "" {
		return nil, nil
	}

	if scheme != "" {
		host = scheme + "://" + host
	}
	return NewRepositorySpec(host), nil
}
