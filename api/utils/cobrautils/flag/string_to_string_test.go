// Copyright 2009 The Go Authors. All rights reserved.
// Use of ths2s source code s2s governed by a BSD-style
// license that can be found in the github.com/spf13/pflag LICENSE file.

// Modifications Copyright  2024 SAP SE or an SAP affiliate company.

package flag_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils/cobrautils/flag"
)

type (
	Flag    = pflag.Flag
	FlagSet = pflag.FlagSet
)

const ContinueOnError = pflag.ContinueOnError

var NewFlagSet = pflag.NewFlagSet

type MyMap map[string]string

func setUpS2SFlagSet(s2sp *MyMap) *FlagSet {
	f := NewFlagSet("test", ContinueOnError)
	flag.StringToStringVar(f, s2sp, "s2s", map[string]string{}, "Command separated ls2st!")
	return f
}

func setUpS2SFlagSetWithDefault(s2sp *MyMap) *FlagSet {
	f := NewFlagSet("test", ContinueOnError)
	flag.StringToStringVar(f, s2sp, "s2s", map[string]string{"da": "1", "db": "2", "de": "5,6"}, "Command separated ls2st!")
	return f
}

func createS2SFlag(vals map[string]string) string {
	records := make([]string, 0, len(vals)>>1)
	for k, v := range vals {
		records = append(records, k+"="+v)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(records); err != nil {
		panic(err)
	}
	w.Flush()
	return strings.TrimSpace(buf.String())
}

func TestEmptyS2S(t *testing.T) {
	var s2s MyMap
	f := setUpS2SFlagSet(&s2s)
	err := f.Parse([]string{})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}

	if len(s2s) != 0 {
		t.Fatalf("got s2s %v with len=%d but expected length=0", s2s, len(s2s))
	}
}

func TestS2S(t *testing.T) {
	var s2s MyMap
	f := setUpS2SFlagSet(&s2s)

	vals := map[string]string{"a": "1", "b": "2", "d": "4", "c": "3", "e": "5,6"}
	arg := fmt.Sprintf("--s2s=%s", createS2SFlag(vals))
	err := f.Parse([]string{arg})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range s2s {
		if vals[k] != v {
			t.Fatalf("expected s2s[%s] to be %s but got: %s", k, vals[k], v)
		}
	}
}

func TestS2SDefault(t *testing.T) {
	var s2s MyMap
	f := setUpS2SFlagSetWithDefault(&s2s)

	vals := map[string]string{"da": "1", "db": "2", "de": "5,6"}

	err := f.Parse([]string{})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range s2s {
		if vals[k] != v {
			t.Fatalf("expected s2s[%s] to be %s but got: %s", k, vals[k], v)
		}
	}
}

func TestS2SWithDefault(t *testing.T) {
	var s2s MyMap
	f := setUpS2SFlagSetWithDefault(&s2s)

	vals := map[string]string{"a": "1", "b": "2", "e": "5,6"}
	arg := fmt.Sprintf("--s2s=%s", createS2SFlag(vals))
	err := f.Parse([]string{arg})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range s2s {
		if vals[k] != v {
			t.Fatalf("expected s2s[%s] to be %s but got: %s", k, vals[k], v)
		}
	}

	flag := f.Lookup("s2s")
	if flag == nil {
		t.Fatal("flag \"s2s\" not found")
	}
	for k, v := range s2s {
		if vals[k] != v {
			t.Fatalf("expected s2s[%s] to be %s, but got: %s", k, vals[k], v)
		}
	}
}

func TestS2SCalledTwice(t *testing.T) {
	var s2s MyMap
	f := setUpS2SFlagSet(&s2s)

	in := []string{"a=1,b=2", "b=3", `"e=5,6"`, `f=7,8`}
	expected := map[string]string{"a": "1", "b": "3", "e": "5,6", "f": "7,8"}
	argfmt := "--s2s=%s"
	arg0 := fmt.Sprintf(argfmt, in[0])
	arg1 := fmt.Sprintf(argfmt, in[1])
	arg2 := fmt.Sprintf(argfmt, in[2])
	arg3 := fmt.Sprintf(argfmt, in[3])
	err := f.Parse([]string{arg0, arg1, arg2, arg3})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	if len(s2s) != len(expected) {
		t.Fatalf("expected %d flags; got %d flags", len(expected), len(s2s))
	}
	for i, v := range s2s {
		if expected[i] != v {
			t.Fatalf("expected s2s[%s] to be %s but got: %s", i, expected[i], v)
		}
	}
}
