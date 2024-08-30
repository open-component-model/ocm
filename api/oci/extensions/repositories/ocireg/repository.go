package ocireg

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/attrs/rootcertsattr"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/refmgmt"
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

type RepositoryImpl struct {
	cpi.RepositoryImplBase
	logger logging.UnboundLogger
	spec   *RepositorySpec
	info   *RepositoryInfo
}

var (
	_ cpi.RepositoryImpl                   = (*RepositoryImpl)(nil)
	_ credentials.ConsumerIdentityProvider = &RepositoryImpl{}
)

func NewRepository(ctx cpi.Context, spec *RepositorySpec, info *RepositoryInfo) (cpi.Repository, error) {
	urs := spec.UniformRepositorySpec()
	if urs.Scheme == "http" {
		ocmlog.Logger(REALM).Warn("using insecure http for oci registry {{host}}", "host", urs.Host)
	}
	i := &RepositoryImpl{
		RepositoryImplBase: cpi.NewRepositoryImplBase(ctx),
		logger:             logging.DynamicLogger(ctx, REALM, logging.NewAttribute(ocmlog.ATTR_HOST, urs.Host)),
		spec:               spec,
		info:               info,
	}
	return cpi.NewRepository(i), nil
}

func GetRepositoryImplementation(r cpi.Repository) (*RepositoryImpl, error) {
	i, err := cpi.GetRepositoryImplementation(r)
	if err != nil {
		return nil, err
	}
	return i.(*RepositoryImpl), nil
}

func (r *RepositoryImpl) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func (r *RepositoryImpl) Close() error {
	return nil
}

func (r *RepositoryImpl) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	if c, ok := utils.Optional(uctx...).(credentials.StringUsageContext); ok {
		return identity.GetConsumerId(r.info.Locator, c.String())
	}
	return identity.GetConsumerId(r.info.Locator, "")
}

func (r *RepositoryImpl) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}

func (r *RepositoryImpl) NamespaceLister() cpi.NamespaceLister {
	return nil
}

func (r *RepositoryImpl) IsReadOnly() bool {
	return false
}

func (r *RepositoryImpl) getCreds(comp string) (credentials.Credentials, error) {
	if r.info.Creds != nil {
		return r.info.Creds, nil
	}
	return identity.GetCredentials(r.GetContext(), r.info.Locator, comp)
}

func (r *RepositoryImpl) getResolver(ref string, comp string) (registry.Repository, error) {
	creds, err := r.getCreds(comp)
	if err != nil {
		if !errors.IsErrUnknownKind(err, credentials.KIND_CONSUMER) {
			return nil, err
		}
	}
	logger := r.logger.BoundLogger().WithValues(ocmlog.ATTR_NAMESPACE, comp)
	if creds == nil {
		logger.Trace("no credentials")
	}
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("error creating oci repository: %w", err)
	}

	authCreds := auth.Credential{}
	if creds != nil {
		pass := creds.GetProperty(credentials.ATTR_IDENTITY_TOKEN)
		if pass == "" {
			pass = creds.GetProperty(credentials.ATTR_PASSWORD)
		}
		authCreds.Username = creds.GetProperty(credentials.ATTR_USERNAME)
		authCreds.Password = pass
	}

	client := retry.DefaultClient
	if r.info.Scheme == "https" {
		// set up TLS
		//nolint:gosec // used like the default, there are OCI servers (quay.io) not working with min version.
		conf := &tls.Config{
			// MinVersion: tls.VersionTLS13,
			RootCAs: func() *x509.CertPool {
				var rootCAs *x509.CertPool
				if creds != nil {
					c := creds.GetProperty(credentials.ATTR_CERTIFICATE_AUTHORITY)
					if c != "" {
						rootCAs = x509.NewCertPool()
						rootCAs.AppendCertsFromPEM([]byte(c))
					}
				}
				if rootCAs == nil {
					rootCAs = rootcertsattr.Get(r.GetContext()).GetRootCertPool(true)
				}
				return rootCAs
			}(),
		}
		client.Transport = &http.Transport{
			TLSClientConfig: conf,
		}
	}
	repo.Client = &auth.Client{
		Client:     client,
		Cache:      auth.NewCache(),
		Credential: auth.StaticCredential(r.info.HostPort(), authCreds),
	}

	return repo, nil
}

func (r *RepositoryImpl) GetRef(comp, vers string) string {
	base := path.Join(r.info.Locator, comp)
	if vers == "" {
		return base
	}
	if ok, d := artdesc.IsDigest(vers); ok {
		return base + "@" + d.String()
	}
	return base + ":" + vers
}

func (r *RepositoryImpl) GetBaseURL() string {
	return r.spec.BaseURL
}

func (r *RepositoryImpl) ExistsArtifact(name string, version string) (bool, error) {
	ref := r.GetRef(name, version)
	res, err := r.getResolver(ref, name)
	if err != nil {
		return false, err
	}

	if _, err = res.Resolve(context.Background(), ref); err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *RepositoryImpl) LookupArtifact(name string, version string) (acc cpi.ArtifactAccess, err error) {
	ns, err := NewNamespace(r, name)
	if err != nil {
		return nil, err
	}
	defer refmgmt.PropagateCloseTemporary(&err, ns) // temporary namespace object not exposed.

	return ns.GetArtifact(version)
}

func (r *RepositoryImpl) LookupNamespace(name string) (cpi.NamespaceAccess, error) {
	return NewNamespace(r, name)
}
