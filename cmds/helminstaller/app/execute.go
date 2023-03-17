// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/cmds/helminstaller/app/driver"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/runtime"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

func Merge(values ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	for _, val := range values {
		for k, v := range val {
			result[k] = v
		}
	}
	return result
}

func Execute(d driver.Driver, action string, ctx ocm.Context, octx out.Context, cv ocm.ComponentVersionAccess, cfg *Config, values map[string]interface{}, kubeconfig []byte) error {
	if action != "install" && action != "uninstall" {
		return errors.ErrNotSupported("action", action)
	}
	cfgv, err := cfg.GetValues()
	if err != nil {
		return err
	}
	values = Merge(cfgv, values)

	out.Outf(octx, "Loading helm chart from resource %s@%s\n", cfg.Chart, common.VersionedElementKey(cv))
	acc, rcv, err := utils.ResolveResourceReference(cv, cfg.Chart, nil)
	if err != nil {
		return errors.ErrNotFoundWrap(err, "chart reference", cfg.Chart.String())
	}
	defer rcv.Close()

	if acc.Meta().Type != resourcetypes.HELM_CHART {
		return errors.Newf("resource type %q required, but found %q", resourcetypes.HELM_CHART, acc.Meta().Type)
	}

	// have to use the OS filesystem here for using the helm library
	file, err := os.CreateTemp("", "helm-*")
	if err != nil {
		return err
	}

	path := file.Name()
	file.Close()
	os.Remove(path)

	fs := osfs.New()
	_, path, err = download.For(ctx).Download(common.NewPrinter(octx.StdOut()), acc, path, fs)
	if err != nil {
		return errors.Wrapf(err, "downloading helm chart")
	}
	defer os.Remove(path)

	if len(cfg.SubCharts) > 0 {
		out.Outf(octx, "  Unpacking chart archive...\n")
		dir := path + ".dir"
		err := os.Mkdir(dir, 0o700)
		if err != nil {
			return errors.Wrapf(err, "cannot mkdir %q", dir)
		}
		defer os.RemoveAll(dir)

		r, err := fs.Open(path)
		if err != nil {
			return errors.Wrapf(err, "cannot read downloaded chart archive %q", path)
		}
		defer r.Close()
		reader, _, err := compression.AutoDecompress(r)
		if err != nil {
			return errors.Wrapf(err, "cannot uncompress downloaded chart archive %q", path)
		}
		chartfs, err := projectionfs.New(fs, dir)
		if err != nil {
			return errors.Wrapf(err, "cannot create projection %q", path)
		}
		err = utils2.ExtractTarToFs(chartfs, reader)
		if err != nil {
			return errors.Wrapf(err, "cannot extract downloaded chart archive %q", path)
		}
		entries, err := vfs.ReadDir(fs, dir)
		if err != nil {
			return errors.Wrapf(err, "cannot find chart folder in", dir)
		}
		if len(entries) != 1 {
			return errors.Wrapf(err, "expected single charts folder in archive, but found %d folders", len(entries))
		}
		path = filepath.Join(dir, entries[0].Name())

		out.Outf(octx, "Loading %d sub charts into %s...\n", len(cfg.SubCharts), path)
		var finalize utils2.Finalizer
		defer finalize.Finalize()
		charts := filepath.Join(path, "charts")
		err = os.Mkdir(charts, 0o700)
		if err != nil {
			return errors.Wrapf(err, "cannot mkdir %q", charts)
		}
		for n, r := range cfg.SubCharts {
			out.Outf(octx, "  Loading sub chart %q from resource %s@%s\n", n, r, common.VersionedElementKey(cv))
			acc, rcv, err := utils.ResolveResourceReference(cv, r, nil)
			if err != nil {
				return errors.ErrNotFoundWrap(err, "chart reference", r.String())
			}
			finalize.Close(rcv)

			if acc.Meta().Type != resourcetypes.HELM_CHART {
				return errors.Newf("%s: resource type %q required, but found %q", r, resourcetypes.HELM_CHART, acc.Meta().Type)
			}

			subpath := filepath.Join(charts, n)
			_, _, err = download.For(ctx).Download(common.NewPrinter(octx.StdOut()), acc, subpath, osfs.New())
			if err != nil {
				return errors.Wrapf(err, "downloading helm chart %s", r)
			}
			finalize.Finalize()
		}
	}

	out.Outf(octx, "Localizing helm chart...\n")

	for i, v := range cfg.ImageMapping {
		acc, rcv, err := utils.ResolveResourceReference(cv, v.ResourceReference, nil)
		if err != nil {
			return errors.ErrNotFoundWrap(err, "mapping", fmt.Sprintf("%d (%s)", i+1, &v.ResourceReference))
		}
		rcv.Close()
		ref, err := utils.GetOCIArtifactRef(ctx, acc)
		if err != nil {
			return errors.Wrapf(err, "mapping %d: cannot resolve resource %s to an OCI Reference", i+1, v)
		}
		ix := strings.Index(ref, ":")
		if ix < 0 {
			ix = strings.Index(ref, "@")
			if ix < 0 {
				return errors.Wrapf(err, "mapping %d: image tag or digest missing (%s)", i+1, ref)
			}
		}
		repo := ref[:ix]
		tag := ref[ix+1:]
		if v.Repository != "" {
			err = Set(values, v.Repository, repo)
			if err != nil {
				return errors.Wrapf(err, "mapping %d: assigning repositry to property %q", v.Repository)
			}
		}
		if v.Tag != "" {
			err = Set(values, v.Tag, tag)
			if err != nil {
				return errors.Wrapf(err, "mapping %d: assigning tag to property %q", v.Tag)
			}
		}
		if v.Image != "" {
			err = Set(values, v.Image, ref)
			if err != nil {
				return errors.Wrapf(err, "mapping %d: assigning image to property %q", v.Image)
			}
		}
	}

	out.Outf(octx, "Installing helm chart...\n")

	ns := "default"
	if cfg.Namespace != "" {
		ns = cfg.Namespace
	}
	if s, ok := values["namespace"].(string); ok && s != "" {
		ns = s
	}
	release := cfg.Release
	if s, ok := values["release"].(string); ok && s != "" {
		release = s
	}
	valuesdata, err := runtime.DefaultYAMLEncoding.Marshal(values)
	if err != nil {
		return errors.Wrapf(err, "marshal values")
	}

	dcfg := &driver.Config{
		ChartPath:       path,
		Release:         release,
		Namespace:       ns,
		CreateNamespace: cfg.CreateNamespace,
		Values:          valuesdata,
		Kubeconfig:      kubeconfig,
	}
	switch action {
	case "install":
		return d.Install(dcfg)
	case "uninstall":
		return d.Uninstall(dcfg)
	default:
		return errors.ErrNotImplemented("action", action)
	}
}
