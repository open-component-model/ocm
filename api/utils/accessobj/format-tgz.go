package accessobj

import (
	"github.com/open-component-model/ocm/api/utils/accessio"
	"github.com/open-component-model/ocm/api/utils/compression"
)

var FormatTGZ = NewTarHandlerWithCompression(accessio.FormatTGZ, compression.Gzip)

func init() {
	RegisterFormat(FormatTGZ)
}
