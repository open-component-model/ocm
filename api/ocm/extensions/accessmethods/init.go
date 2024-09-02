package accessmethods

import (
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/github"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/helm"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/localfsblob"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/localociblob"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/maven"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/none"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/npm"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/ociblob"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/ocm"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/relativeociref"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/s3"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/wget"
)
