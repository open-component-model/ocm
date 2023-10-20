// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/extraid"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/semverutils"
)

func GettingStarted() error {
	// get the default context providing
	// all OCM entry point registrations, like
	// access method, repository types, etc.
	// The context bundles all registrations and
	// configuration settings, like credentials,
	// which should be used when working with the OCM
	// ecosystem.
	ctx := ocm.DefaultContext()

	// The context acts as the central entry
	// point to get access to OCM elements.
	// First, get a repository, to look for
	// component versions. We use the OCM
	// repository providing the standard OCM
	// components.

	spec := ocireg.NewRepositorySpec("ghcr.io/open-component-model/ocm")
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return errors.Wrapf(err, "cannot setup repository")
	}

	// many objects must be closed, if they should not be used
	// anymore, to release potentially allocated temporary resources.
	defer repo.Close()

	// Now, we look up the OCM CLI component.
	// All kinds of repositories, regardless of their type
	// feature the same interface to work with OCM content.
	c, err := repo.LookupComponent("ocm.software/ocmcli")
	if err != nil {
		return errors.Wrapf(err, "cannot lookup component")
	}
	defer repo.Close()

	// Now we look for the versions of the component
	// available in this repository.
	versions, err := c.ListVersions()
	if err != nil {
		return errors.Wrapf(err, "cannot query version names")
	}

	// OCM version names must follow the semver rules.
	err = semverutils.SortVersions(versions)
	if err != nil {
		return errors.Wrapf(err, "cannot sort versions")
	}
	fmt.Printf("versions for component ocm.software/ocmcli: %s\n", strings.Join(versions, ", "))

	// Now, we have a look at the latest version
	cv, err := c.LookupVersion(versions[len(versions)-1])
	if err != nil {
		return errors.Wrapf(err, "cannot get latest version")
	}
	defer cv.Close()

	// Have a look at the component descriptor
	cd := cv.GetDescriptor()
	fmt.Printf("resources of the latest version:\n")
	fmt.Printf("version:  %s\n", cv.GetVersion())
	fmt.Printf("provider: %s\n", cd.Provider.Name)

	// and list all the included resources.
	for i, r := range cv.GetResources() {
		fmt.Printf("%2d: name:           %s\n", i+1, r.Meta().GetName())
		fmt.Printf("    extra identity: %s\n", r.Meta().GetExtraIdentity())
		fmt.Printf("    resource type:  %s\n", r.Meta().GetType())
		acc, err := r.Access()
		if err != nil {
			fmt.Printf("    access:         error: %s\n", err)
		} else {
			fmt.Printf("    access:         %s\n", acc.Describe(ctx))
		}
	}

	// Get the executable for the actual environment.
	// The identity of a resource described by a component version
	// consists of a set of properties. The property name must
	// always be given.
	id := metav1.NewIdentity("ocmcli",
		extraid.ExecutableOperatingSystem, runtime.GOOS,
		extraid.ExecutableArchitecture, runtime.GOARCH,
	)

	res, err := cv.GetResource(id)
	if err != nil {
		return errors.Wrapf(err, "resource %s", id)
	}

	// download to /tmp/ocmcli using basic model
	// operations.
	fmt.Printf("downloading OCM cli resource %s...\n", id)
	reader, err := utils.GetResourceReader(res)
	defer reader.Close()

	file, err := os.OpenFile("/tmp/ocmcli", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0766)
	if err != nil {
		return errors.Wrapf(err, "cannot open output file")
	}
	defer file.Close()

	n, err := io.Copy(file, reader)
	if err != nil {
		return errors.Wrapf(err, "write executable")
	}
	fmt.Printf("%d bytes written\n", n)

	// alternatively, a registered downloader for executables can be used.
	// Download is used to download resources with specific handlers for the
	// selected resource and mime type combinations.
	// The executable downloader is registered by default and automatically
	// sets the X flag.
	_, err = download.DownloadResource(ctx, res, "/tmp/ocmcli", download.WithPrinter(common.NewPrinter(os.Stdout)))
	if err != nil {
		return errors.Wrapf(err, "download failed")
	}

	return nil
}
