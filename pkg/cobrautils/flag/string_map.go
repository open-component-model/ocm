package flag

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

type stringMapValue struct {
	value   *map[string]string
	changed bool
}

func newStringSliceValue(val map[string]string, p *map[string]string) *stringMapValue {
	ssv := new(stringMapValue)
	ssv.value = p
	*ssv.value = val
	return ssv
}

func readAsKSPairs(val string) (map[string]string, error) {
	r := map[string]string{}
	if val == "" {
		return r, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	list, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	for _, e := range list {
		k, v, err := parseAssignment(e)
		if err != nil {
			return nil, err
		}
		r[k] = v
	}
	return r, nil
}

func writeAsKSPairs(vals map[string]string) (string, error) {
	if vals == nil {
		return "", nil
	}

	var list []string
	for k, v := range vals {
		list = append(list, fmt.Sprintf("%s=%s", k, v))
	}
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(list)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func (s *stringMapValue) Set(val string) error {
	m, err := readAsKSPairs(val)
	if err != nil {
		return err
	}
	if !s.changed {
		*s.value = m
	} else {
		if *s.value == nil {
			*s.value = map[string]string{}
		}
		for k, v := range m {
			(*s.value)[k] = v
		}
	}
	s.changed = true
	return nil
}

func (s *stringMapValue) Type() string {
	return "<name>=<value>"
}

func (s *stringMapValue) String() string {
	str, _ := writeAsKSPairs(*s.value)
	return "[" + str + "]"
}

func (s *stringMapValue) GetMap() map[string]string {
	return *s.value
}

func StringMapVar(f *pflag.FlagSet, p *map[string]string, name string, value map[string]string, usage string) {
	f.VarP(newStringSliceValue(value, p), name, "", usage)
}

func StringMapVarP(f *pflag.FlagSet, p *map[string]string, name, shorthand string, value map[string]string, usage string) {
	f.VarP(newStringSliceValue(value, p), name, shorthand, usage)
}

func StringMap(f *pflag.FlagSet, name string, value map[string]string, usage string) *map[string]string {
	p := map[string]string{}
	StringMapVarP(f, &p, name, "", value, usage)
	return &p
}

func StringSliceP(f *pflag.FlagSet, name, shorthand string, value map[string]string, usage string) *map[string]string {
	p := map[string]string{}
	StringMapVarP(f, &p, name, shorthand, value, usage)
	return &p
}

func parseAssignment(s string) (string, string, error) {
	idx := strings.Index(s, "=")
	if idx <= 0 {
		return "", "", fmt.Errorf("expected <name>=<value>")
	}
	return s[:idx], s[idx+1:], nil
}
