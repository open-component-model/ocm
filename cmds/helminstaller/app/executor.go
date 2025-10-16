package app

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"ocm.software/ocm/api/ocm/tools/toi/support"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/helminstaller/app/driver"
)

func New(d driver.Driver) func(o *support.ExecutorOptions) error {
	return func(o *support.ExecutorOptions) error {
		return Executor(d, o)
	}
}

func Executor(d driver.Driver, o *support.ExecutorOptions) error {
	var cfg Config
	err := runtime.DefaultYAMLEncoding.Unmarshal(o.ConfigData, &cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal config")
	}
	var values map[string]interface{}
	err = runtime.DefaultYAMLEncoding.Unmarshal(o.ParameterData, &values)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal parameters")
	}
	if cfg.KubeConfigName == "" {
		cfg.KubeConfigName = "target"
	}
	creds, err := o.CredentialRepo.LookupCredentials(cfg.KubeConfigName)
	if err != nil {
		return errors.Wrapf(err, "cannot get kubeconfig with key %q", cfg.KubeConfigName)
	}
	v := creds.GetProperty("KUBECONFIG")
	if v == "" {
		return errors.Wrapf(err, "property KUBECONFIG missing in credential %q", cfg.KubeConfigName)
	}
	exec := &Execution{
		driver:          d,
		ExecutorOptions: o,
		path:            "",
		fs:              osfs.New(),
	}
	return exec.Execute(&cfg, values, []byte(v))
}
