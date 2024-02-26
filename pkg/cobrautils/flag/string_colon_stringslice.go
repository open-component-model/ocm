// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package flag

import (
	"bytes"
	"encoding/csv"
	"strings"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/errors"
)

type stringColonStringSliceValue[T ~map[string][]string] struct {
	value   *T
	changed bool
}

func newStringColonStringSliceValue[T ~map[string][]string](val map[string][]string, p *T) *stringColonStringSliceValue[T] {
	ssv := new(stringColonStringSliceValue[T])
	ssv.value = p
	*ssv.value = val
	return ssv
}

// Set Format: a:1,2,3
// assumes that keys and values cannot contain = signs or comma.
func (s *stringColonStringSliceValue[T]) Set(val string) error {
	pos := strings.Index(val, ":")
	if pos == -1 {
		return errors.New("option has to contain at least one equal sign (=)")
	}
	key := val[:pos]
	str := val[pos+1:]

	ss := strings.Split(str, ",")
	for i, e := range ss {
		ss[i] = strings.TrimSpace(e)
	}

	if !s.changed {
		*s.value = map[string][]string{key: ss}
	} else {
		if *s.value == nil {
			*s.value = map[string][]string{}
		}
		(*s.value)[key] = ss
	}
	s.changed = true
	return nil
}

func (s *stringColonStringSliceValue[T]) Type() string {
	return "<name>:<value>,<value>,..."
}

func (s *stringColonStringSliceValue[T]) String() string {
	records := make([]string, 0, len(*s.value))
	for k, v := range *s.value {
		records = append(records, k+": ["+strings.Join(v, ",")+"]")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(records); err != nil {
		panic(err)
	}
	w.Flush()
	return "{" + strings.TrimSpace(buf.String()) + "}"
}

// StringColonStringSliceVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a map[string][]string variable in which to store the values of the multiple flags.
func StringColonStringSliceVar[T ~map[string][]string](f *pflag.FlagSet, p *T, name string, value map[string][]string, usage string) {
	f.VarP(newStringColonStringSliceValue(value, p), name, "", usage)
}

// StringColonStringSliceVarP is like StringColonStringSliceVar, but accepts a shorthand letter that can be used after a single dash.
func StringColonStringSliceVarP[T ~map[string][]string](f *pflag.FlagSet, p *T, name, shorthand string, value map[string][]string, usage string) {
	f.VarP(newStringColonStringSliceValue(value, p), name, shorthand, usage)
}

// StringColonStringSliceVarPF is like StringColonStringSliceVarP, but returns the created flag.
func StringColonStringSliceVarPF[T ~map[string][]string](f *pflag.FlagSet, p *T, name, shorthand string, value map[string][]string, usage string) *pflag.Flag {
	return f.VarPF(newStringColonStringSliceValue(value, p), name, shorthand, usage)
}

// StringColonStringSliceVarPFA is like StringColonStringSliceVarPF, but allows to add to a preset map.
func StringColonStringSliceVarPFA[T ~map[string][]string](f *pflag.FlagSet, p *T, name, shorthand string, value map[string][]string, usage string) *pflag.Flag {
	v := newStringColonStringSliceValue(value, p)
	v.changed = true
	return f.VarPF(v, name, shorthand, usage)
}
