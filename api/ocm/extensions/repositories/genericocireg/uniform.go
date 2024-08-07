package genericocireg

import (
	"fmt"
	"strings"

	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/oci/grammar"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/compatattr"
)

func init() {
	cpi.RegisterRepositorySpecHandler(&repospechandler{}, "*")
	cpi.RegisterRefParseHandler(Type, HandleRef)
	cpi.RegisterRefParseHandler(ocireg.ShortType, HandleRef)
}

type repospechandler struct{}

func (h *repospechandler) MapReference(ctx cpi.Context, u *cpi.UniformRepositorySpec) (cpi.RepositorySpec, error) {
	var meta *ComponentRepositoryMeta
	host := u.Host
	subp := u.SubPath

	// This is checked because it can lead to confusion with the ocm notation.
	if strings.Contains(subp, "//") {
		return nil, fmt.Errorf("subpath %q cannot contain double slash (//)", subp)
	}
	if u.Type == Type {
		if u.Info != "" && u.SubPath == "" {
			idx := strings.Index(u.Info, grammar.RepositorySeparator)
			if idx > 0 {
				host = u.Info[:idx]
				subp = u.Info[idx+1:]
			} else {
				host = u.Info
			}
		} else if u.Host == "" {
			return nil, fmt.Errorf("host required for OCI based OCM reference")
		}
	} else {
		if u.Type != "" || u.Info != "" || u.Host == "" {
			return nil, nil
		}
		host = u.Host
	}
	if u.Scheme != "" {
		host = u.Scheme + "://" + host
	}
	if subp != "" {
		meta = NewComponentRepositoryMeta(subp, "")
	}
	if compatattr.Get(ctx) {
		return NewRepositorySpec(ocireg.NewLegacyRepositorySpec(host), meta), nil
	}
	return NewRepositorySpec(ocireg.NewRepositorySpec(host), meta), nil
}

func HandleRef(u *cpi.UniformRepositorySpec) error {
	if u.Host == "" && u.Info != "" && u.SubPath == "" {
		info := u.Info
		scheme := ""
		match := grammar.AnchoredSchemedRegexp.FindStringSubmatch(info)
		if match != nil {
			scheme = match[1]
			info = match[2]
		}
		host := ""
		subp := ""
		idx := strings.Index(info, grammar.RepositorySeparator)
		if idx > 0 {
			host = info[:idx]
			subp = info[idx+1:]
		} else {
			host = info
		}
		if grammar.HostPortRegexp.MatchString(host) || grammar.DomainPortRegexp.MatchString(host) {
			u.Scheme = scheme
			u.Host = host
			u.SubPath = subp
			u.Info = ""
		}
	}
	return nil
}
