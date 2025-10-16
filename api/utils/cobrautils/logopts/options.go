package logopts

import (
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/config"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm"
	loggingopt "ocm.software/ocm/api/utils/cobrautils/logopts/logging"
)

var Description = `
The <code>--log*</code> options can be used to configure the logging behaviour.
For details see <CMD>ocm logging</CMD>.

There is a quick config option <code>--logkeys</code> to configure simple
tag/realm based condition rules. The comma-separated names build an AND rule.
Hereby, names starting with a slash (<code>/</code>) denote a realm (without
the leading slash). A realm is a slash separated sequence of identifiers. If
the realm name starts with a plus (<code>+</code>) character the generated rule 
will match the realm and all its sub-realms, otherwise, only the dedicated
realm is affected. For example <code>/+ocm=trace</code> will enable all log output of the
OCM library.

A tag directly matches the logging tags. Used tags and realms can be found under
topic <CMD>ocm logging</CMD>. The ocm coding basically uses the realm <code>ocm</code>.
The default level to enable is <code>info</code>. Separated by an equal sign (<code>=</code>)
optionally a dedicated level can be specified. Log levels can be (<code>error</code>,
<code>warn</code>, <code>info</code>, <code>debug</code> and <code>trace</code>.
The default level is <code>warn</code>.
The <code>--logconfig*</code> options can be used to configure a complete
logging configuration (yaml/json) via command line. If the argument starts with
an <code>@</code>, the logging configuration is taken from a file.
`

////////////////////////////////////////////////////////////////////////////////

type EvaluatedOptions struct {
	LogForward *config.Config
	LogFile    *loggingopt.LogFile
}

func (o *EvaluatedOptions) Close() error {
	if o.LogFile == nil {
		return nil
	}
	return o.LogFile.Close()
}

type Options struct {
	ConfigFragment
	*EvaluatedOptions
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.ConfigFragment.AddFlags(fs)
}

func (o *Options) Close() error {
	if o.EvaluatedOptions == nil {
		return nil
	}
	return o.EvaluatedOptions.Close()
}

func (o *Options) Configure(ctx ocm.Context, logctx logging.Context) error {
	var err error

	o.EvaluatedOptions, err = o.ConfigFragment.Evaluate(ctx, logctx, true)
	return err
}
