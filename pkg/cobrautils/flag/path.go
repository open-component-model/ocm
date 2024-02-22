package flag

import (
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/spf13/pflag"
)

type pathValue string

func newPathValue(val string, p *string) *pathValue {
	*p = pathConv(val)
	return (*pathValue)(p)
}

func (s *pathValue) Set(val string) error {
	*s = pathValue(pathConv(val))
	return nil
}

func (s *pathValue) Type() string { return "filepath" }

func (s *pathValue) String() string { return string(*s) }

func pathConv(sval string) string {
	vol, paths, rooted := filepath.SplitPath(sval)
	if rooted {
		return vol + "/" + strings.Join(paths, "/")
	}
	return vol + strings.Join(paths, "/")
}

// PathVar defines a filepath flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func PathVar(f *pflag.FlagSet, p *string, name string, value string, usage string) {
	f.VarP(newPathValue(value, p), name, "", usage)
}

// PathVarP is like PathVar, but accepts a shorthand letter that can be used after a single dash.
func PathVarP(f *pflag.FlagSet, p *string, name, shorthand string, value string, usage string) {
	f.VarP(newPathValue(value, p), name, shorthand, usage)
}

// Path defines a filepath flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func Path(f *pflag.FlagSet, name string, value string, usage string) *string {
	p := new(string)
	PathVarP(f, p, name, "", value, usage)
	return p
}

// PathP is like Path, but accepts a shorthand letter that can be used after a single dash.
func PathP(f *pflag.FlagSet, name, shorthand string, value string, usage string) *string {
	p := new(string)
	PathVarP(f, p, name, shorthand, value, usage)
	return p
}
