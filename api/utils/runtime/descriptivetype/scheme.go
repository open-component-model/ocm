package descriptivetype

import (
	"fmt"
	"strings"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

// TypeScheme is the appropriately extended scheme interface based on
// runtime.TypeScheme. Based on the additional type info a complete
// scheme description can be created calling the Describe method.
type TypeScheme[T runtime.TypedObject, R TypedObjectType[T]] interface {
	runtime.TypeScheme[T, R]

	Describe() string
}

type _typeScheme[T runtime.TypedObject, R runtime.TypedObjectType[T]] interface {
	runtime.TypeScheme[T, R] // for goland to be able to accept extender argument type
}

type typeScheme[T runtime.TypedObject, R TypedObjectType[T], S runtime.TypeScheme[T, R]] struct {
	name      string
	extender  DescriptionExtender[R]
	versioned bool
	_typeScheme[T, R]
}

func MustNewDefaultTypeScheme[T runtime.TypedObject, R TypedObjectType[T], S TypeScheme[T, R]](name string, extender DescriptionExtender[R], unknown runtime.Unstructured, acceptUnknown bool, defaultdecoder runtime.TypedObjectDecoder[T], base ...TypeScheme[T, R]) TypeScheme[T, R] {
	scheme := runtime.MustNewDefaultTypeScheme[T, R](unknown, acceptUnknown, defaultdecoder, utils.Optional(base...))
	return &typeScheme[T, R, S]{
		name:        name,
		extender:    extender,
		_typeScheme: scheme,
	}
}

// NewTypeScheme provides an TypeScheme implementation based on the interfaces
// and the default runtime.TypeScheme implementation.
func NewTypeScheme[T runtime.TypedObject, R TypedObjectType[T], S TypeScheme[T, R]](name string, extender DescriptionExtender[R], unknown runtime.Unstructured, acceptUnknown bool, base ...S) TypeScheme[T, R] {
	scheme := runtime.MustNewDefaultTypeScheme[T, R](unknown, acceptUnknown, nil, utils.Optional(base...))
	return &typeScheme[T, R, S]{
		name:        name,
		extender:    extender,
		_typeScheme: scheme,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (t *typeScheme[T, R, S]) KnownTypes() runtime.KnownTypes[T, R] {
	return t._typeScheme.KnownTypes() // Goland
}

func (t *typeScheme[T, R, S]) Describe() string {
	s := ""
	type method struct {
		desc     string
		versions map[string]string
		more     string
	}

	descs := map[string]*method{}

	// gather info for kinds and versions
	for _, n := range t.KnownTypeNames() {
		var kind, vers string
		if t.versioned {
			kind, vers = runtime.KindVersion(n)
		} else {
			kind = n
		}

		info := descs[kind]
		if info == nil {
			info = &method{versions: map[string]string{}}
			descs[kind] = info
		}

		if vers == "" {
			vers = "v1"
		}
		if _, ok := info.versions[vers]; !ok {
			info.versions[vers] = ""
		}

		ty := t.GetType(n)

		if t.extender != nil {
			more := t.extender(ty)
			if more != "" {
				info.more = more
			}
		}
		desc := ty.Description()
		if desc != "" {
			info.desc = desc
		}

		desc = ty.Format()
		if desc != "" {
			info.versions[vers] = desc
		}
	}

	for _, tn := range utils.StringMapKeys(descs) {
		info := descs[tn]
		desc := strings.Trim(info.desc, "\n")
		if desc != "" {
			s = fmt.Sprintf("%s\n- %s <code>%s</code>\n\n%s\n\n", s, t.name, tn, utils.IndentLines(desc, "  "))

			format := ""
			for _, f := range utils.StringMapKeys(info.versions) {
				desc = strings.Trim(info.versions[f], "\n")
				if desc != "" {
					if t.versioned {
						format = fmt.Sprintf("%s\n- Version <code>%s</code>\n\n%s\n", format, f, utils.IndentLines(desc, "  "))
					} else {
						s += utils.IndentLines(desc, "  ")
					}
				}
			}
			if format != "" {
				s += fmt.Sprintf("  The following versions are supported:\n%s\n", strings.Trim(utils.IndentLines(format, "  "), "\n"))
			}
		}
		s += info.more
	}
	return s
}
