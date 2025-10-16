package scriptoption

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "scripts.ocm" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Scripts                     map[string]ScriptSpec `json:"scripts"`
}

type ScriptSpec struct {
	Path       string          `json:"path,omitempty"`
	Script     json.RawMessage `json:"script,omitempty"`
	FileSystem vfs.FileSystem  `json:"-"`
}

// NewConfig creates a new memory ConfigSpec.
func NewConfig() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) AddScriptFile(name, path string, fss ...vfs.FileSystem) {
	var fs vfs.FileSystem
	for _, fs = range fss {
		if fs != nil {
			break
		}
	}
	if a.Scripts == nil {
		a.Scripts = map[string]ScriptSpec{}
	}
	a.Scripts[name] = ScriptSpec{Path: path, FileSystem: fs}
}

func (a *Config) AddScript(name string, data []byte) {
	if a.Scripts == nil {
		a.Scripts = map[string]ScriptSpec{}
	}
	a.Scripts[name] = ScriptSpec{Script: data}
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	t, ok := target.(*Option)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}

	spec, ok := a.Scripts[t.Script]
	if ok {
		if len(spec.Script) > 0 {
			t.ScriptData = spec.Script
		} else {
			if spec.Path == "" {
				return errors.Newf("script or path must be set for entry %q", t.Script)
			}
			data, err := utils.ReadFile(spec.Path, utils.FileSystem(spec.FileSystem, t.FileSystem))
			if err != nil {
				return errors.Wrapf(err, "script file %q", spec.Path)
			}
			t.ScriptData = data
		}
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define transfer scripts:

<pre>
    type: ` + ConfigType + `
    scripts:
      &lt;name>:
        path: &lt;>file path>
      &lt;other name>:
        script: &lt;>nested script as yaml>
</pre>
`
