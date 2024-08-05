package repositories

import (
	_ "ocm.software/ocm/api/credentials/extensions/repositories/aliases"
	_ "ocm.software/ocm/api/credentials/extensions/repositories/directcreds"
	_ "ocm.software/ocm/api/credentials/extensions/repositories/dockerconfig"
	_ "ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig"
	_ "ocm.software/ocm/api/credentials/extensions/repositories/memory"
	_ "ocm.software/ocm/api/credentials/extensions/repositories/memory/config"
	_ "ocm.software/ocm/api/credentials/extensions/repositories/npm"
	_ "ocm.software/ocm/api/credentials/extensions/repositories/vault"
)
