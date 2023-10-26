// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/errors"
)

func TransportingComponentVersions(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	// the context acts as factory for various model types based on
	// specification descriptor serialization formats in YAML or JSON.
	// Access method specifications and repository specification are
	// examples for this feature.
	//
	// Now, we use the repository specification serialization format to
	// determine the target repository for a transport from a yaml
	// configuration file.
	fmt.Printf("target repository is %s\n", string(cfg.Target))
	target, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer target.Close()

	// we just use the component version provided by the last examples
	// in a remote target repository.
	// Therefore, we set up the credentials context, again, as has
	// been shown in example 3.
	id, err := oci.GetConsumerIdForRef(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid consumer")
	}
	creds := ociidentity.SimpleCredentials(cfg.Username, cfg.Password)
	ctx.CredentialsContext().SetCredentialsForConsumer(id, creds)

	// now, we are ready to determine the transportation source.

	// open the source repository.
	spec := ocireg.NewRepositorySpec(cfg.Repository, nil)
	repo, err := ctx.RepositoryForSpec(spec, creds)
	if err != nil {
		return err
	}
	defer repo.Close()

	cv, err := repo.LookupComponentVersion("acme.org/example03", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "added version not found")
	}
	defer cv.Close()

	fmt.Printf("*** source version in source repository\n")
	err = describeVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}

	// transfer the component version with value mode.
	// Here, all resources are transported per value, all external
	// references will be inlined as localBlobs and imported into
	// the target environment, applying blob upload handlers
	// where possible. For a CTF Archive as target, there are no
	// configured handlers, by default.
	err = transfer.Transfer(cv, target, standard.ResourcesByValue(), standard.Overwrite())
	if err != nil {
		return errors.Wrapf(err, "transport failed")
	}

	tcv, err := target.LookupComponentVersion("acme.org/example03", "v0.1.0")
	if err != nil {
		return errors.Wrapf(err, "transported version not found")
	}
	defer tcv.Close()

	// please be aware that the all resources in the target now are localBlobs,
	// if the target is a CTF archive. If it is an OCI registry, all the OCI
	// artifact resources will be uploaded as OCI artifacts into the target
	// repository and the access specifications are adapted to type `ociArtifact`.
	fmt.Printf("*** target version in transportation target\n")
	err = describeVersion(tcv)
	if err != nil {
		return errors.Wrapf(err, "describe failed")
	}
	return nil
}
