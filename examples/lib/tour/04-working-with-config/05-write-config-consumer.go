package main

import (
	"fmt"
	"sync"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci"
	ociidentity "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/examples/lib/helper"
)

// we already have our new acme.org config object type,
// now we want to provide an object, which configures
// itself when used.

// RepositoryProvider should be an object, which is
// able to provide an OCI repository reference.
// It has a setter and a getter (the setter is
// provided by our ad-hoc SimpleRepositoryTarget).
// --- begin type ---
type RepositoryProvider struct {
	lock sync.Mutex
	// cpi.Updater is a utility, which is able to
	// configure an object based on a managed configuration
	// watermark. It remembers which config objects from the
	// config queue are already applied, and replays
	// the config objects applied to the config context
	// after the last update.
	updater cpi.Updater
	SimpleRepositoryTarget
}

// --- end type ---

// --- begin constructor ---
func NewRepositoryProvider(ctx cpi.ContextProvider) *RepositoryProvider {
	p := &RepositoryProvider{}
	// To do its work, the updater needs a connection to
	// the config context to use and the object, which should be
	// configured.
	p.updater = cpi.NewUpdater(ctx.ConfigContext(), p)
	return p
}

// --- end constructor ---

// the magic now happens in the methods provided
// by our configurable object.
// the first step for methods of configurable objects
// dependent on potential configuration is always
// to update itself using the embedded updater.
//
// Please note, the config management reverses the
// request direction. Applying a config object to
// the config context does not configure dependent objects,
// it just manages a config queue, which is used by potential
// configuration targets to configure themselves.
// The actual configuration action is always initiated
// by the object, which wants to be configured.
// The reason for this is to avoid references from the
// management to managed objects. This would prohibit
// the garbage collection of all configurable objects.

// GetRepository returns a repository ref.
// --- begin method ---
func (p *RepositoryProvider) GetRepository() (string, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	err := p.updater.Update()
	if err != nil {
		return "", err
	}
	// now, we can do our regular function, aka
	// providing a repository ref.
	return p.repository, nil
}

// --- end method ---

func WriteConfigTargets(cfg *helper.Config) error {
	// --- begin default context ---
	credctx := credentials.DefaultContext()
	// --- end default context ---

	// after defining or repository provider type
	// we can now use it.
	// --- begin object ---
	prov := NewRepositoryProvider(credctx)
	// --- end object ---

	// If we ask now for a repository we will get the empty
	// answer.
	// --- begin initial query ---
	repo, err := prov.GetRepository()
	if err != nil {
		errors.Wrapf(err, "get repo")
	}
	if repo != "" {
		return fmt.Errorf("Oops, found repository %q", repo)
	}
	// --- end initial query ---

	// Now, we apply our config from the last example.
	// --- begin apply config ---
	ctx := credctx.ConfigContext()
	examplecfg := NewConfig(cfg)
	err = ctx.ApplyConfig(examplecfg, "special acme config")
	if err != nil {
		errors.Wrapf(err, "apply config")
	}
	// --- end apply config ---

	// without any further action, asking for a repository now will return the
	// configured ref.
	// --- begin query ---
	repo, err = prov.GetRepository()
	if err != nil {
		errors.Wrapf(err, "get repo")
	}
	if repo == "" {
		return fmt.Errorf("no repository provided")
	}
	fmt.Printf("using repository: %s\n", repo)
	// --- end query ---

	// now, we should also be prepared to get the credentials,
	// our config object configures the provider as well as
	// the credential context.
	// --- begin credentials ---
	id, err := oci.GetConsumerIdForRef(repo)
	if err != nil {
		return errors.Wrapf(err, "cannot get consumer id")
	}
	fmt.Printf("usage context: %s\n", id)

	creds, err := credentials.CredentialsForConsumer(credctx, id, ociidentity.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "credentials")
	}
	fmt.Printf("credentials: %s\n", obfuscate(creds))
	// --- end credentials ---
	return nil
}
