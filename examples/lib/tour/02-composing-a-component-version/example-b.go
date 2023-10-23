// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/errors"
)

func ComposingAComponentVersionB() error {
	// yes, we need an OCM context, again
	ctx := ocm.DefaultContext()

	// now we compose a component version without a repository.
	// later we add this to a new repository.

	cv := composition.NewComponentVersion(ctx, "acme.org/example2", "v0.1.0")

	// just use the same scomponent version setup from variant A
	err := setupVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "version composition")
	}

	// even on this internal component version, the API is the same
	fmt.Printf("*** composition version ***\n")
	err = describeVersion(cv)

	// Now we can add this version to any OCM repository.
	//
	// Here, we are using an internal composition repository.
	// It has no storage backend and can be used to internally compose
	// a set of component versions, which can then be transferred
	// to any other repository (see example 4)

	repo := composition.NewRepository(ctx)
	err = repo.AddVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "cannot add version")
	}

	var list []string

	// some repository implementations support a component lister,
	// which can be used to list the components with a dedicated prefix.
	// The composition implementation supports this.
	lister := repo.ComponentLister()
	if lister != nil {
		list, err = lister.GetComponents("", true)
		if err != nil {
			return errors.Wrapf(err, "cannot list components")
		}
	} else {
		fmt.Printf("repository does not support listing components\n")
		list = []string{"acme.org/example02"}
	}

	// now we just describe the versions similar to example 1.
	fmt.Printf("*** repository content ***\n")
	return listVersions(repo, list...)
}
