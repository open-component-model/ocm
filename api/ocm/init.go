package ocm

import (
	_ "github.com/open-component-model/ocm/api/datacontext/attrs"
	_ "github.com/open-component-model/ocm/api/oci"
	_ "github.com/open-component-model/ocm/api/ocm/compdesc/normalizations"
	_ "github.com/open-component-model/ocm/api/ocm/compdesc/versions"
	_ "github.com/open-component-model/ocm/api/ocm/config"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/blobhandler/config"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/blobhandler/handlers"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/digester/digesters"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/download/config"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/download/handlers"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/repositories"
	_ "github.com/open-component-model/ocm/api/ocm/valuemergehandler/handlers"
)
