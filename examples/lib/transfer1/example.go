// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	// "github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/config/configutils"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
)

const componentName = "github.com/mandelsoft/test1"
const componentVersion = "0.1.0"

const resourceName = "package"

func setupComponents(repo ocm.Repository) (rerr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	c, err := repo.LookupComponent(componentName)
	if rerr != nil {
		return err
	}
	finalize.Close(c)

	cv, err := c.LookupVersion(componentVersion)
	if err != nil {
		cv, err = c.NewVersion(componentVersion)
		if rerr != nil {
			return err
		}
	}
	finalize.Close(cv)

	cv.GetDescriptor().Provider.Name = "acne.org"
	err = cv.SetResourceBlob(compdesc.NewResourceMeta(resourceName, resourcetypes.PLAIN_TEXT, metav1.LocalRelation),
		accessio.BlobAccessForString(mime.MIME_TEXT, "test data"),
		"", nil)
	if err != nil {
		return errors.Wrapf(err, "cannot add resource test")
	}
	err = c.AddVersion(cv)
	if err != nil {
		return errors.Wrapf(err, "cannot add version")
	}
	return nil
}

func TransferApplication() (rerr error) {

	cfg, err := helper.ReadConfig(CFG)
	if err != nil {
		return err
	}
	// configure default context by evaluating standard config sources
	err = configutils.Configure("")
	if rerr != nil {
		return err
	}

	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	octx := ocm.DefaultContext()

	// create a temporary orchestration environment for a set of
	// component versions. We use a CTF here stored either
	// in a temporary filesystem folder or in memory.
	tmpfs, err := osfs.NewTempFileSystem()
	if err != nil {
		return err
	}
	finalize.With(func() error { return vfs.Cleanup(tmpfs) })

	// if you have not much direct blob content, you could use
	// a memory filesystem instead
	// tmpfs:=memoryfs.New()

	repo, err := ctf.Open(octx, accessobj.ACC_CREATE, "ctf", 0o700, accessio.PathFileSystem(tmpfs))
	if err != nil {
		return errors.Wrapf(err, "cannot create CTF")
	}
	finalize.Close(repo)

	// now setup the components you want to publish
	err = setupComponents(repo)
	if rerr != nil {
		return err
	}

	// prepare transfer to target
	uni, err := ocm.ParseRepo(cfg.Repository)
	if err != nil {
		return errors.Wrapf(err, "invalid repo spec")
	}
	repoSpec, err := octx.MapUniformRepositorySpec(&uni)
	if err != nil {
		return errors.Wrapf(err, "invalid repo spec")
	}

	// if you know you have an OCI registry based OCM repository
	// repoSpec := ocireg.NewRepositorySpec(cfg.Repository)

	// if you want to provide specific credentials....
	// target, err := octx.RepositoryForSpec(repoSpec, cfg.GetCredentials())

	// use credentials from config context (for example initialized by Configure above)
	target, err := octx.RepositoryForSpec(repoSpec)
	if err != nil {
		return err
	}
	finalize.Close(target)

	// if you don't want to create a CTF first you could call
	// setupComponents directly on the target repository
	// instead of transferring the CTF content separately.

	// scan the CTF and transfer all found component versions

	// only available for selected repo types like CTF
	lister := repo.ComponentLister()
	if lister == nil {
		return fmt.Errorf("repo does not support lister")
	}
	comps, err := lister.GetComponents("", true)
	if rerr != nil {
		return errors.Wrapf(err, "cannot list components")
	}

	printer := common.NewPrinter(os.Stdout)
	closure := transfer.TransportClosure{}
	transferHandler, err := standard.New(standard.Overwrite())
	if rerr != nil {
		return err
	}
	for _, cname := range comps {
		loop := finalize.Nested()

		c, err := repo.LookupComponent(cname)
		if err != nil {
			return errors.Wrapf(err, "cannot get component %s", cname)
		}
		loop.Close(c)

		vnames, err := c.ListVersions()
		if err != nil {
			return errors.Wrapf(err, " cannot list versions for component %s", cname)
		}

		for _, vname := range vnames {
			loop := finalize.Nested()

			cv, err := c.LookupVersion(vname)
			if err != nil {
				return errors.Wrapf(err, "cannot get version %s for component %s", vname, cname)
			}
			loop.Close(cv)
			err = transfer.TransferVersion(printer, closure, cv, target, transferHandler)
			if err := loop.Finalize(); err != nil {
				return err
			}
		}
		if err := loop.Finalize(); err != nil {
			return err
		}
	}

	return nil
}
