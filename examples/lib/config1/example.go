package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	ccfg "ocm.software/ocm/api/credentials/config"
	"ocm.software/ocm/api/credentials/extensions/repositories/directcreds"
	"ocm.software/ocm/api/ocm"
	ociid "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/examples/lib/helper"
)

func UsingConfigs() error {
	cfg, err := helper.ReadConfig(CFG)
	if err != nil {
		return err
	}

	cid := credentials.NewConsumerIdentity(ociid.CONSUMER_TYPE,
		ociid.ID_HOSTNAME, "ghcr.io",
		ociid.ID_PATHPREFIX, "mandelsoft",
	)

	octx := ocm.DefaultContext()
	cctx := octx.ConfigContext()

	// create a credential configuration object
	// and configure it to provide some direct consumer credentials.
	creds := ccfg.New()
	creds.AddConsumer(
		cid,
		directcreds.NewRepositorySpec(cfg.GetCredentials().Properties()),
	)

	err = cctx.ApplyConfig(creds, "explicit")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config")
	}

	credctx := octx.CredentialsContext()

	found, err := credctx.GetCredentialsForConsumer(cid, ociid.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "cannot extract credentials")
	}
	got, err := found.Credentials(credctx)
	if err != nil {
		return errors.Wrapf(err, "cannot evaluate credentials")
	}

	fmt.Printf("found: %s\n", got)
	return nil
}
