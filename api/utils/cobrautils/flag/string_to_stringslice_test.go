package flag_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"testing"

	"ocm.software/ocm/api/utils/cobrautils/flag"
)

// MySSM is My String Slice Map
type MySSM map[string][]string

func setUpS2SSFlagSet(s2ssp *MySSM) *FlagSet {
	f := NewFlagSet("test", ContinueOnError)
	flag.StringToStringSliceVar(f, s2ssp, "s2ss", map[string][]string{}, "Command separated ls2st!")
	return f
}

func setUpS2SSFlagSetWithDefault(s2ssp *MySSM) *FlagSet {
	f := NewFlagSet("test", ContinueOnError)
	flag.StringToStringSliceVar(f, s2ssp, "s2ss", map[string][]string{"da": {"1", "2", "3"}, "db": {"2"}, "de": {"5,6", "7"}}, "Command separated ls2sst!")
	return f
}

func createS2SSFlag(vals map[string][]string) string {
	records := make([]string, 0, len(vals))
	for k, v := range vals {
		records = append(records, k+"="+strings.Join(v, ","))
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(records); err != nil {
		panic(err)
	}
	w.Flush()
	return strings.TrimSpace(buf.String())
}

func TestEmptyS2SS(t *testing.T) {
	var s2ss MySSM
	f := setUpS2SSFlagSet(&s2ss)
	err := f.Parse([]string{})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}

	if len(s2ss) != 0 {
		t.Fatalf("got s2ss %v with len=%d but expected length=0", s2ss, len(s2ss))
	}
}

func TestS2SS(t *testing.T) {
	var s2ss MySSM
	f := setUpS2SSFlagSet(&s2ss)

	vals := map[string][]string{"da": {"1", "2", "3"}, "db": {"2"}, "de": {"5,6", "7"}}
	arg := fmt.Sprintf("--s2ss=%s", createS2SSFlag(vals))
	err := f.Parse([]string{arg})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range s2ss {
		for index, value := range vals[k] {
			if value != v[index] {
				t.Fatalf("expected s2ss[%s] to be %s but got: %s", k, vals[k], v)
			}
		}
	}
}

func TestS2SSDefault(t *testing.T) {
	var s2ss MySSM
	f := setUpS2SSFlagSetWithDefault(&s2ss)

	vals := map[string][]string{"da": {"1", "2", "3"}, "db": {"2"}, "de": {"5,6", "7"}}

	err := f.Parse([]string{})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range s2ss {
		for index, value := range vals[k] {
			if value != v[index] {
				t.Fatalf("expected s2ss[%s] to be %s but got: %s", k, vals[k], v)
			}
		}
	}
}

func TestS2SSWithDefault(t *testing.T) {
	var s2ss MySSM
	f := setUpS2SSFlagSetWithDefault(&s2ss)

	vals := map[string][]string{"da": {"1", "2", "3"}, "db": {"2"}, "de": {"5,6", "7"}}
	arg := fmt.Sprintf("--s2ss=%s", createS2SSFlag(vals))
	err := f.Parse([]string{arg})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for k, v := range s2ss {
		for index, value := range vals[k] {
			if value != v[index] {
				t.Fatalf("expected s2ss[%s] to be %s but got: %s", k, vals[k], v)
			}
		}
	}

	flag := f.Lookup("s2ss")
	if flag == nil {
		t.Fatal("flag \"s2s\" not found")
	}
	for k, v := range s2ss {
		for index, value := range vals[k] {
			if value != v[index] {
				t.Fatalf("expected s2ss[%s] to be %s but got: %s", k, vals[k], v)
			}
		}
	}
}

func TestS2SSCalledTwice(t *testing.T) {
	var s2ss MySSM
	f := setUpS2SSFlagSet(&s2ss)

	in := []string{"a=1,2,3,b=2,3", "c=3,d=4,e=5,6", `"f=5"`}
	expected := map[string][]string{"a": {"1", "2", "3"}, "b": {"2", "3"}, "c": {"3"}, "d": {"4"}, "e": {"5", "6"}, "f": {"5"}}
	argfmt := "--s2ss=%s"
	arg0 := fmt.Sprintf(argfmt, in[0])
	arg1 := fmt.Sprintf(argfmt, in[1])
	arg2 := fmt.Sprintf(argfmt, in[2])
	err := f.Parse([]string{arg0, arg1, arg2})
	if err != nil {
		t.Fatal("expected no error; got", err)
	}
	for i, v := range s2ss {
		for index, value := range expected[i] {
			if value != v[index] {
				t.Fatalf("expected s2ss[%s] to be %s but got: %s", i, expected[i], v)
			}
		}
	}
}
