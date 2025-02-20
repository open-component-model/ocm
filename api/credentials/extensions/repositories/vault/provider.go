package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/extensions/repositories/vault/identity"
	"ocm.software/ocm/api/credentials/internal"
	common "ocm.software/ocm/api/utils/misc"
)

const PROVIDER = "ocm.software/credentialprovider/" + Type

const (
	CUSTOM_SECRETS    = "secrets"
	CUSTOM_CONSUMERID = "consumerId"
)

type mapping struct {
	Id   cpi.ConsumerIdentity
	Name string
}

type credentialCache struct {
	creds       cpi.CredentialsSource
	credentials map[string]cpi.DirectCredentials
	consumer    []*mapping
}

func newCredentialCache(creds cpi.CredentialsSource) *credentialCache {
	return &credentialCache{
		creds:       creds,
		credentials: map[string]cpi.DirectCredentials{},
	}
}

type ConsumerProvider struct {
	lock       sync.Mutex
	repository *Repository
	cache      *credentialCache

	updated bool
}

var (
	_ cpi.ConsumerProvider         = (*ConsumerProvider)(nil)
	_ cpi.ConsumerIdentityProvider = (*ConsumerProvider)(nil)
)

func NewConsumerProvider(repo *Repository) (*ConsumerProvider, error) {
	src, err := repo.ctx.GetCredentialsForConsumer(repo.id)
	if err != nil {
		return nil, err
	}
	return &ConsumerProvider{
		cache:      newCredentialCache(src),
		repository: repo,
	}, nil
}

func (p *ConsumerProvider) String() string {
	return p.repository.id.String()
}

func (p *ConsumerProvider) GetConsumerId(uctx ...internal.UsageContext) internal.ConsumerIdentity {
	return p.repository.GetConsumerId()
}

func (p *ConsumerProvider) GetIdentityMatcher() string {
	return p.repository.GetIdentityMatcher()
}

func (p *ConsumerProvider) update(ectx cpi.EvaluationContext) error {
	if p.updated {
		return nil
	}
	credsrc, err := cpi.GetCredentialsForConsumer(p.repository.ctx, ectx, p.repository.id, identity.IdentityMatcher)
	if err != nil {
		return err
	}
	creds, err := credsrc.Credentials(p.repository.ctx)
	if err != nil {
		return err
	}
	err = p.validateCreds(creds)
	if err != nil {
		return err
	}

	ctx := context.Background()

	client, err := vault.New(
		vault.WithAddress(p.repository.spec.ServerURL),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return err
	}

	// vault.WithMountPath("piper/PIPELINE-GROUP-4953/PIPELINE-25042/appRoleCredentials"),
	token, err := p.getToken(ctx, client, creds)
	if err != nil {
		return err
	}

	if err := client.SetToken(token); err != nil {
		return err
	}
	if err := client.SetNamespace(p.repository.spec.Namespace); err != nil {
		return err
	}

	cache := newCredentialCache(credsrc)

	// TODO: support for pure path based access for other secret engine types
	secrets := slices.Clone(p.repository.spec.Secrets)
	if len(secrets) == 0 {
		s, err := client.Secrets.KvV2List(ctx, p.repository.spec.Path,
			vault.WithMountPath(p.repository.spec.MountPath))
		if err != nil {
			p.error(err, "error listing secrets", "")
			return err
		}
		for _, k := range s.Data.Keys {
			if !strings.HasSuffix(k, "/") {
				secrets = append(secrets, k)
			}
		}
	}
	for i := 0; i < len(secrets); i++ {
		n := secrets[i]
		creds, id, list, err := p.read(ctx, client, n)
		p.error(err, "error reading vault secret", n)
		if err == nil {
			for _, a := range list {
				if !slices.Contains(secrets, a) {
					secrets = append(secrets, a)
				}
			}
			if len(id) > 0 {
				cache.consumer = append(cache.consumer, &mapping{
					Id:   cpi.ConsumerIdentity(id),
					Name: n,
				})
			}
			if len(creds) > 0 {
				cache.credentials[n] = cpi.DirectCredentials(creds)
			}
		}
	}
	p.cache = cache
	p.updated = true
	return nil
}

func (p *ConsumerProvider) validateCreds(creds cpi.Credentials) error {
	m := creds.GetProperty(identity.ATTR_AUTHMETH)
	if m == "" {
		return errors.ErrRequired(identity.ATTR_AUTHMETH)
	}
	meth := methods.Get(m)
	if meth == nil {
		return errors.ErrInvalid(identity.ATTR_AUTHMETH, m)
	}
	return meth.Validate(creds)
}

func (p *ConsumerProvider) getToken(ctx context.Context, client *vault.Client, creds cpi.Credentials) (string, error) {
	m := creds.GetProperty(identity.ATTR_AUTHMETH)
	return methods.Get(m).GetToken(ctx, client, p.repository.spec.Namespace, creds)
}

func (p *ConsumerProvider) error(err error, msg string, secret string, keypairs ...interface{}) {
	if err == nil {
		return
	}
	f := log.Info
	var v *vault.ResponseError
	if errors.As(err, &v) && v.StatusCode != http.StatusNotFound {
		f = log.Error
	}
	f(msg, append(keypairs,
		"server", p.repository.spec.ServerURL,
		"namespace", p.repository.spec.Namespace,
		"engine", p.repository.spec.MountPath,
		"path", path.Join(p.repository.spec.Path, secret),
		"error", err.Error(),
	)...,
	)
}

func (p *ConsumerProvider) read(ctx context.Context, client *vault.Client, secret string) (common.Properties, common.Properties, []string, error) {
	// read the secret

	secret = path.Join(p.repository.spec.Path, secret)
	s, err := client.Secrets.KvV2Read(ctx, secret,
		vault.WithMountPath(p.repository.spec.MountPath))
	if err != nil {
		return nil, nil, nil, err
	}

	var id common.Properties
	var list []string
	props := getProps(s.Data.Data)

	if meta, ok := s.Data.Metadata["custom_metadata"].(map[string]interface{}); ok {
		sub := false
		if cid := meta[CUSTOM_CONSUMERID]; cid != nil {
			id = common.Properties{}
			if err := json.Unmarshal([]byte(cid.(string)), &id); err != nil {
				id = nil
			}
			sub = true
		}
		if cid := meta[CUSTOM_SECRETS]; cid != nil {
			if s, ok := meta[CUSTOM_SECRETS].(string); ok {
				for _, e := range strings.Split(s, ",") {
					e = strings.TrimSpace(e)
					if e != "" {
						list = append(list, e)
					}
				}
			}
			sub = true
		}
		if _, ok := meta[cpi.ID_TYPE]; !sub && ok {
			id = getProps(meta)
		}
	}
	return props, id, list, nil
}

func getProps(data map[string]interface{}) common.Properties {
	props := common.Properties{}
	for k, v := range data {
		if s, ok := v.(string); ok {
			props[k] = s
		}
	}
	return props
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ConsumerProvider interface

func (p *ConsumerProvider) Unregister(id cpi.ProviderIdentity) {
}

func (p *ConsumerProvider) Match(ectx cpi.EvaluationContext, req cpi.ConsumerIdentity, cur cpi.ConsumerIdentity, m cpi.IdentityMatcher) (cpi.CredentialsSource, cpi.ConsumerIdentity) {
	return p.get(ectx, req, cur, m)
}

func (p *ConsumerProvider) Get(req cpi.ConsumerIdentity) (cpi.CredentialsSource, bool) {
	creds, _ := p.get(nil, req, nil, cpi.CompleteMatch)
	return creds, creds != nil
}

func (p *ConsumerProvider) get(ectx cpi.EvaluationContext, req cpi.ConsumerIdentity, cur cpi.ConsumerIdentity, m cpi.IdentityMatcher) (cpi.CredentialsSource, cpi.ConsumerIdentity) {
	if req.Equals(p.repository.id) {
		return nil, cur
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	err := p.update(ectx)
	if err != nil {
		log.Info("error accessing credentials provider", "error", err)
	}

	var creds cpi.CredentialsSource

	for _, a := range p.cache.consumer {
		if m(req, cur, a.Id) {
			cur = a.Id
			creds = p.cache.credentials[a.Name]
		}
	}
	return creds, cur
}

////////////////////////////////////////////////////////////////////////////////
// lookup

func (c *ConsumerProvider) ExistsCredentials(name string) (bool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	err := c.update(nil)
	if err != nil {
		return false, err
	}
	_, ok := c.cache.credentials[name]
	return ok, nil
}

func (c *ConsumerProvider) LookupCredentials(name string) (cpi.Credentials, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	err := c.update(nil)
	if err != nil {
		return nil, err
	}
	src, ok := c.cache.credentials[name]
	if ok {
		return src, nil
	}
	return nil, nil
}
