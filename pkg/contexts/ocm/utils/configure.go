package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/defaultconfigregistry"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

const DEFAULT_OCM_CONFIG = ".ocmconfig"

const DEFAULT_OCM_CONFIG_DIR = ".ocm"

func Configure(ctx ocm.Context, path string, fss ...vfs.FileSystem) (ocm.Context, error) {
	fs := utils.FileSystem(fss...)
	if ctx == nil {
		ctx = ocm.DefaultContext()
	}
	h, _ := os.UserHomeDir()
	if path == "" {
		if h != "" {
			cfg := h + "/" + DEFAULT_OCM_CONFIG
			if ok, err := vfs.FileExists(fs, cfg); ok && err == nil {
				path = cfg
			} else {
				cfg := h + "/" + DEFAULT_OCM_CONFIG_DIR + "/ocmconfig"
				if ok, err := vfs.FileExists(fs, cfg); ok && err == nil {
					path = cfg
				} else {
					cfg := h + "/" + DEFAULT_OCM_CONFIG_DIR + "/config"
					if ok, err := vfs.FileExists(fs, cfg); ok && err == nil {
						path = cfg
					}
				}
			}
		}
	}
	if path != "" && path != "None" {
		if strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
			if len(h) == 0 {
				return nil, fmt.Errorf("no home directory found for resolving path of ocm config file %q", path)
			}
			path = h + path[1:]
		}
		data, err := vfs.ReadFile(fs, path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read ocm config file %q", path)
		}

		if err = ConfigureByData(ctx, data, path); err != nil {
			return nil, err
		}
	} else {
		for _, h := range defaultconfigregistry.Get() {
			err := h(ctx.ConfigContext())
			if err != nil {
				return nil, err
			}
		}
	}
	return ctx, nil
}

func ConfigureByData(ctx ocm.Context, data []byte, info string) error {
	var err error

	sctx := spiffing.New().WithFeatures(features.INTERPOLATION, features.CONTROL)
	data, err = spiffing.Process(sctx, spiffing.NewSourceData(info, data))
	if err != nil {
		return errors.Wrapf(err, "processing ocm config %q", info)
	}
	cfg, err := ctx.ConfigContext().GetConfigForData(data, nil)
	if err != nil {
		return errors.Wrapf(err, "invalid ocm config file %q", info)
	}
	err = ctx.ConfigContext().ApplyConfig(cfg, info)
	if err != nil {
		return errors.Wrapf(err, "cannot apply ocm config %q", info)
	}
	return nil
}
