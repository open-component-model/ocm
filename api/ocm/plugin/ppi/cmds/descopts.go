package cmds

import (
	"io"

	"github.com/mandelsoft/goutils/optionutils"
)

type Option = optionutils.Option[*PluginCommand]

type stdout struct {
	io.Writer
}

func StdOut(w io.Writer) Option {
	return &stdout{w}
}

func (o *stdout) ApplyTo(p *PluginCommand) {
	p.command.SetOut(o.Writer)
}

////////////////////////////////////////////////////////////////////////////////

type stderr struct {
	io.Writer
}

func StdErr(w io.Writer) Option {
	return &stderr{w}
}

func (o *stderr) ApplyTo(p *PluginCommand) {
	p.command.SetErr(o.Writer)
}

////////////////////////////////////////////////////////////////////////////////

type stdin struct {
	io.Reader
}

func StdIn(r io.Reader) Option {
	return &stdin{r}
}

func (o *stdin) ApplyTo(p *PluginCommand) {
	p.command.SetIn(o.Reader)
}
