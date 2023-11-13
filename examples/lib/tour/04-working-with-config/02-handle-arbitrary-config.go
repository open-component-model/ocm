// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	configcfg "github.com/open-component-model/ocm/pkg/contexts/config/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	credcfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/directcreds"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/errors"
)

func credConfig(cfg *helper.Config) (config.Config, error) {
	creds := credcfg.New()

	// here we can configure credential settings:
	// credential repositories and consumer is mappings.
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid consumer")
	}
	creds.AddConsumer(
		id,
		directcreds.NewRepositorySpec(cfg.GetCredentials().Properties()),
	)
	return creds, nil
}

func HandleArbitraryConfiguration(cfg *helper.Config) error {
	// The configuration management provides a configuration object
	// for it own.

	generic := configcfg.New()

	// the generic config holds a list of config objects,
	// or their specification formats.
	// Additionally, it is possible to configure names sets
	// of configurations, which can later be enabled
	// on-demand at the config context.

	// we recycle our credential config from the last example.
	creds, err := credConfig(cfg)
	if err != nil {
		return err
	}
	err = generic.AddConfig(creds)
	if err != nil {
		return errors.Wrapf(err, "adding config")
	}

	// credential objects are typically serializable and deserializable.
	// this also holds for the generic config object of the config context.

	spec, err := json.MarshalIndent(generic, "  ", "  ")
	if err != nil {
		return errors.Wrapf(err, "marshal credential config")
	}

	// the result is a config object hosting a list (with 1 entry)
	// of other config object specifications.
	fmt.Printf("this a a generic configuration object:\n%s\n", string(spec))

	// the generic config object can be added to a config context, again.
	ctx := config.DefaultContext()
	err = ctx.ApplyConfig(creds, "generic setting")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config")
	}
	credctx := credentials.DefaultContext()

	// query now works, also.
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	found, err := credentials.CredentialsForConsumer(credctx, id)
	if err != nil {
		return errors.Wrapf(err, "cannot get credentials")
	}
	fmt.Printf("consumer id: %s\n", id)
	fmt.Printf("credentials: %s\n", obfuscate(found))
	return nil
}
