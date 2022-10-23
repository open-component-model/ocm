package flag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

type valueMapValue struct {
	value   *map[string]interface{}
	changed bool
}

func newValueMapValue(val map[string]interface{}, p *map[string]interface{}) *valueMapValue {
	ssv := new(valueMapValue)
	ssv.value = p
	*ssv.value = val
	return ssv
}

func (s *valueMapValue) Set(val string) error {
	k, v, err := parseAssignment(val)
	if err != nil {
		return err
	}
	y, err := parseValue(v)
	if err != nil {
		return err
	}
	if !s.changed {
		*s.value = map[string]interface{}{k: y}
	} else {
		if *s.value == nil {
			*s.value = map[string]interface{}{}
		}
		(*s.value)[k] = y
	}
	s.changed = true
	return nil
}

func (s *valueMapValue) Type() string {
	return "<name>=<YAML>"
}

func (s *valueMapValue) String() string {
	if *s.value == nil {
		return ""
	}
	var list []string
	for k, v := range *s.value {
		//nolint: errchkjson // initialized by unmarshal
		s, _ := json.Marshal(v)
		list = append(list, fmt.Sprintf("%s=%s", k, string(s)))
	}
	return "[" + strings.Join(list, ", ") + "]"
}

func (s *valueMapValue) GetMap() map[string]interface{} {
	return *s.value
}

func ValueMapVar(f *pflag.FlagSet, p *map[string]interface{}, name string, value map[string]interface{}, usage string) {
	f.VarP(newValueMapValue(value, p), name, "", usage)
}

func ValueMapVarP(f *pflag.FlagSet, p *map[string]interface{}, name, shorthand string, value map[string]interface{}, usage string) {
	f.VarP(newValueMapValue(value, p), name, shorthand, usage)
}

func ValueMap(f *pflag.FlagSet, name string, value map[string]interface{}, usage string) *map[string]interface{} {
	p := map[string]interface{}{}
	ValueMapVarP(f, &p, name, "", value, usage)
	return &p
}

func VaueSliceP(f *pflag.FlagSet, name, shorthand string, value map[string]interface{}, usage string) *map[string]interface{} {
	p := map[string]interface{}{}
	ValueMapVarP(f, &p, name, shorthand, value, usage)
	return &p
}
