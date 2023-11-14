// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	configcfg "github.com/open-component-model/ocm/pkg/contexts/config/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	credcfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/dockerconfig"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"sigs.k8s.io/yaml"
)

func HandleOCMConfig(cfg *helper.Config) error {

	// Although the configuration of an OCM context can
	// be done by a sequence of explicit calls the mechanism
	// shown in the example before, is used to provide a simple
	// library function, which can be used to configure an OCM
	// context and all related other contexts with a single call
	// based on a central configuration file (~/.ocmconfig)
	ctx := ocm.DefaultContext()
	_, err := utils.Configure(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "configuration")
	}

	// It is typically such a generic configuration specification,
	// enriched with specialized config specifications for
	// credentials, default repositories signing keys and any
	// other configuration specification.
	// Most important are here the credentials.
	// Because OCM embraces lots of storage technologies
	// for artifact storage as well as storing OCM meta data,
	// tzere are typically multiple technology specific ways
	// to configure credentials for command line tools.
	// Using the credentials settings shown in the previous examples,
	// it ius possible to specify credentials for all
	// required purposes, but the configuration mangement provides
	// an extensible way to embed native technology specific ways
	// to provide credentials just by adding an appropriate type
	// of config objects, which reads the specialized stoarge and
	// feeds it into the credential context.
	//
	// One such config object type is the docker config type. It
	// reads a dockerconfig.json file and fed in the credentials.
	// because it is sed for a dedicated purpose (credentials for
	// OCI registries), it not only can feed the credentials, but
	// also their mapping to consumer ids.

	// create the specification for a new credential repository of
	// type dockerconfig.
	credspec := dockerconfig.NewRepositorySpec("~/.docker/config.json", true)

	// add this repository specification to a credential configuration.
	ccfg := credcfg.New()
	err = ccfg.AddRepository(credspec)
	if err != nil {
		return errors.Wrapf(err, "invalid credential config")
	}

	// By adding the default location for the standard docker config
	// file, all credentials provided by the <code>docker login</code>
	// are available in the OCM toolset, also.

	// A typical minimal <code>.ocmconfig</code> file can be composed as follows.

	ocmcfg := configcfg.New()
	err = ocmcfg.AddConfig(ccfg)

	spec, err := yaml.Marshal(ocmcfg)
	if err != nil {
		return errors.Wrapf(err, "marshal ocm config")
	}

	// the result is a typical minimal ocm configuration file
	// just providing the credentials configured with
	// <code>doicker login</code>.
	fmt.Printf("this a typical ocm config file:\n%s\n", string(spec))

	// Besides from a file, such a config can be provided as data, also,
	// taken from any other source, for example from a Kubernetes secret

	err = utils.ConfigureByData(ctx, spec, "from data")
	if err != nil {
		return errors.Wrapf(err, "configuration")
	}

	// If you have provided your OCI credentials with
	// docker login, they should now be available.

	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	found, err := credentials.CredentialsForConsumer(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "cannot get credentials")
	}
	fmt.Printf("consumer id: %s\n", id)
	fmt.Printf("credentials: %s\n", obfuscate(found))
	return nil
}
