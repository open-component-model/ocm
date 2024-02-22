package configutils

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/open-component-model/ocm/pkg/contexts/datacontext/config"

	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/errors"
)

func Configure(path string) error {
	return ConfigureContext(config.DefaultContext(), path)
}

func ConfigureContext(ctxp config.ContextProvider, path string) error {
	ctx := config.FromProvider(ctxp)

	h, _ := os.UserHomeDir()
	if path == "" {
		if h != "" {
			cfg := h + "/.ocmconfig"
			if ok, err := vfs.FileExists(osfs.New(), cfg); ok && err == nil {
				path = cfg
			}
		}
	}

	if path != "" {
		if strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
			if len(h) == 0 {
				return fmt.Errorf("no home directory found for resolving path of ocm config file %q", path)
			}
			path = h + path[1:]
		}
		data, err := vfs.ReadFile(osfs.New(), path)
		if err != nil {
			return errors.Wrapf(err, "cannot read ocm config file %q", path)
		}

		sctx := spiffing.New().WithFeatures(features.INTERPOLATION, features.CONTROL)
		data, err = spiffing.Process(sctx, spiffing.NewSourceData(path, data))
		if err != nil {
			return errors.Wrapf(err, "processing ocm config %q", path)
		}
		cfg, err := ctx.GetConfigForData(data, nil)
		if err != nil {
			return errors.Wrapf(err, "invalid ocm config file %q", path)
		}
		err = ctx.ApplyConfig(cfg, path)
		if err != nil {
			return errors.Wrapf(err, "cannot apply ocm config %q", path)
		}
	}
	return nil
}
