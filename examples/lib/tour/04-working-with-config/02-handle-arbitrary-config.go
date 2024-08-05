package main

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/config"
	configcfg "ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/credentials"
	credcfg "ocm.software/ocm/api/credentials/config"
	"ocm.software/ocm/api/credentials/extensions/repositories/directcreds"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/examples/lib/helper"
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

	// --- begin config config ---
	generic := configcfg.New()
	// --- end config config ---

	// the generic config holds a list of any other config objects,
	// or their specification formats.
	// Additionally, it is possible to configure named sets
	// of configurations, which can later be enabled
	// on-demand by their name at the config context.

	// we recycle our credential config from the last example.
	// --- begin sub config ---
	creds, err := credConfig(cfg)
	if err != nil {
		return err
	}
	// --- end sub config ---

	// now, we can add this credential config object to
	// our generic config list.
	// --- begin add config ---
	err = generic.AddConfig(creds)
	if err != nil {
		return errors.Wrapf(err, "adding config")
	}
	// --- end add config ---

	// as we have seen in the previous example, config objects are typically
	// serializable and deserializable.
	// this also holds for the generic config object of the config context.

	// --- begin serialized ---
	spec, err := json.MarshalIndent(generic, "  ", "  ")
	if err != nil {
		return errors.Wrapf(err, "marshal credential config")
	}

	fmt.Printf("this a a generic configuration object:\n%s\n", string(spec))
	// --- end serialized ---
	// the result is a config object hosting a list (with 1 entry)
	// of other config object specifications.

	// The generic config object can be added to a config context, again, like
	// any other config object. If it is asked to configure a configuration
	// context it uses the methods of the configuration context to apply the
	// contained list of config objects (and the named set of config lists).
	// Therefore, all config objects applied to a configuration context are
	// asked to configure the configuration context itself when queued to the
	// list of applied configuration objects.

	// If we now ask the default credential context (which uses the default
	// configuration context to configure itself) for credentials for our OCI registry,
	// the credential mapping provided by the config object added to the generic one,
	// will be found.

	// --- begin query ---
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
	// --- end query ---
	return nil
}
