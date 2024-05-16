package npm

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"github.com/open-component-model/ocm/pkg/common"
	npmCredentials "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/npm/identity"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/utils"
)

const PROVIDER = "ocm.software/credentialprovider/" + Type

type Repository struct {
	ctx       cpi.Context
	path      string
	propagate bool
	npmrc     npmConfig
}

func NewRepository(ctx cpi.Context, path string, prop ...bool) (*Repository, error) {
	return newRepository(ctx, path, utils.OptionalDefaultedBool(true, prop...))
}

func newRepository(ctx cpi.Context, path string, prop bool) (*Repository, error) {
	r := &Repository{
		ctx:       ctx,
		path:      path,
		propagate: prop,
	}
	err := r.Read(true)
	return r, err
}

var _ cpi.Repository = &Repository{}

func (r *Repository) ExistsCredentials(name string) (bool, error) {
	err := r.Read(false)
	if err != nil {
		return false, err
	}
	return r.npmrc[name] != "", nil
}

func (r *Repository) LookupCredentials(name string) (cpi.Credentials, error) {
	exists, err := r.ExistsCredentials(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.ErrNotFound("credentials", name, Type)
	}
	return newCredentials(r.npmrc[name]), nil
}

func (r *Repository) WriteCredentials(_ string, _ cpi.Credentials) (cpi.Credentials, error) {
	return nil, errors.ErrNotSupported("write", "credentials", Type)
}

func (r *Repository) Read(force bool) error {
	if !force && r.npmrc != nil {
		return nil
	}

	if r.path == "" {
		return fmt.Errorf("npmrc path not provided")
	}
	cfg, path, err := readNpmConfigFile(r.path)
	if err != nil {
		return fmt.Errorf("failed to load npmrc: %w", err)
	}
	id := cpi.ProviderIdentity(PROVIDER + "/" + path)

	if r.propagate {
		r.ctx.RegisterConsumerProvider(id, &ConsumerProvider{r.path})
	}
	r.npmrc = cfg
	return nil
}

func newCredentials(token string) cpi.Credentials {
	props := common.Properties{
		npmCredentials.ATTR_TOKEN: token,
	}
	return cpi.NewCredentials(props)
}
