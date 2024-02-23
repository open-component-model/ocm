package flag

import (
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

type pathArrayValue struct {
	value   *[]string
	changed bool
}

func newPathArrayValue(val []string, p *[]string) *pathArrayValue {
	ssv := new(pathArrayValue)
	ssv.value = p
	*ssv.value = pathArrayConv(val)
	return ssv
}

func (s *pathArrayValue) Set(val string) error {
	if !s.changed {
		*s.value = pathStringListConv(val)
		s.changed = true
	} else {
		*s.value = append(*s.value, pathConv(val))
	}
	return nil
}

func (s *pathArrayValue) Append(val string) error {
	*s.value = append(*s.value, pathConv(val))
	return nil
}

func (s *pathArrayValue) Replace(val []string) error {
	out := make([]string, len(val))
	for i, d := range val {
		var err error
		out[i] = pathConv(d)
		if err != nil {
			return err
		}
	}
	*s.value = out
	return nil
}

func (s *pathArrayValue) GetSlice() []string {
	out := make([]string, len(*s.value))
	for i, d := range *s.value {
		out[i] = d
	}
	return out
}

func (s *pathArrayValue) Type() string {
	return "stringArray"
}

func (s *pathArrayValue) String() string {
	str := new(string)
	*str = strings.Join(*s.value, string(filepath.ListSeparator))
	return *str
}

// Converts every string into correct filepath format. See pathConv for more details.
func pathArrayConv(sval []string) []string {
	for i, val := range sval {
		sval[i] = pathConv(val)
	}
	return sval
}

// pathStringListConv converts a string containing multiple filepaths seperated by filepath.ListSeparator into a list
// of filepaths.
func pathStringListConv(sval string) []string {
	values := filepath.SplitList(sval)
	values = pathArrayConv(values)
	return values
}

// PathArrayVar defines a filepath flag with specified name, default value, and usage string.
// The argument p points to a []string variable in which to store the values of the multiple flags.
func PathArrayVar(f *pflag.FlagSet, p *[]string, name string, value []string, usage string) {
	f.VarP(newPathArrayValue(value, p), name, "", usage)
}

// PathArrayVarP is like PathArrayVar, but accepts a shorthand letter that can be used after a single dash.
func PathArrayVarP(f *pflag.FlagSet, p *[]string, name, shorthand string, value []string, usage string) {
	f.VarP(newPathArrayValue(value, p), name, shorthand, usage)
}

// PathArray defines a filepath flag with specified name, default value, and usage string.
// The return value is the address of a []string variable that stores the value of the flag.
func PathArray(f *pflag.FlagSet, name string, value []string, usage string) *[]string {
	p := []string{}
	PathArrayVarP(f, &p, name, "", value, usage)
	return &p
}

// PathArrayP is like PathArray, but accepts a shorthand letter that can be used after a single dash.
func PathArrayP(f *pflag.FlagSet, name, shorthand string, value []string, usage string) *[]string {
	p := []string{}
	PathArrayVarP(f, &p, name, shorthand, value, usage)
	return &p
}

// PathArrayVarPF is like PathArrayVarP, but returns the created flag.
func PathArrayVarPF(f *pflag.FlagSet, p *[]string, name, shorthand string, value []string, usage string) *pflag.Flag {
	PathArrayVarP(f, p, name, shorthand, value, usage)
	return f.Lookup(name)
}
