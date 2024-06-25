package accessmethods

import (
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/github"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/helm"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/localblob"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/localfsblob"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/localociblob"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/maven"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/none"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/npm"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/ociartifact"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/ociblob"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/relativeociref"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/s3"
	_ "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/wget"
)
