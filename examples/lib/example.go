// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/utils"
)

const componentName = "github.com/mandelsoft/ocmhelmdemo"
const componentVersion = "0.1.0-dev"

const resourceName = "package"

func MyFirstOCMApplication() error {
	octx := ocm.DefaultContext()

	repoSpec := ocireg.NewRepositorySpec("ghcr.io/mandelsoft/ocm", nil)

	repo, err := octx.RepositoryForSpec(repoSpec)
	if err != nil {
		return err
	}
	defer repo.Close()

	compvers, err := repo.LookupComponentVersion(componentName, componentVersion)
	if err != nil {
		return err
	}
	defer compvers.Close()

	cd := compvers.GetDescriptor()
	data, err := compdesc.Encode(cd)
	if err != nil {
		return err
	}

	fmt.Printf("component descriptor:\n%s\n", string(data))

	res, err := compvers.GetResource(metav1.NewIdentity(resourceName))
	if err != nil {
		return err
	}

	fmt.Printf("resource %s:\n  type: %s\n", resourceName, res.Meta().Type)

	meth, err := res.AccessMethod()
	if err != nil {
		return err
	}
	defer meth.Close()

	fmt.Printf("  mime: %s\n", meth.MimeType())

	data, err = meth.Get()
	if err != nil {
		return err
	}

	fmt.Printf("  content:\n%s\n", utils.IndentLines(string(data), "    "))

	return nil
}
