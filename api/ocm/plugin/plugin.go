package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	mlog "github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/datacontext/attrs/clicfgattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/plugin/cache"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/accessmethod"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/accessmethod/compose"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/accessmethod/get"
	accval "ocm.software/ocm/api/ocm/plugin/ppi/cmds/accessmethod/validate"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/action"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/action/execute"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/command"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/download"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/mergehandler"
	merge "ocm.software/ocm/api/ocm/plugin/ppi/cmds/mergehandler/execute"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/upload"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/upload/put"
	uplval "ocm.software/ocm/api/ocm/plugin/ppi/cmds/upload/validate"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/valueset"
	vscompose "ocm.software/ocm/api/ocm/plugin/ppi/cmds/valueset/compose"
	vsval "ocm.software/ocm/api/ocm/plugin/ppi/cmds/valueset/validate"
	"ocm.software/ocm/api/ocm/valuemergehandler"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/cobrautils/logopts/logging"
	"ocm.software/ocm/api/utils/runtime"
)

type Plugin = *pluginImpl

type impl = cache.Plugin

// //nolint: errname // is no error.
type pluginImpl struct {
	lock sync.RWMutex
	ctx  ocm.Context
	impl
	config                   json.RawMessage
	disableAutoConfiguration bool
}

func NewPlugin(ctx ocm.Context, impl cache.Plugin, config json.RawMessage) Plugin {
	return &pluginImpl{
		ctx:    ctx,
		impl:   impl,
		config: config,
	}
}

func (p *pluginImpl) Context() ocm.Context {
	return p.ctx
}

func (p *pluginImpl) DisableAutoConfiguration(flag bool) {
	p.disableAutoConfiguration = flag
}

func (p *pluginImpl) IsAutoConfigurationEnabled() bool {
	return !p.disableAutoConfiguration
}

func (p *pluginImpl) SetConfig(config json.RawMessage) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.config = config
}

func (p *pluginImpl) Exec(r io.Reader, w io.Writer, args ...string) (result []byte, rerr error) {
	var (
		finalize finalizer.Finalizer
		err      error
		logfile  *os.File
	)

	defer finalize.FinalizeWithErrorPropagationf(&rerr, "error processing plugin command %s", args[0])

	if p.GetDescriptor().ForwardLogging {
		logfile, err = os.CreateTemp("", "ocm-plugin-log-*")
		if rerr != nil {
			return nil, err
		}
		logfile.Close()
		finalize.With(func() error {
			return os.Remove(logfile.Name())
		}, "failed to remove temporary log file %s", logfile.Name())

		lcfg := &logging.LoggingConfiguration{}
		_, err = p.Context().ConfigContext().ApplyTo(0, lcfg)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot extract plugin logging configuration")
		}
		lcfg.LogFileName = logfile.Name()
		data, err := json.Marshal(lcfg)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal plugin logging configuration")
		}
		args = append([]string{"--" + ppi.OptPlugingLogConfig, string(data)}, args...)
	}

	if p.ctx.Logger(TAG).Enabled(mlog.DebugLevel) {
		// Plainly kill any credentials found in the logger.
		// Stupidly match for "credentials" arg.
		// Not totally safe, but better than nothing.
		logargs := make([]string, len(args))
		for i, arg := range args {
			if logargs[i] != "" {
				continue
			}
			if strings.Contains(arg, "credentials") {
				if strings.Contains(arg, "=") {
					logargs[i] = "***"
				} else if i+1 < len(args)-1 {
					logargs[i+1] = "***"
				}
			}
			logargs[i] = arg
		}

		if len(p.config) == 0 {
			p.ctx.Logger(TAG).Debug("execute plugin action", "path", p.Path(), "args", logargs)
		} else {
			p.ctx.Logger(TAG).Debug("execute plugin action", "path", p.Path(), "args", logargs, "config", p.config)
		}
	}

	data, err := cache.Exec(p.ctx, p.Path(), p.config, r, w, args...)

	if logfile != nil {
		r, oerr := os.OpenFile(logfile.Name(), vfs.O_RDONLY, 0o600)
		if oerr == nil {
			finalize.Close(r, "plugin logfile", logfile.Name())
			w := p.ctx.LoggingContext().Tree().LogWriter()
			if w == nil {
				if logging.GlobalLogFile != nil {
					w = logging.GlobalLogFile.File()
				}
				if w == nil {
					w = os.Stderr
				}
			}

			// weaken the sync problem when merging log files.
			// If a SyncWriter is used, the copy is done under a write lock.
			// This is only a solution, if the log records are written
			// by single write calls.
			// The underlying logging apis do not expose their
			// sync mechanism for writing log records.
			if writer, ok := w.(io.ReaderFrom); ok {
				writer.ReadFrom(r)
			} else {
				io.Copy(w, r)
			}
		}
	}
	return data, err
}

func (p *pluginImpl) MergeValue(specification *valuemergehandler.Specification, local, inbound valuemergehandler.Value) (bool, *valuemergehandler.Value, error) {
	desc := p.GetValueMappingDescriptor(specification.Algorithm)
	if desc == nil {
		return false, nil, errors.ErrNotSupported(valuemergehandler.KIND_VALUE_MERGE_ALGORITHM, specification.Algorithm, KIND_PLUGIN, p.Name())
	}
	input, err := json.Marshal(ppi.ValueMergeData{
		Local:   local,
		Inbound: inbound,
	})
	if err != nil {
		return false, nil, err
	}

	args := []string{mergehandler.Name, merge.Name, specification.Algorithm}
	if len(specification.Config) > 0 {
		args = append(args, string(specification.Config))
	}

	var buf bytes.Buffer
	_, err = p.Exec(bytes.NewReader(input), &buf, args...)
	if err != nil {
		return false, nil, errors.Wrapf(err, "plugin %s", p.Name())
	}
	var r ppi.ValueMergeResult

	err = json.Unmarshal(buf.Bytes(), &r)
	if err != nil {
		if r.Message != "" {
			return false, nil, fmt.Errorf("%w: %s", err, r.Message)
		}
		return false, nil, err
	}
	return r.Modified, &r.Value, nil
}

func (p *pluginImpl) Action(spec ppi.ActionSpec, creds json.RawMessage) (ppi.ActionResult, error) {
	desc := p.GetActionDescriptor(spec.GetKind())
	if desc == nil {
		return nil, errors.ErrNotSupported(KIND_ACTION, spec.GetKind(), KIND_PLUGIN, p.Name())
	}
	if desc.ConsumerType != "" {
		cid := spec.GetConsumerAttributes()
		cid[cpi.ID_TYPE] = desc.ConsumerType
		c, err := credentials.CredentialsForConsumer(p.Context(), credentials.ConsumerIdentity(cid), hostpath.Matcher)
		if err != nil || c == nil {
			return nil, errors.ErrNotFound(credentials.KIND_CREDENTIALS, cid.String())
		}
		creds, err = json.Marshal(c.Properties())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal credentials")
		}
	}

	data, err := p.ctx.GetActions().GetActionTypes().EncodeActionSpec(spec, runtime.DefaultJSONEncoding)
	if err != nil {
		return nil, err
	}

	args := []string{action.Name, execute.Name, string(data)}
	if creds != nil {
		args = append(args, "--"+get.OptCreds, string(creds))
	}

	result, err := p.Exec(nil, nil, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s", p.Name())
	}

	info, err := p.ctx.GetActions().GetActionTypes().DecodeActionResult(result, runtime.DefaultJSONEncoding)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s: cannot unmarshal action result", p.Name())
	}
	return info, nil
}

func (p *pluginImpl) ValidateAccessMethod(spec []byte) (*ppi.AccessSpecInfo, error) {
	result, err := p.Exec(nil, nil, accessmethod.Name, accval.Name, string(spec))
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s", p.Name())
	}

	var info ppi.AccessSpecInfo
	err = json.Unmarshal(result, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s: cannot unmarshal access spec info", p.Name())
	}
	return &info, nil
}

func (p *pluginImpl) ComposeAccessMethod(name string, opts flagsets.ConfigOptions, base flagsets.Config) error {
	cfg := flagsets.Config{}
	for _, o := range opts.Options() {
		cfg[o.GetName()] = o.Value()
	}
	optsdata, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal option values")
	}
	basedata, err := json.Marshal(base)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal access specification base value")
	}
	result, err := p.Exec(nil, nil, accessmethod.Name, compose.Name, name, string(optsdata), string(basedata))
	if err != nil {
		return err
	}
	var r flagsets.Config
	err = json.Unmarshal(result, &r)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal composition result")
	}

	for k := range base {
		delete(base, k)
	}
	for k, v := range r {
		base[k] = v
	}
	return nil
}

func (p *pluginImpl) ValidateUploadTarget(name string, spec []byte) (*ppi.UploadTargetSpecInfo, error) {
	result, err := p.Exec(nil, nil, upload.Name, uplval.Name, name, string(spec))
	if err != nil {
		return nil, errors.Wrapf(err, "plugin uploader %s/%s", p.Name(), name)
	}

	var info ppi.UploadTargetSpecInfo
	err = json.Unmarshal(result, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin uploader %s/%s: cannot unmarshal upload target info", p.Name(), name)
	}
	return &info, nil
}

func (p *pluginImpl) Get(w io.Writer, creds, spec json.RawMessage) error {
	args := []string{accessmethod.Name, get.Name, string(spec)}
	if creds != nil {
		args = append(args, "--"+get.OptCreds, string(creds))
	}
	_, err := p.Exec(nil, w, args...)
	return err
}

func (p *pluginImpl) Put(name string, r io.Reader, artType, mimeType, hint, digest string, creds, target json.RawMessage) (ocm.AccessSpec, error) {
	args := []string{upload.Name, put.Name, name, string(target)}

	if creds != nil {
		args = append(args, "--"+put.OptCreds, string(creds))
	}
	if hint != "" {
		args = append(args, "--"+put.OptHint, hint)
	}
	if mimeType != "" {
		args = append(args, "--"+put.OptMedia, mimeType)
	}
	if artType != "" {
		args = append(args, "--"+put.OptArt, artType)
	}
	if digest != "" {
		args = append(args, "--"+put.OptDigest, digest)
	}
	result, err := p.Exec(r, nil, args...)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(result, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal put result")
	}
	if len(m) == 0 {
		return nil, nil // not used
	}
	return p.ctx.AccessSpecForConfig(result, runtime.DefaultJSONEncoding)
}

func (p *pluginImpl) Download(name string, r io.Reader, artType, mimeType, target string, config json.RawMessage) (bool, string, error) {
	args := []string{download.Name, name, target}

	if mimeType != "" {
		args = append(args, "--"+download.OptMedia, mimeType)
	}
	if artType != "" {
		args = append(args, "--"+download.OptArt, artType)
	}

	// new attribute can only be set for extended plugin format version
	// so, omitting config if not set is compatible with former CLI.
	if d := p.GetDescriptor().Downloaders.Get(name); len(config) > 0 && d != nil && d.ConfigScheme != "" {
		args = append(args, "--"+download.OptConfig, string(config))
	}
	result, err := p.Exec(r, nil, args...)
	if err != nil {
		return true, "", err
	}
	var m download.Result
	err = json.Unmarshal(result, &m)
	if err != nil {
		return true, "", errors.Wrapf(err, "cannot unmarshal put result")
	}
	if m.Error != "" {
		return true, "", fmt.Errorf("%s", m.Error)
	}
	return m.Path != "", m.Path, nil
}

func (p *pluginImpl) ValidateValueSet(purpose string, spec []byte) (*ppi.ValueSetInfo, error) {
	result, err := p.Exec(nil, nil, valueset.Name, vsval.Name, purpose, string(spec))
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s", p.Name())
	}

	var info ppi.ValueSetInfo
	err = json.Unmarshal(result, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin %s: cannot unmarshal value set info", p.Name())
	}
	return &info, nil
}

func (p *pluginImpl) ComposeValueSet(purpose, name string, opts flagsets.ConfigOptions, base flagsets.Config) error {
	cfg := flagsets.Config{}
	for _, o := range opts.Options() {
		cfg[o.GetName()] = o.Value()
	}
	optsdata, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal option values")
	}
	basedata, err := json.Marshal(base)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal access specification base value")
	}
	result, err := p.Exec(nil, nil, valueset.Name, vscompose.Name, purpose, name, string(optsdata), string(basedata))
	if err != nil {
		return err
	}
	var r flagsets.Config
	err = json.Unmarshal(result, &r)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal composition result")
	}

	for k := range base {
		delete(base, k)
	}
	for k, v := range r {
		base[k] = v
	}
	return nil
}

func (p *pluginImpl) Command(name string, reader io.Reader, writer io.Writer, cmdargs []string) (rerr error) {
	var finalize finalizer.Finalizer
	cmd := p.GetDescriptor().Commands.Get(name)
	if cmd == nil {
		return errors.ErrNotFound("command", name)
	}

	defer finalize.FinalizeWithErrorPropagation(&rerr)

	var f vfs.File

	args := []string{command.Name}

	a := clicfgattr.Get(p.Context())
	if a != nil && cmd.CLIConfigRequired {
		cfgdata, err := json.Marshal(a)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal CLI config")
		}
		// cannot use a vfs here, since it's not possible to pass it to the plugin
		f, err = os.CreateTemp("", "cli-om-config-*")
		if err != nil {
			return err
		}
		finalize.With(func() error {
			return os.Remove(f.Name())
		}, "failed to remove temporary config file %s", f.Name())

		_, err = f.Write(cfgdata)
		if err != nil {
			f.Close()
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
		args = append(args, "--"+command.OptCliConfig, f.Name())
	}
	args = append(append(args, name), cmdargs...)

	_, err := p.Exec(reader, writer, args...)
	return err
}
