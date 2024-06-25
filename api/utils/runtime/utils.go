package runtime

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"sigs.k8s.io/yaml"
)

func MustProtoType(proto interface{}) reflect.Type {
	t, err := ProtoType(proto)
	if err != nil {
		panic(err.Error())
	}
	return t
}

func ProtoType(proto interface{}) (reflect.Type, error) {
	if proto == nil {
		return nil, errors.New("prototype required")
	}
	t := reflect.TypeOf(proto)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.Newf("prototype %q must be a struct", t)
	}
	return t, nil
}

func ToYAML(data interface{}) ([]byte, error) {
	var m interface{}

	if bytes, ok := data.([]byte); ok {
		err := yaml.Unmarshal(bytes, &m)
		if err != nil {
			return nil, err
		}
	} else {
		m = data
	}
	return yaml.Marshal(m)
}

func TypedObjectFactory(proto TypedObject) func() TypedObject {
	return func() TypedObject { return reflect.New(MustProtoType(proto)).Interface().(TypedObject) }
}

func TypeNames[T TypedObject, R TypedObjectDecoder[T]](scheme Scheme[T, R]) []string {
	types := []string{}
	for t := range scheme.KnownTypes() {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func KindNames[T TypedObject, R TypedObjectDecoder[T]](scheme KnownTypesProvider[T, R]) []string {
	types := []string{}
	for t := range scheme.KnownTypes() {
		if !strings.Contains(t, VersionSeparator) {
			types = append(types, t)
		}
	}
	sort.Strings(types)
	return types
}

func KindToVersionList(types []string, excludes ...string) map[string]string {
	tmp := map[string][]string{}
outer:
	for _, t := range types {
		k, v := KindVersion(t)
		for _, e := range excludes {
			if k == e {
				continue outer
			}
		}
		if _, ok := tmp[k]; !ok {
			tmp[k] = []string{}
		}
		if v != "" {
			tmp[k] = append(tmp[k], v)
		}
	}
	result := map[string]string{}
	for k, v := range tmp {
		result[k] = strings.Join(v, ", ")
	}
	return result
}

func Nil[T any]() T {
	var _nil T
	return _nil
}

// --- begin check ---

// CheckSpecification checks a byte sequence to describe a
// valid minimum specification object.
func CheckSpecification(data []byte) error {
	var obj ObjectTypedObject

	err := DefaultYAMLEncoding.Unmarshal(data, &obj)
	if err != nil {
		return errors.ErrInvalidWrap(err, "repository specification", string(data))
	}
	if obj.GetType() == "" {
		return errors.ErrInvalidWrap(fmt.Errorf("non-empty type field required"), "repository specification", string(data))
	}
	return nil
}

// --- end check ---

func CompleteSpecWithType(typ string, data []byte) ([]byte, error) {
	var m map[string]interface{}
	err := DefaultJSONEncoding.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	if typ != "" {
		if m["type"] != nil && m["type"] != typ {
			return nil, fmt.Errorf("type mismatch between type in reference \"%s\" and type in json spec \"%s\"", typ, m["type"])
		}
		m["type"] = typ
		return DefaultJSONEncoding.Marshal(m)
	} else if m["type"] == nil {
		return nil, fmt.Errorf("type missing")
	}
	return data, nil
}
