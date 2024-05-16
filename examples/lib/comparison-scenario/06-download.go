package main

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

func Download(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	err := RegisterCredentials(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot register credentials")
	}

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

	fs := memoryfs.New()

	path, err := DownloadHelmChart(cv, "chart", fs)
	if err != nil {
		return errors.Wrapf(err, "download failed")
	}

	// report found files
	files, err := tarutils.ListArchiveContent(path, fs)
	if err != nil {
		return errors.Wrapf(err, "cannot list files for helm chart")
	}
	fmt.Printf("files for helm chart:\n")
	for _, f := range files {
		fmt.Printf("- %s\n", f)
	}
	return nil
}

func DownloadHelmChart(cv ocm.ComponentVersionAccess, path string, fss ...vfs.FileSystem) (string, error) {
	fmt.Printf("*** download helm chart\n")

	res, err := cv.GetResource(metav1.NewIdentity("helmchart"))
	if err != nil {
		return "", errors.Wrapf(err, "resource for helmchart not found")
	}

	targetfs := utils.FileSystem(fss...)

	// helm downloader registered by default.
	effPath, err := download.DownloadResource(cv.GetContext(), res, path, download.WithFileSystem(targetfs))
	if err != nil {
		return "", errors.Wrapf(err, "cannot download helm chart")
	}

	return effPath, nil
}
