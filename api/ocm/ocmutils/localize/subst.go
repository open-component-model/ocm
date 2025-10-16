package localize

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/utils/subst"
)

func Substitute(subs Substitutions, fs vfs.FileSystem) error {
	files := map[string]subst.SubstitutionTarget{}

	for i, s := range subs {
		file, err := vfs.Canonical(fs, s.FilePath, true)
		if err != nil {
			return errors.Wrapf(err, "entry %d", i)
		}

		fi, ok := files[file]
		if !ok {
			s, err := subst.ParseFile(file, fs)
			if err != nil {
				return errors.Wrapf(err, "entry %d", i)
			}
			files[file], fi = s, s
		}

		if err = fi.SubstituteByData(s.ValuePath, s.Value); err != nil {
			return errors.Wrapf(err, "entry %d: cannot substitute value", i+1)
		}
	}

	for file, fi := range files {
		data, err := fi.Content()
		if err != nil {
			return errors.Wrapf(err, "cannot marshal %q after substitution ", file)
		}

		if err := vfs.WriteFile(fs, file, data, vfs.ModePerm); err != nil {
			return errors.Wrapf(err, "file %q", file)
		}
	}
	return nil
}

// SubstituteMappings substitutes value mappings for a dedicated substitution target.
func SubstituteMappings(subs ValueMappings, target subst.SubstitutionTarget) error {
	for i, s := range subs {
		if err := target.SubstituteByData(s.ValuePath, s.Value); err != nil {
			return errors.Wrapf(err, "entry %d: cannot substitute value", i+1)
		}
	}
	return nil
}

// SubstituteMappingsForData substitutes value mappings for some data.
func SubstituteMappingsForData(subs ValueMappings, data []byte) ([]byte, error) {
	target, err := subst.Parse(data)
	if err != nil {
		return nil, err
	}
	err = SubstituteMappings(subs, target)
	if err != nil {
		return nil, err
	}
	return target.Content()
}

/*
2024-04-28 Don't see any use of this in either ocm or ocm-controller
As it exposes the implementation detail of what YAML model we use
_and_ we're switching to yqlib from goccy comment it out.
func Set(content *ast.File, path string, value *ast.File) error {
	p, err := yaml.PathString("$." + path)
	if err != nil {
		return errors.Wrapf(err, "invalid substitution path")
	}
	return p.ReplaceWithFile(content, value)
}
*/
