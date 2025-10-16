package main

import (
	"fmt"
	"io/ioutil"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm"
	ociid "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils/runtime"
)

const CFGFILE = "examples/lib/config2/config.yaml"

func UsingConfigs() error {
	data, err := ioutil.ReadFile(CFGFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot read configuration file %s", CFGFILE)
	}

	octx := ocm.DefaultContext()
	cctx := octx.ConfigContext()

	_, err = cctx.ApplyData(data, runtime.DefaultYAMLEncoding, CFGFILE)
	if err != nil {
		return errors.Wrapf(err, "cannot apply config data")
	}

	cid := credentials.NewConsumerIdentity(ociid.CONSUMER_TYPE,
		ociid.ID_HOSTNAME, "ghcr.io",
		ociid.ID_PATHPREFIX, "mandelsoft",
	)

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
