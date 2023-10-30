// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	ocmutils "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

func TransportTo(target ocm.Repository, src string) error {
	ctx := target.GetContext()

	// get the access to the source repository
	fmt.Printf("source OCI repository is %s\n", string(src))
	spec := ocireg.NewRepositorySpec(src, nil)
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return errors.Wrapf(err, "cannot get repository access for %s", src)
	}
	defer repo.Close()
	PrintConsumerId(repo, "source repository")

	// lookup component version to be transported
	cv, err := repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION)
	if err != nil {
		return errors.Wrapf(err, "cannot get component version from %s", src)
	}
	defer cv.Close()

	err = transfer.Transfer(cv, target,
		standard.ResourcesByValue(),
		standard.Overwrite(),
		transfer.WithPrinter(common.StdoutPrinter))
	if err != nil {
		return errors.Wrapf(err, "transfer failed")
	}
	return nil
}

func Consumer(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()
	err := ReadConfiguration(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot read ocm configuration")
	}

	// Open the local OCM repository

	// use the generic form here to enable the specification of any
	// supported repository type as target.
	fmt.Printf("local repository is %s\n", string(cfg.Target))
	repo, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open local repository")
	}
	defer repo.Close()
	PrintConsumerId(repo, "local repository")

	////////////////////////////////////////////////////////////////////////////
	fmt.Printf("*** transfer compoment version\n")

	// first, get the version into the local environment
	err = TransportTo(repo, cfg.Repository)

	// lookup component in local repo
	cv, err := repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION)
	if err != nil {
		return errors.Wrapf(err, "cannot get component version from %s", cfg.Target)
	}
	defer cv.Close()

	PrintSignatures(cv)
	PrintPublicKey(ctx, "acme.org")

	// second, verify signature
	fmt.Printf("*** verify signature\n")
	_, err = signing.VerifyComponentVersion(cv, "acme.org")
	if err != nil {
		return errors.Wrapf(err, "verification failed")
	} else {
		fmt.Printf("  verification successful\n")
	}

	////////////////////////////////////////////////////////////////////////////
	fmt.Printf("*** download helm chart\n")

	res, err := cv.GetResource(metav1.NewIdentity("helmchart"))
	if err != nil {
		return errors.Wrapf(err, "resource for helmchart not found")
	}

	targetfs := memoryfs.New()

	// helm downloader registered by default.
	effPath, err := download.DownloadResource(cv.GetContext(), res, "chart", download.WithFileSystem(targetfs))
	if err != nil {
		return errors.Wrapf(err, "cannot download helm chart")
	}
	// report found files
	files, err := tarutils.ListArchiveContent(effPath, targetfs)
	if err != nil {
		return errors.Wrapf(err, "cannot list files for helm chart")
	}
	fmt.Printf("files for helm chart:\n")
	for _, f := range files {
		fmt.Printf("- %s\n", f)
	}

	////////////////////////////////////////////////////////////////////////////
	fmt.Printf("*** get local image reference\n")

	// get resource and get blob content
	res, err = cv.GetResource(metav1.NewIdentity("podinfo-image"))
	if err != nil {
		return errors.Wrapf(err, "resource for podinfo-image not found")
	}

	acc, err := res.Access()
	if err == nil {
		data, _ := json.Marshal(acc)
		fmt.Printf("access: %s\n", string(data))
	}
	ref, err := ocmutils.GetOCIArtifactRef(ctx, res)
	if err != nil {
		return errors.Wrapf(err, "cannot get OCI reference for resource")
	}
	if ref != "" {
		fmt.Printf("OCI reference: %s\n", ref)
	} else {
		fmt.Printf("no OCI reference found\n")
	}

	////////////////////////////////////////////////////////////////////////////
	fmt.Printf("*** download deploy script\n")

	// get resource and blob content
	res, err = cv.GetResource(metav1.NewIdentity(RSC_DEPLOY))
	if err != nil {
		return errors.Wrapf(err, "resource for podinfo-image not found")
	}

	data, err := ocmutils.GetResourceData(res)
	if err != nil {
		return errors.Wrapf(err, "cannot get deployscript")
	}

	fmt.Printf("deploy script:\n%s\n", string(data))
	return nil
}
