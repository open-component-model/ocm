package repositories

import (
	_ "github.com/open-component-model/ocm/api/credentials/extensions/repositories/aliases"
	_ "github.com/open-component-model/ocm/api/credentials/extensions/repositories/directcreds"
	_ "github.com/open-component-model/ocm/api/credentials/extensions/repositories/dockerconfig"
	_ "github.com/open-component-model/ocm/api/credentials/extensions/repositories/gardenerconfig"
	_ "github.com/open-component-model/ocm/api/credentials/extensions/repositories/memory"
	_ "github.com/open-component-model/ocm/api/credentials/extensions/repositories/memory/config"
	_ "github.com/open-component-model/ocm/api/credentials/extensions/repositories/npm"
	_ "github.com/open-component-model/ocm/api/credentials/extensions/repositories/vault"
)
