// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
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

	repoSpec := ocireg.NewRepositorySpec("ghcr.io/mandelsoft/ocm", nil)

	repo, err := octx.RepositoryForSpec(repoSpec, cfg.GetCredentials())
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
	err= compvers.SetResourceBlob(
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
	if err= comp.AddVersion(compvers); err != nil {
		return errors.Wrapf(err, "cannot add new version")
	}
	fmt.Printf("added component %s version %s\n", cfg.Component, cfg.Version)
	return nil
}
