package app

import (
	"strings"

	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
)

// Prepare pre-prepares CLI options by evaluation the main options
// and providing an appropriately configured cli context.
func Prepare(ctx clictx.Context, args []string) (*CLIOptions, []string, error) {
	if ctx == nil {
		ctx = clictx.DefaultContext()
	}
	opts := &CLIOptions{
		Context: ctx,
	}
	flagset := pflag.NewFlagSet("dummy", pflag.ContinueOnError)

	opts.AddFlags(flagset)

	flags := []string{}
	inFlag := false
	help := false
	for i, arg := range args {
		switch {
		case arg == "--help" || arg == "-h":
			help = true
			continue
		// A long flag with a space separated value
		case strings.HasPrefix(arg, "--") && !strings.Contains(arg, "="):
			// TODO: this isn't quite right, we should really check ahead for 'true' or 'false'
			inFlag = !hasNoOptDefVal(arg[2:], flagset)
			flags = append(flags, arg)
			continue
		// A short flag with a space separated value
		case strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") && len(arg) == 2 && !shortHasNoOptDefVal(arg[1:], flagset):
			inFlag = true
			flags = append(flags, arg)
			continue
		// The value for a flag
		case inFlag:
			inFlag = false
			flags = append(flags, arg)
			continue
		// A flag without a value, or with an `=` separated value
		case isFlagArg(arg):
			flags = append(flags, arg)
			continue
		}
		args = args[i:]
		break
	}
	err := flagset.Parse(flags)
	if err != nil {
		return nil, nil, err
	}
	if help {
		args = append([]string{"--help"}, args...)
	}

	err = opts.Complete()
	if err != nil {
		return nil, nil, err
	}
	return opts, args, nil
}

func hasNoOptDefVal(name string, fs *pflag.FlagSet) bool {
	flag := fs.Lookup(name)
	if flag == nil {
		return false
	}
	return flag.NoOptDefVal != ""
}

func shortHasNoOptDefVal(name string, fs *pflag.FlagSet) bool {
	if len(name) == 0 {
		return false
	}

	flag := fs.ShorthandLookup(name[:1])
	if flag == nil {
		return false
	}
	return flag.NoOptDefVal != ""
}

func isFlagArg(arg string) bool {
	return ((len(arg) >= 3 && arg[1] == '-') ||
		(len(arg) >= 2 && arg[0] == '-' && arg[1] != '-'))
}
