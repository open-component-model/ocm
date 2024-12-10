package ocireg

import (
	"context"
	"path"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	regconfig "github.com/regclient/regclient/config"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/tech/regclient"
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
	logger := logging.DynamicLogger(ctx, REALM, logging.NewAttribute(ocmlog.ATTR_HOST, urs.Host))
	if urs.Scheme == "http" {
		logger.Warn("using insecure http for oci registry {{host}}", "host", urs.Host)
	}
	i := &RepositoryImpl{
		RepositoryImplBase: cpi.NewRepositoryImplBase(ctx),
		logger:             logger,
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

func (r *RepositoryImpl) getResolver(comp string) (regclient.Resolver, error) {
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

	var (
		password string
		username string
	)

	if creds != nil {
		password = creds.GetProperty(credentials.ATTR_IDENTITY_TOKEN)
		if password == "" {
			password = creds.GetProperty(credentials.ATTR_PASSWORD)
		}
		username = creds.GetProperty(credentials.ATTR_USERNAME)
	}

	opts := regclient.ClientOptions{
		Host: &regconfig.Host{
			Name: "ghcr.io", //TODO: Need to figure out how to set the host.
			User: username,
			Pass: password,
		},
		Version: comp,
	}
	//opts := docker.ResolverOptions{
	//	Hosts: docker.ConvertHosts(config.ConfigureHosts(context.Background(), config.HostOptions{
	//		UpdateClient: func(client *http.Client) error {
	//			// copy from http.DefaultTransport with a roundtripper injection
	//			client.Transport = ocmlog.NewRoundTripper(client.Transport, logger)
	//			return nil
	//		},
	//		Credentials: func(host string) (string, string, error) {
	//			if creds != nil {
	//				p := creds.GetProperty(credentials.ATTR_IDENTITY_TOKEN)
	//				if p == "" {
	//					p = creds.GetProperty(credentials.ATTR_PASSWORD)
	//				}
	//				pw := ""
	//				if p != "" {
	//					pw = "***"
	//				}
	//				logger.Trace("query credentials", ocmlog.ATTR_USER, creds.GetProperty(credentials.ATTR_USERNAME), "pass", pw)
	//				return creds.GetProperty(credentials.ATTR_USERNAME), p, nil
	//			}
	//			logger.Trace("no credentials")
	//			return "", "", nil
	//		},
	//		DefaultScheme: r.info.Scheme,
	//		//nolint:gosec // used like the default, there are OCI servers (quay.io) not working with min version.
	//		DefaultTLS: func() *tls.Config {
	//			if r.info.Scheme == "http" {
	//				return nil
	//			}
	//			return &tls.Config{
	//				// MinVersion: tls.VersionTLS13,
	//				RootCAs: func() *x509.CertPool {
	//					var rootCAs *x509.CertPool
	//					if creds != nil {
	//						c := creds.GetProperty(credentials.ATTR_CERTIFICATE_AUTHORITY)
	//						if c != "" {
	//							rootCAs = x509.NewCertPool()
	//							rootCAs.AppendCertsFromPEM([]byte(c))
	//						}
	//					}
	//					if rootCAs == nil {
	//						rootCAs = rootcertsattr.Get(r.GetContext()).GetRootCertPool(true)
	//					}
	//					return rootCAs
	//				}(),
	//			}
	//		}(),
	//	})),
	//}

	return regclient.New(opts), nil
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
	res, err := r.getResolver(name)
	if err != nil {
		return false, err
	}
	ref := r.GetRef(name, version)
	_, _, err = res.Resolve(context.Background(), ref)
	if err != nil {
		if errdefs.IsNotFound(err) {
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
