// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"sync"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/errors"
)

// we already have our new acme.org config object type,
// now we want to provide an object, which configures
// itself when used.

// RepositoryProvider should be an object, which is
// able to provide an OCI repository reference.
// It has a setter and a getter (the setter is
// provided by our ad-hoc SimpleRepositoryTarget).
type RepositoryProvider struct {
	lock sync.Mutex
	// updater is a utility, which ia able to
	// configure an object basesd a a managed configuration
	// watermark. It remembers which config objects from the
	// config queue are already applies, and replays
	// the config objects applied to the config context
	// after the last update.
	updater cpi.Updater
	SimpleRepositoryTarget
}

func NewRepositoryProvider(ctx cpi.ContextProvider) *RepositoryProvider {
	p := &RepositoryProvider{}
	// To do its work, the updater needs a connection to
	// the config context to use and the object, which should be
	// configured.
	p.updater = cpi.NewUpdater(ctx.ConfigContext(), p)
	return p
}

// GetRepository returns a repository ref.
func (p *RepositoryProvider) GetRepository() (string, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// the first step for methods of configurable objects
	// dependent on potential configuration is always
	// to update itself using the embedded updater.
	// Please remember, the config management reverses the
	// request direction. Applying a config object to
	// the config context does not configure dependent objects,
	// it just manages a config queue, which is used by potential
	// configuration targets to configure themselves.
	// The reason for this is to avoid references from the
	// management to managed objects. This would prohibit
	// the garbage collection of all configurable objects.
	err := p.updater.Update()
	if err != nil {
		return "", err
	}
	// now, we can do our regular function, aka
	// providing a repository ref.
	return p.repository, nil
}

func WriteConfigTargets(cfg *helper.Config) error {
	credctx := credentials.DefaultContext()

	// after defining or repository provider type
	// we can now use it.
	prov := NewRepositoryProvider(credctx)

	// If we ask now for a repository we will get the empty
	// answer.
	repo, err := prov.GetRepository()
	if err != nil {
		errors.Wrapf(err, "get repo")
	}
	if repo != "" {
		return fmt.Errorf("Oops, found repository %q", repo)
	}

	// Now, we apply our config from the last example.
	ctx := credctx.ConfigContext()
	examplecfg := NewConfig(cfg)
	err = ctx.ApplyConfig(examplecfg, "special acme config")
	if err != nil {
		errors.Wrapf(err, "apply config")
	}

	// asking for a repository now will return the configured
	// ref.
	repo, err = prov.GetRepository()
	if err != nil {
		errors.Wrapf(err, "get repo")
	}
	if repo == "" {
		return fmt.Errorf("no repository provided")
	}
	fmt.Printf("using repository: %s\n", repo)

	// now, we should also be prepared to get the credentials,
	// our config object configures the provider as well as
	// the credential context.
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

	return nil
}
