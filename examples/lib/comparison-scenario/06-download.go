// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/errors"
)

func Download(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	err := RegisterCredentials(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot register credentials")
	}

	err = DownloadHelmChart(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "download failed")
	}
	return nil
}

func DownloadHelmChart(ctx ocm.Context, cfg *helper.Config) error {
	fmt.Printf("*** download helm chart\n")

	// use the generic form here to enable the specification of any
	// supported repository type as target.
	fmt.Printf("target repository is %s\n", string(cfg.Target))
	repo, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open repository")
	}
	defer repo.Close()

	// lookup component version to be transported
	cv, err := repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION)
	if err != nil {
		return errors.Wrapf(err, "cannot get component version from %s", cfg.Target)
	}
	defer cv.Close()

	res, err := cv.GetResource(metav1.NewIdentity("helmchart"))
	if err != nil {
		return errors.Wrapf(err, "resource for helmchart not found")
	}

	targetfs := memoryfs.New()

	// helm downloader registered by default.
	path, err := download.DownloadResource(ctx, res, "chart", download.WithFileSystem(targetfs))
	if err != nil {
		return errors.Wrapf(err, "cannot download helm chart")
	}

	// report found files
	files, err := ListFiles(path, targetfs)
	if err != nil {
		return errors.Wrapf(err, "cannot list files for helm chart")
	}
	fmt.Printf("files for helm chart:\n")
	for _, f := range files {
		fmt.Printf("- %s\n", f)
	}
	return nil
}
