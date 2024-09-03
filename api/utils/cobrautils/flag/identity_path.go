package flag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

type identityPath struct {
	value   *[]map[string]string
	changed bool
}

func newIdentityPathValue(val []map[string]string, p *[]map[string]string) *identityPath {
	ssv := new(identityPath)
	ssv.value = p
	*ssv.value = val
	return ssv
}

func (s *identityPath) Set(val string) error {
	k, v, err := parseAssignment(val)
	if err != nil {
		return err
	}
	if !s.changed {
		if k != "name" {
			return fmt.Errorf("first identity attribute must be the name attribute")
		}
		*s.value = []map[string]string{{k: v}}
	} else {
		if k == "name" {
			*s.value = append(*s.value, map[string]string{k: v})
		} else {
			(*s.value)[len(*s.value)-1][k] = v
		}
	}
	s.changed = true
	return nil
}

func (s *identityPath) Type() string {
	return "{<name>=<value>}"
}

func (s *identityPath) String() string {
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

func (s *identityPath) GetPath() []map[string]string {
	return *s.value
}

func IdentityPathVar(f *pflag.FlagSet, p *[]map[string]string, name string, value []map[string]string, usage string) {
	f.VarP(newIdentityPathValue(value, p), name, "", usage)
}

func IdentityPathVarP(f *pflag.FlagSet, p *[]map[string]string, name, shorthand string, value []map[string]string, usage string) {
	f.VarP(newIdentityPathValue(value, p), name, shorthand, usage)
}

func IdentityPathVarPF(f *pflag.FlagSet, p *[]map[string]string, name, shorthand string, value []map[string]string, usage string) *pflag.Flag {
	return f.VarPF(newIdentityPathValue(value, p), name, shorthand, usage)
}

func IdentityPath(f *pflag.FlagSet, name string, value []map[string]string, usage string) *[]map[string]string {
	p := []map[string]string{}
	IdentityPathVarP(f, &p, name, "", value, usage)
	return &p
}

func IdentityPathP(f *pflag.FlagSet, name, shorthand string, value []map[string]string, usage string) *[]map[string]string {
	p := []map[string]string{}
	IdentityPathVarP(f, &p, name, shorthand, value, usage)
	return &p
}
