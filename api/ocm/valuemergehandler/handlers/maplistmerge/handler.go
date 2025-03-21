package maplistmerge

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
)

const ALGORITHM = "mapListMerge"

func init() {
	hpi.Register(New())
}

type (
	// Value is the minimal structure of values usable with the merge algorithm.
	Value = []Entry
	Entry = map[string]interface{}
)

func New() hpi.Handler {
	return hpi.New(ALGORITHM, desc, merge)
}

var desc = `
This handler merges values with a list of map values by observing a key field
to identify similar map entries.
The default entry key is taken from map field <code>name</code>.

It supports the following config structure:
- *<code>keyField</code>* *string* (optional)

  the key field to identify entries in the maps.

- *<code>overwrite</code>* *string* (optional) determines how to handle conflicts.

  - <code>none</code> (default) no change possible, if entry differs the merge is rejected.
  - <code>local</code> the local value is preserved.
  - <code>inbound</code> the inbound value overwrites the local one.

- *<code>entries</code> *merge spec* (optional)

  The merge specification (<code>algorithm</code> and <code>config</code>) used to merge conflicting
  changes in list entries.
`

func merge(ctx cpi.Context, c *Config, lv Value, tv *Value) (bool, error) {
	var err error

	subm := false
	modified := false
	for _, le := range lv {
		key := le[c.KeyField]
		if key != nil {
			found := -1
			for i, entry := range *tv {
				if entry[c.KeyField] == key {
					found = i
					if !reflect.DeepEqual(le, entry) {
						switch c.Overwrite {
						case MODE_DEFAULT:
							if c.Entries != nil {
								subm, entry, err = hpi.GenericMerge(ctx, c.Entries, "", le, entry)
								if err != nil {
									return false, errors.Wrapf(err, "entry identity %q", key)
								}
								if subm {
									(*tv)[i] = entry
									modified = true
								}
								break
							}
							fallthrough
						case MODE_NONE:
							return false, fmt.Errorf("target value for %q changed", key)
						case MODE_LOCAL:
							(*tv)[i] = le
							modified = true
						}
					}
				}
			}
			if found < 0 {
				*tv = append(*tv, le)
				modified = true
			}
		}
	}
	return modified, nil
}
