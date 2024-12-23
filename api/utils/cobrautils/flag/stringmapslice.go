package flag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

type stringMapSlice struct {
	main    string
	value   *[]map[string]string
	changed bool
}

func newStringMapSliceValue(main string, val []map[string]string, p *[]map[string]string) *stringMapSlice {
	ssv := new(stringMapSlice)
	ssv.main = main
	ssv.value = p
	*ssv.value = val
	return ssv
}

func (s *stringMapSlice) Set(val string) error {
	k, v, err := parseAssignment(val)
	if err != nil {
		k = s.main
		v = val
	}
	if !s.changed {
		if k != s.main {
			return fmt.Errorf("first attribute must be the %q attribute", s.main)
		}
		*s.value = []map[string]string{{k: v}}
	} else {
		if k == s.main {
			*s.value = append(*s.value, map[string]string{k: v})
		} else {
			(*s.value)[len(*s.value)-1][k] = v
		}
	}
	s.changed = true
	return nil
}

func (s *stringMapSlice) Type() string {
	return "{<name[" + s.main + "]>=<value>}"
}

func (s *stringMapSlice) String() string {
	if *s.value == nil {
		return ""
	}
	var list []string
	for _, v := range *s.value {
		//nolint: errchkjson // initialized by unmarshal
		s, _ := json.Marshal(v)
		list = append(list, string(s))
	}
	return "[" + strings.Join(list, ", ") + "]"
}

func (s *stringMapSlice) GetSlice() []map[string]string {
	return *s.value
}

func StringMapSliceVar(f *pflag.FlagSet, p *[]map[string]string, main string, name string, value []map[string]string, usage string) {
	f.VarP(newStringMapSliceValue(main, value, p), name, "", usage)
}

func StringMapSliceVarP(f *pflag.FlagSet, p *[]map[string]string, main, name, shorthand string, value []map[string]string, usage string) {
	f.VarP(newStringMapSliceValue(main, value, p), name, shorthand, usage)
}

func StringMapSliceVarPF(f *pflag.FlagSet, p *[]map[string]string, main, name, shorthand string, value []map[string]string, usage string) *pflag.Flag {
	return f.VarPF(newStringMapSliceValue(main, value, p), name, shorthand, usage)
}

func StringMapSlice(f *pflag.FlagSet, name string, main string, value []map[string]string, usage string) *[]map[string]string {
	p := []map[string]string{}
	StringMapSliceVarP(f, &p, main, name, "", value, usage)
	return &p
}

func StringMapSliceP(f *pflag.FlagSet, main, name, shorthand string, value []map[string]string, usage string) *[]map[string]string {
	p := []map[string]string{}
	StringMapSliceVarP(f, &p, main, name, shorthand, value, usage)
	return &p
}
