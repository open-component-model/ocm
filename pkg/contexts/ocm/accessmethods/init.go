package accessmethods

import (
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/helm"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localociblob"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/none"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/npm"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/relativeociref"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/s3"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/wget"
)
