// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocireg

import (
	"context"
	"path"
	"strings"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes/docker/config"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/docker"
	"github.com/open-component-model/ocm/pkg/docker/resolve"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/utils"
)

type RepositoryInfo struct {
	Scheme  string
	Locator string
	Creds   credentials.Credentials
	Legacy  bool
}

func (r *RepositoryInfo) HostPort() string {
	i := strings.Index(r.Locator, "/")
	if i < 0 {
		return r.Locator
	} else {
		return r.Locator[:i]
	}
}

func (r *RepositoryInfo) HostInfo() (string, string, string) {
	return utils.SplitLocator(r.Locator)
}

type Repository struct {
	ctx  cpi.Context
	spec *RepositorySpec
	info *RepositoryInfo
}

var _ cpi.Repository = &Repository{}

func NewRepository(ctx cpi.Context, spec *RepositorySpec, info *RepositoryInfo) (*Repository, error) {
	return &Repository{
		ctx:  ctx,
		spec: spec,
		info: info,
	}, nil
}

func (r *Repository) NamespaceLister() cpi.NamespaceLister {
	return nil
}

func (r *Repository) IsReadOnly() bool {
	return false
}

func (r *Repository) IsClosed() bool {
	return false
}

func (r *Repository) getCreds(comp string) (credentials.Credentials, error) {
	if r.info.Creds != nil {
		return r.info.Creds, nil
	}
	return GetCredentials(r.ctx, r.info.Locator, comp)
}

func (r *Repository) getResolver(comp string) (resolve.Resolver, error) {
	creds, err := r.getCreds(comp)
	if err != nil {
		if !errors.IsErrUnknownKind(err, credentials.KIND_CONSUMER) {
			return nil, err
		}
	}
	if creds == nil {
		logging.Logger(REALM).Trace("no credentials", "comp", comp, "base", r.spec.BaseURL)
	}

	opts := docker.ResolverOptions{
		Hosts: docker.ConvertHosts(config.ConfigureHosts(context.Background(), config.HostOptions{
			Credentials: func(host string) (string, string, error) {
				if creds != nil {
					p := creds.GetProperty(credentials.ATTR_IDENTITY_TOKEN)
					if p == "" {
						p = creds.GetProperty(credentials.ATTR_PASSWORD)
					}
					pw := ""
					if pw != "" {
						pw = "***"
					}
					logging.Logger(REALM).Trace("query credentials", "host", host, "user", creds.GetProperty(credentials.ATTR_USERNAME), "pass", pw)
					return creds.GetProperty(credentials.ATTR_USERNAME), p, nil
				}
				logging.Logger(REALM).Trace("no credentials", "host", host)
				return "", "", nil
			},
			DefaultScheme: r.info.Scheme,
		})),
	}

	return docker.NewResolver(opts), nil
}

func (r *Repository) getRef(comp, vers string) string {
	base := path.Join(r.info.Locator, comp)
	if vers == "" {
		return base
	}
	if ok, d := artdesc.IsDigest(vers); ok {
		return base + "@" + d.String()
	}
	return base + ":" + vers
}

func (r *Repository) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func (r *Repository) GetBaseURL() string {
	return r.spec.BaseURL
}

func (r *Repository) ExistsArtifact(name string, version string) (bool, error) {
	res, err := r.getResolver(name)
	if err != nil {
		return false, err
	}
	ref := r.getRef(name, version)
	_, _, err = res.Resolve(context.Background(), ref)

	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) LookupArtifact(name string, version string) (cpi.ArtifactAccess, error) {
	n, err := r.LookupNamespace(name)
	if err != nil {
		return nil, err
	}
	return n.GetArtifact(version)
}

func (r *Repository) LookupNamespace(name string) (cpi.NamespaceAccess, error) {
	return NewNamespace(r, name)
}

func (r *Repository) Close() error {
	return nil
}
