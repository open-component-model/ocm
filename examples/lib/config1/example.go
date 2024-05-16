package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ociid "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	ccfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/directcreds"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
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
