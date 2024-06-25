package blobaccess

import (
	"github.com/open-component-model/ocm/api/utils/blobaccess/blobaccess"
)

// DataAccessForBytes wraps a bytes slice into a DataAccess.
// Deprecated: used DataAccessForData.
func DataAccessForBytes(data []byte, origin ...string) DataSource {
	return blobaccess.DataAccessForData(data, origin...)
}
