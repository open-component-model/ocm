package spiff

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils"
)

func init() {
	transferhandler.RegisterHandler(100, &TransferOptionsCreator{})
}

type Options struct {
	standard.Options
	script []byte
	fs     vfs.FileSystem
}

var (
	_ transferhandler.TransferOption = (*Options)(nil)

	_ ScriptOption           = (*Options)(nil)
	_ ScriptFilesystemOption = (*Options)(nil)
)

type TransferOptionsCreator = transferhandler.SpecializedOptionsCreator[*Options, Options]

func (o *Options) NewOptions() transferhandler.TransferHandlerOptions {
	return &Options{}
}

func (o *Options) NewTransferHandler() (transferhandler.TransferHandler, error) {
	return New(o)
}

func (o *Options) ApplyTransferOption(target transferhandler.TransferOptions) error {
	if len(o.script) > 0 {
		if opts, ok := target.(ScriptOption); ok {
			opts.SetScript(o.script)
		}
	}
	if o.fs != nil {
		if opts, ok := target.(ScriptFilesystemOption); ok {
			opts.SetScriptFilesystem(o.fs)
		}
	}
	return o.Options.ApplyTransferOption(target)
}

func (o *Options) SetScript(data []byte) {
	o.script = data
}

func (o *Options) GetScript() []byte {
	return o.script
}

func (o *Options) SetScriptFilesystem(fs vfs.FileSystem) {
	o.fs = fs
}

func (o *Options) GetScriptFilesystem() vfs.FileSystem {
	return o.fs
}

///////////////////////////////////////////////////////////////////////////////

type ScriptOption interface {
	SetScript(data []byte)
	GetScript() []byte
}

type scriptOption struct {
	TransferOptionsCreator
	source string
	script func() ([]byte, error)
}

func (o *scriptOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if o.script == nil {
		return nil
	}
	script, err := o.script()
	if err != nil {
		return err
	}
	_, err = spiffing.New().Unmarshal(o.source, script)
	if err != nil {
		return err
	}

	if eff, ok := to.(ScriptOption); ok {
		eff.SetScript(script)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "script")
	}
}

func Script(data []byte) transferhandler.TransferOption {
	if data == nil {
		return &scriptOption{
			source: "script",
		}
	}
	return &scriptOption{
		source: "script",
		script: func() ([]byte, error) { return data, nil },
	}
}

func ScriptByFile(path string, fss ...vfs.FileSystem) transferhandler.TransferOption {
	path, _ = utils.ResolvePath(path)
	return &scriptOption{
		source: path,
		script: func() ([]byte, error) { return vfs.ReadFile(utils.FileSystem(fss...), path) },
	}
}

///////////////////////////////////////////////////////////////////////////////

type ScriptFilesystemOption interface {
	SetScriptFilesystem(fs vfs.FileSystem)
	GetScriptFilesystem() vfs.FileSystem
}

type filesystemOption struct {
	TransferOptionsCreator
	fs vfs.FileSystem
}

func (o *filesystemOption) ApplyTransferOption(to transferhandler.TransferOptions) error {
	if eff, ok := to.(ScriptFilesystemOption); ok {
		eff.SetScriptFilesystem(o.fs)
		return nil
	} else {
		return errors.ErrNotSupported(transferhandler.KIND_TRANSFEROPTION, "script filesystem")
	}
}

func ScriptFilesystem(fs vfs.FileSystem) transferhandler.TransferOption {
	return &filesystemOption{
		fs: fs,
	}
}
