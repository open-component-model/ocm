// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/ioutil"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ociid "github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const CFGFILE = "config.yaml"

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

	cid := credentials.ConsumerIdentity{
		ociid.ID_TYPE:       ociid.CONSUMER_TYPE,
		ociid.ID_HOSTNAME:   "ghcr.io",
		ociid.ID_PATHPREFIX: "mandelsoft",
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
