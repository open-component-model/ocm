package general

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"slices"

	"github.com/gowebpki/jcs"
	"github.com/modern-go/reflect2"
)

func Optional[T any](args ...T) T {
	var _nil T
	for _, e := range args {
		if !reflect.DeepEqual(e, _nil) {
			return e
		}
	}
	return _nil
}

func OptionalDefaulted[T any](def T, args ...T) T {
	var _nil T
	for _, e := range args {
		if !reflect.DeepEqual(e, _nil) {
			return e
		}
	}
	return def
}

// OptionalDefaultedBool checks all args for true. If no args are given
// the given default is returned.
func OptionalDefaultedBool(def bool, list ...bool) bool {
	if len(list) == 0 {
		return def
	}
	for _, e := range list {
		if e {
			return e
		}
	}
	return false
}

func HashData(d interface{}) string {
	if reflect2.IsNil(d) {
		return ""
	}
	var err error
	var data []byte
	switch b := d.(type) {
	case []byte:
		data = b
	case string:
		data = []byte(b)
	default:
		data, err = json.Marshal(d)
		if err != nil {
			panic(err)
		}
		data, err = jcs.Transform(data)
		if err != nil {
			panic(err)
		}
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func Cycle[T comparable](id T, stack ...T) []T {
	i := slices.Index(stack, id)
	if i < 0 {
		return nil
	}
	return append(slices.Clone(stack[i:]), id)
}

type description interface {
	Description() string
}
type getdescription interface {
	GetDescription() string
}
type getversion interface {
	GetVersion() string
}

func DescribeObject(o any) string {
	if d, ok := o.(getdescription); ok {
		return d.GetDescription()
	}
	if d, ok := o.(description); ok {
		return d.Description()
	}
	if d, ok := o.(getversion); ok {
		return d.GetVersion()
	}
	return "<no description>"
}

// Conditional is the tenary operator, BUT
// there is no conditional/lazy evaluation
// of the cases, because they are handled as arguments.
func Conditional[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
