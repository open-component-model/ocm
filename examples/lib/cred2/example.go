// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
)

func SimpleWriteWithCredentials() error {
	cfg, err := helper.ReadConfig(CFG)
	if err != nil {
		return err
	}

	octx := ocm.DefaultContext()

	octx.CredentialsContext().SetCredentialsForConsumer(
		credentials.ConsumerIdentity{
			identity.ID_TYPE:       identity.CONSUMER_TYPE,
			identity.ID_HOSTNAME:   "ghcr.io",
			identity.ID_PATHPREFIX: "mandelsoft",
		},
		cfg.GetCredentials(),
	)
	repoSpec := ocireg.NewRepositorySpec(cfg.Repository, nil)

	repo, err := octx.RepositoryForSpec(repoSpec)
	if err != nil {
		return err
	}
	defer repo.Close()

	comp, err := repo.LookupComponent(cfg.Component)
	if err != nil {
		return errors.Wrapf(err, "cannot lookup component %s", cfg.Component)
	}
	defer comp.Close()

	compvers, err := comp.NewVersion(cfg.Version, true)
	if err != nil {
		return errors.Wrapf(err, "cannot create new version %s", cfg.Version)
	}
	defer compvers.Close()

	// add provider information
	compvers.GetDescriptor().Provider = metav1.Provider{Name: "mandelsoft"}

	// add a new resource artifact with the local identity `name="test"`.
	err = compvers.SetResourceBlob(
		&compdesc.ResourceMeta{
			ElementMeta: compdesc.ElementMeta{
				Name: "test",
			},
			Type:     resourcetypes.BLOB,
			Relation: metav1.LocalRelation,
		},
		accessio.BlobAccessForString(mime.MIME_TEXT, "testdata"),
		"", nil,
	)
	if err != nil {
		return errors.Wrapf(err, "cannot add resource")
	}

	// finally push the new component version
	if err = comp.AddVersion(compvers); err != nil {
		return errors.Wrapf(err, "cannot add new version")
	}
	fmt.Printf("added component %s version %s\n", cfg.Component, cfg.Version)
	return nil
}
