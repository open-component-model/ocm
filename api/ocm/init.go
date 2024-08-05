package ocm

import (
	_ "ocm.software/ocm/api/datacontext/attrs"
	_ "ocm.software/ocm/api/oci"
	_ "ocm.software/ocm/api/ocm/compdesc/normalizations"
	_ "ocm.software/ocm/api/ocm/compdesc/versions"
	_ "ocm.software/ocm/api/ocm/config"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods"
	_ "ocm.software/ocm/api/ocm/extensions/blobhandler/config"
	_ "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers"
	_ "ocm.software/ocm/api/ocm/extensions/digester/digesters"
	_ "ocm.software/ocm/api/ocm/extensions/download/config"
	_ "ocm.software/ocm/api/ocm/extensions/download/handlers"
	_ "ocm.software/ocm/api/ocm/extensions/pubsub/providers"
	_ "ocm.software/ocm/api/ocm/extensions/pubsub/types"
	_ "ocm.software/ocm/api/ocm/extensions/repositories"
	_ "ocm.software/ocm/api/ocm/valuemergehandler/handlers"
)
