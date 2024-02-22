package repositories

import (
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/aliases"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/directcreds"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/dockerconfig"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory/config"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/npm"
	_ "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/vault"
)
