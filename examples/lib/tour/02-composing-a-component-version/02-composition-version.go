package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
)

func ComposingAComponentVersionB() error {
	// yes, we need an OCM context, again
	// --- begin default context ---
	ctx := ocm.DefaultContext()
	// --- end default context ---

	// now we compose a component version without a repository.
	// later, we add this to a new repository.
	// --- begin new version ---
	cv := composition.NewComponentVersion(ctx, "acme.org/example2", "v0.1.0")
	// --- end new version ---

	// just use the same component version setup from variant A
	// --- begin setup version ---
	err := setupVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "version composition")
	}
	// --- end setup version ---

	// even on this internal component version, the API is the same
	fmt.Printf("*** composition version ***\n")
	err = describeVersion(cv)

	// Now we can add this version to any OCM repository.
	//
	// Here, we are using an internal composition repository.
	// It has no storage backend and can be used to internally compose
	// a set of component versions, which can then be transferred
	// to any other repository (see tour 5)

	// --- begin create composition repository ---
	repo := composition.NewRepository(ctx)
	// --- end create composition repository ---
	// --- begin add version ---
	err = repo.AddComponentVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "cannot add version")
	}
	// --- end add version ---

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
