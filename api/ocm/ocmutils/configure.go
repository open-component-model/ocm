package ocmutils

import (
	"fmt"
	"os"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/pkgutils"
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/config"
	configcfg "ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/ocmutils/defaultconfigregistry"
	"ocm.software/ocm/api/utils"
)

const DEFAULT_OCM_CONFIG = ".ocmconfig"

const DEFAULT_OCM_CONFIG_DIR = ".ocm"

func Configure(ctxp config.ContextProvider, path string, fss ...vfs.FileSystem) (ocm.Context, error) {
	ctx, _, err := Configure2(ctxp, path, fss...)
	return ctx, err
}

func Configure2(ctx config.ContextProvider, path string, fss ...vfs.FileSystem) (ocm.Context, config.Config, error) {
	var ocmctx ocm.Context

	cfg, err := configcfg.NewAggregator(false)
	if err != nil {
		return nil, nil, err
	}
	fs := utils.FileSystem(fss...)
	if ctx == nil {
		ocmctx = ocm.DefaultContext()
		ctx = ocmctx
	} else {
		if c, ok := ctx.(ocm.Context); ok {
			ocmctx = c
		}
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
				return nil, nil, fmt.Errorf("no home directory found for resolving path of ocm config file %q", path)
			}
			path = h + path[1:]
		}
		data, err := vfs.ReadFile(fs, path)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot read ocm config file %q", path)
		}

		if eff, err := ConfigureByData2(ctx, data, path); err != nil {
			return nil, nil, err
		} else {
			err = cfg.AddConfig(eff)
			if err != nil {
				return nil, nil, err
			}
		}
	} else {
		for _, h := range defaultconfigregistry.Get() {
			desc, def, err := h(ctx.ConfigContext())
			if err != nil {
				return nil, nil, err
			}
			if def != nil {
				name, err := pkgutils.GetPackageName(h)
				if err != nil {
					name = "unknown handler"
				}
				err = ctx.ConfigContext().ApplyConfig(def, fmt.Sprintf("%s: %s", name, desc))
				if err != nil {
					return nil, nil, errors.Wrapf(err, "cannot apply default config from %s(%s)", name, desc)
				}
				err = cfg.AddConfig(def)
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}
	return ocmctx, cfg.Get(), nil
}

func ConfigureByData(ctx config.ContextProvider, data []byte, info string) error {
	_, err := ConfigureByData2(ctx, data, info)
	return err
}

func ConfigureByData2(ctx config.ContextProvider, data []byte, info string) (config.Config, error) {
	var err error

	sctx := spiffing.New().WithFeatures(features.INTERPOLATION, features.CONTROL)
	data, err = spiffing.Process(sctx, spiffing.NewSourceData(info, data))
	if err != nil {
		return nil, errors.Wrapf(err, "processing ocm config %q", info)
	}
	cfg, err := ctx.ConfigContext().GetConfigForData(data, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid ocm config file %q", info)
	}
	err = ctx.ConfigContext().ApplyConfig(cfg, info)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot apply ocm config %q", info)
	}
	return cfg, nil
}
