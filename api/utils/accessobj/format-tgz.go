package accessobj

import (
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/compression"
)

var FormatTGZ = NewTarHandlerWithCompression(accessio.FormatTGZ, compression.Gzip)

func init() {
	RegisterFormat(FormatTGZ)
}
