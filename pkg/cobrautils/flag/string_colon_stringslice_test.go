// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package flag_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"testing"

	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
)

func setUpSCSSFlagSet(scssp *MySSM) *FlagSet {
	f := NewFlagSet("test", ContinueOnError)
	flag.StringColonStringSliceVar(f, scssp, "scss", nil, "Command separated lscst!")
	return f
}

func setUpSCSSFlagSetWithDefault(scssp *MySSM) *FlagSet {
	f := NewFlagSet("test", ContinueOnError)
	flag.StringColonStringSliceVar(f, scssp, "scss", map[string][]string{"da": {"1", "2", "3"}, "db": {"2"}, "de": {"5,6", "7"}}, "Command separated lscsst!")
	return f
}

func createSCSSFlag(vals map[string][]string) string {
	records := make([]string, 0, len(vals))
	for k, v := range vals {
		records = append(records, k+":"+strings.Join(v, ","))
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(records); err != nil {
		panic(err)
	}
	w.Flush()
	return strings.TrimSpace(buf.String())
}

func TestEmptySCSS(t *testing.T) {
	var scss MySSM
	f := setUpSCSSFlagSet(&scss)
	err := f.Parse([]string{})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}

	if len(scss) != 0 {
		t.Fatalf("got scss %v with len=%d but expected length=0", scss, len(scss))
	}
}

func TestSCSS(t *testing.T) {
	var scss MySSM
	f := setUpSCSSFlagSet(&scss)

	vals := map[string][]string{"da": {"1", "2", "3"}, "db": {"2"}, "de": {"5,6", "7"}}
	arg := fmt.Sprintf("--scss=%s", createSCSSFlag(vals))
	err := f.Parse([]string{arg})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range scss {
		for index, value := range vals[k] {
			if value != v[index] {
				t.Fatalf("expected scss[%s] to be %s but got: %s", k, vals[k], v)
			}
		}
	}
}

func TestSCSSDefault(t *testing.T) {
	var scss MySSM
	f := setUpSCSSFlagSetWithDefault(&scss)

	vals := map[string][]string{"da": {"1", "2", "3"}, "db": {"2"}, "de": {"5,6", "7"}}
	err := f.Parse([]string{})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range scss {
		for index, value := range vals[k] {
			if value != v[index] {
				t.Fatalf("expected scss[%s] to be %s but got: %s", k, vals[k], v)
			}
		}
	}
}

func TestSCSSWithDefault(t *testing.T) {
	var scss MySSM
	f := setUpSCSSFlagSetWithDefault(&scss)

	vals := map[string][]string{"da": {"1", "2", "3"}, "db": {"2"}, "de": {"5,6", "7"}}
	arg := fmt.Sprintf("--scss=%s", createSCSSFlag(vals))
	err := f.Parse([]string{arg})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range scss {
		for index, value := range vals[k] {
			if value != v[index] {
				t.Fatalf("expected scss[%s] to be %s but got: %s", k, vals[k], v)
			}
		}
	}

	flag := f.Lookup("scss")
	if flag == nil {
		t.Fatal("flag \"scs\" not found")
	}
	for k, v := range scss {
		for index, value := range vals[k] {
			if value != v[index] {
				t.Fatalf("expected scss[%s] to be %s but got: %s", k, vals[k], v)
			}
		}
	}
}

func TestSCSSCalledTwice(t *testing.T) {
	var scss MySSM
	f := setUpSCSSFlagSet(&scss)

	in := []string{"a:1,2,3,b:2,3", "c:3,d:4,e:5,6", `"f:5"`}
	expected := map[string][]string{"a": {"1", "2", "3"}, "b": {"2", "3"}, "c": {"3"}, "d": {"4"}, "e": {"5", "6"}, "f": {"5"}}
	argfmt := "--scss=%s"
	arg0 := fmt.Sprintf(argfmt, in[0])
	arg1 := fmt.Sprintf(argfmt, in[1])
	arg2 := fmt.Sprintf(argfmt, in[2])
	err := f.Parse([]string{arg0, arg1, arg2})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for i, v := range scss {
		for index, value := range expected[i] {
			if value != v[index] {
				t.Fatalf("expected scss[%s] to be %s but got: %s", i, expected[i], v)
			}
		}
	}
}
