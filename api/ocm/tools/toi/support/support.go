package support

import (
	"fmt"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/extensions/repositories/memory"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	ocmutils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/tools/toi/install"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
)

type ExecutorOptions struct {
	Context              ocm.Context
	Logger               logging.Logger
	OutputContext        out.Context
	Action               string
	ComponentVersionName string
	Root                 string
	Inputs               string
	Outputs              string
	OCMConfig            string
	Config               string
	ConfigData           []byte
	Parameters           string
	ParameterData        []byte
	RepoPath             string
	Repository           ocm.Repository
	CredentialRepo       credentials.Repository
	ComponentVersion     ocm.ComponentVersionAccess
	Closer               func() error
}

func (o *ExecutorOptions) FileSystem() vfs.FileSystem {
	return vfsattr.Get(o.Context)
}

func (o *ExecutorOptions) Complete() error {
	if o.ComponentVersionName == "" {
		return fmt.Errorf("component version required")
	}
	compvers, err := common.ParseNameVersion(o.ComponentVersionName)
	if err != nil {
		return fmt.Errorf("unable to parse component name and version: %w", err)
	}

	if o.OutputContext == nil {
		o.OutputContext = out.New()
	}

	if o.Action == "" {
		o.Action = "install"
	}

	if o.Root == "" {
		o.Root = install.PathTOI
	}

	if o.Inputs == "" {
		o.Inputs = o.Root + "/" + install.Inputs
	}

	if o.Outputs == "" {
		o.Outputs = o.Root + "/" + install.Outputs
	}

	if o.RepoPath == "" {
		o.RepoPath = o.Inputs + "/" + install.InputOCMRepo
	}

	if o.Config == "" {
		cfg := o.Inputs + "/" + install.InputConfig
		if ok, err := vfs.FileExists(o.FileSystem(), cfg); ok && err == nil {
			o.Config = cfg
		}
	}

	if o.Config != "" && o.ConfigData == nil {
		o.ConfigData, err = utils.ReadFile(o.Config, o.FileSystem())
		if err != nil {
			return errors.Wrapf(err, "cannot read config %q", o.Config)
		}
	}

	if o.OCMConfig == "" {
		cfg, err := utils.ResolvePath(o.Inputs + "/" + install.InputOCMConfig)
		if err != nil {
			return errors.Wrapf(err, "cannot resolve OCM config %q", o.Inputs)
		}
		if ok, err := vfs.FileExists(o.FileSystem(), cfg); ok && err == nil {
			o.OCMConfig = cfg
		}
	}

	o.Context, err = ocmutils.Configure(o.Context, o.OCMConfig)
	if err != nil {
		return fmt.Errorf("unable to configure context: %w", err)
	}

	if o.Parameters == "" {
		p, err := utils.ResolvePath(o.Inputs + "/" + install.InputParameters)
		if err != nil {
			return errors.Wrapf(err, "cannot resolve path %q", o.Inputs)
		}
		if ok, err := vfs.FileExists(o.FileSystem(), p); ok && err == nil {
			o.Parameters = p
		}
	}

	if o.Parameters != "" && o.ParameterData == nil {
		o.ParameterData, err = utils.ReadFile(o.Parameters, o.FileSystem())
		if err != nil {
			return errors.Wrapf(err, "cannot read parameters %q", o.Config)
		}
	}

	var repoCloser io.Closer
	if o.Repository == nil {
		repo, err := ctf.Open(o.Context, accessobj.ACC_READONLY, o.RepoPath, 0, accessio.PathFileSystem(o.FileSystem()))
		if err != nil {
			return errors.Wrapf(err, "cannot open ctf %q", o.RepoPath)
		}
		o.Repository = repo
		repoCloser = repo
	}

	var versCloser io.Closer

	if o.ComponentVersion == nil {
		cv, err := o.Repository.LookupComponentVersion(compvers.GetName(), compvers.GetVersion())
		if err != nil {
			return fmt.Errorf("failed component version lookup: %w", err)
		}
		o.ComponentVersion = cv
		versCloser = cv
	}

	old := o.Closer
	o.Closer = func() error {
		list := errors.ErrListf("closing")
		if versCloser != nil {
			list.Add(errors.Wrapf(versCloser.Close(), "component version"))
		}
		if repoCloser != nil {
			list.Add(errors.Wrapf(repoCloser.Close(), "repository"))
		}
		if old != nil {
			list.Add(errors.Wrapf(old(), "external closer"))
		}
		return list.Result()
	}

	if o.CredentialRepo == nil {
		c, err := o.Context.CredentialsContext().RepositoryForSpec(memory.NewRepositorySpec("default"))
		if err != nil {
			return errors.Wrapf(err, "cannot get default memory based credential repository")
		}
		o.CredentialRepo = c
	}
	return nil
}

type Executor struct {
	Completed bool
	Options   *ExecutorOptions
	Run       func(o *ExecutorOptions) error
}

func (e *Executor) Execute() error {
	if e.Options == nil {
		e.Completed = false
		e.Options = &ExecutorOptions{}
	}
	if !e.Completed {
		err := e.Options.Complete()
		if err != nil {
			return fmt.Errorf("unable to complete options: %w", err)
		}
	}
	list := errors.ErrListf("executor:")
	list.Add(e.Run(e.Options))
	if e.Options.Closer != nil {
		list.Add(e.Options.Closer())
	}
	return list.Result()
}
