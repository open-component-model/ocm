// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	configcfg "github.com/open-component-model/ocm/pkg/contexts/config/config"
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"sigs.k8s.io/yaml"
)

// TYPE is the name of our new configuration object type.
// To be globally unique, it should always end with a
// DNS domain owned by the provider of the new type.
const TYPE = "example.config.acme.org"

// ExampleConfigSpec is a new type of config specification
// covering our example configuration.
type ExampleConfigSpec struct {
	// ObjectVersionedType is the base type providing the type feature
	// form config specifications.
	runtime.ObjectVersionedType `json:",inline"`
	// Config is our example config representation.
	helper.Config `json:",inline"`
}

// NewConfig provides a config object for out helper configuration.
func NewConfig(cfg *helper.Config) cpi.Config {
	return &ExampleConfigSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(TYPE),
		Config:              *cfg,
	}
}

// RepositoryTarget consumes a repository name.
type RepositoryTarget interface {
	SetRepository(r string)
}

// ApplyTo is used to apply the provided configuration settings
// to a dedicated object, which wants to be configured.
func (c *ExampleConfigSpec) ApplyTo(_ cpi.Context, tgt interface{}) error {

	switch t := tgt.(type) {
	// if the target is a credentials context
	// configure the credentials to be used for the
	// described OCI repository.
	case credentials.Context:
		// determine the consumer id for our target repository-
		id, err := oci.GetConsumerIdForRef(c.Repository)
		if err != nil {
			return errors.Wrapf(err, "invalid consumer")
		}
		// create the credentials.
		creds := c.GetCredentials()

		// configure the targeted credential context with
		// the provided credentials (see previous examples).
		t.SetCredentialsForConsumer(id, creds)

	// if the target consumes an OCI repository, propagate
	// the provided OCI repository ref.
	case RepositoryTarget:
		t.SetRepository(c.Repository)

	// all other targets are ignored, we don't have
	// something to set at these objects.
	default:
		return cpi.ErrNoContext(TYPE)
	}
	return nil
}

func init() {
	// register the new config type, so that is can be used
	// by the config management to deserialize appropriately
	// typed specifications.
	cpi.RegisterConfigType(cpi.NewConfigType[*ExampleConfigSpec](TYPE, "this ia config object type based on the example config data."))
}

func WriteConfigType(cfg *helper.Config) error {

	// after preparing aout new special config type
	// we can feed it into the config management.

	credctx := credentials.DefaultContext()

	// the credential context is based on a config context
	// used to configure it.
	ctx := credctx.ConfigContext()

	// create our new config based on the actual settings
	// and apply it to the config context.
	examplecfg := NewConfig(cfg)
	ctx.ApplyConfig(examplecfg, "special acme config")
	// If you omit the above call, no credentials
	// will be found later.
	// _, _ = ctx, examplecfg

	// now we should be prepared to get the credentials
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "cannot get consumer id")
	}
	fmt.Printf("usage context: %s\n", id)

	// the returned credentials are provided via an interface, which might change its
	// content, if the underlying credential source changes.
	creds, err := credentials.CredentialsForConsumer(credctx, id, ociidentity.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "credentials")
	}
	fmt.Printf("credentials: %s\n", obfuscate(creds))

	// Because of the new credential type, such a specification can
	// now be added to the ocm config, also.
	// So, we could use our special tour config file content
	// directly as part of the ocm config.

	ocmcfg := configcfg.New()
	err = ocmcfg.AddConfig(examplecfg)

	spec, err := yaml.Marshal(ocmcfg)
	if err != nil {
		return errors.Wrapf(err, "marshal ocm config")
	}

	// the result is a minimal ocm configuration file
	// just providing our new example configuration.
	fmt.Printf("this a typical ocm config file:\n%s\n", string(spec))

	// above, we added a new kind of target, the RepositoryTarget interface.
	// Just by providing an implementation for this interface, we can
	// configure such an object using the config management.
	target := &SimpleRepositoryTarget{}

	_, err = ctx.ApplyTo(0, target)
	if err != nil {
		return errors.Wrapf(err, "applying to new target")
	}
	fmt.Printf("repository for target: %s\n", target.repository)

	// This way any specialized configuration object can be added
	// by a user of the OCM library. It can be used to configure
	// existing objects or even new object types, even in combination.
	//
	// What is still required is a way
	// to implement new config targets, objects, which want
	// to be configured and which autoconfigure themselves when
	// used. Our simple repository target is just an example
	// for some kind of ad-hoc configuration.
	// This is shown in the next example.
	return nil
}

type SimpleRepositoryTarget struct {
	repository string
}

var _ RepositoryTarget = (*SimpleRepositoryTarget)(nil)

func (t *SimpleRepositoryTarget) SetRepository(repo string) {
	t.repository = repo
}
